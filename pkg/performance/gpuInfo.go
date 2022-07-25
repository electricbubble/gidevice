package entity

import (
	"encoding/json"
	"fmt"
	"strings"
)

type GPUInfo struct {
	// 设备利用率
	DeviceUtilization int `json:"Device Utilization,omitempty"`
	// 渲染器利用率
	RendererUtilization int `json:"Renderer Utilization,omitempty"`
	// 处理顶点的GPU时间占比
	TilerUtilization int `json:"Tiler Utilization,omitempty"`
	TimeStamp        int `json:"time,omitempty"`
}

func (gpuInfo GPUInfo) ToString() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("Device Utilization:%d Renderer Utilization:%d Tiler Utilization:%d time:%d\n", gpuInfo.DeviceUtilization, gpuInfo.RendererUtilization, gpuInfo.TilerUtilization, gpuInfo.TimeStamp))
	return s.String()
}

func (gpuInfo GPUInfo) ToJson() string {
	result, _ := json.Marshal(gpuInfo)
	return string(result)
}

func (gpuInfo GPUInfo) ToFormat() string {
	result, _ := json.MarshalIndent(gpuInfo, "", "\t")
	return string(result)
}
