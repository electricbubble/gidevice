package perfEntity

import (
	"encoding/json"
	"fmt"
	"strings"
)

type CPUInfo struct {
	// 上下文切换数
	AttrCtxSwitch 			int `json:"attrCtxSwitch,omitempty"`
	// 唤醒数
	AttrIntWakeups		 	int `json:"attrIntWakeups,omitempty"`
	// CPU总数
	CPUCount 				int `json:"cpuCount,omitempty"`

	Pid 					string `json:"pid,omitempty"`

	TimeStamp        		int64 `json:"time,omitempty"`
	// 单个进程的CPU使用率
	CPUUsage 				float64 `json:"cpuUsage,omitempty"`
	// 系统总体CPU占用
	SysCpuUsage 			float64 `json:"sysCpuUsage,omitempty"`
}

func (cpuInfo CPUInfo) ToString() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("CtxSwitch:%d IntWakeups:%d CPUCount:%d CPUUsage:%f SysCPUUsage:%f PID:%s time:%d\n",
		cpuInfo.AttrCtxSwitch, cpuInfo.AttrIntWakeups,
		cpuInfo.CPUCount, cpuInfo.CPUUsage,
		cpuInfo.SysCpuUsage, cpuInfo.Pid,
		cpuInfo.TimeStamp))
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
