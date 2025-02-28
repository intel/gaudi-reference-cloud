// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import "google.golang.org/protobuf/types/known/timestamppb"

const (
	cloudAccountIdLenLimit   = 12
	recordIdLenLimit         = 50
	resourceIdLenLimit       = 50
	transactionIdLenLimit    = 50
	regionLenLimit           = 50
	availabilityZoneLenLimit = 20
)

func IsValidRecordId(recordId string) bool {
	//TODO: Add any other validation check here
	return len(recordId) == 0 || len(recordId) > recordIdLenLimit
}

func IsValidCloudAccountId(cloudAccountId string) bool {
	//TODO: Add any other validation check here
	return len(cloudAccountId) != cloudAccountIdLenLimit
}

func IsValidResourceId(resourceId string) bool {
	//TODO: Add any other validation check here
	return len(resourceId) == 0 || len(resourceId) > resourceIdLenLimit
}

func IsValidRegion(region string) bool {
	//TODO: Add any other validation check here
	return len(region) == 0 || len(region) > regionLenLimit
}

func IsValidTransactionId(transactionId string) bool {
	//TODO: Add any other validation check here
	return len(transactionId) == 0 || len(transactionId) > transactionIdLenLimit
}

func IsValidTimestamp(timestamp *timestamppb.Timestamp) bool {
	//TODO: Add any other validation check here
	return timestamp == nil
}

func IsValidAvailibilityZone(availabilityZone string) bool {
	//TODO: Add any other validation check here
	return len(availabilityZone) > 0 && len(availabilityZone) <= availabilityZoneLenLimit
}
