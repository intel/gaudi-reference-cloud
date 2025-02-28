// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewNamespace(namespace string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func NewVPC(vpcName, namespace, cidrBlock string) (*pb.VPCPrivate, error) {
	return &pb.VPCPrivate{
		Metadata: &pb.VPCMetadataPrivate{
			CloudAccountId:  namespace,
			Name:            vpcName,
			ResourceId:      vpcName,
			ResourceVersion: "12",
		},
		Spec: &pb.VPCSpecPrivate{
			CidrBlock: cidrBlock,
		},
		Status: &pb.VPCStatusPrivate{
			Phase:   pb.VPCPhase_VPCPhase_Provisioning,
			Message: "Provisioning",
		},
	}, nil
}
