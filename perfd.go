package giDevice

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
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
		stop:        make(chan struct{}),
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

	if c.option.fps {
		cancel, err := c.startGetFPS(context.Background())
		if err != nil {
			return nil, err
		}
		c.cancels = append(c.cancels, cancel)
	}

	if c.option.gpu {
		cancel, err := c.startGetGPU(context.Background())
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
				if ok {
					fmt.Println("cpu: ", string(cpuBytes))
					outCh <- cpuBytes
				}
			case memBytes, ok := <-c.chanMem:
				if ok {
					fmt.Println("mem: ", string(memBytes))
					outCh <- memBytes
				}
			case gpuBytes, ok := <-c.chanGPU:
				if ok {
					fmt.Println("gpu: ", string(gpuBytes))
					outCh <- gpuBytes
				}
			case fpsBytes, ok := <-c.chanFPS:
				if ok {
					fmt.Println("fps: ", string(fpsBytes))
					outCh <- fpsBytes
				}
			case networkBytes, ok := <-c.chanNetwork:
				if ok {
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

	if _, err = c.i.call(
		instrumentsServiceNetworking,
		"replayLastRecordedSession",
	); err != nil {
		return nil, err
	}

	if _, err = c.i.call(
		instrumentsServiceNetworking,
		"startMonitoring",
	); err != nil {
		return nil, err
	}

	ctx, cancel = context.WithCancel(ctx)
	c.i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			return
		default:
			c.parseNetworking(m.Obj)
		}
	})

	return
}

func (c *perfdClient) parseNetworking(data interface{}) {
	raw, ok := data.([]interface{})
	if !ok || len(raw) != 2 {
		fmt.Printf("invalid networking data: %v\n", data)
		return
	}

	var netBytes []byte
	msgType := raw[0].(uint64)
	msgValue := raw[1].([]interface{})
	if msgType == 0 {
		// interface-detection
		// ['InterfaceIndex', "Name"]
		// e.g. [0, [14, 'en0']]
		netData := NetworkDataInterfaceDetection{
			NetworkData: NetworkData{
				Type:      "network-interface-detection",
				TimeStamp: time.Now().Unix(),
			},
			InterfaceIndex: convert2Int64(msgValue[0]),
			Name:           msgValue[1].(string),
		}
		netBytes, _ = json.Marshal(netData)
	} else if msgType == 1 {
		// connection-detected
		// ['LocalAddress', 'RemoteAddress', 'InterfaceIndex', 'Pid',
		// 'RecvBufferSize', 'RecvBufferUsed', 'SerialNumber', 'Kind']
		// e.g. [1 [[16 2 211 158 192 168 100 101 0 0 0 0 0 0 0 0]
		//       [16 2 0 53 183 221 253 100 0 0 0 0 0 0 0 0]
		//       14 -2 786896 0 133 2]]

		localAddr, err := parseSocketAddr(msgValue[0].([]byte))
		if err != nil {
			fmt.Printf("parse local socket address err: %v\n", err)
		}
		remoteAddr, err := parseSocketAddr(msgValue[1].([]byte))
		if err != nil {
			fmt.Printf("parse remote socket address err: %v\n", err)
		}
		netData := NetworkDataConnectionDetected{
			NetworkData: NetworkData{
				Type:      "network-connection-detected",
				TimeStamp: time.Now().Unix(),
			},
			LocalAddress:   localAddr,
			RemoteAddress:  remoteAddr,
			InterfaceIndex: convert2Int64(msgValue[2]),
			Pid:            convert2Int64(msgValue[3]),
			RecvBufferSize: convert2Int64(msgValue[4]),
			RecvBufferUsed: convert2Int64(msgValue[5]),
			SerialNumber:   convert2Int64(msgValue[6]),
			Kind:           convert2Int64(msgValue[7]),
		}
		netBytes, _ = json.Marshal(netData)
	} else if msgType == 2 {
		// connection-update
		// ['RxPackets', 'RxBytes', 'TxPackets', 'TxBytes',
		// 'RxDups', 'RxOOO', 'TxRetx', 'MinRTT', 'AvgRTT', 'ConnectionSerial']
		// e.g. [2, [21, 1708, 22, 14119, 309, 0, 5830, 0.076125, 0.076125, 54, -1]]
		netData := NetworkDataConnectionUpdate{
			NetworkData: NetworkData{
				Type:      "network-connection-update",
				TimeStamp: time.Now().Unix(),
			},
			RxBytes:   convert2Int64(msgValue[0]),
			RxPackets: convert2Int64(msgValue[1]),
			TxBytes:   convert2Int64(msgValue[2]),
			TxPackets: convert2Int64(msgValue[3]),
		}
		if value, ok := msgValue[4].(uint64); ok {
			netData.RxDups = int64(value)
		}
		if value, ok := msgValue[5].(uint64); ok {
			netData.RxOOO = int64(value)
		}
		if value, ok := msgValue[6].(uint64); ok {
			netData.TxRetx = int64(value)
		}
		if value, ok := msgValue[7].(uint64); ok {
			netData.MinRTT = int64(value)
		}
		if value, ok := msgValue[8].(uint64); ok {
			netData.AvgRTT = int64(value)
		}
		if value, ok := msgValue[9].(uint64); ok {
			netData.ConnectionSerial = int64(value)
		}

		netBytes, _ = json.Marshal(netData)
	}
	c.chanNetwork <- netBytes
}

