// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCredit(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	billingAccts := GetBillingAccounts(t, ctx)
	ariaDriverCreditClient := pb.NewBillingCreditServiceClient(ariaDriverClientConn)

	for _, billingAcct := range billingAccts {
		currentDate := time.Now()
		newDate := currentDate.AddDate(0, 0, config.Cfg.PremiumDefaultCreditExpirationDays)
		billingCredit := &pb.BillingCredit{
			Expiration:     timestamppb.New(newDate),
			CloudAccountId: billingAcct.CloudAccountId,
			Reason:         pb.BillingCreditReason_CREDIT_INITIAL,
			OriginalAmount: 100,
			CouponCode:     "SomeCode",
		}
		_, err := ariaDriverCreditClient.Create(ctx, billingCredit)
		if err != nil {
			t.Fatalf("failed to create cloud credit: %v", err)
		}

		unAppliedCreditBalance, err := ariaDriverCreditClient.ReadUnappliedCreditBalance(ctx, billingAcct)
		if err != nil {
			t.Fatalf("failed to get unapplied credit: %v", err)
		}
		if unAppliedCreditBalance.UnappliedAmount != billingCredit.OriginalAmount {
			t.Fatalf("un applied amount does not match initial amount")
		}

		billingCreditReadClient, err := ariaDriverCreditClient.ReadInternal(ctx, billingAcct)
		if err != nil {
			t.Fatalf("failed to get client for reading billing credit: %v", err)
		}
		for {
			billingCreditR, err := billingCreditReadClient.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("failed to read billing credit: %v", err)
			}
			if billingCreditR.OriginalAmount != billingCredit.OriginalAmount {
				t.Fatalf("original amount does not match")
			}
		}
	}
}

func TestCreditPremium(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditPremium")
	logger.Info("BEGIN")
	defer logger.Info("End")
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateCloudAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	_, clientPlanId, err := tests.CreateAcctWithPlanDefaultProd(ctx, client.GetAccountClientId(cloudAcct.Id))
	if err != nil {
		t.Fatalf("failed to create account with default prod")
	}
	ariaDriverCreditClient := pb.NewBillingCreditServiceClient(ariaDriverClientConn)
	currentDate := time.Now()
	newDate := currentDate.AddDate(0, 0, config.Cfg.PremiumDefaultCreditExpirationDays)
	billingCredit := &pb.BillingCredit{
		Expiration:     timestamppb.New(newDate),
		CloudAccountId: cloudAcct.Id,
		Reason:         pb.BillingCreditReason_CREDIT_INITIAL,
		OriginalAmount: 10000,
		CouponCode:     "SomeCode",
	}
	_, err = ariaDriverCreditClient.Create(ctx, billingCredit)
	billingCredit1 := &pb.BillingCredit{
		Expiration:     timestamppb.New(newDate.AddDate(0, 1, 1)),
		CloudAccountId: cloudAcct.Id,
		Reason:         pb.BillingCreditReason_CREDIT_INITIAL,
		OriginalAmount: 10,
		CouponCode:     "SomeCode",
	}
	_, err = ariaDriverCreditClient.Create(ctx, billingCredit1)
	_, err = ariaDriverCreditClient.Create(ctx, billingCredit)
	if err != nil {
		t.Fatalf("failed to create cloud credit: %v", err)
	}
	err = tests.CreateTestUsageRecord(ctx,
		client.GetAccountClientId(cloudAcct.Id),
		clientPlanId, client.GetMinsUsageTypeCode(), 100000)
	if err != nil {
		t.Fatalf("failed to add default usage record")
	}
	billingCreditResponse, err := ariaDriverCreditClient.Read(ctx, &pb.BillingCreditFilter{CloudAccountId: cloudAcct.Id})
	if err != nil {
		t.Fatalf("failed to get client for reading billing credit: %v", err)
	}
	logger.Info("total remaining", "amount", billingCreditResponse.TotalRemainingAmount)
}

