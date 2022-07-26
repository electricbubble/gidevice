package giDevice

import (
	"context"
	"encoding/json"
	"fmt"
	perfEntity "github.com/electricbubble/gidevice/pkg/performance"
	"time"

	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
)

var _ Instruments = (*instruments)(nil)

func newInstruments(client *libimobiledevice.InstrumentsClient) *instruments {
	return &instruments{
		client: client,
	}
}

type instruments struct {
	client *libimobiledevice.InstrumentsClient
}

func (i *instruments) notifyOfPublishedCapabilities() (err error) {
	_, err = i.client.NotifyOfPublishedCapabilities()
	return
}

func (i *instruments) requestChannel(channel string) (id uint32, err error) {
	return i.client.RequestChannel(channel)
}

func (i *instruments) AppLaunch(bundleID string, opts ...AppLaunchOption) (pid int, err error) {
	opt := new(appLaunchOption)
	opt.appPath = ""
	opt.options = map[string]interface{}{
		"StartSuspendedKey": uint64(0),
		"KillExisting":      uint64(0),
	}
	if len(opts) != 0 {
		for _, optFunc := range opts {
			optFunc(opt)
		}
	}

	var id uint32
	if id, err = i.requestChannel("com.apple.instruments.server.services.processcontrol"); err != nil {
		return 0, err
	}

	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(opt.appPath); err != nil {
		return 0, err
	}
	if err = args.AppendObject(bundleID); err != nil {
		return 0, err
	}
	if err = args.AppendObject(opt.environment); err != nil {
		return 0, err
	}
	if err = args.AppendObject(opt.arguments); err != nil {
		return 0, err
	}
	if err = args.AppendObject(opt.options); err != nil {
		return 0, err
	}

	var result *libimobiledevice.DTXMessageResult
	selector := "launchSuspendedProcessWithDevicePath:bundleIdentifier:environment:arguments:options:"
	if result, err = i.client.Invoke(selector, args, id, true); err != nil {
		return 0, err
	}

	if nsErr, ok := result.Obj.(libimobiledevice.NSError); ok {
		return 0, fmt.Errorf("%s", nsErr.NSUserInfo.(map[string]interface{})["NSLocalizedDescription"])
	}

	return int(result.Obj.(uint64)), nil
}

func (i *instruments) appProcess(bundleID string) (err error) {
	var id uint32
	if id, err = i.requestChannel("com.apple.instruments.server.services.processcontrol"); err != nil {
		return err
	}

	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(bundleID); err != nil {
		return err
	}

	selector := "processIdentifierForBundleIdentifier:"
	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return err
	}

	return
}

func (i *instruments) startObserving(pid int) (err error) {
	var id uint32
	if id, err = i.requestChannel("com.apple.instruments.server.services.processcontrol"); err != nil {
		return err
	}

	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(pid); err != nil {
		return err
	}

	var result *libimobiledevice.DTXMessageResult
	selector := "startObservingPid:"
	if result, err = i.client.Invoke(selector, args, id, true); err != nil {
		return err
	}

	if nsErr, ok := result.Obj.(libimobiledevice.NSError); ok {
		return fmt.Errorf("%s", nsErr.NSUserInfo.(map[string]interface{})["NSLocalizedDescription"])
	}
	return
}

func (i *instruments) AppKill(pid int) (err error) {
	var id uint32
	if id, err = i.requestChannel("com.apple.instruments.server.services.processcontrol"); err != nil {
		return err
	}

	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(pid); err != nil {
		return err
	}

	selector := "killPid:"
	if _, err = i.client.Invoke(selector, args, id, false); err != nil {
		return err
	}

	return
}

