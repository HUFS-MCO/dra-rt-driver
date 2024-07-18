// /*
//  * Copyright 2023 The Kubernetes Authors.
//  *
//  * Licensed under the Apache License, Version 2.0 (the "License");
//  * you may not use this file except in compliance with the License.
//  * You may obtain a copy of the License at
//  *
//  *     http://www.apache.org/licenses/LICENSE-2.0
//  *
//  * Unless required by applicable law or agreed to in writing, software
//  * distributed under the License is distributed on an "AS IS" BASIS,
//  * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  * See the License for the specific language governing permissions and
//  * limitations under the License.
//  */

package main

import (
	"fmt"
	"sync"

	nascrd "github.com/nasim-samimi/dra-rt-driver/api/example.com/resource/rt/nas/v1alpha1"
)

type AllocatableRtCpus map[string]*AllocatableCpusetInfo
type PreparedClaims map[string]*PreparedRtCpu

// //	type GpuInfo struct {
// //		uuid  string
// //		model string
// //	}
type RtCpuInfo struct {
	id   int
	util float64
}

// //	type PreparedGpus struct {
// //		Devices []*GpuInfo
// //	}
type PreparedCpuset struct {
	Cpuset []*RtCpuInfo
}

type PreparedRtCpu struct {
	RtCpu *PreparedCpuset
}

func (d PreparedRtCpu) Type() string {
	if d.RtCpu != nil {
		return nascrd.RtCpuType
	}
	return nascrd.UnknownDeviceType
}

type AllocatableCpusetInfo struct {
	*RtCpuInfo
}

type RtState struct {
	sync.Mutex
	allocatable     AllocatableRtCpus
	prepared        PreparedClaims
	cgroup          *CgroupManager
	containerToUtil map[string]float64
	cpuToUtil       map[int]float64
}

func NewRtState(config *Config) (*RtState, error) {
	allocatable, err := enumerateAllPossibleDevices()
	if err != nil {
		return nil, fmt.Errorf("error enumerating all possible devices: %v", err)
	}

	// 	// here we must create the cgroup file? mkdir?

	state := &RtState{
		allocatable:     allocatable,
		prepared:        make(PreparedClaims),
		containerToUtil: make(map[string]float64),
	}
	// 	// here we don't have the cpuset functions we need to use s.th else
	// 	state.cpuToUtil = make(map[int]float64, s.GetDefaultCPUSet().Size())
	// 	for _, cpu := range s.GetDefaultCPUSet().UnsortedList() {
	// 		state.cpuToUtil[cpu] = 0
	// 	}

	// 	err = state.syncPreparedDevicesFromCRDSpec(&config.nascr.Spec)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("unable to sync prepared devices from CRD: %v", err)
	// 	}

	return state, nil
}
