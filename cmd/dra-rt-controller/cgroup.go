package main

import (
	"fmt"
	"strconv"
	"strings"

	nascrd "github.com/nasim-samimi/dra-rt-driver/api/example.com/resource/rt/nas/v1alpha1"

	rtcrd "github.com/nasim-samimi/dra-rt-driver/api/example.com/resource/rt/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func (rt *rtdriver) containerCgroups(podCgroup map[string]nascrd.PodCgroup, allocated []nascrd.AllocatedCpu, podClaimName string, pod *corev1.Pod, claimParams *rtcrd.RtClaimParametersSpec) error {

	claimRuntime := claimParams.Runtime
	claimPeriod := claimParams.Period

	// containerCgroup := make(map[string]nascrd.ContainerCgroup)
	var builder strings.Builder
	for i, allocatedCpu := range allocated {
		if i > 0 {
			builder.WriteString("-") // TODO: change this later to comma
		}
		builder.WriteString(strconv.Itoa(allocatedCpu.ID))
	}
	claimCpuset := builder.String()

	cgroup := nascrd.ClaimCgroup{
		ContainerRuntime: claimRuntime,
		ContainerPeriod:  claimPeriod,
		ContainerCpuset:  claimCpuset,
	}
	for _, c := range pod.Spec.Containers {
		for _, n := range c.Resources.Claims {
			if n.Name == podClaimName {
				if _, exists := podCgroup[string(pod.UID)].Containers[c.Name]; exists {
					fmt.Println("Container already exists:", podCgroup[string(pod.UID)].Containers[c.Name])
					break
				}
				podCgroup[string(pod.UID)].Containers[c.Name] = cgroup
				break
				/////////////this code is for when we need to have podClaimName in the cgroup struct
				// if _, exists := podCgroup[string(pod.UID)].Containers[c.Name][podClaimName]; exists {
				// 	fmt.Println("Container already exists:", podCgroup[string(pod.UID)].Containers[c.Name][podClaimName])
				// 	break
				// }
				// if len(podCgroup[string(pod.UID)].Containers[c.Name]) != 0 {
				// 	podCgroup[string(pod.UID)].Containers[c.Name][podClaimName] = cgroup
				// 	break
				// }
				// podCgroup[string(pod.UID)].Containers[c.Name] = make(map[string]nascrd.ClaimCgroup)
				// podCgroup[string(pod.UID)].Containers[c.Name][podClaimName] = cgroup
				// break
			}
		}
	}

	return nil
}

func setAnnotations(podCG map[string]nascrd.PodCgroup, pod *corev1.Pod) map[string]string {
	annotations := pod.GetAnnotations()
	p := pod.ObjectMeta.Annotations
	fmt.Println("Pod metadate annotations:", p)
	if pod.GetAnnotations() == nil {
		annotations = make(map[string]string)
	}
	fmt.Println("old Pod annotations:", annotations)
	if _, exists := podCG[string(pod.UID)]; exists {
		annotations["RTDevice"] = "exists"
		for c, cg := range podCG[string(pod.UID)].Containers {
			runtime := strconv.Itoa(cg.ContainerRuntime)
			period := strconv.Itoa(cg.ContainerPeriod)
			cpuset := cg.ContainerCpuset
			annotations[c+"-runtime"] = runtime
			annotations[c+"-period"] = period
			annotations[c+"-CPUs"] = cpuset
		}
	}

	fmt.Println("Annotations:", annotations)
	pod.SetAnnotations(annotations)
	pod.ObjectMeta.Annotations = annotations
	fmt.Println("Pod get annotations:", pod.GetAnnotations())
	fmt.Println("Pod annotations:", pod.Annotations)
	fmt.Println("Pod get metadate annotations:", pod.ObjectMeta.Annotations)
	return annotations
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
