// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"time"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	"github.com/google/uuid"
)

func TestPremiumServicesTermination(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPremiumServicesTermination")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	createAccForServicesTermination(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)

	logger.Info("and done with testing services termination for premium customers")
}

func TestIntelServicesTermination(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelServicesTermination")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	createAccForServicesTermination(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL)

	logger.Info("and done with testing services termination for intel customers")
}

func TestIntelServicesTerminated(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelServicesTerminated")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	cloudAcct := createAccForServicesTermination(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL)

	creditsDepletionDate := time.Now().AddDate(0, 0, -2)
	creditsDepletedDate := timestamppb.New(creditsDepletionDate)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	meteringSvcClient := pb.NewMeteringServiceClient(meteringClientConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	meteringClient := billingCommon.NewMeteringClientForTest(meteringSvcClient)
	_, err := cloudAcctClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, CreditsDepleted: creditsDepletedDate})
	if err != nil {
		t.Fatalf("failed to update acct: %v", err)
	}
	servicesTerminationScheduler := NewServicesTerminationScheduler(testSchedulerCloudAccountState, cloudAccountSvcClient, meteringClient)
	servicesTerminationScheduler.servicesTermination(ctx)

	VerifyCloudAcctServicesTermination(t, cloudAcct.Id)
	logger.Info("and done with testing services termination for intel customers")
}

func createAccForServicesTermination(t *testing.T, ctx context.Context, user string, acctType pb.AccountType) *pb.CloudAccount {
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  acctType,
	})
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	meteringSvcClient := pb.NewMeteringServiceClient(meteringClientConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	meteringClient := billingCommon.NewMeteringClientForTest(meteringSvcClient)
	servicesTerminationScheduler := NewServicesTerminationScheduler(testSchedulerCloudAccountState, cloudAccountSvcClient, meteringClient)
	servicesTerminationScheduler.servicesTermination(ctx)
	VerifyCloudAcctServicesNoTermination(t, cloudAcct.Id)
	return cloudAcct
}

func TestMultiAccountServicesTermination(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestMultiAccountPremiumServicesTermination")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	createAccForServicesTermination(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	cloudAcct := createAccForServicesTermination(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL)

	creditsDepletionDate := time.Now().AddDate(0, 0, -2)
	creditsDepletedDate := timestamppb.New(creditsDepletionDate)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	meteringSvcClient := pb.NewMeteringServiceClient(meteringClientConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	meteringClient := billingCommon.NewMeteringClientForTest(meteringSvcClient)
	_, err := cloudAcctClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, CreditsDepleted: creditsDepletedDate})
	if err != nil {
		t.Fatalf("failed to update acct: %v", err)
	}
	servicesTerminationScheduler := NewServicesTerminationScheduler(testSchedulerCloudAccountState, cloudAccountSvcClient, meteringClient)
	servicesTerminationScheduler.servicesTermination(ctx)
	VerifyCloudAcctServicesTermination(t, cloudAcct.Id)

}
