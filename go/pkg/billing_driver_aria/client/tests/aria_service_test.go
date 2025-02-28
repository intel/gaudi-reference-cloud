// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/stretchr/testify/assert"
)

var serviceId = uuid.New().String()
var serviceName = uuid.New().String()

func TestCreateAriaService(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCreateAriaService")
	logger.Info("testing create aria service")

	usageType := GetUsageType(t)
	ariaServiceClient := common.GetAriaServiceClient()
	_, err := ariaServiceClient.CreateAriaService(context.Background(), serviceId, serviceName, usageType.UsageTypeNo)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
}

func TestUpdateAriaService(t *testing.T) {
	var serviceId = uuid.New().String()
	logger := log.FromContext(context.Background()).WithName("TestUpdateAriaService")
	logger.Info("testing update aria service")

	ariaServiceClient := common.GetAriaServiceClient()
	usageType := GetUsageType(t)
	createResp, err := ariaServiceClient.CreateAriaService(context.Background(), serviceId, serviceName, usageType.UsageTypeNo)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	serviceDetails, err := ariaServiceClient.GetServiceDetailsForServiceNo(context.Background(), createResp.ServiceNo)
	if err != nil {
		t.Fatalf("failed to get services: %v", err)
	}
	assert.Equal(t, serviceId, serviceDetails.ClientServiceId)
	assert.Equal(t, serviceName, serviceDetails.ServiceName)
	assert.Equal(t, usageType.UsageTypeNo, serviceDetails.UsageType)
	testServieName := "TestServiceName"
	updateResp, err := ariaServiceClient.UpdateAriaService(context.Background(), serviceId, testServieName, usageType.UsageTypeNo)
	if err != nil {
		t.Fatalf("failed to update service: %v", err)
	}
	assert.Equal(t, createResp.ServiceNo, updateResp.ServiceNo)
	updateServiceDetails, err := ariaServiceClient.GetServiceDetailsForServiceNo(context.Background(), createResp.ServiceNo)
	if err != nil {
		t.Fatalf("failed to get services: %v", err)
	}
	assert.Equal(t, testServieName, updateServiceDetails.ServiceName)
}
