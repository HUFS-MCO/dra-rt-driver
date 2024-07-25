/*
 * Copyright 2023 The Kubernetes Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"sort"

	nascrd "github.com/nasim-samimi/dra-rt-driver/api/example.com/resource/rt/nas/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1alpha2"
	"k8s.io/dynamic-resource-allocation/controller"

	rtcrd "github.com/nasim-samimi/dra-rt-driver/api/example.com/resource/rt/v1alpha1"
)

type rtdriver struct {
	PendingAllocatedClaims *PerNodeAllocatedClaims
}

func NewRtDriver() *rtdriver {
	return &rtdriver{
		PendingAllocatedClaims: NewPerNodeAllocatedClaims(),
	}
}

func (g *rtdriver) ValidateClaimParameters(claimParams *rtcrd.RtClaimParametersSpec) error {
	if claimParams.Count < 1 {
		return fmt.Errorf("invalid number of GPUs requested: %v", claimParams.Count)
	}
	return nil
}

func (g *rtdriver) Allocate(crd *nascrd.NodeAllocationState, claim *resourcev1.ResourceClaim, claimParams *rtcrd.RtClaimParametersSpec, class *resourcev1.ResourceClass, classParams *rtcrd.DeviceClassParametersSpec, selectedNode string) (OnSuccessCallback, error) {
	claimUID := string(claim.UID)

	if !g.PendingAllocatedClaims.Exists(claimUID, selectedNode) {
		return nil, fmt.Errorf("no allocations generated for claim '%v' on node '%v' yet", claim.UID, selectedNode)
	}

	crd.Spec.AllocatedClaims[claimUID] = g.PendingAllocatedClaims.Get(claimUID, selectedNode)
	onSuccess := func() {
		g.PendingAllocatedClaims.Remove(claimUID)
	}
	crd.Spec.AllocatedUtilToCpu = g.PendingAllocatedClaims.GetUtil(selectedNode)

	return onSuccess, nil
}

func (g *rtdriver) Deallocate(crd *nascrd.NodeAllocationState, claim *resourcev1.ResourceClaim) error {
	g.PendingAllocatedClaims.Remove(string(claim.UID))
	return nil
}

func (rt *rtdriver) UnsuitableNode(crd *nascrd.NodeAllocationState, pod *corev1.Pod, rtcas []*controller.ClaimAllocation, allcas []*controller.ClaimAllocation, potentialNode string) error {
	rt.PendingAllocatedClaims.VisitNode(potentialNode, func(claimUID string, allocation nascrd.AllocatedCpuset, utilisation []nascrd.AllocatedUtilset) {
		if _, exists := crd.Spec.AllocatedClaims[claimUID]; exists {
			rt.PendingAllocatedClaims.Remove(claimUID)
		} else {
			crd.Spec.AllocatedClaims[claimUID] = allocation
			crd.Spec.AllocatedUtilToCpu = utilisation
		}
	})

	allocated, allocatedUtil := rt.allocate(crd, pod, rtcas, allcas, potentialNode)

	for _, ca := range rtcas {
		claimUID := string(ca.Claim.UID)
		claimParams, _ := ca.ClaimParameters.(*rtcrd.RtClaimParametersSpec)

		if claimParams.Count != len(allocated[claimUID]) {
			for _, ca := range allcas {
				ca.UnsuitableNodes = append(ca.UnsuitableNodes, potentialNode)
			}
			return nil
		}

		var devices []nascrd.AllocatedCpu
		var allocatedUtilisations []nascrd.AllocatedUtilset
		for _, cpu := range allocated[claimUID] {
			device := cpu
			devices = append(devices, device)
		}

		allocatedDevices := nascrd.AllocatedCpuset{
			RtCpu: &nascrd.AllocatedRtCpu{
				Cpuset: devices,
			},
		}

		for _, ut := range allocatedUtil {
			allocatedUtilisations = append(allocatedUtilisations, nascrd.AllocatedUtilset{
				RtUtil: ut,
			})
		}
		rt.PendingAllocatedClaims.Set(claimUID, potentialNode, allocatedDevices)
		rt.PendingAllocatedClaims.SetUtil(potentialNode, allocatedUtilisations)
	}

	return nil
}

func (g *rtdriver) allocate(crd *nascrd.NodeAllocationState, pod *corev1.Pod, cpucas []*controller.ClaimAllocation, allcas []*controller.ClaimAllocation, node string) (map[string][]nascrd.AllocatedCpu, map[int]*nascrd.AllocatedUtil) {
	available := make(map[int]*nascrd.AllocatableCpu)
	util := make(map[int]*nascrd.AllocatedUtil)
	currUtil := 0

	for _, device := range crd.Spec.AllocatableCpuset {
		switch device.Type() {
		case nascrd.RtCpuType:
			available[device.RtCpu.ID] = device.RtCpu
		default:
			// skip other devices
		}
	}
	if crd.Spec.AllocatedUtilToCpu == nil {
		for _, device := range crd.Spec.AllocatableCpuset {
			util[device.RtCpu.ID] = &nascrd.AllocatedUtil{
				ID:   device.RtCpu.ID,
				Util: 0,
			}

		}
	} else {
		fmt.Println("crd.Spec.AllocatedUtilToCpu not nil in allocate")
		for _, device := range crd.Spec.AllocatedUtilToCpu {
			util[device.RtUtil.ID] = device.RtUtil
			fmt.Println("util from spec:", device.RtUtil.Util)
		}
	}

	// for _, allocation := range crd.Spec.AllocatedClaims {
	// 	switch allocation.Type() {
	// 	case nascrd.RtCpuType:
	// 		for _, device := range allocation.RtCpu.Cpuset {
	// 			delete(available, device.ID)
	// 		}
	// 	default:
	// 		// skip other devices
	// 	}
	// }

	allocated := make(map[string][]nascrd.AllocatedCpu)
	for _, ca := range cpucas {
		claimUID := string(ca.Claim.UID)
		if _, exists := crd.Spec.AllocatedClaims[claimUID]; exists {
			devices := crd.Spec.AllocatedClaims[claimUID].RtCpu.Cpuset
			for _, device := range devices {
				allocated[claimUID] = append(allocated[claimUID], device)
			}
			continue
		}

		claimParams, _ := ca.ClaimParameters.(*rtcrd.RtClaimParametersSpec)
		var devices []nascrd.AllocatedCpu
		for i := 0; i < claimParams.Count; i++ {
			// for _, device := range available {
			fmt.Println("Available CPUs:", util)
			bestFitCpus := bestFit(util, (claimParams.Runtime * 1000 / claimParams.Period), claimParams.Count)
			fmt.Println("Best fit CPUs:", bestFitCpus)
			claimUtil := (claimParams.Runtime * 1000 / claimParams.Period)
			for id, ut := range util {
				fmt.Println("util:", ut.Util)
				fmt.Println("util:", id)
			}

			if _, exist := util[bestFitCpus[0]]; !exist {
				fmt.Println("AllocatedUtilToCpu is nil (function:allocate)")
			} else {
				currUtil = util[bestFitCpus[0]].Util
			}

			if claimUtil+currUtil <= 1000 {
				d := nascrd.AllocatedCpu{
					ID:      bestFitCpus[0],
					Runtime: claimParams.Runtime,
					Period:  claimParams.Period,
				}
				devices = append(devices, d)
				util[d.ID].Util = util[d.ID].Util + claimUtil
				if util[d.ID].Util >= 1000 {
					delete(available, d.ID)
				}
				break
			}
			// }
		}
		allocated[claimUID] = devices
	}

	var utilisations []nascrd.AllocatedUtilset
	for _, device := range util {
		utilslice := nascrd.AllocatedUtilset{
			RtUtil: device,
		}
		fmt.Println("let's see the utils that are computed:", device)
		utilisations = append(utilisations, utilslice)
	}
	crd.Spec.AllocatedUtilToCpu = utilisations
	for _, ut := range crd.Spec.AllocatedUtilToCpu {
		fmt.Println("let's see the spec utils that are computed:", ut.RtUtil.Util)
	}

	return allocated, util
}

func worstFit(spec map[int]*nascrd.AllocatedUtil, reqUtil int, reqCpus int) []int {
	type scoredCpu struct {
		cpu   int
		score int
	}

	var scoredCpus []scoredCpu
	for _, cpuinfo := range spec {
		score := cpuinfo.Util - reqUtil
		if score > 0 {
			scoredCpus = append(scoredCpus, scoredCpu{
				cpu:   cpuinfo.ID,
				score: score,
			})
		}
	}

	if int(len(scoredCpus)) < reqCpus {
		return nil
	}

	sort.SliceStable(scoredCpus, func(i, j int) bool {
		if scoredCpus[i].score > scoredCpus[j].score {
			return true
		}
		return false
	})

	var fittingCpus []int
	for i := int(0); i < reqCpus; i++ {
		fittingCpus = append(fittingCpus, scoredCpus[i].cpu)
	}

	return fittingCpus
}

func bestFit(spec map[int]*nascrd.AllocatedUtil, reqUtil int, reqCpus int) []int {
	type scoredCpu struct {
		cpu   int
		score int
	}

	var scoredCpus []scoredCpu
	for _, cpuinfo := range spec {
		score := 1000 - cpuinfo.Util - reqUtil
		if score > 0 {
			scoredCpus = append(scoredCpus, scoredCpu{
				cpu:   cpuinfo.ID,
				score: score,
			})
		}
	}

	if int(len(scoredCpus)) < reqCpus {
		return nil
	}

	sort.SliceStable(scoredCpus, func(i, j int) bool {
		if scoredCpus[i].score < scoredCpus[j].score {
			return true
		}
		return false
	})

	var fittingCpus []int
	for i := int(0); i < reqCpus; i++ {
		fittingCpus = append(fittingCpus, scoredCpus[i].cpu)
	}

	return fittingCpus
}
