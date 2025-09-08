package controller

import (
	resourcev1 "k8s.io/api/resource/v1"
)

// 1.34 호환 ClaimAllocation 타입 정의
type ClaimAllocation struct {
	Claim           *resourcev1.ResourceClaim
	ClaimParameters interface{}
	UnsuitableNodes []string
}

type OnSuccessCallback func()
