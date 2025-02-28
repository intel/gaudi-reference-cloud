// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// This test calls through the billing service auto-generated proxy code
// to the Aria driver. This tests the auto-generated streaming code (with
// no data streamed just yet)
func TestReadBillingOption(t *testing.T) {
	t.Skip("Test disabled for billing and included to be part of driver")
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestReadBillingOption")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "example.com"

	acct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  premiumUser,
		Owner: premiumUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
	})

	client := pb.NewBillingOptionServiceClient(clientConn)
	res, err := client.Read(ctx,
		&pb.BillingOptionFilter{CloudAccountId: &acct.Id})

	//TODO: delete the above created account
	if err != nil {
		t.Fatalf("error reading billing options: %v", err)
	}
	t.Logf("billing option: %v", res)
}
