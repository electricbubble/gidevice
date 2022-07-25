package perfmorance

import (
	"encoding/json"
	"fmt"
	"strings"
)

type NetWorkingInfo struct {
	RxBytes   int64 `json:"rxBytes,omitempty"`
	RxPackets int64 `json:"rxPackets,omitempty"`
	TxBytes   int64 `json:"txBytes,omitempty"`
	TxPackets int64 `json:"txPackets,omitempty"`
	TimeStamp int64 `json:"timeStamp,omitempty"`
}

func (netWorkInfo NetWorkingInfo) ToString() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("rx.bytes:%d rx.packets:%d tx.bytes:%d tx.packets:%d time:%d\n", netWorkInfo.RxBytes, netWorkInfo.RxPackets, netWorkInfo.TxBytes, netWorkInfo.TxPackets, netWorkInfo.TimeStamp))
	return s.String()
}

func (netWorkInfo NetWorkingInfo) ToJson() string {
	result, _ := json.Marshal(netWorkInfo)
	return string(result)
}

func (netWorkInfo NetWorkingInfo) ToFormat() string {
	result, _ := json.MarshalIndent(netWorkInfo, "", "\t")
	return string(result)
}
