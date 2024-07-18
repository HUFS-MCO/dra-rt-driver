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
	"math/rand"
	"os"
)

func enumerateAllPossibleDevices() (AllocatableRtCpus, error) {
	numGPUs := 8
	seed := os.Getenv("NODE_NAME")
	ids := generateIDs(seed, numGPUs)

	alldevices := make(AllocatableRtCpus)
	for _, id := range ids {
		deviceInfo := &AllocatableCpusetInfo{
			RtCpuInfo: &RtCpuInfo{
				id:   id,
				util: 10,
			},
		}
		alldevices[string(id)] = deviceInfo
	}
	return alldevices, nil
}

func generateIDs(seed string, count int) []int {
	rand := rand.New(rand.NewSource(hash(seed)))

	ids := make([]int, count)
	for i := 0; i < count; i++ {
		id := rand.Int()
		ids[i] = id
	}

	return ids
}

func hash(s string) int64 {
	h := int64(0)
	for _, c := range s {
		h = 31*h + int64(c)
	}
	return h
}
