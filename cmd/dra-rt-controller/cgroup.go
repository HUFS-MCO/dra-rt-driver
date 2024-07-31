package main

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"

	nascrd "github.com/nasim-samimi/dra-rt-driver/api/example.com/resource/rt/nas/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

func (rt *rtdriver) containerCgroups(containerCgroup map[string]nascrd.ContainerCgroup, allocated []nascrd.AllocatedCpu, podClaimName string, pod *corev1.Pod) error {

	runtime := make(nascrd.MappedCgroup)
	period := make(nascrd.MappedCgroup)
	claimCgroup := nascrd.ContainerCgroup{}
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
				if _, exists := containerCgroup[c.Name][podClaimName]; exists {
					break
				}
				if _, exists := containerCgroup[c.Name]; exists {
					containerCgroup[c.Name][podClaimName] = cgroup
					break
				}
				claimCgroup[podClaimName] = cgroup
				containerCgroup[c.Name] = claimCgroup
				break
			}
		}
	}

	return nil
}

func (rt *rtdriver) podCgroups(containerCgroups map[string]nascrd.ContainerCgroup, crd *nascrd.NodeAllocationState, pod *corev1.Pod) nascrd.PodCgroup {
	// cgroupUID:=cgroupUIDGenerator()
	if _, exists := crd.Spec.AllocatedPodCgroups[string(pod.UID)]; exists {
		fmt.Println("Pod already exists")
		return crd.Spec.AllocatedPodCgroups[string(pod.UID)]

	}
	fmt.Println("in pod cgroups function:", containerCgroups)
	return nascrd.PodCgroup{
		Containers: containerCgroups,
		PodName:    pod.Name,
		PodUID:     string(pod.UID),
	}
	// return nil
}

//can we have a separate struct for cgroups to keep cgroup data?

func cgroupUIDGenerator() string {
	return uuid.NewString()
}
