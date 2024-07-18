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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/utils/cpuset"
)

const ()

type CgroupName []string

type CgroupManager struct {
}

// func (cdi *CDIHandler) CreateClaimSpecFile(claimUID string, devices *PreparedDevices) error {

// func (cdi *CDIHandler) DeleteClaimSpecFile(claimUID string) error {

// func (cdi *CDIHandler) GetClaimDevices(claimUID string, devices *PreparedDevices) ([]string, error) {

func writeCpuRtMultiRuntimeFile(cgroupFs string, cpuSet cpuset.CPUSet, rtRuntime int64) error {
	// TODO(stefano.fiori): can we write with opencontainer approach?
	const (
		CpuRtMultiRuntimeFile = "cpu.rt_multi_runtime_us"
	)

	if cpuSet.IsEmpty() {
		return nil
	}

	if err := os.MkdirAll(cgroupFs, os.ModePerm); err != nil {
		return fmt.Errorf("creating the container cgroupFs %s: %v", cgroupFs, err)
	}

	filePath := filepath.Join(cgroupFs, CpuRtMultiRuntimeFile)
	// BUG: write 0 gives error
	if rtRuntime == 0 {
		rtRuntime = 2
	}

	rtRuntimeStr := strconv.FormatInt(rtRuntime, 10)
	str := cpuSet.String() + " " + rtRuntimeStr

	if err := os.WriteFile(filePath, []byte(str), os.ModePerm); err != nil {
		return fmt.Errorf("writing %s in cpu.rt_multi_runtime_us, path %s: %v", str, filePath, err)
	}
	return nil
}

func writeRtFile(cgroupFs string, value int64) error {

	if err := os.MkdirAll(filepath.Dir(cgroupFs), os.ModePerm); err != nil {
		return fmt.Errorf("creating the container cgroupFs %s: %v", cgroupFs, err)
	}

	str := strconv.FormatInt(value, 10)

	if err := os.WriteFile(cgroupFs, []byte(str), os.ModePerm); err != nil {
		return fmt.Errorf("writing %v in cpu.rt_multi_runtime_us, path %v: %v", str, value, err)
	}
	return nil
}

func readCpuRtMultiRuntimeFile(cgroupFs string) ([]int64, error) {
	const (
		CpuRtMultiRuntimeFile = "cpu.rt_multi_runtime_us"
	)

	filePath := filepath.Join(cgroupFs, CpuRtMultiRuntimeFile)
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	runtimeStrings := strings.Split(string(buf), " ")
	runtimeStrings = runtimeStrings[:len(runtimeStrings)-2]

	runtimes := make([]int64, 0, len(runtimeStrings))
	for _, runtimeStr := range runtimeStrings {
		v, err := strconv.ParseInt(runtimeStr, 10, 32)
		if err != nil {
			panic(fmt.Errorf("error parsing runtime %s in file %s: %v", runtimeStr, filePath, err))
		}
		runtimes = append(runtimes, v)
	}
	return runtimes, nil
}