func TestCreditPremiumVerifyOrder(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditPremium")
	logger.Info("BEGIN")
	defer logger.Info("End")
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateCloudAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	_, clientPlanId, err := tests.CreateAcctWithPlanDefaultProd(ctx, client.GetAccountClientId(cloudAcct.Id))
	if err != nil {
		t.Fatalf("failed to create account with default prod")
	}
	ariaDriverCreditClient := pb.NewBillingCreditServiceClient(ariaDriverClientConn)
	currentDate := time.Now()
	newDate := currentDate.AddDate(0, 0, config.Cfg.PremiumDefaultCreditExpirationDays)
	billingCredit := &pb.BillingCredit{
		Expiration:     timestamppb.New(newDate),
		CloudAccountId: cloudAcct.Id,
		Reason:         pb.BillingCreditReason_CREDIT_INITIAL,
		OriginalAmount: 100,
		CouponCode:     "EarlierCredit",
	}
	_, err = ariaDriverCreditClient.Create(ctx, billingCredit)
	if err != nil {
		t.Fatalf("failed to create cloud credit: %v", err)
	}

	newDate1 := currentDate.AddDate(0, 0, config.Cfg.PremiumDefaultCreditExpirationDays+1)
	billingCredit1 := &pb.BillingCredit{
		Expiration:     timestamppb.New(newDate1),
		CloudAccountId: cloudAcct.Id,
		Reason:         pb.BillingCreditReason_CREDIT_INITIAL,
		OriginalAmount: 100,
		CouponCode:     "LaterCredit",
	}
	_, err = ariaDriverCreditClient.Create(ctx, billingCredit1)
	if err != nil {
		t.Fatalf("failed to create second cloud credit: %v", err)
	}
	err = tests.CreateTestUsageRecord(ctx,
		client.GetAccountClientId(cloudAcct.Id),
		clientPlanId, client.GetMinsUsageTypeCode(), tests.DefaultUsageAmount)
	if err != nil {
		t.Fatalf("failed to add default usage record")
	}
	billingCreditResponse, err := ariaDriverCreditClient.Read(ctx, &pb.BillingCreditFilter{CloudAccountId: cloudAcct.Id})
	if err != nil {
		t.Fatalf("failed to get client for reading billing credit: %v", err)
	}
	logger.Info("total remaining", "amount", billingCreditResponse.TotalRemainingAmount)
}

func TestCreateCreditInvalidCloudAcct(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	ariaDriverCreditClient := pb.NewBillingCreditServiceClient(ariaDriverClientConn)
	currentDate := time.Now()
	newDate := currentDate.AddDate(0, 0, config.Cfg.PremiumDefaultCreditExpirationDays)
	billingCredit := &pb.BillingCredit{
		Expiration:     timestamppb.New(newDate),
		CloudAccountId: uuid.NewString(),
		Reason:         pb.BillingCreditReason_CREDIT_INITIAL,
		OriginalAmount: 100,
		CouponCode:     "SomeCode",
	}
	_, err := ariaDriverCreditClient.Create(ctx, billingCredit)
	if err == nil {
		t.Fatalf("cloud credit should not have been created")
	}
	if !strings.Contains(err.Error(), client.FailedToCreateCloudCreditError) {
		t.Fatalf("error code does not match")
	}
	if !strings.Contains(err.Error(), FailedToCreateBillingCredit) {
		t.Fatalf("error code does not match")
	}
}

func TestCreateCreditInvalidExpiration(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	ariaDriverCreditClient := pb.NewBillingCreditServiceClient(ariaDriverClientConn)
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	billingAcct := CreateBillingAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	billingCredit := &pb.BillingCredit{
		Expiration:     timestamppb.New(time.Now()),
		CloudAccountId: billingAcct.CloudAccountId,
		Reason:         pb.BillingCreditReason_CREDIT_INITIAL,
		OriginalAmount: 100,
		CouponCode:     "SomeCode",
	}
	_, err := ariaDriverCreditClient.Create(ctx, billingCredit)
	if err == nil {
		t.Fatalf("cloud credit should not have been created")
	}
	if !strings.Contains(err.Error(), client.FailedToCreateCloudCreditError) {
		t.Fatalf("error code does not match")
	}
	if !strings.Contains(err.Error(), FailedToCreateBillingCredit) {
		t.Fatalf("error code does not match")
	}
}

func TestUnAppliedCreditAmountInvalidCloudAccount(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	ariaDriverCreditClient := pb.NewBillingCreditServiceClient(ariaDriverClientConn)
	_, err := ariaDriverCreditClient.ReadUnappliedCreditBalance(ctx, &pb.BillingAccount{CloudAccountId: uuid.NewString()})
	if err == nil {
		t.Fatalf("get unapplied credit amount should have failed")
	}
	if !strings.Contains(err.Error(), client.FailedToGetUnappliedServiceCreditsError) {
		t.Fatalf("error code does not match")
	}
	if !strings.Contains(err.Error(), FailedToReadUnappliedCreditBalance) {
		t.Fatalf("error code does not match")
	}
}
