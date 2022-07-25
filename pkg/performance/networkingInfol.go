package entity

import (
	"encoding/json"
	"fmt"
	"strings"
)

type NetWorkingInfo struct {
	RxBytes   int `json:"rx.bytes,omitempty"`
	RxPackets int `json:"rx.packets,omitempty"`
	TxBytes   int `json:"tx.bytes,omitempty"`
	TxPackets int `json:"tx.packets,omitempty"`
	TimeStamp int `json:"time,omitempty"`
}

func (gpuInfo NetWorkingInfo) ToString() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("rx.bytes:%d rx.packets:%d tx.bytes:%d tx.packets:%d time:%d\n", gpuInfo.RxBytes, gpuInfo.RxPackets, gpuInfo.TxBytes, gpuInfo.TxPackets, gpuInfo.TimeStamp))
	return s.String()
}

func (gpuInfo NetWorkingInfo) ToJson() string {
	result, _ := json.Marshal(gpuInfo)
	return string(result)
}

func (gpuInfo NetWorkingInfo) ToFormat() string {
	result, _ := json.MarshalIndent(gpuInfo, "", "\t")
	return string(result)
}
