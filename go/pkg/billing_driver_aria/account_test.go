// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestAccount(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestAccount")
	logger.Info("BEGIN")
	defer logger.Info("End")
	cloudAccts := GetCloudAccounts(t, ctx)
	ariaAccountClient := pb.NewBillingAccountServiceClient(ariaDriverClientConn)
	for _, cloudAccount := range cloudAccts {
		billingAcct := &pb.BillingAccount{
			CloudAccountId: cloudAccount.Id,
		}
		_, err := ariaAccountClient.Create(ctx, billingAcct)
		if err != nil {
			t.Fatalf("failed to create billing account: %v", err)
		}
	}
}