func (i *instruments) AppRunningProcesses() (processes []Process, err error) {
	var id uint32
	if id, err = i.requestChannel("com.apple.instruments.server.services.deviceinfo"); err != nil {
		return nil, err
	}

	selector := "runningProcesses"

	var result *libimobiledevice.DTXMessageResult
	if result, err = i.client.Invoke(selector, libimobiledevice.NewAuxBuffer(), id, true); err != nil {
		return nil, err
	}

	objs := result.Obj.([]interface{})

	processes = make([]Process, 0, len(objs))

	for _, v := range objs {
		m := v.(map[string]interface{})

		var data []byte
		if data, err = json.Marshal(m); err != nil {
			debugLog(fmt.Sprintf("process marshal: %v\n%v\n", err, m))
			err = nil
			continue
		}

		var tp Process
		if err = json.Unmarshal(data, &tp); err != nil {
			debugLog(fmt.Sprintf("process unmarshal: %v\n%v\n", err, m))
			err = nil
			continue
		}

		processes = append(processes, tp)
	}

	return
}

func (i *instruments) AppList(opts ...AppListOption) (apps []Application, err error) {
	opt := new(appListOption)
	opt.updateToken = ""
	opt.appsMatching = make(map[string]interface{})
	if len(opts) != 0 {
		for _, optFunc := range opts {
			optFunc(opt)
		}
	}

	var id uint32
	if id, err = i.requestChannel("com.apple.instruments.server.services.device.applictionListing"); err != nil {
		return nil, err
	}

	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(opt.appsMatching); err != nil {
		return nil, err
	}
	if err = args.AppendObject(opt.updateToken); err != nil {
		return nil, err
	}

	selector := "installedApplicationsMatching:registerUpdateToken:"

	var result *libimobiledevice.DTXMessageResult
	if result, err = i.client.Invoke(selector, args, id, true); err != nil {
		return nil, err
	}

	objs := result.Obj.([]interface{})

	for _, v := range objs {
		m := v.(map[string]interface{})

		var data []byte
		if data, err = json.Marshal(m); err != nil {
			debugLog(fmt.Sprintf("application marshal: %v\n%v\n", err, m))
			err = nil
			continue
		}

		var app Application
		if err = json.Unmarshal(data, &app); err != nil {
			debugLog(fmt.Sprintf("application unmarshal: %v\n%v\n", err, m))
			err = nil
			continue
		}
		apps = append(apps, app)
	}

	return
}

func (i *instruments) DeviceInfo() (devInfo *DeviceInfo, err error) {
	var id uint32
	if id, err = i.requestChannel("com.apple.instruments.server.services.deviceinfo"); err != nil {
		return nil, err
	}

	selector := "systemInformation"

	var result *libimobiledevice.DTXMessageResult
	if result, err = i.client.Invoke(selector, libimobiledevice.NewAuxBuffer(), id, true); err != nil {
		return nil, err
	}

	data, err := json.Marshal(result.Obj)
	if err != nil {
		return nil, err
	}
	devInfo = new(DeviceInfo)
	err = json.Unmarshal(data, devInfo)

	return
}

func (i *instruments) registerCallback(obj string, cb func(m libimobiledevice.DTXMessageResult)) {
	i.client.RegisterCallback(obj, cb)
}

