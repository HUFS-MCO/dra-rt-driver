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
	"fmt"

	nascrd "github.com/HUFS-MCO/dra-rt-driver/api/example.com/resource/rt/nas/v1alpha1"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GroupName = "rt.resource.example.com" // nas. 제거
	Version   = "v1alpha1"

	DriverName        = "rt.resource.example.com" // 추가
	RtCpuType         = "rtcpu"
	UnknownDeviceType = "unknown"

	NodeAllocationStateStatusReady    = "Ready"
	NodeAllocationStateStatusNotReady = "NotReady"
	RtClaimParametersKind             = "RtClaimParameters"
)

type DeviceClassConfig struct {
	Name      string
	Namespace string
	Driver    string
}

func NewDeviceClass(config *DeviceClassConfig) *resourcev1.DeviceClass {
	return &resourcev1.DeviceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
		},
		Spec: resourcev1.DeviceClassSpec{
			Selectors: []resourcev1.DeviceSelector{
				{
					CEL: &resourcev1.CELDeviceSelector{
						Expression: fmt.Sprintf("device.driver == '%s'", config.Driver),
					},
				},
			},
		},
	}
}

func DefaultDeviceClassParametersSpec() *DeviceClassParametersSpec {
	return &DeviceClassParametersSpec{
		DeviceSelector: []DeviceSelector{
			{
				Type: nascrd.RtCpuType,
				Name: "*",
			},
		},
	}
}

func DefaultRtClaimParametersSpec() *RtClaimParametersSpec {
	return &RtClaimParametersSpec{
		Count:   1,
		Runtime: 10, //should we put the default as miliseconds?
		Period:  100,
	}
}
