package main

import (
	"fmt"
	"strconv"

	nascrd "github.com/nasim-samimi/dra-rt-driver/api/example.com/resource/rt/nas/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

func (rt *rtdriver) containerCgroups(podCgroup map[string]nascrd.PodCgroup, allocated []nascrd.AllocatedCpu, podClaimName string, pod *corev1.Pod) error {

	runtime := make(nascrd.MappedCgroup)
	period := make(nascrd.MappedCgroup)

	// containerCgroup := make(map[string]nascrd.ContainerCgroup)
	for _, allocatedCpu := range allocated {
		ID := strconv.Itoa(allocatedCpu.ID)
		runtime[ID] = allocatedCpu.Runtime
		period[ID] = allocatedCpu.Period
	}
	cgroup := nascrd.ClaimCgroup{
		ContainerRuntime: runtime,
		ContainerPeriod:  period,
	}
	for _, c := range pod.Spec.Containers {
		for _, n := range c.Resources.Claims {
			if n.Name == podClaimName {
				if _, exists := podCgroup[string(pod.UID)].Containers[c.Name][podClaimName]; exists {
					fmt.Println("Container already exists:", podCgroup[string(pod.UID)].Containers[c.Name][podClaimName])
					break
				}
				if len(podCgroup[string(pod.UID)].Containers[c.Name]) != 0 {
					podCgroup[string(pod.UID)].Containers[c.Name][podClaimName] = cgroup
					break
				}
				podCgroup[string(pod.UID)].Containers[c.Name] = make(map[string]nascrd.ClaimCgroup)
				podCgroup[string(pod.UID)].Containers[c.Name][podClaimName] = cgroup
				break
			}
		}
	}

	return nil
}

// func (rt *rtdriver) podCgroups(containerCgroups map[string]nascrd.ContainerCgroup, crd *nascrd.NodeAllocationState, pod *corev1.Pod) nascrd.PodCgroup {
// 	// cgroupUID:=cgroupUIDGenerator()
// 	if _, exists := crd.Spec.AllocatedPodCgroups[string(pod.UID)]; exists {
// 		fmt.Println("Pod already exists")
// 		fmt.Println("Pod already exists:", crd.Spec.AllocatedPodCgroups[string(pod.UID)])
// 		return crd.Spec.AllocatedPodCgroups[string(pod.UID)]

// 	}
// 	fmt.Println("in pod cgroups function:", containerCgroups)
// 	if len(containerCgroups) == 0 {
// 		return nascrd.PodCgroup{}
// 	}
// 	return nascrd.PodCgroup{
// 		Containers: containerCgroups,
// 		PodName:    pod.Name,
// 	}
// 	// return nil
// }

//can we have a separate struct for cgroups to keep cgroup data?

// func cgroupUIDGenerator() string {
// 	return uuid.NewString()
// }
