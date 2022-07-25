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

func (i *instruments) StartSysmontapServer(pid string) (chanMem chan perfEntity.MEMInfo, chanCPU chan perfEntity.CPUInfo, cancel context.CancelFunc, err error) {
	var id uint32
	ctx, cancelFunc := context.WithCancel(context.TODO())
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
		mess := m.Obj
		chanCPUAndMEMData(mess,_outMEM,_outCPU,pid)
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
	return _outMEM, _outCPU, cancelFunc, err
}

func chanCPUAndMEMData(mess interface{},_outMEM chan perfEntity.MEMInfo, _outCPU chan perfEntity.CPUInfo,pid string)  {
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
				infoCPU.CPUCount = cpuCount.(int)
				infoCPU.SysCpuUsage = cpuTotalLoad.(float64)
				//finalCpuInfo["attrCpuTotal"] = cpuTotalLoad
				infoCPU.TimeStamp = time.Now().Unix()

				var cpuUsage = 0.0
				pidMess := pinfolist["Processes"].(map[string]interface{})[pid]
				if pidMess != nil {
					processInfo := sysmonPorceAttrs(pidMess)
					cpuUsage = processInfo["cpuUsage"].(float64)
					infoCPU.CPUUsage = cpuUsage
					infoCPU.Pid = pid
					infoCPU.AttrCtxSwitch = processInfo["ctxSwitch"].(int)
					infoCPU.AttrIntWakeups = processInfo["intWakeups"].(int)
					_outCPU <- infoCPU

					infoMEM.Vss = processInfo["memVirtualSize"].(int)
					infoMEM.Rss = processInfo["memResidentSize"].(int)
					infoMEM.Anon = processInfo["memAnon"].(int)
					infoMEM.PhysMemory = processInfo["physFootprint"].(int)
					infoMEM.TimeStamp = time.Now().Unix()
					_outMEM <- infoMEM
				}
			}
		}
	}
}

func (i *instruments) StopSysmontapServer() {
	id, err := i.requestChannel("com.apple.instruments.server.services.sysmontap")
	if err != nil {
		return
	}
	selector := "stop"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return
	}
	return
}

// todo 获取单进程流量情况，看情况做不做
// 目前只获取到系统全局的流量情况，单进程需要用到set，go没有，并且实际用python测试单进程的流量情况不准
func (i *instruments) StartNetWorkingServer() (chanNetWorking chan perfEntity.NetWorkingInfo, cancel context.CancelFunc, err error) {
	var id uint32
	ctx, cancelFunc := context.WithCancel(context.TODO())
	netWorkData := make(chan perfEntity.NetWorkingInfo)
	if id, err = i.requestChannel("com.apple.instruments.server.services.networking"); err != nil {
		return nil, cancelFunc, err
	}
	selector := "startMonitoring"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return nil, cancelFunc, err
	}
	i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		receData, ok := m.Obj.([]interface{})
		if ok && len(receData) == 2 {
			sendAndReceiveData, ok := receData[1].([]interface{})
			if ok {
				var netData perfEntity.NetWorkingInfo
				netData.RxBytes = sendAndReceiveData[0].(int)
				netData.RxPackets = sendAndReceiveData[1].(int)
				netData.TxBytes = sendAndReceiveData[2].(int)
				netData.TxPackets = sendAndReceiveData[3].(int)
				netWorkData <- netData
			}
		}
	})
	go func() {
		i.registerCallback("_Golang-iDevice_Over", func(_ libimobiledevice.DTXMessageResult) {
			cancelFunc()
		})
		select {
		case <-ctx.Done():
			_, isOpen := <-netWorkData
			if isOpen {
				close(netWorkData)
			}
		}
		return
	}()
	return netWorkData, cancelFunc, err
}

func (i *instruments) StopNetWorkingServer() {
	var id uint32
	id, err := i.requestChannel("com.apple.instruments.server.services.networking")
	if err != nil {
		return
	}
	selector := "stopMonitoring"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return
	}
}

func (i *instruments) StartOpenglServer() (chanFPS chan perfEntity.FPSInfo,chanGPU chan perfEntity.GPUInfo, cancel context.CancelFunc, err error) {
	var id uint32
	ctx, cancelFunc := context.WithCancel(context.TODO())
	_outFPS := make(chan interface{})
	_outGPU := make(chan interface{})
	if id, err = i.requestChannel("com.apple.instruments.server.services.graphics.opengl"); err != nil {
		return nil,nil, cancelFunc, err
	}

	selector := "availableStatistics"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return nil,nil, cancelFunc, err
	}

	selector = "setSamplingRate:"
	if err = args.AppendObject(10); err != nil {
		return nil,nil, cancelFunc, err
	}
	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return nil,nil, cancelFunc, err
	}

	selector = "startSamplingAtTimeInterval:processIdentifier:"
	args = libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(0); err != nil {
		return _out, cancelFunc, err
	}
	if err = args.AppendObject(0); err != nil {
		return _out, cancelFunc, err
	}
	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return _out, cancelFunc, err
	}

	i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		_out <- m.Obj
	})
	go func() {
		i.registerCallback("_Golang-iDevice_Over", func(_ libimobiledevice.DTXMessageResult) {
			cancelFunc()
		})
		select {
		case <-ctx.Done():
			_, isOpen := <-_out
			if isOpen {
				close(_out)
			}
		}
		return
	}()
	return _out, cancelFunc, err
}

func (i *instruments) StopOpenglServer() {

	id, err := i.requestChannel("com.apple.instruments.server.services.graphics.opengl")
	if err != nil {
		return
	}
	selector := "stop"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return
	}
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
