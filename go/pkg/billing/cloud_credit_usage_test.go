// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	aria "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	intel "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	standard "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_standard"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	"github.com/google/uuid"
)

func TestPremiumCloudCreditUsage(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPremiumCloudCreditUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM, cloudCreditUsageScheduler)
	logger.Info("and done with testing cloud credit usage for premium customers")
}

func TestPremiumCloudCreditUsageThreshold(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPremiumCloudCreditUsageThreshold")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM, cloudCreditUsageScheduler)

	aria.PostUsageForAccount(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id}, 0.81*(DefaultCloudCreditAmount+config.Cfg.PremiumDefaultCreditAmount))
	cloudCreditUsageScheduler.cloudCreditUsages(ctx)
	// cloud credit usages do not get posted until invoices are generated.
	// the code needs to be changed for Aria to calculate cloud credit usages without invoicing.
	//VerifyCloudAcctHasLowCredits(t, cloudAcct.Id)
	logger.Info("and done with testing cloud credit usage for premium customers with usage threshold")
}

// todo: add a test for verifying the account is not disabled if has a payment method for both
// premium and enterprise.
func TestPremiumCloudCreditUsageComplete(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPremiumCloudCreditUsageComplete")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM, cloudCreditUsageScheduler)

	aria.PostUsageForAccount(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id}, 1.01*(DefaultCloudCreditAmount+config.Cfg.PremiumDefaultCreditAmount))
	cloudCreditUsageScheduler.cloudCreditUsages(ctx)
	// cloud credit usages do not get posted until invoices are generated.
	// the code needs to be changed for Aria to calculate cloud credit usages without invoicing.
	//VerifyCloudAcctHasNoCredits(t, cloudAcct.Id)
	logger.Info("and done with testing cloud credit usage for premium customers with usage completely used")
}

func TestIntelCloudCreditUsage(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditUsageScheduler)

	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for intel customers")
}

func TestIntelCloudCreditUsageThreshold(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditUsageThreshold")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditUsageScheduler)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

	intelCredits := GetBillingCredits(t, ctx, billingAcct)
	for _, intelCredit := range intelCredits {
		if intelCredit.CloudAccountId == cloudAcct.Id {
			intel.UpdateIntelCreditWithRemainingAmount(t, ctx, billingAcct, intelCredit.CloudAccountId, 0.19*DefaultCloudCreditAmount)
		}
	}

	cloudCreditUsageScheduler.cloudCreditUsages(ctx)
	VerifyCloudAcctHasLowCredits(t, cloudAcct.Id)
	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for intel customers with usage threshold")
}

func TestIntelCloudCreditUsageComplete(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditUsageComplete")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditUsageScheduler)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

	intel.UpdateIntelCreditUsed(t, ctx, billingAcct, GetBillingCredits(t, ctx, billingAcct))
	cloudCreditUsageScheduler.cloudCreditUsages(ctx)
	VerifyCloudAcctHasNoCredits(t, cloudAcct.Id)
	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for intel customers with usage completely used")
}

func TestIntelCloudCreditsUsageComplete(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditsUsageComplete")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditUsageScheduler)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

	intelUser1 := "intel_" + uuid.NewString() + "@example.com"
	cloudAcct1 := createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, intelUser1, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditUsageScheduler)
	billingAcct1 := &pb.BillingAccount{CloudAccountId: cloudAcct1.Id}

	CreateBillingCredit(t, ctx, cloudAcct, billingAcct)
	CreateBillingCredit(t, ctx, cloudAcct1, billingAcct1)
	cloudCreditUsageScheduler.cloudCreditUsages(ctx)

	VerifyCloudAcctHasCredits(t, cloudAcct.Id)
	VerifyCloudAcctHasCredits(t, cloudAcct1.Id)

	intel.UpdateIntelCreditUsed(t, ctx, billingAcct, GetBillingCredits(t, ctx, billingAcct))
	intel.UpdateIntelCreditUsed(t, ctx, billingAcct1, GetBillingCredits(t, ctx, billingAcct1))
	cloudCreditUsageScheduler.cloudCreditUsages(ctx)

	VerifyCloudAcctHasNoCredits(t, cloudAcct.Id)
	VerifyCloudAcctHasNoCredits(t, cloudAcct1.Id)
	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit usages for intel customers with usage completely used")
}

func TestStandardCloudCreditUsage(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCloudCreditUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")

	standardUser := "standard_" + uuid.NewString() + "@example.com"
	cloudCreditUsageScheduler := getCloudCreditUsageScheduler()

	createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, standardUser, pb.AccountType_ACCOUNT_TYPE_STANDARD, cloudCreditUsageScheduler)

	standard.DeleteStandardCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for standard customers")
}

func TestStandardCloudCreditUsageThreshold(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCloudCreditUsageThreshold")
	logger.Info("BEGIN")
	defer logger.Info("End")

	standardUser := "standard_" + uuid.NewString() + "@example.com"
	cloudCreditUsageScheduler := getCloudCreditUsageScheduler()

	cloudAcct := createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, standardUser, pb.AccountType_ACCOUNT_TYPE_STANDARD, cloudCreditUsageScheduler)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

	standardCredits := GetBillingCredits(t, ctx, billingAcct)
	for _, standardCredit := range standardCredits {
		if standardCredit.CloudAccountId == cloudAcct.Id {
			standard.UpdateStandardCreditWithRemainingAmount(t, ctx, billingAcct, standardCredit.CloudAccountId, 0.19*DefaultCloudCreditAmount)
		}
	}

	cloudCreditUsageScheduler.cloudCreditUsages(ctx)
	VerifyCloudAcctHasLowCredits(t, cloudAcct.Id)
	standard.DeleteStandardCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for standard customers with usage threshold")
}

func TestStandardCloudCreditUsageComplete(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCloudCreditUsageComplete")
	logger.Info("BEGIN")
	defer logger.Info("End")

	standardUser := "standard_" + uuid.NewString() + "@example.com"
	cloudCreditUsageScheduler := getCloudCreditUsageScheduler()

	cloudAcct := createAcctWithCreditsAndVerifyUsageScheduler(t, ctx, standardUser, pb.AccountType_ACCOUNT_TYPE_STANDARD, cloudCreditUsageScheduler)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

	standard.UpdateStandardCreditUsed(t, ctx, billingAcct, GetBillingCredits(t, ctx, billingAcct))
	cloudCreditUsageScheduler.cloudCreditUsages(ctx)
	VerifyCloudAcctHasNoCredits(t, cloudAcct.Id)
	standard.DeleteStandardCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for standard customers with usage completely used")
}

func getCloudCreditUsageScheduler() *CloudCreditUsageScheduler {
	// eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	return NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
}

func createAcctWithCreditsAndVerifyUsageScheduler(t *testing.T, ctx context.Context, user string, acctType pb.AccountType, scheduler *CloudCreditUsageScheduler) *pb.CloudAccount {
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  acctType,
	})
	CreateBillingCredit(t, ctx, cloudAcct, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
	scheduler.cloudCreditUsages(ctx)
	VerifyCloudAcctHasCredits(t, cloudAcct.Id)
	return cloudAcct
}
