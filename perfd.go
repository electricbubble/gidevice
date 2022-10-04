package giDevice

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
)

type perfOption struct {
	bundleID string
	pid      string
	gpu      bool
	cpu      bool
	mem      bool
	fps      bool
	network  bool
}

func defaulPerfOption() *perfOption {
	return &perfOption{
		gpu:     false,
		cpu:     true, // default on
		mem:     true, // default on
		fps:     false,
		network: false,
	}
}

type PerfOption func(*perfOption)

func WithPerfBundleID(bundleID string) PerfOption {
	return func(opt *perfOption) {
		opt.bundleID = bundleID
	}
}

func WithPerfPID(pid string) PerfOption {
	return func(opt *perfOption) {
		opt.pid = pid
	}
}

func WithPerfGPU(b bool) PerfOption {
	return func(opt *perfOption) {
		opt.gpu = b
	}
}

func WithPerfCPU(b bool) PerfOption {
	return func(opt *perfOption) {
		opt.cpu = b
	}
}

func WithPerfMem(b bool) PerfOption {
	return func(opt *perfOption) {
		opt.mem = b
	}
}

func WithPerfFPS(b bool) PerfOption {
	return func(opt *perfOption) {
		opt.fps = b
	}
}

func WithPerfNetwork(b bool) PerfOption {
	return func(opt *perfOption) {
		opt.network = b
	}
}

type perfdClient struct {
	option      *perfOption
	i           *instruments
	stop        chan struct{}        // used to stop perf client
	cancels     []context.CancelFunc // used to cancel all iterators
	chanCPU     chan []byte          // cpu channel
	chanMem     chan []byte          // mem channel
	chanGPU     chan []byte          // gpu channel
	chanFPS     chan []byte          // fps channel
	chanNetwork chan []byte          // network channel
}

func (d *device) newPerfdClient(i Instruments, opts ...PerfOption) *perfdClient {
	perfOption := defaulPerfOption()
	for _, fn := range opts {
		fn(perfOption)
	}

	if perfOption.bundleID != "" {
		pid, err := d.getPidByBundleID(perfOption.bundleID)
		if err == nil {
			perfOption.pid = strconv.Itoa(pid)
		}
	}

	return &perfdClient{
		i:           i.(*instruments),
		option:      perfOption,
		chanCPU:     make(chan []byte, 10),
		chanMem:     make(chan []byte, 10),
		chanGPU:     make(chan []byte, 10),
		chanFPS:     make(chan []byte, 10),
		chanNetwork: make(chan []byte, 10),
	}
}

func (c *perfdClient) Start() (data <-chan []byte, err error) {
	outCh := make(chan []byte, 100)

	if c.option.cpu || c.option.mem {
		cancel, err := c.registerSysmontap(c.option.pid, context.Background())
		if err != nil {
			return nil, err
		}
		c.cancels = append(c.cancels, cancel)
	}

	if c.option.gpu || c.option.fps {
		cancel, err := c.registerGraphicsOpengl(context.Background())
		if err != nil {
			return nil, err
		}
		c.cancels = append(c.cancels, cancel)
	}

	if c.option.network {
		cancel, err := c.registerNetworking(context.Background())
		if err != nil {
			return nil, err
		}
		c.cancels = append(c.cancels, cancel)
	}

	go func() {
		for {
			select {
			case <-c.stop:
				return
			case cpuBytes, ok := <-c.chanCPU:
				if ok && c.option.cpu {
					fmt.Println("cpu: ", string(cpuBytes))
					outCh <- cpuBytes
				}
			case memBytes, ok := <-c.chanMem:
				if ok && c.option.mem {
					fmt.Println("mem: ", string(memBytes))
					outCh <- memBytes
				}
			case gpuBytes, ok := <-c.chanGPU:
				if ok && c.option.gpu {
					fmt.Println("gpu: ", string(gpuBytes))
					outCh <- gpuBytes
				}
			case fpsBytes, ok := <-c.chanFPS:
				if ok && c.option.fps {
					fmt.Println("fps: ", string(fpsBytes))
					outCh <- fpsBytes
				}
			case networkBytes, ok := <-c.chanNetwork:
				if ok && c.option.network {
					fmt.Println("network: ", string(networkBytes))
					outCh <- networkBytes
				}
			}
		}
	}()

	return outCh, nil
}

func (c *perfdClient) Stop() {
	close(c.stop)
	for _, cancel := range c.cancels {
		cancel()
	}
}

