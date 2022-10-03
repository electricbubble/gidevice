package giDevice

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
)

// instruments services
const (
	instrumentsServiceDeviceInfo              = "com.apple.instruments.server.services.deviceinfo"
	instrumentsServiceProcessControl          = "com.apple.instruments.server.services.processcontrol"
	instrumentsServiceDeviceApplictionListing = "com.apple.instruments.server.services.device.applictionListing"
	instrumentsServiceGraphicsOpengl          = "com.apple.instruments.server.services.graphics.opengl"        // 获取FPS
	instrumentsServiceSysmontap               = "com.apple.instruments.server.services.sysmontap"              // 获取 CPU/Mem 性能数据
	instrumentsServiceXcodeNetworkStatistics  = "com.apple.xcode.debug-gauge-data-providers.NetworkStatistics" // 获取单进程网络数据
	instrumentsServiceXcodeEnergyStatistics   = "com.apple.xcode.debug-gauge-data-providers.Energy"            // 获取功耗数据
	instrumentsServiceNetworking              = "com.apple.instruments.server.services.networking"             // 获取全局网络数据
	instrumentsServiceMobileNotifications     = "com.apple.instruments.server.services.mobilenotifications"    // 监控应用状态
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
	if id, err = i.requestChannel(instrumentsServiceProcessControl); err != nil {
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
	if id, err = i.requestChannel(instrumentsServiceProcessControl); err != nil {
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
	if id, err = i.requestChannel(instrumentsServiceProcessControl); err != nil {
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
	if id, err = i.requestChannel(instrumentsServiceProcessControl); err != nil {
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
	if id, err = i.requestChannel(instrumentsServiceDeviceInfo); err != nil {
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
	if id, err = i.requestChannel(instrumentsServiceDeviceApplictionListing); err != nil {
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
	if id, err = i.requestChannel(instrumentsServiceDeviceInfo); err != nil {
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

func (i *instruments) StartSysmontapServer(pid string, ctxParent context.Context) (
	chanCPU chan CPUData, chanMem chan MemData, cancel context.CancelFunc, err error) {

	var id uint32
	if ctxParent == nil {
		return nil, nil, nil, fmt.Errorf("missing context")
	}
	ctx, cancelFunc := context.WithCancel(ctxParent)
	_outMEM := make(chan MemData)
	_outCPU := make(chan CPUData)
	if id, err = i.requestChannel(instrumentsServiceSysmontap); err != nil {
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

func chanCPUAndMEMData(mess interface{}, _outMEM chan MemData, _outCPU chan CPUData, pid string) {
	switch mess.(type) {
	case []interface{}:
		var infoCPU CPUData
		var infoMEM MemData
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
				infoCPU.SysCPUUsageTotalLoad = cpuTotalLoad.(float64)
				//finalCpuInfo["attrCpuTotal"] = cpuTotalLoad

				var cpuUsage = 0.0
				pidMess := pinfolist["Processes"].(map[string]interface{})[pid]
				if pidMess != nil {
					processInfo := convertProcessData(pidMess)
					cpuUsage = processInfo["cpuUsage"].(float64)
					infoCPU.ProcCPUUsage = cpuUsage
					infoCPU.ProcPID = pid
					infoCPU.ProcAttrCtxSwitch = convert2Int64(processInfo["ctxSwitch"])
					infoCPU.ProcAttrIntWakeups = convert2Int64(processInfo["intWakeups"])

					infoMEM.Vss = convert2Int64(processInfo["memVirtualSize"])
					infoMEM.Rss = convert2Int64(processInfo["memResidentSize"])
					infoMEM.Anon = convert2Int64(processInfo["memAnon"])
					infoMEM.PhysMemory = convert2Int64(processInfo["physFootprint"])

				} else {
					infoCPU.Msg = "invalid PID"
					infoMEM.Msg = "invalid PID"

					infoMEM.Vss = -1
					infoMEM.Rss = -1
					infoMEM.Anon = -1
					infoMEM.PhysMemory = -1
				}

				infoMEM.TimeStamp = time.Now().UnixNano()
				infoCPU.TimeStamp = time.Now().UnixNano()

				_outMEM <- infoMEM
				_outCPU <- infoCPU
			}
		}
	}
}

func (i *instruments) StopSysmontapServer() (err error) {
	id, err := i.requestChannel(instrumentsServiceSysmontap)
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
func (i *instruments) StartNetWorkingServer(ctxParent context.Context) (chanNetWorking chan NetWorkingInfo, cancel context.CancelFunc, err error) {
	var id uint32
	if ctxParent == nil {
		return nil, nil, fmt.Errorf("missing context")
	}
	ctx, cancelFunc := context.WithCancel(ctxParent)
	_outNetWork := make(chan NetWorkingInfo)
	if id, err = i.requestChannel(instrumentsServiceNetworking); err != nil {
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
					var netData NetWorkingInfo
					// 有时候是uint8，有时候是uint64。。。恶心
					netData.RxBytes = convert2Int64(sendAndReceiveData[0])
					netData.RxPackets = convert2Int64(sendAndReceiveData[1])
					netData.TxBytes = convert2Int64(sendAndReceiveData[2])
					netData.TxPackets = convert2Int64(sendAndReceiveData[3])
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

func (i *instruments) StopNetWorkingServer() (err error) {
	var id uint32
	id, err = i.requestChannel(instrumentsServiceNetworking)
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

func (i *instruments) StartOpenglServer(ctxParent context.Context) (chanFPS chan FPSInfo, chanGPU chan GPUInfo, cancel context.CancelFunc, err error) {
	var id uint32
	if ctxParent == nil {
		return nil, nil, nil, fmt.Errorf("missing context")
	}
	ctx, cancelFunc := context.WithCancel(ctxParent)
	_outFPS := make(chan FPSInfo)
	_outGPU := make(chan GPUInfo)
	if id, err = i.requestChannel(instrumentsServiceGraphicsOpengl); err != nil {
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

			var infoGPU GPUInfo

			infoGPU.DeviceUtilization = convert2Int64(deviceUtilization)
			infoGPU.TilerUtilization = convert2Int64(tilerUtilization)
			infoGPU.RendererUtilization = convert2Int64(rendererUtilization)
			infoGPU.TimeStamp = time.Now().UnixNano()
			_outGPU <- infoGPU

			var infoFPS FPSInfo
			var fps = mess.(map[string]interface{})["CoreAnimationFramesPerSecond"]
			infoFPS.FPS = int(convert2Int64(fps))
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

	id, err := i.requestChannel(instrumentsServiceGraphicsOpengl)
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

type FPSInfo struct {
	FPS       int   `json:"fps"`
	TimeStamp int64 `json:"timeStamp"`
}

type GPUInfo struct {
	TilerUtilization    int64  `json:"tilerUtilization"` // 处理顶点的GPU时间占比
	TimeStamp           int64  `json:"timeStamp"`
	Mess                string `json:"mess,omitempty"`      // 提示信息，当PID没输入时提示
	DeviceUtilization   int64  `json:"deviceUtilization"`   // 设备利用率
	RendererUtilization int64  `json:"rendererUtilization"` // 渲染器利用率
}

type NetWorkingInfo struct {
	RxBytes   int64 `json:"rxBytes"`
	RxPackets int64 `json:"rxPackets"`
	TxBytes   int64 `json:"txBytes"`
	TxPackets int64 `json:"txPackets"`
	TimeStamp int64 `json:"timeStamp"`
}
