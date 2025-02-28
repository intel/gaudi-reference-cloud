// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	aria "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria"
	intel "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	standard "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_standard"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	"context"
	"testing"

	ariaConfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	credits "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	"github.com/google/uuid"
)

func TestPremiumCloudCreditUsageEvent(t *testing.T) {
	if ariaConfig.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPremiumCloudCreditUsageEvent")
	logger.Info("BEGIN")
	defer logger.Info("End")
	premiumUser := "premium_" + uuid.NewString() + "@example.com"

	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditUsageEventScheduler := credits.NewCloudCreditUsageEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	createAcctWithCreditsUsageEventScheduler(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM, cloudCreditUsageEventScheduler)
	logger.Info("and done with testing cloud credit usage for premium customers")
}

func TestPremiumCloudCreditUsageEventThreshold(t *testing.T) {
	if ariaConfig.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPremiumCloudCreditUsageEventThreshold")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditUsageEventScheduler := credits.NewCloudCreditUsageEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsUsageEventScheduler(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM, cloudCreditUsageEventScheduler)

	aria.PostUsageForAccount(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id}, 0.81*(DefaultCloudCreditAmount+ariaConfig.Cfg.PremiumDefaultCreditAmount))
	cloudCreditUsageEventScheduler.CloudCreditUsageEvent(ctx)
	logger.Info("and done with testing cloud credit usage for premium customers with usage threshold")
}

// todo: add a test for verifying the account is not disabled if has a payment method for both
// premium and enterprise.
func TestPremiumCloudCreditUsageEventComplete(t *testing.T) {
	if ariaConfig.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPremiumCloudCreditUsageEventComplete")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "@example.com"

	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditUsageEventScheduler := credits.NewCloudCreditUsageEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsUsageEventScheduler(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM, cloudCreditUsageEventScheduler)
	aria.PostUsageForAccount(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id}, 1.01*(DefaultCloudCreditAmount+ariaConfig.Cfg.PremiumDefaultCreditAmount))
	cloudCreditUsageEventScheduler.CloudCreditUsageEvent(ctx)
	logger.Info("and done with testing cloud credit usage for premium customers with usage completely used")
}

func TestIntelCloudCreditUsageEvent(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditUsageEvent")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"

	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditUsageEventScheduler := credits.NewCloudCreditUsageEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	createAcctWithCreditsUsageEventScheduler(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditUsageEventScheduler)

	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for intel customers")
}

func TestIntelCloudCreditUsageEventThreshold(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditUsageEventThreshold")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"

	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditUsageEventScheduler := credits.NewCloudCreditUsageEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsUsageEventScheduler(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditUsageEventScheduler)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

	intelCredits := GetBillingCredits(t, ctx, billingAcct)
	for _, intelCredit := range intelCredits {
		if intelCredit.CloudAccountId == cloudAcct.Id {
			intel.UpdateIntelCreditWithRemainingAmount(t, ctx, billingAcct, intelCredit.CloudAccountId, 0.19*DefaultCloudCreditAmount)
		}
	}

	cloudCreditUsageEventScheduler.CloudCreditUsageEvent(ctx)
	VerifyCloudAcctHasLowCredits(t, cloudAcct.Id)
	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for intel customers with usage threshold")
}

func TestIntelCloudCreditUsageEventComplete(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditUsageEventComplete")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"

	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditUsageEventScheduler := credits.NewCloudCreditUsageEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsUsageEventScheduler(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditUsageEventScheduler)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

	intel.UpdateIntelCreditUsed(t, ctx, billingAcct, GetBillingCredits(t, ctx, billingAcct))
	cloudCreditUsageEventScheduler.CloudCreditUsageEvent(ctx)
	VerifyCloudAcctHasNoCredits(t, cloudAcct.Id)
	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for intel customers with usage completely used")
}

