package perfEntity

import (
	"encoding/json"
	"fmt"
	"strings"
)

type FPSInfo struct {
	FPS       		int `json:"fps,omitempty"`
	TimeStamp       int64 `json:"time,omitempty"`
}

func (fpsInfo FPSInfo) ToString() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("FPS:%d time:%d\n", fpsInfo.FPS, fpsInfo.TimeStamp))
	return s.String()
}

func (fpsInfo FPSInfo) ToJson() string {
	result, _ := json.Marshal(fpsInfo)
	return string(result)
}

func (fpsInfo FPSInfo) ToFormat() string {
	result, _ := json.MarshalIndent(fpsInfo, "", "\t")
	return string(result)
}