func (i *instruments) StartSysmontapServer(pid string, ctxParent context.Context) (chanCPU chan perfEntity.CPUInfo, chanMem chan perfEntity.MEMInfo, cancel context.CancelFunc, err error) {
	var id uint32
	if ctxParent == nil {
		return nil, nil, nil, fmt.Errorf("missing context")
	}
	ctx, cancelFunc := context.WithCancel(ctxParent)
	_outMEM := make(chan perfEntity.MEMInfo)
	_outCPU := make(chan perfEntity.CPUInfo)
	if id, err = i.requestChannel("com.apple.instruments.server.services.sysmontap"); err != nil {
		return nil, nil, cancelFunc, err
	}

	selector := "setConfig:"
	args := libimobiledevice.NewAuxBuffer()

	var config map[string]interface{}
	config = make(map[string]interface{})
	{
		config["bm"] = 0
		config["cpuUsage"] = true

		config["procAttrs"] = []string{
			"memVirtualSize", "cpuUsage", "ctxSwitch", "intWakeups",
			"physFootprint", "memResidentSize", "memAnon", "pid"}

		config["sampleInterval"] = 1000000000
		// 系统信息字段
		config["sysAttrs"] = []string{
			"vmExtPageCount", "vmFreeCount", "vmPurgeableCount",
			"vmSpeculativeCount", "physMemSize"}
		// 刷新频率
		config["ur"] = 1000
	}

	args.AppendObject(config)
	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return nil, nil, cancelFunc, err
	}
	selector = "start"
	args = libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return nil, nil, cancelFunc, err
	}

	i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			return
		default:
			mess := m.Obj
			chanCPUAndMEMData(mess, _outMEM, _outCPU, pid)
		}
	})

	go func() {
		i.registerCallback("_Golang-iDevice_Over", func(_ libimobiledevice.DTXMessageResult) {
			cancelFunc()
		})
		select {
		case <-ctx.Done():
			var isOpen bool
			if _outCPU != nil {
				_, isOpen = <-_outMEM
				if isOpen {
					close(_outMEM)
				}
			}
			if _outMEM != nil {
				_, isOpen = <-_outCPU
				if isOpen {
					close(_outCPU)
				}
			}
		}
		return
	}()
	return _outCPU, _outMEM, cancelFunc, err
}

func chanCPUAndMEMData(mess interface{}, _outMEM chan perfEntity.MEMInfo, _outCPU chan perfEntity.CPUInfo, pid string) {
	switch mess.(type) {
	case []interface{}:
		var infoCPU perfEntity.CPUInfo
		var infoMEM perfEntity.MEMInfo
		messArray := mess.([]interface{})
		if len(messArray) == 2 {
			var sinfo = messArray[0].(map[string]interface{})
			var pinfolist = messArray[1].(map[string]interface{})
			if sinfo["CPUCount"] == nil {
				var temp = sinfo
				sinfo = pinfolist
				pinfolist = temp
			}
			if sinfo["CPUCount"] != nil && pinfolist["Processes"] != nil {
				var cpuCount = sinfo["CPUCount"]
				var sysCpuUsage = sinfo["SystemCPUUsage"].(map[string]interface{})
				var cpuTotalLoad = sysCpuUsage["CPU_TotalLoad"]
				// 构建返回信息
				infoCPU.CPUCount = int(cpuCount.(uint64))
				infoCPU.SysCpuUsage = cpuTotalLoad.(float64)
				//finalCpuInfo["attrCpuTotal"] = cpuTotalLoad
				infoCPU.TimeStamp = time.Now().UnixNano()

				var cpuUsage = 0.0
				pidMess := pinfolist["Processes"].(map[string]interface{})[pid]
				infoCPU.Pid = "invalid PID"
				if pidMess != nil {
					processInfo := sysmonPorceAttrs(pidMess)
					cpuUsage = processInfo["cpuUsage"].(float64)
					infoCPU.CPUUsage = cpuUsage
					infoCPU.Pid = pid
					infoCPU.AttrCtxSwitch = uIntToInt64(processInfo["ctxSwitch"])
					infoCPU.AttrIntWakeups = uIntToInt64(processInfo["intWakeups"])
					_outCPU <- infoCPU

					infoMEM.Vss = uIntToInt64(processInfo["memVirtualSize"])
					infoMEM.Rss = uIntToInt64(processInfo["memResidentSize"])
					infoMEM.Anon = uIntToInt64(processInfo["memAnon"])
					infoMEM.PhysMemory = uIntToInt64(processInfo["physFootprint"])
					infoMEM.TimeStamp = time.Now().UnixNano()
					_outMEM <- infoMEM

				}
			}
		}
	}
}

