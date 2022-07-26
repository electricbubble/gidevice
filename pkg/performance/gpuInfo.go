package perfmorance

import (
	"encoding/json"
	"fmt"
	"strings"
)

type GPUInfo struct {
	// 设备利用率
	DeviceUtilization int64 `json:"deviceUtilization"`
	// 渲染器利用率
	RendererUtilization int64 `json:"rendererUtilization"`
	// 处理顶点的GPU时间占比
	TilerUtilization int64 `json:"tilerUtilization"`
	TimeStamp        int64 `json:"timeStamp"`
}

func (gpuInfo GPUInfo) ToString() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("deviceUtilization:%d rendererUtilization:%d tilerUtilization:%d timeStamp:%d\n", gpuInfo.DeviceUtilization, gpuInfo.RendererUtilization, gpuInfo.TilerUtilization, gpuInfo.TimeStamp))
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