func (c *perfdClient) registerNetworking(ctx context.Context) (
	cancel context.CancelFunc, err error) {

	chanID, err := c.i.requestChannel(instrumentsServiceNetworking)
	if err != nil {
		return nil, err
	}

	selector := "startMonitoring"
	args := libimobiledevice.NewAuxBuffer()
	if _, err = c.i.client.Invoke(selector, args, chanID, true); err != nil {
		return nil, err
	}
	c.i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			return
		default:
			c.parseNetworking(m.Obj)
		}
	})

	return nil, nil
}

func (c *perfdClient) parseNetworking(data interface{}) {
	// data example (3 types):
	// [
	//   2
	//   [
	//     36756    // RxBytes
	//     11180213 // RxPackets
	//     32837    // TxBytes
	//     8365982  // TxPackets
	//     map[$class:9] map[$class:9] map[$class:9] map[$class:9] map[$class:9] 144 -1]
	// ]
	// [2 [205 435704 0 0 0 0 0 0.005 0.005 95 166]]
	// [1 [[16 2 197 19 192 168 100 103 0 0 0 0 0 0 0 0] [16 2 1 187 117 174 183 75 0 0 0 0 0 0 0 0] 14 -2 262144 0 21 1]]
	raw, ok := data.([]interface{})
	if !ok || len(raw) != 2 {
		return
	}
	if raw[0].(uint64) == 1 {
		// TODO
		return
	}

	rtxData, ok := raw[1].([]interface{})
	if !ok {
		return
	}

	netData := NetworkData{
		Type:      "network",
		TimeStamp: time.Now().Unix(),
		RxBytes:   convert2Int64(rtxData[0]),
		RxPackets: convert2Int64(rtxData[1]),
		TxBytes:   convert2Int64(rtxData[2]),
		TxPackets: convert2Int64(rtxData[3]),
	}
	netBytes, _ := json.Marshal(netData)
	c.chanNetwork <- netBytes
}

type NetworkData struct {
	Type      string `json:"type"` // network
	TimeStamp int64  `json:"timestamp"`
	RxBytes   int64  `json:"rxBytes"`
	RxPackets int64  `json:"rxPackets"`
	TxBytes   int64  `json:"txBytes"`
	TxPackets int64  `json:"txPackets"`
}

func (c *perfdClient) registerGraphicsOpengl(ctx context.Context) (
	cancel context.CancelFunc, err error) {

	chanID, err := c.i.requestChannel(instrumentsServiceGraphicsOpengl)
	if err != nil {
		return nil, err
	}

	selector := "availableStatistics"
	args := libimobiledevice.NewAuxBuffer()
	if _, err = c.i.client.Invoke(selector, args, chanID, true); err != nil {
		return nil, err
	}

	selector = "setSamplingRate:"
	if err = args.AppendObject(0.0); err != nil {
		return nil, err
	}
	if _, err = c.i.client.Invoke(selector, args, chanID, true); err != nil {
		return nil, err
	}

	selector = "startSamplingAtTimeInterval:processIdentifier:"
	args = libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(0); err != nil {
		return nil, err
	}
	if err = args.AppendObject(0); err != nil {
		return nil, err
	}
	if _, err = c.i.client.Invoke(selector, args, chanID, true); err != nil {
		return nil, err
	}

	c.i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			return
		default:
			c.parseGpuFps(m.Obj)
		}
	})

	return
}

func (c *perfdClient) parseGpuFps(data interface{}) {
	// data example:
	// map[
	//   Alloc system memory:50167808
	//   Allocated PB Size:1179648
	//   CoreAnimationFramesPerSecond:0
	//   Device Utilization %:0
	//   IOGLBundleName:Built-In
	//   In use system memory:10633216
	//   Renderer Utilization %:0
	//   SplitSceneCount:0
	//   TiledSceneBytes:0
	//   Tiler Utilization %:0
	//   XRVideoCardRunTimeStamp:1010679
	//   recoveryCount:0
	// ]
	var deviceUtilization = data.(map[string]interface{})["Device Utilization %"]     // Device Utilization
	var tilerUtilization = data.(map[string]interface{})["Tiler Utilization %"]       // Tiler Utilization
	var rendererUtilization = data.(map[string]interface{})["Renderer Utilization %"] // Renderer Utilization

	gpuInfo := GPUData{
		Type:                "gpu",
		TimeStamp:           time.Now().Unix(),
		DeviceUtilization:   convert2Int64(deviceUtilization),
		TilerUtilization:    convert2Int64(tilerUtilization),
		RendererUtilization: convert2Int64(rendererUtilization),
	}
	gpuBytes, _ := json.Marshal(gpuInfo)
	c.chanGPU <- gpuBytes

	var fps = data.(map[string]interface{})["CoreAnimationFramesPerSecond"]
	fpsInfo := FPSData{
		Type:      "fps",
		TimeStamp: time.Now().Unix(),
		FPS:       int(convert2Int64(fps)),
	}
	fpsBytes, _ := json.Marshal(fpsInfo)
	c.chanFPS <- fpsBytes
}

