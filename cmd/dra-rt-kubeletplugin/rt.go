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
	nascrd "github.com/nasim-samimi/dra-rt-driver/api/example.com/resource/rt/nas/v1alpha1"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/topology"
	"k8s.io/utils/cpuset"
)

type AllocatableRtCpus map[string]*AllocatableCpusetInfo
type PreparedClaims map[string]*PreparedRtCpu

// //	type GpuInfo struct {
// //		uuid  string
// //		model string
// //	}
type RtCpuInfo struct {
	id   int
	util int
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

func (s *DeviceState) SetDefaultCPUSet() {

	for _, cpus := range s.allocatable {
		cpus.RtCpuInfo.util = 0
	}
}

type realTimePolicy struct {
	topology *topology.CPUTopology
	// allocable utilization
	allocableRtUtil int
	// number of reserved cpus
	numReservedCpus int
	// unassignable cpus
	reservedCpus cpuset.CPUSet
}

// func (p *realTimePolicy) worstFit(cpuToUtil map[int]float64, reqUtil float64, reqCpus int64) []int {
// 	type scoredCpu struct {
// 		cpu   int
// 		score float64
// 	}

// 	var scoredCpus []scoredCpu
// 	for cpu, util := range cpuToUtil {
// 		score := p.allocableRtUtil - util - reqUtil
// 		if score > 0 {
// 			scoredCpus = append(scoredCpus, scoredCpu{
// 				cpu:   cpu,
// 				score: score,
// 			})
// 		}
// 	}

// 	if int64(len(scoredCpus)) < reqCpus {
// 		return nil
// 	}

// 	sort.SliceStable(scoredCpus, func(i, j int) bool {
// 		if scoredCpus[i].score > scoredCpus[j].score {
// 			return true
// 		}
// 		return false
// 	})

// 	var fittingCpus []int
// 	for i := int64(0); i < reqCpus; i++ {
// 		fittingCpus = append(fittingCpus, scoredCpus[i].cpu)
// 	}

// 	return fittingCpus
// }

// func (p *realTimePolicy) bestFit(cpuToUtil map[int]float64, reqUtil float64, reqCpus int64) []int {
// 	type scoredCpu struct {
// 		cpu   int
// 		score float64
// 	}

// 	var scoredCpus []scoredCpu
// 	for cpu, util := range cpuToUtil {
// 		score := p.allocableRtUtil - util - reqUtil
// 		if score > 0 {
// 			scoredCpus = append(scoredCpus, scoredCpu{
// 				cpu:   cpu,
// 				score: score,
// 			})
// 		}
// 	}

// 	if int64(len(scoredCpus)) < reqCpus {
// 		return nil
// 	}

// 	sort.SliceStable(scoredCpus, func(i, j int) bool {
// 		if scoredCpus[i].score < scoredCpus[j].score {
// 			return true
// 		}
// 		return false
// 	})

// 	var fittingCpus []int
// 	for i := int64(0); i < reqCpus; i++ {
// 		fittingCpus = append(fittingCpus, scoredCpus[i].cpu)
// 	}

// 	return fittingCpus
// }
