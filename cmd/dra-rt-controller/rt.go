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
	"strconv"

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
		return fmt.Errorf("invalid number of HCBS requested: %v", claimParams.Count)
	}
	return nil
}

func (g *rtdriver) Allocate(crd *nascrd.NodeAllocationState, claim *resourcev1.ResourceClaim, claimParams *rtcrd.RtClaimParametersSpec, class *resourcev1.ResourceClass, classParams *rtcrd.DeviceClassParametersSpec, selectedNode string) (OnSuccessCallback, error) {
	claimUID := string(claim.UID)

	if !g.PendingAllocatedClaims.Exists(claimUID, selectedNode) {
		return nil, fmt.Errorf("no allocations generated for claim '%v' on node '%v' yet", claim.UID, selectedNode)
	}

	crd.Spec.AllocatedClaims[claimUID] = g.PendingAllocatedClaims.Get(claimUID, selectedNode)
	crd.Spec.AllocatedUtilToCpu = g.PendingAllocatedClaims.GetUtil(selectedNode)
	crd.Spec.AllocatedPodCgroups = g.PendingAllocatedClaims.GetCgroup(selectedNode)
	fmt.Println("Allocate, crd.Spec.AllocatedPodCgroups:", crd.Spec.AllocatedPodCgroups)
	fmt.Println("Allocate, crd.Spec.AllocatedClaims:", crd.Spec.AllocatedClaims)

	onSuccess := func() {
		g.PendingAllocatedClaims.RemoveUtil(claimUID)
		g.PendingAllocatedClaims.RemoveCgroup(claimUID)
		g.PendingAllocatedClaims.Remove(claimUID)
		fmt.Println("what happens in remove cgroups:", g.PendingAllocatedClaims.GetCgroup(selectedNode))
	}

	return onSuccess, nil
}

func (g *rtdriver) Deallocate(crd *nascrd.NodeAllocationState, claim *resourcev1.ResourceClaim) error {
	claimUID := string(claim.UID)
	g.PendingAllocatedClaims.RemoveUtil(claimUID)
	g.PendingAllocatedClaims.RemoveCgroup(claimUID)
	g.PendingAllocatedClaims.Remove(claimUID)
	return nil
}

func (rt *rtdriver) UnsuitableNode(crd *nascrd.NodeAllocationState, pod *corev1.Pod, rtcas []*controller.ClaimAllocation, allcas []*controller.ClaimAllocation, potentialNode string) error {
	rt.PendingAllocatedClaims.VisitNode(potentialNode, func(claimUID string, allocation nascrd.AllocatedCpuset, utilisation nascrd.AllocatedUtilset, cgroups nascrd.PodCgroup) {
		if _, exists := crd.Spec.AllocatedClaims[claimUID]; exists {
			// rt.PendingAllocatedClaims.RemoveUtil(claimUID)
			// rt.PendingAllocatedClaims.RemoveCgroup(claimUID)
			rt.PendingAllocatedClaims.Remove(claimUID)
		} else {
			fmt.Println("what is assigned to crd in unsuitable nodes:", cgroups)
			crd.Spec.AllocatedClaims[claimUID] = allocation
			crd.Spec.AllocatedUtilToCpu = utilisation
			crd.Spec.AllocatedPodCgroups[string(pod.UID)] = cgroups
		}
	})
	cgroupUID := string(pod.UID)

	allocated, allocatedUtil, podCgroup := rt.allocate(crd, pod, rtcas, allcas, potentialNode)

	for _, ca := range rtcas {
		claimUID := string(ca.Claim.UID)
		claimParams, _ := ca.ClaimParameters.(*rtcrd.RtClaimParametersSpec)
		if claimParams.Count != len(allocated[claimUID]) {
			for _, ca := range allcas {
				ca.UnsuitableNodes = append(ca.UnsuitableNodes, potentialNode)
			}
			return nil
		} // it puts everything on only one node

		var devices []nascrd.AllocatedCpu
		for _, cpu := range allocated[claimUID] {
			device := cpu
			devices = append(devices, device)
		}

		allocatedDevices := nascrd.AllocatedCpuset{
			RtCpu: &nascrd.AllocatedRtCpu{
				Cpuset:    devices,
				CgroupUID: cgroupUID,
			},
		}
		fmt.Println("allocatedUtil:", allocatedUtil)
		allocatedUtilisations := nascrd.AllocatedUtilset{
			Cpus: allocatedUtil,
		}
		fmt.Println("allocatedUtilisations:", allocatedUtilisations)

		rt.PendingAllocatedClaims.Set(claimUID, potentialNode, allocatedDevices)
		rt.PendingAllocatedClaims.SetUtil(potentialNode, allocatedUtilisations)
	}
	fmt.Println("UnsuitableNode, podCgroup:", podCgroup)
	if len(podCgroup[cgroupUID].Containers) > 0 {
		rt.PendingAllocatedClaims.SetCgroup(cgroupUID, potentialNode, podCgroup[cgroupUID])
	}
	fmt.Println("what is assigned to pending after setcgroup:", rt.PendingAllocatedClaims.GetCgroup(potentialNode))
	return nil
}