type GPUData struct {
	Type                string `json:"type"` // gpu
	TimeStamp           int64  `json:"timestamp"`
	TilerUtilization    int64  `json:"tilerUtilization"`    // 处理顶点的GPU时间占比
	DeviceUtilization   int64  `json:"deviceUtilization"`   // 设备利用率
	RendererUtilization int64  `json:"rendererUtilization"` // 渲染器利用率
	Msg                 string `json:"msg,omitempty"`       // 提示信息
}

type FPSData struct {
	Type      string `json:"type"` // fps
	TimeStamp int64  `json:"timestamp"`
	FPS       int    `json:"fps"`
}

func (c *perfdClient) registerSysmontap(pid string, ctx context.Context) (
	cancel context.CancelFunc, err error) {

	chanID, err := c.i.requestChannel(instrumentsServiceSysmontap)
	if err != nil {
		return nil, err
	}

	// set config
	args := libimobiledevice.NewAuxBuffer()
	config := map[string]interface{}{
		"bm":             0,
		"cpuUsage":       true,
		"sampleInterval": time.Second * 1, // 1s
		"ur":             1000,            // 刷新频率
		"procAttrs": []string{
			"memVirtualSize", // vss
			"cpuUsage",
			"ctxSwitch",       // the number of context switches by process each second
			"intWakeups",      // the number of threads wakeups by process each second
			"physFootprint",   // real memory (物理内存)
			"memResidentSize", // rss
			"memAnon",         // anonymous memory
			"pid",
		},
		"sysAttrs": []string{ // 系统信息字段
			"vmExtPageCount",
			"vmFreeCount",
			"vmPurgeableCount",
			"vmSpeculativeCount",
			"physMemSize",
		},
	}
	args.AppendObject(config)
	if _, err = c.i.client.Invoke("setConfig:", args, chanID, true); err != nil {
		return nil, err
	}

	// start
	args = libimobiledevice.NewAuxBuffer()
	if _, err = c.i.client.Invoke("start", args, chanID, true); err != nil {
		return nil, err
	}

	// register listener
	ctx, cancel = context.WithCancel(ctx)
	c.i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			return
		default:
			c.parseCPUMem(m.Obj, pid)
		}
	})

	return cancel, err
}

func (c *perfdClient) parseCPUMem(data interface{}, pid string) {
	messArray, ok := data.([]interface{})
	if !ok || len(messArray) != 2 {
		return
	}

	var systemInfo = messArray[0].(map[string]interface{})
	var processInfoList = messArray[1].(map[string]interface{})
	if systemInfo["CPUCount"] == nil {
		systemInfo, processInfoList = processInfoList, systemInfo
	}
	if systemInfo["CPUCount"] == nil {
		fmt.Printf("invalid system info: %v\n", systemInfo)
		return
	}
	// systemInfo example:
	// map[
	//   CPUCount:2
	//   EnabledCPUs:2
	//   PerCPUUsage:[
	//     map[CPU_NiceLoad:0 CPU_SystemLoad:-1 CPU_TotalLoad:4.587155963302749 CPU_UserLoad:-1]
	//     map[CPU_NiceLoad:0 CPU_SystemLoad:-1 CPU_TotalLoad:0.9174311926605441 CPU_UserLoad:-1]
	//   ]
	//   System:[70117 2850 465 1579 128643]
	//   SystemCPUUsage:map[
	//     CPU_NiceLoad:0
	//     CPU_SystemLoad:-1
	//     CPU_TotalLoad:5.504587155963293
	//     CPU_UserLoad:-1
	//   ]
	//   StartMachAbsTime:2514085834016
	//   EndMachAbsTime:2514111855034
	//   Type:41
	// ]
	var cpuCount = systemInfo["CPUCount"]
	var sysCpuUsage = systemInfo["SystemCPUUsage"].(map[string]interface{})

	if processInfoList["Processes"] == nil {
		fmt.Printf("invalid process info list: %v\n", processInfoList)
		return
	}
	// processInfoList example:
	// map[
	//   Processes:map[
	//     0:[108940918784 0.35006396059439243 11867680 6069179 147456 294600704 167346176 0]
	//     100:[417741438976 0 65418 21019 1819088 7045120 1671168 100]
	//     107:[417746960384 0.06996075980775063 71187 21226 3342800 9420800 3178496 107]
	//   ]
	// 	 StartMachAbsTime:2514086593642
	//   EndMachAbsTime:2514112708690
	//   Type:5
	// ]
	processes := processInfoList["Processes"].(map[string]interface{})

	cpuInfo := CPUData{
		Type:                 "cpu",
		TimeStamp:            time.Now().Unix(),
		CPUCount:             int(cpuCount.(uint64)),
		SysCPUUsageTotalLoad: sysCpuUsage["CPU_TotalLoad"].(float64),
		ProcPID:              pid,
	}
	memInfo := MemData{
		Type:      "mem",
		TimeStamp: time.Now().Unix(),
		ProcPID:   pid,
	}

	procData, ok := processes[pid]
	processInfo := convertProcessData(procData)
	if ok && processInfo != nil {
		// cpu
		cpuInfo.ProcCPUUsage = processInfo["cpuUsage"].(float64)
		cpuInfo.ProcAttrCtxSwitch = convert2Int64(processInfo["ctxSwitch"])
		cpuInfo.ProcAttrIntWakeups = convert2Int64(processInfo["intWakeups"])
		// mem
		memInfo.Vss = convert2Int64(processInfo["memVirtualSize"])
		memInfo.Rss = convert2Int64(processInfo["memResidentSize"])
		memInfo.Anon = convert2Int64(processInfo["memAnon"])
		memInfo.PhysMemory = convert2Int64(processInfo["physFootprint"])
	} else {
		// cpu
		cpuInfo.Msg = "invalid PID"
		// mem
		memInfo.Msg = "invalid PID"
		memInfo.Vss = -1
		memInfo.Rss = -1
		memInfo.Anon = -1
		memInfo.PhysMemory = -1
	}

	cpuBytes, _ := json.Marshal(cpuInfo)
	memBytes, _ := json.Marshal(memInfo)
	c.chanCPU <- cpuBytes
	c.chanMem <- memBytes
}

