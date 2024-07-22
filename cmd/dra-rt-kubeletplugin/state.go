package main

import (
	"fmt"
	"sync"

	nascrd "github.com/nasim-samimi/dra-rt-driver/api/example.com/resource/rt/nas/v1alpha1"
)

type DeviceState struct {
	sync.Mutex
	cdi         *CDIHandler
	allocatable AllocatableRtCpus
	prepared    PreparedClaims
}

func NewDeviceState(config *Config) (*DeviceState, error) {
	allocatable, err := enumerateAllPossibleDevices()
	if err != nil {
		return nil, fmt.Errorf("error enumerating all possible devices: %v", err)
	}

	cdi, err := NewCDIHandler(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create CDI handler: %v", err)
	}

	err = cdi.CreateCommonSpecFile()
	if err != nil {
		return nil, fmt.Errorf("unable to create CDI spec file for common edits: %v", err)
	}

	state := &DeviceState{
		cdi:         cdi,
		allocatable: allocatable,
		prepared:    make(PreparedClaims),
	}

	err = state.syncPreparedCpusetFromCRDSpec(&config.nascr.Spec)
	if err != nil {
		return nil, fmt.Errorf("unable to sync prepared devices from CRD: %v", err)
	}

	return state, nil
}

func (s *DeviceState) Prepare(claimUID string, allocation nascrd.AllocatedCpuset) ([]string, error) {
	s.Lock()
	defer s.Unlock()

	if s.prepared[claimUID] != nil {
		cdiDevices, err := s.cdi.GetClaimDevices(claimUID, s.prepared[claimUID])
		if err != nil {
			return nil, fmt.Errorf("unable to get CDI devices names: %v", err)
		}
		return cdiDevices, nil
	}

	prepared := &PreparedCpuset{}

	var err error
	switch allocation.Type() {
	case nascrd.RtCpuType:
		prepared.RtCpu, err = s.prepareRtCpus(claimUID, allocation.RtCpu)
	default:
		err = fmt.Errorf("unknown device type: %v", allocation.Type())
	}
	if err != nil {
		return nil, fmt.Errorf("allocation failed: %v", err)
	}

	err = s.cdi.CreateClaimSpecFile(claimUID, prepared)
	if err != nil {
		return nil, fmt.Errorf("unable to create CDI spec file for claim: %v", err)
	}

	s.prepared[claimUID] = prepared

	cdiDevices, err := s.cdi.GetClaimDevices(claimUID, s.prepared[claimUID])
	if err != nil {
		return nil, fmt.Errorf("unable to get CDI devices names: %v", err)
	}
	return cdiDevices, nil
}

func (s *DeviceState) Unprepare(claimUID string) error {
	s.Lock()
	defer s.Unlock()

	if s.prepared[claimUID] == nil {
		return nil
	}

	switch s.prepared[claimUID].Type() {
	case nascrd.RtCpuType:
		err := s.unprepareRtCpus(claimUID, s.prepared[claimUID])
		if err != nil {
			return fmt.Errorf("unprepare failed: %v", err)
		}
	default:
		return fmt.Errorf("unknown device type: %v", s.prepared[claimUID].Type())
	}

	err := s.cdi.DeleteClaimSpecFile(claimUID)
	if err != nil {
		return fmt.Errorf("unable to delete CDI spec file for claim: %v", err)
	}

	delete(s.prepared, claimUID)

	return nil
}

func (s *DeviceState) GetUpdatedSpec(inspec *nascrd.NodeAllocationStateSpec) (*nascrd.NodeAllocationStateSpec, error) {
	s.Lock()
	defer s.Unlock()

	outspec := inspec.DeepCopy()
	err := s.syncAllocatableRtCpusToCRDSpec(outspec)
	if err != nil {
		return nil, fmt.Errorf("synching allocatable devices to CR spec: %v", err)
	}

	err = s.syncPreparedRtCpuToCRDSpec(outspec)
	if err != nil {
		return nil, fmt.Errorf("synching prepared devices to CR spec: %v", err)
	}

	return outspec, nil
}

func (s *DeviceState) prepareRtCpus(claimUID string, allocated *nascrd.AllocatedRtCpu) (*PreparedRtCpu, error) {
	prepared := &PreparedRtCpu{}

	for _, device := range allocated.Cpuset {
		cpuInfo := s.allocatable[device.ID].RtCpuInfo

		if _, exists := s.allocatable[device.ID]; !exists {
			return nil, fmt.Errorf("requested CPU does not exist: %v", device.ID)
		}
		prepared.Cpuset = append(prepared.Cpuset, cpuInfo)
	}

	return prepared, nil
}

func (s *DeviceState) unprepareRtCpus(claimUID string, devices *PreparedCpuset) error {
	return nil
}

func (s *DeviceState) syncAllocatableRtCpusToCRDSpec(spec *nascrd.NodeAllocationStateSpec) error {
	cpus := make(map[int]nascrd.AllocatableRtCpu)
	for _, device := range s.allocatable {
		cpus[device.id] = nascrd.AllocatableRtCpu{
			RtCpu: &nascrd.AllocatableCpu{
				ID:   device.id,
				Util: device.util,
			},
		}
	}

	var allocatable []nascrd.AllocatableRtCpu
	for _, device := range cpus {
		allocatable = append(allocatable, device)
	}

	spec.AllocatableRtCpu = allocatable

	return nil
}

func (s *DeviceState) syncPreparedCpusetFromCRDSpec(spec *nascrd.NodeAllocationStateSpec) error {
	cpus := s.allocatable

	prepared := make(PreparedClaims)
	for claim, devices := range spec.PreparedClaims {
		switch devices.Type() {
		case nascrd.RtCpuType:
			prepared[claim] = &PreparedCpuset{RtCpu: &PreparedRtCpu{}}
			for _, d := range devices.RtCpu.Cpuset {
				prepared[claim].RtCpu.Cpuset = append(prepared[claim].RtCpu.Cpuset, cpus[d.ID].RtCpuInfo)
			}
		default:
			return fmt.Errorf("unknown device type: %v", devices.Type())
		}
	}

	s.prepared = prepared

	return nil
}

func (s *DeviceState) syncPreparedRtCpuToCRDSpec(spec *nascrd.NodeAllocationStateSpec) error {
	outcas := make(map[string]nascrd.PreparedCpuset)
	for claim, devices := range s.prepared {
		var prepared nascrd.PreparedCpuset
		switch devices.Type() {
		case nascrd.RtCpuType:
			prepared.RtCpu = &nascrd.PreparedRtCpu{}
			for _, device := range devices.RtCpu.Cpuset {
				outdevice := nascrd.PreparedCpu{
					ID: device.id,
				}
				prepared.RtCpu.Cpuset = append(prepared.RtCpu.Cpuset, outdevice)
			}
		default:
			return fmt.Errorf("unknown device type: %v", devices.Type())
		}
		outcas[claim] = prepared
	}

	spec.PreparedClaims = outcas

	return nil
}
