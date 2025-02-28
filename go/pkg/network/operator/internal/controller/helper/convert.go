// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package helper

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DerefTime(t *metav1.Time) metav1.Time {
	if t == nil {
		return metav1.Time{}
	}
	return *t
}

func FromPbTimestamp(t *timestamppb.Timestamp) *metav1.Time {
	if t == nil {
		return nil
	}
	mt := metav1.NewTime(t.AsTime())
	return &mt
}
