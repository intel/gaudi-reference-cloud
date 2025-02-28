// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	defaultVMClusterId = "default_cluster_id"
)

func GetComputeUsageRecordCreate(cloudAccountId string, resourceId string, transactionId string,
	region string, serviceName string, instanceType string, quantity float32) *pb.UsageCreate {

	return &pb.UsageCreate{
		CloudAccountId: cloudAccountId,
		ResourceId:     resourceId,
		TransactionId:  transactionId,
		Properties: map[string]string{
			"availabilityZone":    region,
			"clusterId":           defaultVMClusterId,
			"deleted":             "false",
			"firstReadyTimestamp": "2023-02-16T14:53:29Z",
			// this has to be a valid service type
			"serviceType":    "ComputeAsAService",
			"service":        serviceName,
			"instanceType":   instanceType,
			"region":         region,
			"runningSeconds": fmt.Sprintf("%f", quantity),
		},
		Timestamp: timestamppb.Now(),
	}
}

func GetStorageUsageRecordCreate(cloudAccountId string, serviceType string, resourceId string, transactionId string,
	region string, serviceName string, timeQuantity float32, storageQuantity float32) *pb.UsageCreate {

	return &pb.UsageCreate{
		CloudAccountId: cloudAccountId,
		ResourceId:     resourceId,
		TransactionId:  transactionId,
		Properties: map[string]string{
			"availabilityZone":    region,
			"deleted":             "false",
			"firstReadyTimestamp": "2023-02-16T14:53:29Z",
			// this has to be a valid service type
			"serviceType":             serviceType,
			"service":                 serviceName,
			"region":                  region,
			StorageTimeMetricUnitType: fmt.Sprintf("%f", timeQuantity),
			StorageMetricUnitType:     fmt.Sprintf("%f", storageQuantity),
		},
		Timestamp: timestamppb.Now(),
	}
}