func TestIntelCloudCreditsUsageEventComplete(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditsUsageComplete")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"

	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditUsageEventScheduler := credits.NewCloudCreditUsageEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsUsageEventScheduler(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditUsageEventScheduler)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

	intelUser1 := "intel_" + uuid.NewString() + "@example.com"
	cloudAcct1 := createAcctWithCreditsUsageEventScheduler(t, ctx, intelUser1, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditUsageEventScheduler)
	billingAcct1 := &pb.BillingAccount{CloudAccountId: cloudAcct1.Id}

	CreateBillingCredit(t, ctx, cloudAcct, billingAcct)
	CreateBillingCredit(t, ctx, cloudAcct1, billingAcct1)
	cloudCreditUsageEventScheduler.CloudCreditUsageEvent(ctx)

	VerifyCloudAcctHasCredits(t, cloudAcct.Id)
	VerifyCloudAcctHasCredits(t, cloudAcct1.Id)

	intel.UpdateIntelCreditUsed(t, ctx, billingAcct, GetBillingCredits(t, ctx, billingAcct))
	intel.UpdateIntelCreditUsed(t, ctx, billingAcct1, GetBillingCredits(t, ctx, billingAcct1))
	cloudCreditUsageEventScheduler.CloudCreditUsageEvent(ctx)

	VerifyCloudAcctHasNoCredits(t, cloudAcct.Id)
	VerifyCloudAcctHasNoCredits(t, cloudAcct1.Id)
	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit usages for intel customers with usage completely used")
}

func TestStandardCloudCreditUsageEvent(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCloudCreditUsageEvent")
	logger.Info("BEGIN")
	defer logger.Info("End")

	standardUser := "standard_" + uuid.NewString() + "@example.com"
	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditUsageEventScheduler := credits.NewCloudCreditUsageEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	createAcctWithCreditsUsageEventScheduler(t, ctx, standardUser, pb.AccountType_ACCOUNT_TYPE_STANDARD, cloudCreditUsageEventScheduler)

	standard.DeleteStandardCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for standard customers")
}

func TestStandardCloudCreditUsageEventThreshold(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCloudCreditUsageEventThreshold")
	logger.Info("BEGIN")
	defer logger.Info("End")

	standardUser := "standard_" + uuid.NewString() + "@example.com"
	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditUsageEventScheduler := credits.NewCloudCreditUsageEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)

	cloudAcct := createAcctWithCreditsUsageEventScheduler(t, ctx, standardUser, pb.AccountType_ACCOUNT_TYPE_STANDARD, cloudCreditUsageEventScheduler)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

	standardCredits := GetBillingCredits(t, ctx, billingAcct)
	for _, standardCredit := range standardCredits {
		if standardCredit.CloudAccountId == cloudAcct.Id {
			standard.UpdateStandardCreditWithRemainingAmount(t, ctx, billingAcct, standardCredit.CloudAccountId, 0.19*DefaultCloudCreditAmount)
		}
	}

	cloudCreditUsageEventScheduler.CloudCreditUsageEvent(ctx)
	VerifyCloudAcctHasLowCredits(t, cloudAcct.Id)
	standard.DeleteStandardCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for standard customers with usage threshold")
}

func TestStandardCloudCreditUsageEventComplete(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCloudCreditUsageEventComplete")
	logger.Info("BEGIN")
	defer logger.Info("End")

	standardUser := "standard_" + uuid.NewString() + "@example.com"
	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditUsageEventScheduler := credits.NewCloudCreditUsageEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)

	cloudAcct := createAcctWithCreditsUsageEventScheduler(t, ctx, standardUser, pb.AccountType_ACCOUNT_TYPE_STANDARD, cloudCreditUsageEventScheduler)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

	standard.UpdateStandardCreditUsed(t, ctx, billingAcct, GetBillingCredits(t, ctx, billingAcct))
	cloudCreditUsageEventScheduler.CloudCreditUsageEvent(ctx)
	VerifyCloudAcctHasNoCredits(t, cloudAcct.Id)
	standard.DeleteStandardCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for standard customers with usage completely used")
}

func createAcctWithCreditsUsageEventScheduler(t *testing.T, ctx context.Context, user string, acctType pb.AccountType, scheduler *credits.CloudCreditUsageEventScheduler) *pb.CloudAccount {
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  acctType,
	})
	CreateBillingCredit(t, ctx, cloudAcct, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
	scheduler.CloudCreditUsageEvent(ctx)
	return cloudAcct
}
