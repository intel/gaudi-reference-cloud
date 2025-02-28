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
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestEnterpriseAcctLinked(t *testing.T) {
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestEnterpriseAcctLinked")
	user := "enterprise_" + uuid.NewString() + "example.com"
	cloudAcctId, err := AriaService.cloudAccountClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  user,
			Owner: user,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING,
		})
	if err != nil {
		t.Fatalf("failed to create cloud account: %v", err)
	}

	ariaController := NewAriaController(clientTestCommon.GetAriaClient(), clientTestCommon.GetAriaAdminClient(), clientTestCommon.GetAriaCredentials())
	err = ariaController.InitAria(ctx)
	if err != nil {
		t.Fatalf("failed to init Aria: %v", err)
	}

	_, err = clientTestCommon.GetAriaAccountClient().CreateAriaAccount(ctx, client.GetAccountClientId(cloudAcctId.Id), client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_ENTERPRISE_PENDING)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	enterpriseAcctLinkScheduler := NewEnterpriseAcctLinkScheduler(ctx, eventManager, AriaService.cloudAccountClient, clientTestCommon.GetAriaAccountClient(), clientTestCommon.GetServiceCreditClient())
	err = enterpriseAcctLinkScheduler.handleLinkedEnterpriseAcct(ctx)

	if err != nil {
		t.Fatalf("failed to handle enterprise acct linked: %v", err)
	}

	verifyCloudAcctType(t, ctx, cloudAcctId, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING)

	parentAccountId := uuid.NewString()
	_, err = clientTestCommon.GetAriaAccountClient().CreateAriaAccount(ctx, client.GetAccountClientId(parentAccountId), client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_ENTERPRISE_PENDING)
	if err != nil {
		t.Fatalf("failed to create parent account: %v", err)
	}

	getChildAcctNoResponse, err := clientTestCommon.GetAriaAccountClient().GetAccountNoFromUserId(ctx, client.GetAccountClientId(cloudAcctId.Id))
	if err != nil {
		t.Fatalf("failed to get account number for child account: %v", err)
	}

	getParentAcctNoResponse, err := clientTestCommon.GetAriaAccountClient().GetAccountNoFromUserId(ctx, client.GetAccountClientId(parentAccountId))
	if err != nil {
		t.Fatalf("failed to get account number for parent account: %v", err)
	}

	_, err = clientTestCommon.GetAriaAccountClient().UpdateAriaAccount(ctx, getChildAcctNoResponse.AcctNo, getParentAcctNoResponse.AcctNo)
	if err != nil {
		t.Fatalf("failed to update child account to parent: %v", err)
	}

	err = enterpriseAcctLinkScheduler.handleLinkedEnterpriseAcct(ctx)
	if err != nil {
		t.Fatalf("failed to handle enterprise acct linked: %v", err)
	}

	verifyCloudAcctType(t, ctx, cloudAcctId, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE)
	verifyNoDefaultCloudCreditsAssignedForEnterprise(t, ctx, ariaController, cloudAcctId.Id)

	logger.Info("and done with testing syncing to plan activation status")
}

func verifyCloudAcctType(t *testing.T, ctx context.Context, cloudAcctId *pb.CloudAccountId, accountType pb.AccountType) {
	cloudAcct, err := AriaService.cloudAccountClient.GetById(ctx, cloudAcctId)
	if err != nil {
		t.Fatalf("failed to get cloud account: %v", err)
	}
	if cloudAcct.Type != accountType {
		t.Fatal("cloud account should still have been enterprise pending")
	}
}

func verifyNoDefaultCloudCreditsAssignedForEnterprise(t *testing.T, ctx context.Context, ariaController *AriaController, cloudAccountId string) {
	acctCredits, err := ariaController.GetAccountCredits(ctx, cloudAccountId)
	if err != nil {
		t.Fatalf("failed to get account credits: %v", err)
	}
	for _, acctCredit := range acctCredits {
		if acctCredit.ReasonCode == kFixMeWeHaventImplementedReasonCodeYet {
			t.Fatalf("found default cloud credits assigned")
		}
	}
}

// disabling default credits #TWC4726-1391
func verifyDefaultCloudCreditsAssignedForEnterprise(t *testing.T, ctx context.Context, ariaController *AriaController, cloudAccountId string) {
	acctCredits, err := ariaController.GetAccountCredits(ctx, cloudAccountId)
	if err != nil {
		t.Fatalf("failed to get account credits: %v", err)
	}
	for _, acctCredit := range acctCredits {
		if acctCredit.ReasonCode == kFixMeWeHaventImplementedReasonCodeYet {
			if config.Cfg.EntDefaultCreditAmount != float64(acctCredit.Amount) {
				t.Fatalf("incorrect credit amount: %v", err)
			}
			return
		}
	}
	t.Fatalf("did not find default cloud credits assigned")
}
