// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"time"

	"context"
	"testing"

	"github.com/google/uuid"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	ariaConfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	intel "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	standard "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_standard"
	credits "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPremiumCloudCreditExpiryEvent(t *testing.T) {
	if ariaConfig.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPremiumCloudCreditExpiry")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditExpiryEventScheduler := credits.NewCloudCreditExpiryEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	createAcctWithCreditsExpiryEventScheduler(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM, cloudCreditExpiryEventScheduler)

	logger.Info("and done with testing cloud credit expiry for premium customers")
}

func TestIntelCloudCreditExpiryEvent(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditExpiry")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditExpiryEventScheduler := credits.NewCloudCreditExpiryEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	createAcctWithCreditsExpiryEventScheduler(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL, cloudCreditExpiryEventScheduler)
	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit expiry for intel customers")
}

func TestIntelCloudCreditExpiredEvent(t *testing.T) {
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
	BillingCredit := &pb.BillingCredit{
		CloudAccountId:  cloudAcct.Id,
		Created:         timestamppb.New(time.Now()),
		OriginalAmount:  DefaultCloudCreditAmount,
		RemainingAmount: DefaultCloudCreditAmount,
		Reason:          DefaultCreditReason,
		CouponCode:      DefaultCreditCoupon,
		Expiration:      expirationDate}
	//Create
	driver, err := billingCommon.GetDriver(ctx, cloudAcct.Id)
	if err != nil {
		t.Fatalf("failed to get driver: %v", err)
	}
	_, err = driver.BillingCredit.Create(context.Background(), BillingCredit)
	if err != nil {
		t.Fatalf("failed to create intel credits: %v", err)
	}

	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditExpiryEventScheduler := credits.NewCloudCreditExpiryEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	cloudCreditExpiryEventScheduler.PublishCloudCreditExpiryEvent(ctx)
	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit expired for intel customers")
}

func TestStandardCloudCreditExpiryEvent(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCloudCreditExpiry")
	logger.Info("BEGIN")
	defer logger.Info("End")

	standardUser := "standard_" + uuid.NewString() + "@example.com"
	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditExpiryEventScheduler := credits.NewCloudCreditExpiryEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	createAcctWithCreditsExpiryEventScheduler(t, ctx, standardUser, pb.AccountType_ACCOUNT_TYPE_STANDARD, cloudCreditExpiryEventScheduler)

	standard.DeleteStandardCredits(t, ctx)
	logger.Info("and done with testing cloud credit expiry for standard customers")
}

func TestStandardCloudCreditExpiredEvent(t *testing.T) {
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
	BillingCredit := &pb.BillingCredit{
		CloudAccountId:  cloudAcct.Id,
		Created:         timestamppb.New(time.Now()),
		OriginalAmount:  DefaultCloudCreditAmount,
		RemainingAmount: DefaultCloudCreditAmount,
		Reason:          DefaultCreditReason,
		CouponCode:      DefaultCreditCoupon,
		Expiration:      expirationDate}
	//Create
	driver, err := billingCommon.GetDriver(ctx, cloudAcct.Id)
	if err != nil {
		t.Fatalf("failed to get driver: %v", err)
	}
	_, err = driver.BillingCredit.Create(context.Background(), BillingCredit)
	if err != nil {
		t.Fatalf("failed to create standard credits: %v", err)
	}
	notificationClient, _ := billingCommon.NewNotificationGatewayTestClient(ctx)
	cloudCreditExpiryEventScheduler := credits.NewCloudCreditExpiryEventScheduler(notificationClient, TestSchedulerCloudAccountState, TestCloudAccountSvcClient)
	cloudCreditExpiryEventScheduler.PublishCloudCreditExpiryEvent(ctx)
	standard.DeleteStandardCredits(t, ctx)
	logger.Info("and done with testing cloud credit expired for standard customers")
}

func createAcctWithCreditsExpiryEventScheduler(t *testing.T, ctx context.Context, user string, acctType pb.AccountType, scheduler *credits.CloudCreditExpiryEventScheduler) *pb.CloudAccount {
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  acctType,
	})
	CreateBillingCredit(t, ctx, cloudAcct, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
	scheduler.PublishCloudCreditExpiryEvent(ctx)
	return cloudAcct
}