// 获取进程相关信息
func sysmonPorceAttrs(cpuMess interface{}) (outCpuInfo map[string]interface{}) {
	if cpuMess == nil {
		return nil
	}
	cpuMessArray, ok := cpuMess.([]interface{})
	if !ok {
		return nil
	}
	if len(cpuMessArray) != 8 {
		return nil
	}
	if outCpuInfo == nil {
		outCpuInfo = map[string]interface{}{}
	}
	// 虚拟内存
	outCpuInfo["memVirtualSize"] = cpuMessArray[0]
	// CPU
	outCpuInfo["cpuUsage"] = cpuMessArray[1]
	// 每秒进程的上下文切换次数
	outCpuInfo["ctxSwitch"] = cpuMessArray[2]
	// 每秒进程唤醒的线程数
	outCpuInfo["intWakeups"] = cpuMessArray[3]
	// 物理内存
	outCpuInfo["physFootprint"] = cpuMessArray[4]

	outCpuInfo["memResidentSize"] = cpuMessArray[5]
	// 匿名内存
	outCpuInfo["memAnon"] = cpuMessArray[6]

	outCpuInfo["pid"] = cpuMessArray[7]
	return
}

func (i *instruments) StopSysmontapServer() (err error) {
	id, err := i.requestChannel("com.apple.instruments.server.services.sysmontap")
	if err != nil {
		return err
	}
	selector := "stop"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return err
	}
	return nil
}

// todo 获取单进程流量情况，看情况做不做
// 目前只获取到系统全局的流量情况，单进程需要用到set，go没有，并且实际用python测试单进程的流量情况不准
func (i *instruments) StartNetWorkingServer(ctxParent context.Context) (chanNetWorking chan perfEntity.NetWorkingInfo, cancel context.CancelFunc, err error) {
	var id uint32
	if ctxParent == nil {
		return nil, nil, fmt.Errorf("missing context")
	}
	ctx, cancelFunc := context.WithCancel(ctxParent)
	_outNetWork := make(chan perfEntity.NetWorkingInfo)
	if id, err = i.requestChannel("com.apple.instruments.server.services.networking"); err != nil {
		return nil, cancelFunc, err
	}
	selector := "startMonitoring"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return nil, cancelFunc, err
	}
	i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			return
		default:
			receData, ok := m.Obj.([]interface{})
			if ok && len(receData) == 2 {
				sendAndReceiveData, ok := receData[1].([]interface{})
				if ok {
					var netData perfEntity.NetWorkingInfo
					// 有时候是uint8，有时候是uint64。。。恶心
					netData.RxBytes = uIntToInt64(sendAndReceiveData[0])
					netData.RxPackets = uIntToInt64(sendAndReceiveData[1])
					netData.TxBytes = uIntToInt64(sendAndReceiveData[2])
					netData.TxPackets = uIntToInt64(sendAndReceiveData[3])
					netData.TimeStamp = time.Now().UnixNano()
					_outNetWork <- netData
				}
			}
		}
	})
	go func() {
		i.registerCallback("_Golang-iDevice_Over", func(_ libimobiledevice.DTXMessageResult) {
			cancelFunc()
		})
		select {
		case <-ctx.Done():
			_, isOpen := <-_outNetWork
			if isOpen {
				close(_outNetWork)
			}
		}
		return
	}()
	return _outNetWork, cancelFunc, err
}

func uIntToInt64(num interface{}) (cnum int64) {
	switch num.(type) {
	case uint64:
		return int64(num.(uint64))
	case uint32:
		return int64(num.(uint32))
	case uint16:
		return int64(num.(uint16))
	case uint8:
		return int64(num.(uint8))
	case uint:
		return int64(num.(uint))
	}
	return -1
}

func (i *instruments) StopNetWorkingServer() (err error) {
	var id uint32
	id, err = i.requestChannel("com.apple.instruments.server.services.networking")
	if err != nil {
		return err
	}
	selector := "stopMonitoring"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return err
	}
	return nil
}

