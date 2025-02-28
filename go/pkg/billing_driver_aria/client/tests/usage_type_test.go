// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestGetUsageUnitTypes(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetUsageUnitTypes")
	logger.Info("testing get usage unit types")

	usageTypeClient := common.GetAriaUsageTypeClient()
	_, err := usageTypeClient.GetUsageUnitTypes(context.Background())
	if err != nil {
		t.Fatalf("failed to get usage unit types: %v", err)
	}
}

func TestGetUsageTypes(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetUsageTypes")
	logger.Info("testing get usage types")

	usageTypeClient := common.GetAriaUsageTypeClient()
	_, err := usageTypeClient.GetUsageTypes(context.Background())
	if err != nil {
		t.Fatalf("Failed to get usage types: %v", err)
	}
}

func TestUsageType(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetUsageTypes")
	logger.Info("testing create usage types")
	testUsagesTypeCode := uuid.New().String()[:30]
	usageTypeClient := common.GetAriaUsageTypeClient()
	usageUnitTypeNo := 2
	_, err := usageTypeClient.CreateUsageType(context.Background(), client.USAGE_TYPE_NAME, client.USAGE_TYPE_DESC, usageUnitTypeNo, testUsagesTypeCode)
	if err != nil {
		t.Fatalf("failed to create usage type: %v", err)
	}
	usageTypeDetails, err := usageTypeClient.GetUsageTypeDetails(context.Background(), testUsagesTypeCode)
	if err != nil {
		t.Fatalf("failed to get usage types: %v", err)
	}
	assert.Equal(t, testUsagesTypeCode, usageTypeDetails.UsageTypeCode)
	testUsagesTypeName := "test_minutes"
	_, err = usageTypeClient.UpdateUsageType(context.Background(), testUsagesTypeName, client.USAGE_TYPE_DESC, usageUnitTypeNo, testUsagesTypeCode)
	if err != nil {
		t.Fatalf("failed to get usage types: %v", err)
	}
	usageTypeDetails, err = usageTypeClient.GetUsageTypeDetails(context.Background(), testUsagesTypeCode)
	if err != nil {
		t.Fatalf("failed to get usage types: %v", err)
	}
	assert.Equal(t, testUsagesTypeName, usageTypeDetails.UsageTypeName)
}
