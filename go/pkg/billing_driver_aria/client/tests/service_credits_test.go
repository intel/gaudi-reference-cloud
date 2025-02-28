// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/stretchr/testify/assert"
)

var currentDate = time.Now()
var newDate = currentDate.AddDate(0, 0, 100)
var expirationDate = fmt.Sprintf("%04d-%02d-%02d", newDate.Year(), newDate.Month(), newDate.Day())

func TestCreateCredits(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCreateCredits")
	logger.Info("testing create service credits")
	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	ariaAccountClient.CreateAriaAccount(context.Background(), id, client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_PREMIUM)
	serviceCreditClient := common.GetServiceCreditClient()
	// TODO: reason code should be 101 for initial cloud credits
	// and 102 for purchased cloud credits
	code := int64(1)
	_, err := serviceCreditClient.CreateServiceCredits(context.Background(), id, DefaultCloudCreditAmount, code, expirationDate, "test_Credit")
	if err != nil {
		t.Fatalf("failed to create credits: %v", err)
	}
	_, err = serviceCreditClient.GetUnappliedServiceCredits(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get unapplied credits: %v", err)
	}
}

// Test to apply_service_credit_m aria api
func TestApplyServiceCredit(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestApplyServiceCredit")
	logger.Info("testing apply service credit to aria account")
	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	serviceCreditClient := common.GetServiceCreditClient()
	clientPlanId := client.GetDefaultPlanClientId()
	ariaPlanClient := common.GetAriaPlanClient()
	err := ariaPlanClient.CreateDefaultPlan(context.Background())
	if err != nil && !strings.Contains(err.Error(), "error code:1001") {
		t.Fatalf("failed to create default plan: %v", err)
	}
	resp, err := ariaAccountClient.CreateAriaAccount(context.Background(), id, clientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(resp))

	// Aria api call for applying service credits
	_resp, err := serviceCreditClient.ApplyCreditService(context.Background(), id, 100)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	assert.Equal(t, int64(ERROR_CODE_OK), _resp.GetErrorCode())
	assert.Equal(t, ERROR_MESSAGE_OK, _resp.GetErrorMsg())
}