func parseSocketAddr(data []byte) (string, error) {
	len := data[0]                             // length of address
	_ = data[1]                                // family
	port := binary.BigEndian.Uint16(data[2:4]) // port

	// network, data[4:4+len]
	if len == 0x10 {
		// IPv4, 4 bytes
		ip := net.IP(data[4:8])
		return fmt.Sprintf("%s:%d", ip, port), nil
	} else if len == 0x1c {
		// IPv6, 16 bytes
		ip := net.IP(data[4:20])
		return fmt.Sprintf("%s:%d", ip, port), nil
	}
	return "", fmt.Errorf("invalid socket address: %v", data)
}

type NetworkData struct {
	Type      string `json:"type"` // network
	TimeStamp int64  `json:"timestamp"`
}

// network-interface-detection
type NetworkDataInterfaceDetection struct {
	NetworkData
	InterfaceIndex int64  `json:"interface_index"` // 0
	Name           string `json:"name"`            // 1
}

// network-connection-detected
type NetworkDataConnectionDetected struct {
	NetworkData
	LocalAddress   string `json:"local_address"`    // 0
	RemoteAddress  string `json:"remote_address"`   // 1
	InterfaceIndex int64  `json:"interface_index"`  // 2
	Pid            int64  `json:"pid"`              // 3
	RecvBufferSize int64  `json:"recv_buffer_size"` // 4
	RecvBufferUsed int64  `json:"recv_buffer_used"` // 5
	SerialNumber   int64  `json:"serial_number"`    // 6
	Kind           int64  `json:"kind"`             // 7
}

// network-connection-update
type NetworkDataConnectionUpdate struct {
	NetworkData
	RxBytes          int64 `json:"rx_bytes"`          // 0
	RxPackets        int64 `json:"rx_packets"`        // 1
	TxBytes          int64 `json:"tx_bytes"`          // 2
	TxPackets        int64 `json:"tx_packets"`        // 3
	RxDups           int64 `json:"rx_dups,omitempty"` // 4
	RxOOO            int64 `json:"rx_000,omitempty"`  // 5
	TxRetx           int64 `json:"tx_retx,omitempty"` // 6
	MinRTT           int64 `json:"min_rtt,omitempty"` // 7
	AvgRTT           int64 `json:"avg_rtt,omitempty"` // 8
	ConnectionSerial int64 `json:"connection_serial"` // 9
}

