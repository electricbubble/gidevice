package perfmorance

import (
	"encoding/json"
	"fmt"
	"strings"
)

type FPSInfo struct {
	FPS       int   `json:"fps"`
	TimeStamp int64 `json:"timeStamp"`
}

func (fpsInfo FPSInfo) ToString() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("fps:%d timeStamp:%d\n", fpsInfo.FPS, fpsInfo.TimeStamp))
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
