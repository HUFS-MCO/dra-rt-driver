package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
)

// ClaimAllocation 타입 정의
type ClaimAllocation struct {
	Claim           *resourcev1.ResourceClaim
	ClaimParameters interface{}
	ClassParameters interface{}
	Class           *resourcev1.DeviceClass
	UnsuitableNodes []string
	Allocation      *resourcev1.AllocationResult
	Error           error
}

// Driver 인터페이스 정의
type Driver interface {
	GetClassParameters(ctx context.Context, class *resourcev1.DeviceClass) (interface{}, error)
	GetClaimParameters(ctx context.Context, claim *resourcev1.ResourceClaim, class *resourcev1.DeviceClass, classParameters interface{}) (interface{}, error)
	Allocate(ctx context.Context, cas []*ClaimAllocation, selectedNode string)
	Deallocate(ctx context.Context, claim *resourcev1.ResourceClaim) error
	UnsuitableNodes(ctx context.Context, pod *corev1.Pod, cas []*ClaimAllocation, potentialNodes []string) error
}

type OnSuccessCallback func()