func (c *perfdClient) startGetFPS(ctx context.Context) (
	cancel context.CancelFunc, err error) {

	if _, err = c.i.call(
		instrumentsServiceGraphicsOpengl,
		"setSamplingRate:",
		float64(1000)/100, // TODO: make it configurable
	); err != nil {
		return nil, err
	}

	if _, err = c.i.call(
		instrumentsServiceGraphicsOpengl,
		"startSamplingAtTimeInterval:",
		0,
	); err != nil {
		return nil, err
	}

	ctx, cancel = context.WithCancel(ctx)
	c.i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			return
		default:
			c.parseFPS(m.Obj)
		}
	})

	return
}

func (c *perfdClient) parseFPS(data interface{}) {
	// data example:
	// map[
	//   Alloc system memory:50167808
	//   Allocated PB Size:1179648
	//   CoreAnimationFramesPerSecond:0  // fps from GPU
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

	fpsInfo := FPSData{
		Type:      "fps",
		TimeStamp: time.Now().Unix(),
	}

	defer func() {
		fpsBytes, _ := json.Marshal(fpsInfo)
		c.chanFPS <- fpsBytes
	}()

	raw, ok := data.(map[string]interface{})
	if !ok {
		fpsInfo.Msg = fmt.Sprintf("invalid graphics.opengl data: %v", data)
		return
	}

	// fps
	fpsInfo.FPS = int(convert2Int64(raw["CoreAnimationFramesPerSecond"]))
}

func (c *perfdClient) startGetGPU(ctx context.Context) (
	cancel context.CancelFunc, err error) {

	if _, err = c.i.call(
		instrumentsServiceGraphicsOpengl,
		"setSamplingRate:",
		float64(1000)/100, // TODO: make it configurable
	); err != nil {
		return nil, err
	}

	if _, err = c.i.call(
		instrumentsServiceGraphicsOpengl,
		"startSamplingAtTimeInterval:",
		0,
	); err != nil {
		return nil, err
	}

	ctx, cancel = context.WithCancel(ctx)
	c.i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			return
		default:
			c.parseGPU(m.Obj)
		}
	})

	return
}

func (c *perfdClient) parseGPU(data interface{}) {
	// data example:
	// map[
	//   Alloc system memory:50167808
	//   Allocated PB Size:1179648
	//   CoreAnimationFramesPerSecond:0
	//   Device Utilization %:0          // device
	//   IOGLBundleName:Built-In
	//   In use system memory:10633216
	//   Renderer Utilization %:0        // renderer
	//   SplitSceneCount:0
	//   TiledSceneBytes:0
	//   Tiler Utilization %:0           // tiler
	//   XRVideoCardRunTimeStamp:1010679
	//   recoveryCount:0
	// ]

	gpuInfo := GPUData{
		Type:      "gpu",
		TimeStamp: time.Now().Unix(),
	}

	defer func() {
		gpuBytes, _ := json.Marshal(gpuInfo)
		c.chanGPU <- gpuBytes
	}()

	raw, ok := data.(map[string]interface{})
	if !ok {
		gpuInfo.Msg = fmt.Sprintf("invalid graphics.opengl data: %v", data)
		return
	}

	// gpu
	gpuInfo.DeviceUtilization = convert2Int64(raw["Device Utilization %"])
	gpuInfo.TilerUtilization = convert2Int64(raw["Tiler Utilization %"])
	gpuInfo.RendererUtilization = convert2Int64(raw["Renderer Utilization %"])
}

type GPUData struct {
	Type                string `json:"type"` // gpu
	TimeStamp           int64  `json:"timestamp"`
	TilerUtilization    int64  `json:"tiler_utilization"`    // 处理顶点的 GPU 时间占比
	DeviceUtilization   int64  `json:"device_utilization"`   // 设备利用率
	RendererUtilization int64  `json:"renderer_utilization"` // 渲染器利用率
	Msg                 string `json:"msg,omitempty"`        // message for invalid data
}

