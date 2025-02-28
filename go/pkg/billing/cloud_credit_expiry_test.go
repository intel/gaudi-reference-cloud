// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"time"

	"context"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	intel "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	standard "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_standard"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	"github.com/google/uuid"
)

func TestPremiumCloudCreditExpiry(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPremiumCloudCreditExpiry")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditExpiryScheduler := NewCloudCreditExpiryScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	createAcctWithCreditsAndVerifyExpiryScheduler(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM, cloudCreditExpiryScheduler)

	logger.Info("and done with testing cloud credit expiry for premium customers")
}

func TestIntelCloudCreditExpiry(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditExpiry")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditExpiryScheduler := NewCloudCreditExpiryScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	createAcctWithCreditsAndVerifyExpiryScheduler(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditExpiryScheduler)

	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit expiry for intel customers")
}

func TestIntelCloudCreditExpired(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditExpired")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  intelUser,
		Owner: intelUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_INTEL,
	})
	expiredDate := time.Now().AddDate(0, 0, -10)
	expirationDate := timestamppb.New(expiredDate)
	billingCredit := &pb.BillingCredit{
		CloudAccountId:  cloudAcct.Id,
		Created:         timestamppb.New(time.Now()),
		OriginalAmount:  DefaultCloudCreditAmount,
		RemainingAmount: DefaultCloudCreditAmount,
		Reason:          DefaultCreditReason,
		CouponCode:      DefaultCreditCoupon,
		Expiration:      expirationDate}
	//Create
	_, err := intelDriver.billingCredit.Create(context.Background(), billingCredit)
	if err != nil {
		t.Fatalf("failed to create intel credits: %v", err)
	}
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditExpiryScheduler := NewCloudCreditExpiryScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	cloudCreditExpiryScheduler.cloudCreditExpiry(ctx)

	VerifyCloudAcctHasNoCredits(t, cloudAcct.Id)
	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit expired for intel customers")
}

func TestStandardCloudCreditExpiry(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCloudCreditExpiry")
	logger.Info("BEGIN")
	defer logger.Info("End")

	standardUser := "standard_" + uuid.NewString() + "@example.com"
	cloudCreditExpiryScheduler := getCloudCreditExpiryScheduler()

	createAcctWithCreditsAndVerifyExpiryScheduler(t, ctx, standardUser, pb.AccountType_ACCOUNT_TYPE_STANDARD, cloudCreditExpiryScheduler)

	standard.DeleteStandardCredits(t, ctx)
	logger.Info("and done with testing cloud credit expiry for standard customers")
}

func TestStandardCloudCreditExpired(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCloudCreditExpired")
	logger.Info("BEGIN")
	defer logger.Info("End")

	standardUser := "standard_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  standardUser,
		Owner: standardUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	})
	expiredDate := time.Now().AddDate(0, 0, -10)
	expirationDate := timestamppb.New(expiredDate)
	billingCredit := &pb.BillingCredit{
		CloudAccountId:  cloudAcct.Id,
		Created:         timestamppb.New(time.Now()),
		OriginalAmount:  DefaultCloudCreditAmount,
		RemainingAmount: DefaultCloudCreditAmount,
		Reason:          DefaultCreditReason,
		CouponCode:      DefaultCreditCoupon,
		Expiration:      expirationDate}
	//Create
	_, err := standardDriver.billingCredit.Create(context.Background(), billingCredit)
	if err != nil {
		t.Fatalf("failed to create standard credits: %v", err)
	}
	cloudCreditExpiryScheduler := getCloudCreditExpiryScheduler()
	cloudCreditExpiryScheduler.cloudCreditExpiry(ctx)

	VerifyCloudAcctHasNoCredits(t, cloudAcct.Id)
	standard.DeleteStandardCredits(t, ctx)
	logger.Info("and done with testing cloud credit expired for standard customers")
}

func getCloudCreditExpiryScheduler() *CloudCreditExpiryScheduler {
	// eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	return NewCloudCreditExpiryScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
}

func createAcctWithCreditsAndVerifyExpiryScheduler(t *testing.T, ctx context.Context, user string, acctType pb.AccountType, scheduler *CloudCreditExpiryScheduler) *pb.CloudAccount {
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  acctType,
	})
	CreateBillingCredit(t, ctx, cloudAcct, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
	scheduler.cloudCreditExpiry(ctx)
	VerifyCloudAcctHasCreditsForExpiry(t, cloudAcct.Id)
	return cloudAcct
}
