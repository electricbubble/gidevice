package giDevice

import (
	"context"
	"encoding/json"
	"fmt"
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

func (i *instruments) SysmontapServer() (out <-chan interface{}, cancel context.CancelFunc, err error) {
	var id uint32
	_, cancelFunc := context.WithCancel(context.TODO())
	_out := make(chan interface{})
	if id, err = i.requestChannel("com.apple.instruments.server.services.sysmontap"); err != nil {
		return _out, cancelFunc, err
	}

	selector := "setConfig:"
	args := libimobiledevice.NewAuxBuffer()

	var config map[string]interface{}
	config = make(map[string]interface{})
	{
		config["bm"] = 0
		config["cpuUsage"] = true
		// 输出所有进程信息字段，字段顺序与自定义相同（全量自字段，按需使用）
		//config["procAttrs"] = []string{
		//	"memVirtualSize", "cpuUsage", "procStatus", "appSleep",
		//	"uid", "vmPageIns", "memRShrd", "ctxSwitch", "memCompressed",
		//	"intWakeups", "cpuTotalSystem", "responsiblePID", "physFootprint",
		//	"cpuTotalUser", "sysCallsUnix", "memResidentSize", "sysCallsMach",
		//	"memPurgeable", "diskBytesRead", "machPortCount", "__suddenTerm", "__arch",
		//	"memRPrvt", "msgSent", "ppid", "threadCount", "memAnon", "diskBytesWritten",
		//	"pgid", "faults", "msgRecv", "__restricted", "pid", "__sandbox"}

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
		return _out, cancelFunc, err
	}
	ctx, cancelFunc := context.WithCancel(context.TODO())
	selector = "start"
	args = libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return _out, cancelFunc, err
	}

	i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		_out <- m.Obj
	})

	go func() {
		i.registerCallback("_Golang-iDevice_Over", func(_ libimobiledevice.DTXMessageResult) {
			args = libimobiledevice.NewAuxBuffer()
			i.client.Invoke("stop", args, id, true)
			cancelFunc()
		})

		<-ctx.Done()
		// time.Sleep(time.Second)
		close(_out)
		return
	}()
	return _out, cancelFunc, err
}

func (i *instruments) SystemNetWorkServer() (out <-chan map[string]interface{}, cancel context.CancelFunc, err error) {
	var id uint32
	ctx, cancelFunc := context.WithCancel(context.TODO())
	netWorkData := make(chan map[string]interface{})
	if id, err = i.requestChannel("com.apple.instruments.server.services.networking"); err != nil {
		return netWorkData, cancelFunc, err
	}
	selector := "startMonitoring"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return netWorkData, cancelFunc, err
	}
	i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		receData, ok := m.Obj.([]interface{})
		if ok && len(receData) == 2 {
			sendAndReceiveData, ok := receData[1].([]interface{})
			if ok {
				data := make(map[string]interface{})
				data["rx.packets"] = sendAndReceiveData[0]
				data["rx.bytes"] = sendAndReceiveData[1]
				data["tx.packets"] = sendAndReceiveData[2]
				data["tx.bytes"] = sendAndReceiveData[3]
				netWorkData <- data
			}
		}
	})
	go func() {
		i.registerCallback("_Golang-iDevice_Over", func(_ libimobiledevice.DTXMessageResult) {
			cancelFunc()
		})

		<-ctx.Done()
		// time.Sleep(time.Second)
		close(netWorkData)
		return
	}()
	return netWorkData, cancelFunc, err
}

func (i *instruments) OpenglServer() (out <-chan interface{}, cancel context.CancelFunc, err error) {
	var id uint32
	ctx, cancelFunc := context.WithCancel(context.TODO())
	_out := make(chan interface{})
	if id, err = i.requestChannel("com.apple.instruments.server.services.graphics.opengl"); err != nil {
		return _out, cancelFunc, err
	}

	selector := "availableStatistics"
	args := libimobiledevice.NewAuxBuffer()

	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return _out, cancelFunc, err
	}

	selector = "setSamplingRate:"
	if err = args.AppendObject(10); err != nil {
		return _out, cancelFunc, err
	}
	if _, err = i.client.Invoke(selector, args, id, true); err != nil {
		return _out, cancelFunc, err
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

		<-ctx.Done()
		// time.Sleep(time.Second)
		close(_out)
		return
	}()
	return _out, cancelFunc, err
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