type FPSData struct {
	Type      string `json:"type"` // fps
	TimeStamp int64  `json:"timestamp"`
	FPS       int    `json:"fps"`
	Msg       string `json:"msg,omitempty"` // message for invalid data
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
	if _, err = c.i.client.Invoke("start", args, chanID, false); err != nil {
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
	timestamp := time.Now().Unix()
	cpuInfo := CPUData{
		Type:      "cpu",
		TimeStamp: timestamp,
		ProcPID:   pid,
	}
	memInfo := MemData{
		Type:      "mem",
		TimeStamp: timestamp,
		ProcPID:   pid,
	}

	defer func() {
		if c.option.cpu {
			cpuBytes, _ := json.Marshal(cpuInfo)
			c.chanCPU <- cpuBytes
		}
		if c.option.mem {
			memBytes, _ := json.Marshal(memInfo)
			c.chanMem <- memBytes
		}
	}()

	messArray, ok := data.([]interface{})
	if !ok || len(messArray) != 2 {
		cpuInfo.Msg = fmt.Sprintf("invalid sysmontap data: %v", data)
		memInfo.Msg = fmt.Sprintf("invalid sysmontap data: %v", data)
		return
	}

	var systemInfo = messArray[0].(map[string]interface{})
	var processInfoList = messArray[1].(map[string]interface{})
	if systemInfo["CPUCount"] == nil {
		systemInfo, processInfoList = processInfoList, systemInfo
	}
	if systemInfo["CPUCount"] == nil {
		cpuInfo.Msg = fmt.Sprintf("invalid system info: %v", systemInfo)
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
		cpuInfo.Msg = fmt.Sprintf("invalid process info list: %v", processInfoList)
		memInfo.Msg = fmt.Sprintf("invalid process info list: %v", processInfoList)
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

	// cpu
	cpuInfo.CPUCount = int(cpuCount.(uint64))
	cpuInfo.SysCPUUsageTotalLoad = sysCpuUsage["CPU_TotalLoad"].(float64)

	processes := processInfoList["Processes"].(map[string]interface{})
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
		cpuInfo.Msg = fmt.Sprintf("pid %s not found", pid)
		// mem
		memInfo.Msg = fmt.Sprintf("pid %s not found", pid)
		memInfo.Vss = -1
		memInfo.Rss = -1
		memInfo.Anon = -1
		memInfo.PhysMemory = -1
	}
}

type CPUData struct {
	Type      string `json:"type"` // cpu
	TimeStamp int64  `json:"timestamp"`
	Msg       string `json:"msg,omitempty"` // message for invalid data
	// system
	CPUCount             int     `json:"cpu_count"`     // CPU总数
	SysCPUUsageTotalLoad float64 `json:"sys_cpu_usage"` // 系统总体CPU占用
	// process
	ProcPID            string  `json:"pid"`                             // 进程 PID
	ProcCPUUsage       float64 `json:"proc_cpu_usage,omitempty"`        // 单个进程的CPU使用率
	ProcAttrCtxSwitch  int64   `json:"proc_attr_ctx_switch,omitempty"`  // 上下文切换数
	ProcAttrIntWakeups int64   `json:"proc_attr_int_wakeups,omitempty"` // 唤醒数
}

type MemData struct {
	Type       string `json:"type"` // mem
	TimeStamp  int64  `json:"timestamp"`
	Anon       int64  `json:"anon"`          // 虚拟内存
	PhysMemory int64  `json:"phys_memory"`   // 物理内存
	Rss        int64  `json:"rss"`           // 总内存
	Vss        int64  `json:"vss"`           // 虚拟内存
	ProcPID    string `json:"pid"`           // 进程 PID
	Msg        string `json:"msg,omitempty"` // message for invalid data
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
	case int64:
		return value
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
