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
package perfEntity

import (
	"encoding/json"
	"fmt"
	"strings"
)

type PerfMonData struct {
	CPUInfo        *CPUInfo        `json:"CPU,omitempty"`
	MEMInfo        *MEMInfo        `json:"MEM,omitempty"`
	GPUInfo        *GPUInfo        `json:"GPU,omitempty"`
	FPSInfo        *FPSInfo        `json:"FPS,omitempty"`
	NetWorkingInfo *NetWorkingInfo `json:"NetWorking,omitempty"`
}

func (PerfMonData PerfMonData) ToString() string {
	var s strings.Builder
	if PerfMonData.CPUInfo != nil {
		s.WriteString(fmt.Sprintf("CPU:\n %s", PerfMonData.CPUInfo.ToString()))
	}
	if PerfMonData.MEMInfo != nil {
		s.WriteString(fmt.Sprintf("MEM:\n %s", PerfMonData.MEMInfo.ToString()))
	}
	if PerfMonData.FPSInfo != nil {
		s.WriteString(fmt.Sprintf("FPS:\n %s", PerfMonData.FPSInfo.ToString()))
	}
	if PerfMonData.GPUInfo != nil {
		s.WriteString(fmt.Sprintf("GPU:\n %s", PerfMonData.GPUInfo.ToString()))
	}
	if PerfMonData.NetWorkingInfo != nil {
		s.WriteString(fmt.Sprintf("Networking:\n %s", PerfMonData.NetWorkingInfo.ToString()))
	}
	return s.String()
}

func (PerfMonData PerfMonData) ToJson() string {
	result, _ := json.Marshal(PerfMonData)
	return string(result)
}

func (PerfMonData PerfMonData) ToFormat() string {
	result, _ := json.MarshalIndent(PerfMonData, "", "\t")
	return string(result)
}