func (i *instruments) StartOpenglServer(ctxParent context.Context) (chanFPS chan perfEntity.FPSInfo, chanGPU chan perfEntity.GPUInfo, cancel context.CancelFunc, err error) {
	var id uint32
	if ctxParent == nil {
		return nil, nil, nil, fmt.Errorf("missing context")
	}
	ctx, cancelFunc := context.WithCancel(ctxParent)
	_outFPS := make(chan perfEntity.FPSInfo)
	_outGPU := make(chan perfEntity.GPUInfo)
	if id, err = i.requestChannel("com.apple.instruments.server.services.graphics.opengl"); err != nil {
		return nil, nil, cancelFunc, err
	}

	selector := "availableStatistics"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return nil, nil, cancelFunc, err
	}

	selector = "setSamplingRate:"
	if err = args.AppendObject(0.0); err != nil {
		return nil, nil, cancelFunc, err
	}
	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return nil, nil, cancelFunc, err
	}

	selector = "startSamplingAtTimeInterval:processIdentifier:"
	args = libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(0); err != nil {
		return nil, nil, cancelFunc, err
	}
	if err = args.AppendObject(0); err != nil {
		return nil, nil, cancelFunc, err
	}
	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return nil, nil, cancelFunc, err
	}

	i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			return
		default:
			mess := m.Obj
			var deviceUtilization = mess.(map[string]interface{})["Device Utilization %"]     // Device Utilization
			var tilerUtilization = mess.(map[string]interface{})["Tiler Utilization %"]       // Tiler Utilization
			var rendererUtilization = mess.(map[string]interface{})["Renderer Utilization %"] // Renderer Utilization

			var infoGPU perfEntity.GPUInfo

			infoGPU.DeviceUtilization = uIntToInt64(deviceUtilization)
			infoGPU.TilerUtilization = uIntToInt64(tilerUtilization)
			infoGPU.RendererUtilization = uIntToInt64(rendererUtilization)
			infoGPU.TimeStamp = time.Now().UnixNano()
			_outGPU <- infoGPU

			var infoFPS perfEntity.FPSInfo
			var fps = mess.(map[string]interface{})["CoreAnimationFramesPerSecond"]
			infoFPS.FPS = int(uIntToInt64(fps))
			infoFPS.TimeStamp = time.Now().UnixNano()
			_outFPS <- infoFPS
		}
	})
	go func() {
		i.registerCallback("_Golang-iDevice_Over", func(_ libimobiledevice.DTXMessageResult) {
			cancelFunc()
		})
		select {
		case <-ctx.Done():
			var isOpen bool
			if _outGPU != nil {
				_, isOpen = <-_outGPU
				if isOpen {
					close(_outGPU)
				}
			}
			if _outFPS != nil {
				_, isOpen = <-_outFPS
				if isOpen {
					close(_outFPS)
				}
			}
		}
		return
	}()
	return _outFPS, _outGPU, cancelFunc, err
}

func (i *instruments) StopOpenglServer() (err error) {

	id, err := i.requestChannel("com.apple.instruments.server.services.graphics.opengl")
	if err != nil {
		return err
	}
	selector := "stop"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return err
	}
	return nil
}

type Application struct {
	AppExtensionUUIDs         []string `json:"AppExtensionUUIDs,omitempty"`
	BundlePath                string   `json:"BundlePath"`
	CFBundleIdentifier        string   `json:"CFBundleIdentifier"`
	ContainerBundleIdentifier string   `json:"ContainerBundleIdentifier,omitempty"`
	ContainerBundlePath       string   `json:"ContainerBundlePath,omitempty"`
	DisplayName               string   `json:"DisplayName"`
	ExecutableName            string   `json:"ExecutableName,omitempty"`
	Placeholder               bool     `json:"Placeholder,omitempty"`
	PluginIdentifier          string   `json:"PluginIdentifier,omitempty"`
	PluginUUID                string   `json:"PluginUUID,omitempty"`
	Restricted                int      `json:"Restricted"`
	Type                      string   `json:"Type"`
	Version                   string   `json:"Version"`
}

type DeviceInfo struct {
	Description       string `json:"_deviceDescription"`
	DisplayName       string `json:"_deviceDisplayName"`
	Identifier        string `json:"_deviceIdentifier"`
	Version           string `json:"_deviceVersion"`
	ProductType       string `json:"_productType"`
	ProductVersion    string `json:"_productVersion"`
	XRDeviceClassName string `json:"_xrdeviceClassName"`
}
