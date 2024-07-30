package main

import (
	"github.com/google/uuid"

	nascrd "github.com/nasim-samimi/dra-rt-driver/api/example.com/resource/rt/nas/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

func (rt *rtdriver) containerCgroups(claimCgroups map[string]map[string]cgroups, allocated []nascrd.AllocatedCpu, podClaimName string, pod *corev1.Pod) error {

	runtime := make(map[int]int)
	period := make(map[int]int)
	for _, allocatedCpu := range allocated {
		runtime[allocatedCpu.ID] = allocatedCpu.Runtime
		period[allocatedCpu.ID] = allocatedCpu.Period
	}
	cgroup := cgroups{
		runtime: runtime,
		period:  period,
	}
	for _, c := range pod.Spec.Containers {
		for _, n := range c.Resources.Claims {
			if n.Name == podClaimName {
				if _, exists := claimCgroups[c.Name][podClaimName]; exists {
					break
				}
				claimCgroups[c.Name][podClaimName] = cgroup
				break
			}
		}
	}

	return nil
}

// func (rt *rtdriver) podCgroups(claimCgroups claimCgroup, crd *nascrd.NodeAllocationState, pod *corev1.Pod) error {
// 	podCG:=nascrd.AllocatedPodCgroup{}
// 	containersCG:=[]nascrd.ContainerCgroup{}
// 	for c,claims := range claimCgroups{
// 		for _, cgroup := range claims{
// 			containerCG=append(containerCG,nascrd.ContainerCgroup{
// 				ContainerName: c,
// 				ContainerRuntime: ,
// 			)

// 	return nil
// } //can we have a separate struct for cgroups to keep cgroup data?

func cgroupUIDGenerator() string {
	return uuid.NewString()
}

type claimCgroup map[string]map[string]cgroups

type cgroups struct {
	runtime map[int]int
	period  map[int]int
}
