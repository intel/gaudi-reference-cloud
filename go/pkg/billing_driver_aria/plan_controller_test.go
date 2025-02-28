// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	clientTestCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestSyncPlanActivationToCloudAccount(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestSyncPlanActivationToCloudAccount")
	// Do the pre-requisites of the testing
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	premiumCloudAcctId := createDelinquentAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	enterpriseUser := "enterprise_" + uuid.NewString() + "@example.com"
	enterpriseCloudAcctId := createDelinquentAccount(t, ctx, enterpriseUser, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE)

	ariaController := NewAriaController(clientTestCommon.GetAriaClient(), clientTestCommon.GetAriaAdminClient(), clientTestCommon.GetAriaCredentials())
	err := ariaController.InitAria(ctx)
	if err != nil {
		t.Fatalf("failed to init Aria: %v", err)
	}

	planController := NewPlanController(AriaService.cloudAccountClient, clientTestCommon.GetAriaAccountClient())
	planController.syncPlanActivationToCloudAccount(context.Background())

	verifySyncActivatedAccount(t, ctx, premiumCloudAcctId, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	verifySyncActivatedAccount(t, ctx, enterpriseCloudAcctId, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE)
	logger.Info("and done with testing syncing to plan activation status")
}

func createDelinquentAccount(t *testing.T, ctx context.Context, user string, acctType pb.AccountType) *pb.CloudAccountId {
	billingAcctCreated := true
	cloudAccountDelinquent := true
	terminatePaidServices := true
	cloudAcctId, err := AriaService.cloudAccountClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:                  user,
			Owner:                 user,
			Tid:                   uuid.NewString(),
			Oid:                   uuid.NewString(),
			Type:                  acctType,
			BillingAccountCreated: &billingAcctCreated,
			Delinquent:            &cloudAccountDelinquent,
			TerminatePaidServices: &terminatePaidServices,
		})
	if err != nil {
		t.Fatalf("failed to create premium cloud account: %v for acct type: %s", err, acctType.String())
	}
	_, err = clientTestCommon.GetAriaAccountClient().CreateAriaAccount(ctx, client.GetAccountClientId(cloudAcctId.Id), client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v for acct type: %s", err, acctType.String())
	}
	return cloudAcctId
}

func verifySyncActivatedAccount(t *testing.T, ctx context.Context, cloudAccountId *pb.CloudAccountId, acctType pb.AccountType) {
	cloudAcct, err := AriaService.cloudAccountClient.GetById(ctx, cloudAccountId)
	if err != nil {
		t.Fatalf("failed to get cloud account: %v for acct type: %s", err, acctType.String())
	}
	// check if cloud account is updated
	if cloudAcct.Delinquent {
		t.Fatal("delinquent should be false for acct type: ", acctType.String())
	}
	if cloudAcct.TerminatePaidServices {
		t.Fatal("terminate paid services should be false for acct type: ", acctType.String())
	}
}
