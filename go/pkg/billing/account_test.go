// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStandardCreateAccount(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCreateAccount")
	logger.Info("BEGIN")
	defer logger.Info("End")
	user := "stduser_" + uuid.NewString() + "example.com"
	acct := CreateAndGetCloudAccount(t, ctx, user, pb.AccountType_ACCOUNT_TYPE_STANDARD)

	billingClient := pb.NewBillingAccountServiceClient(clientConn)
	_, err := billingClient.Create(ctx, &pb.BillingAccount{
		CloudAccountId: acct.GetId(),
	})
	if err != nil {
		t.Fatalf("failed to create billing account for standard user: %v", err)
	}
}

func TestPremiumCreateAccount(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPremiumCreateAccount")
	logger.Info("BEGIN")
	defer logger.Info("End")
	user := "premium_" + uuid.NewString() + "example.com"
	acct := CreateAndGetCloudAccount(t, ctx, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM)

	billingClient := pb.NewBillingAccountServiceClient(clientConn)
	_, err := billingClient.Create(ctx, &pb.BillingAccount{
		CloudAccountId: acct.GetId(),
	})
	if err != nil {
		t.Fatalf("failed to create billing account for premium user: %v", err)
	}
}

func TestEntPendingCreateAccount(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestEntPendingCreateAccount")
	logger.Info("BEGIN")
	defer logger.Info("End")
	user := "ent_" + uuid.NewString() + "example.com"
	acct := CreateAndGetCloudAccount(t, ctx, user, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING)

	billingClient := pb.NewBillingAccountServiceClient(clientConn)
	_, err := billingClient.Create(ctx, &pb.BillingAccount{
		CloudAccountId: acct.GetId(),
	})
	if err != nil {
		t.Fatalf("failed to create billing account for enterprise pending user: %v", err)
	}
}

func TestFailsInvalidCloudAccountId(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestFailsInvalidCloudAccountId")
	logger.Info("BEGIN")
	defer logger.Info("End")
	billingClient := pb.NewBillingAccountServiceClient(clientConn)
	_, err := billingClient.Create(ctx, &pb.BillingAccount{
		CloudAccountId: uuid.NewString(),
	})
	if err == nil {
		t.Fatalf("billing  account creation did not fail as expected: %v", err)
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got wrong  error code: %v", err)
	}
	if !strings.Contains(err.Error(), InvalidCloudAccountId) {
		t.Fatalf("error code does not match")
	}
}
