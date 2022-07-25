/*
 *  Copyright (C) [SonicCloudOrg] Sonic Project
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */
package perfmorance

import (
	"encoding/json"
	"fmt"
	"strings"
)

type PerfMonData struct {
	CPUInfo        *CPUInfo        `json:"cpuInfo,omitempty"`
	MEMInfo        *MEMInfo        `json:"menInfo,omitempty"`
	GPUInfo        *GPUInfo        `json:"gpuInfo,omitempty"`
	FPSInfo        *FPSInfo        `json:"fpsInfo,omitempty"`
	NetWorkingInfo *NetWorkingInfo `json:"netWorking,omitempty"`
}

func (perfMonData PerfMonData) ToString() string {
	var s strings.Builder
	if perfMonData.CPUInfo != nil {
		s.WriteString(fmt.Sprintf("CPU:\n %s", perfMonData.CPUInfo.ToString()))
	}
	if perfMonData.MEMInfo != nil {
		s.WriteString(fmt.Sprintf("MEM:\n %s", perfMonData.MEMInfo.ToString()))
	}
	if perfMonData.FPSInfo != nil {
		s.WriteString(fmt.Sprintf("FPS:\n %s", perfMonData.FPSInfo.ToString()))
	}
	if perfMonData.GPUInfo != nil {
		s.WriteString(fmt.Sprintf("GPU:\n %s", perfMonData.GPUInfo.ToString()))
	}
	if perfMonData.NetWorkingInfo != nil {
		s.WriteString(fmt.Sprintf("Networking:\n %s", perfMonData.NetWorkingInfo.ToString()))
	}
	return s.String()
}

func (perfMonData PerfMonData) ToJson() string {
	result, _ := json.Marshal(perfMonData)
	return string(result)
}

func (perfMonData PerfMonData) ToFormat() string {
	result, _ := json.MarshalIndent(perfMonData, "", "\t")
	return string(result)
}

func (perfMonData PerfMonData) IsDataNull() bool {
	return perfMonData.FPSInfo == nil && perfMonData.GPUInfo == nil &&
		perfMonData.MEMInfo == nil && perfMonData.CPUInfo == nil &&
		perfMonData.NetWorkingInfo == nil
}
