package perfmorance

import (
	"encoding/json"
	"fmt"
	"strings"
)

type CPUInfo struct {
	// 上下文切换数
	AttrCtxSwitch 	int64 	`json:"attrCtxSwitch,omitempty"`
	// 唤醒数
	AttrIntWakeups 	int64 	`json:"attrIntWakeups,omitempty"`
	// CPU总数
	CPUCount 		int 	`json:"cpuCount"`

	Pid 			string 	`json:"pid"`

	TimeStamp 		int64 	`json:"timeStamp"`
	// 单个进程的CPU使用率
	CPUUsage 		float64 `json:"cpuUsage,omitempty"`
	// 系统总体CPU占用
	SysCpuUsage 	float64 `json:"sysCpuUsage,omitempty"`
}

func (cpuInfo CPUInfo) ToString() string {
	var s strings.Builder
	if cpuInfo.Pid!="" {
		s.WriteString(fmt.Sprintf("ctxSwitch:%d intWakeups:%d cpuCount:%d cpuUsage:%f sysCPUUsage:%f pid:%s timeStamp:%d\n",
			cpuInfo.AttrCtxSwitch, cpuInfo.AttrIntWakeups,
			cpuInfo.CPUCount, cpuInfo.CPUUsage,
			cpuInfo.SysCpuUsage, cpuInfo.Pid,
			cpuInfo.TimeStamp))
	}else {
		s.WriteString(fmt.Sprintf("cpuCount:%d  sysCPUUsage:%f pid:%s timeStamp:%d\n",
			cpuInfo.CPUCount, cpuInfo.SysCpuUsage,
			cpuInfo.Pid,
			cpuInfo.TimeStamp))
	}

	return s.String()
}

func (cpuInfo CPUInfo) ToJson() string {
	result, _ := json.Marshal(cpuInfo)
	return string(result)
}

func (cpuInfo CPUInfo) ToFormat() string {
	result, _ := json.MarshalIndent(cpuInfo, "", "\t")
	return string(result)
}