func (rt *rtdriver) allocate(crd *nascrd.NodeAllocationState, pod *corev1.Pod, cpucas []*controller.ClaimAllocation, allcas []*controller.ClaimAllocation, node string) (map[string][]nascrd.AllocatedCpu, map[string]nascrd.AllocatedUtil, map[string]nascrd.PodCgroup) {
	available := make(map[int]*nascrd.AllocatableCpu)
	util := crd.Spec.AllocatedUtilToCpu.Cpus
	// util := make(map[string]nascrd.AllocatedUtil)
	allocated := make(map[string][]nascrd.AllocatedCpu)
	fmt.Println("do we have the cgroups in pending and crd?")
	fmt.Println("pending:", rt.PendingAllocatedClaims.GetCgroup(node))
	fmt.Println("crd:", crd.Spec.AllocatedPodCgroups)
	podCG := make(map[string]nascrd.PodCgroup)
	podCG[string(pod.UID)] = nascrd.PodCgroup{
		Containers: make(map[string]nascrd.ClaimCgroup),
		PodName:    pod.Name,
	}
	// if _, exists := crd.Spec.AllocatedPodCgroups[string(pod.UID)]; exists {
	// 	containerCG = crd.Spec.AllocatedPodCgroups[string(pod.UID)].Containers
	// }

	for _, device := range crd.Spec.AllocatableCpuset {
		switch device.Type() {
		case nascrd.RtCpuType:
			available[device.RtCpu.ID] = device.RtCpu
		default:
			// skip other devices
		}
	}

	for _, ca := range cpucas {
		claimUID := string(ca.Claim.UID)
		if _, exists := crd.Spec.AllocatedClaims[claimUID]; exists {
			devices := crd.Spec.AllocatedClaims[claimUID].RtCpu.Cpuset
			cgroupUID := crd.Spec.AllocatedClaims[claimUID].RtCpu.CgroupUID
			for _, device := range devices {
				allocated[claimUID] = append(allocated[claimUID], device)
			}
			if _, exists := crd.Spec.AllocatedPodCgroups[cgroupUID]; exists {
				podCG[cgroupUID] = crd.Spec.AllocatedPodCgroups[cgroupUID]
			}

			continue
		}

		claimParams, _ := ca.ClaimParameters.(*rtcrd.RtClaimParametersSpec)
		claimUtil := (claimParams.Runtime * 1000 / claimParams.Period)
		var devices []nascrd.AllocatedCpu
		for i := 0; i < claimParams.Count; i++ {
			// for _, device := range available {
			worstFitCpus, err := cpuPartitioning(util, claimUtil, 1, "worstFit") //must get the policy from the user
			if err != nil {
				return nil, nil, nil
			}
			fmt.Println("worstFitCpus:", worstFitCpus)
			worstFitCpusStr, _ := strconv.Atoi(worstFitCpus[0])
			d := nascrd.AllocatedCpu{
				ID:      worstFitCpusStr,
				Runtime: claimParams.Runtime,
				Period:  claimParams.Period,
			}
			util[strconv.Itoa(d.ID)] = nascrd.AllocatedUtil{
				Util: util[strconv.Itoa(d.ID)].Util + claimUtil,
			}
			if util[strconv.Itoa(d.ID)].Util >= 1000 {
				delete(available, d.ID)
			}
			devices = append(devices, d)
		}
		allocated[claimUID] = devices

		CCgroup, _ := rt.containerCgroups(podCG, devices, ca.PodClaimName, pod, claimParams)
		setClaimAnnotations(CCgroup, pod, ca.Claim)
		fmt.Println("allocate, CCgroup:", CCgroup)

	}
	// adding to pod annotations
	setPodAnnotations(podCG, pod) // not working
	fmt.Println("allocate, podCG:", podCG)

	return allocated, util, podCG
}

func cpuPartitioning(spec map[string]nascrd.AllocatedUtil, reqUtil int, reqCpus int, policy string) ([]string, error) {
	type scoredCpu struct {
		cpu   string
		score int
	}
	var scoredCpus []scoredCpu
	for id, cpuinfo := range spec {
		score := 1000 - cpuinfo.Util - reqUtil
		if score > 0 {
			scoredCpus = append(scoredCpus, scoredCpu{
				cpu:   id,
				score: score,
			})
		}
	}

	if int(len(scoredCpus)) < reqCpus {
		return nil, fmt.Errorf("not enough cpus to allocate")
	}
	switch policy {
	case "worstFit":
		sort.SliceStable(scoredCpus, func(i, j int) bool {
			if scoredCpus[i].score > scoredCpus[j].score {
				return true
			}
			return false
		})
	case "bestFit":
		sort.SliceStable(scoredCpus, func(i, j int) bool {
			if scoredCpus[i].score < scoredCpus[j].score {
				return true
			}
			return false
		})
	default:
		sort.SliceStable(scoredCpus, func(i, j int) bool {
			if scoredCpus[i].score > scoredCpus[j].score {
				return true
			}
			return false
		}) //default is worstFit
	}

	var fittingCpus []string
	for i := int(0); i < reqCpus; i++ {
		fittingCpus = append(fittingCpus, scoredCpus[i].cpu)
	}

	return fittingCpus, nil
}
