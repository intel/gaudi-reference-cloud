// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestOptions(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	billingAccts := GetBillingAccounts(t, ctx)
	billingOptionClient := pb.NewBillingOptionServiceClient(ariaDriverClientConn)

	for _, billingAcct := range billingAccts {
		billingOptionFilter := &pb.BillingOptionFilter{
			CloudAccountId: &billingAcct.CloudAccountId,
		}
		billingOption, err := billingOptionClient.Read(ctx, billingOptionFilter)
		if err != nil {
			t.Fatalf("failed to get client for reading billing options: %v", err)
		}
		t.Logf("billing option: %v", billingOption)
	}
}

func TestOptionsForPremium(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	premiumBillingAcct := CreateBillingAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	billingOptionClient := pb.NewBillingOptionServiceClient(ariaDriverClientConn)
	AddTestPaymentMethod(t, ctx, premiumBillingAcct.CloudAccountId)
	billingOptionFilter := &pb.BillingOptionFilter{
		CloudAccountId: &premiumBillingAcct.CloudAccountId,
	}
	billingOption, err := billingOptionClient.Read(ctx, billingOptionFilter)
	if err != nil {
		t.Fatalf("failed to get client for reading billing options: %v", err)
	}

	if billingOption.CloudAccountId != premiumBillingAcct.CloudAccountId {
		t.Fatalf("cloud account id does not match")
	}

}

func TestOptionsInvalidFilter(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	cloudAccountId := uuid.NewString()
	billingOptionFilter := &pb.BillingOptionFilter{
		CloudAccountId: &cloudAccountId,
	}
	billingOptionClient := pb.NewBillingOptionServiceClient(ariaDriverClientConn)
	_, err := billingOptionClient.Read(ctx, billingOptionFilter)

	if err == nil {
		t.Fatalf("should have failed to read options")
	}
	if !strings.Contains(err.Error(), client.FailedToGetPaymentMethodsAllError) {
		t.Fatalf("error code does not match")
	}
	if !strings.Contains(err.Error(), FailedToReadBillingOptions) {
		t.Fatalf("error code does not match")
	}
}
