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

func TestAddPaymentPreProcessing(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestAddPaymentPreProcessing")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "example.com"
	//TODO: delete the created account
	acct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  premiumUser,
		Owner: premiumUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
	})

	client := pb.NewPaymentServiceClient(clientConn)
	_, err := client.AddPaymentPreProcessing(ctx, &pb.PrePaymentRequest{CloudAccountId: acct.Id})
	if err != nil {
		t.Fatalf("add payment preprocessing failed: %v", err)
	}

}