type CPUData struct {
	Type      string `json:"type"` // cpu
	TimeStamp int64  `json:"timestamp"`
	Msg       string `json:"msg,omitempty"` // 提示信息
	// system
	CPUCount             int     `json:"cpuCount"`             // CPU总数
	SysCPUUsageTotalLoad float64 `json:"sysCpuUsageTotalLoad"` // 系统总体CPU占用
	// process
	ProcPID            string  `json:"procPID"`                      // 进程 PID
	ProcCPUUsage       float64 `json:"procCpuUsage,omitempty"`       // 单个进程的CPU使用率
	ProcAttrCtxSwitch  int64   `json:"procAttrCtxSwitch,omitempty"`  // 上下文切换数
	ProcAttrIntWakeups int64   `json:"procAttrIntWakeups,omitempty"` // 唤醒数
}

type MemData struct {
	Type       string `json:"type"` // mem
	TimeStamp  int64  `json:"timestamp"`
	Anon       int64  `json:"anon"`          // 虚拟内存
	PhysMemory int64  `json:"physMemory"`    // 物理内存
	Rss        int64  `json:"rss"`           // 总内存
	Vss        int64  `json:"vss"`           // 虚拟内存
	ProcPID    string `json:"procPID"`       // 进程 PID
	Msg        string `json:"msg,omitempty"` // 提示信息
}

func convertProcessData(procData interface{}) map[string]interface{} {
	if procData == nil {
		return nil
	}
	procDataArray, ok := procData.([]interface{})
	if !ok {
		return nil
	}
	if len(procDataArray) != 8 {
		return nil
	}

	// procDataArray example:
	// [417741438976 0 65418 21019 1819088 7045120 1671168 100]
	// corresponds to procAttrs:
	// ["memVirtualSize", "cpuUsage", "ctxSwitch", "intWakeups",
	// "physFootprint", "memResidentSize", "memAnon", "pid"]
	return map[string]interface{}{
		"memVirtualSize":  procDataArray[0],
		"cpuUsage":        procDataArray[1],
		"ctxSwitch":       procDataArray[2],
		"intWakeups":      procDataArray[3],
		"physFootprint":   procDataArray[4],
		"memResidentSize": procDataArray[5],
		"memAnon":         procDataArray[6],
		"PID":             procDataArray[7],
	}
}

func convert2Int64(num interface{}) int64 {
	switch value := num.(type) {
	case uint64:
		return int64(value)
	case uint32:
		return int64(value)
	case uint16:
		return int64(value)
	case uint8:
		return int64(value)
	case uint:
		return int64(value)
	}
	fmt.Printf("convert2Int64 failed: %v, %T\n", num, num)
	return -1
}
