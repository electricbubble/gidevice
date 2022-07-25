package entity

import (
	"encoding/json"
	"fmt"
	"strings"
)

type MEMInfo struct {
	// 虚拟内存
	Anon int `json:"anon,omitempty"`
	// 物理内存
	PhysMemory int `json:"physMemory,omitempty"`
	// 总内存
	Rss int `json:"rss,omitempty"`
	// 虚拟内存
	Vss       int `json:"vss,omitempty"`
	TimeStamp int `json:"time,omitempty"`
}

func (memInfo MEMInfo) ToString() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("anon:%d physMemory:%d rss:%d vss:%d time:%d\n", memInfo.Anon, memInfo.PhysMemory, memInfo.Rss, memInfo.Vss, memInfo.TimeStamp))
	return s.String()
}

func (memInfo MEMInfo) ToJson() string {
	result, _ := json.Marshal(memInfo)
	return string(result)
}

func (memInfo MEMInfo) ToFormat() string {
	result, _ := json.MarshalIndent(memInfo, "", "\t")
	return string(result)
}
