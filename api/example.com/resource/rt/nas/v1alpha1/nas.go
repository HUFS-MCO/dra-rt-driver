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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AllocatableGpu represents an allocatable GPU on a node.
type AllocatableCpu struct {
	ID   int `json:"id"`
	Util int `json:"util"`
	// ProductName string `json:"productName"` // let's assume that the UUID is enough for now
}

// AllocatableDevice represents an allocatable device on a node.
type AllocatableCpuset struct {
	RtCpu *AllocatableCpu `json:"rtcpu,omitempty"`
}

// Type returns the type of AllocatableDevice this represents.
func (d AllocatableCpuset) Type() string {
	if d.RtCpu != nil {
		return RtCpuType
	}
	return UnknownDeviceType
}

// AllocatedGpu represents an allocated GPU.
type AllocatedCpu struct {
	ID      int `json:"id,omitempty"`
	Runtime int `json:"runtime,omitempty"`
	Period  int `json:"period,omitempty"`
}

// AllocatedCpuset represents a set of allocated CPUs.
type AllocatedRtCpu struct {
	Cpuset    []AllocatedCpu `json:"cpuset"`
	CgroupUID string         `json:"cgroupUID,omitempty"`
}

// AllocatedRtCpu represents a set of allocated CPUs.
type AllocatedCpuset struct {
	RtCpu *AllocatedRtCpu `json:"rtcpu,omitempty"`
}

// Type returns the type of AllocatedDevices this represents.
func (r AllocatedCpuset) Type() string {
	if r.RtCpu != nil {
		return RtCpuType
	}
	return UnknownDeviceType
}

// PreparedGpu represents a prepared GPU on a node.
type PreparedCpu struct {
	ID   int `json:"id"`
	Util int `json:"util"`
}

// PreparedGpus represents a set of prepared GPUs on a node.
type PreparedRtCpu struct {
	Cpuset []PreparedCpu `json:"cpuset"`
}

// PreparedDevices represents a set of prepared devices on a node.
type PreparedCpuset struct {
	RtCpu *PreparedRtCpu `json:"rtcpu,omitempty"`
}

// AllocatedUtil represents an allocated utilisation to a CPU.
type AllocatedUtil struct {
	Util int `json:"util"`
}

// MappedUtil represents a mapping of utilisation to CPUs.
type MappedUtil map[string]AllocatedUtil

// AllocatedUtilset represents a set of allocated utilisations to CPUs.
type AllocatedUtilset struct {
	Cpus MappedUtil `json:"cpus,omitempty"`
}

type ClaimCgroup struct {
	ContainerRuntime int    `json:"containerRuntime,omitempty"`
	ContainerPeriod  int    `json:"containerPeriod,omitempty"`
	ContainerCpuset  string `json:"containerCpuset,omitempty"`
}
type ContainerCgroup map[string]ClaimCgroup // key is the container Name

type PodCgroup struct {
	PodName     string          `json:"podName,omitempty"`
	Containers  ContainerCgroup `json:"containers,omitempty"` // key is the container Name
	PodRuntimes []int           `json:"podRuntimes,omitempty"`
}

const (
	AllocatedPodCgroupStatus   = "Allocated"
	UnallocatedPodCgroupStatus = "Unallocated"
)

// Type returns the type of PreparedDevices this represents.
func (d PreparedCpuset) Type() string {
	if d.RtCpu != nil {
		return RtCpuType
	}
	return UnknownDeviceType
}

// NodeAllocationStateSpec is the spec for the NodeAllocationState CRD.
type NodeAllocationStateSpec struct {
	AllocatableCpuset   []AllocatableCpuset        `json:"allocatableCpuset,omitempty"`
	AllocatedClaims     map[string]AllocatedCpuset `json:"allocatedClaims,omitempty"`
	PreparedClaims      map[string]PreparedCpuset  `json:"preparedClaims,omitempty"`
	AllocatedUtilToCpu  AllocatedUtilset           `json:"allocatedUtilToCpu,omitempty"`
	AllocatedPodCgroups map[string]PodCgroup       `json:"allocatedPodCgroups,omitempty"` // key is the cgroup UID
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:resource:singular=nas

// NodeAllocationState holds the state required for allocation on a node.
type NodeAllocationState struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeAllocationStateSpec `json:"spec,omitempty"`
	Status string                  `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeAllocationStateList represents the "plural" of a NodeAllocationState CRD object.
type NodeAllocationStateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []NodeAllocationState `json:"items"`
}
