// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"fmt"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"golang.org/x/exp/slices"
)

const meteringQuantityPropertyKey = "runningSeconds"
const meteringServiceTypeKey = "serviceType"

func GetStorageMeteringMetric(meteringRecord *pb.MeteringRecord) (float64, error) {
	storageQty, ok := meteringRecord.Properties[Cfg.StorageMetricUnitType]
	if !ok {
		return 0, fmt.Errorf("metering %v missing %v", meteringRecord.Id, Cfg.StorageMetricUnitType)
	}

	floatVal, err := strconv.ParseFloat(storageQty, 64)
	if err != nil {
		return 0, fmt.Errorf("metering %v: invalid %v: %w", meteringRecord.Id, Cfg.StorageMetricUnitType, err)
	}

	return floatVal, nil
}

func GetStorageMeteringTimeMetric(meteringRecord *pb.MeteringRecord) (float64, error) {
	timeQty, ok := meteringRecord.Properties[Cfg.StorageTimeMetricUnitType]
	if !ok {
		return 0, fmt.Errorf("metering %v missing %v", meteringRecord.Id, Cfg.StorageTimeMetricUnitType)
	}

	floatVal, err := strconv.ParseFloat(timeQty, 64)
	if err != nil {
		return 0, fmt.Errorf("metering %v: invalid %v: %w", meteringRecord.Id, Cfg.StorageTimeMetricUnitType, err)
	}

	return floatVal, nil
}

func GetStorageMeteringQuantity(meteringRecord *pb.MeteringRecord) (float64, error) {
	timeQty, err := GetStorageMeteringTimeMetric(meteringRecord)
	if err != nil {
		return 0, err
	}

	storageQty, err := GetStorageMeteringMetric(meteringRecord)
	if err != nil {
		return 0, err
	}

	return storageQty * timeQty, nil
}

func GetStorageMeteringRelativeQuantity(currentMeteringRecord *pb.MeteringRecord, previousMeteringRecord *pb.MeteringRecord) (float64, error) {
	currentTimeQty, err := GetStorageMeteringTimeMetric(currentMeteringRecord)
	if err != nil {
		return 0, err
	}

	prevTimeQty, err := GetStorageMeteringTimeMetric(previousMeteringRecord)
	if err != nil {
		return 0, err
	}

	if currentTimeQty < prevTimeQty {
		return 0, fmt.Errorf("metering %v: less than previous %v: %w", currentMeteringRecord.Id, Cfg.StorageTimeMetricUnitType, err)
	}

	previousStorageQty, err := GetStorageMeteringMetric(previousMeteringRecord)
	if err != nil {
		return 0, err
	}

	return previousStorageQty * (currentTimeQty - prevTimeQty), nil
}

func GetStorageUsageQuantity(meteringRecords []*pb.MeteringRecord) (float64, error) {

	if len(meteringRecords) == 1 {
		return GetStorageMeteringQuantity(meteringRecords[0])
	}

	if len(meteringRecords) == 2 {
		return GetStorageMeteringRelativeQuantity(meteringRecords[0], meteringRecords[1])
	}

	var totalQty float64 = 0
	var meteringRecordsIter = 0
	lengthOfMeteringRecords := len(meteringRecords)

	for meteringRecordsIter <= (lengthOfMeteringRecords - 2) {

		qty, err := GetStorageMeteringRelativeQuantity(meteringRecords[meteringRecordsIter], meteringRecords[meteringRecordsIter+1])

		if err != nil {
			return 0, err
		}

		totalQty = totalQty + qty
		meteringRecordsIter = meteringRecordsIter + 1
	}

	return totalQty, nil
}

func GetUsageQuantity(meteringRecords []*pb.MeteringRecord) (float64, error) {
	if _, foundServiceTypeKey := meteringRecords[0].Properties[meteringServiceTypeKey]; foundServiceTypeKey {
		serviceType := meteringRecords[0].Properties[meteringServiceTypeKey]
		if slices.Contains(Cfg.StorageServiceTypes, serviceType) {
			return GetStorageUsageQuantity(meteringRecords)
		}
	}
	return GetMeteringQuantityInMins(meteringRecords[0])
}

func GetMeteringQuantityInMins(meteringRecord *pb.MeteringRecord) (float64, error) {
	// This should not be running seconds and instead should be time
	val, ok := meteringRecord.Properties[meteringQuantityPropertyKey]
	if !ok {
		return 0, fmt.Errorf("metering %v missing %v", meteringRecord.Id, meteringQuantityPropertyKey)
	}
	numVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("usage %v: invalid quantity: %w", meteringRecord.Id, err)
	}
	return numVal / 60, nil
}

func GetMeteringQuantityMultipleMetrics(meteringRecords []*pb.MeteringRecord) (float64, error) {
	return 0, nil
}
