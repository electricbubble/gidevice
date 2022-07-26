package perfmorance

import (
	"encoding/json"
	"fmt"
	"strings"
)

type NetWorkingInfo struct {
	RxBytes   int64 `json:"rxBytes"`
	RxPackets int64 `json:"rxPackets"`
	TxBytes   int64 `json:"txBytes"`
	TxPackets int64 `json:"txPackets"`
	TimeStamp int64 `json:"timeStamp"`
}

func (netWorkInfo NetWorkingInfo) ToString() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("rxBytes:%d rxPackets:%d txBytes:%d txPackets:%d timeStamp:%d\n", netWorkInfo.RxBytes, netWorkInfo.RxPackets, netWorkInfo.TxBytes, netWorkInfo.TxPackets, netWorkInfo.TimeStamp))
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
