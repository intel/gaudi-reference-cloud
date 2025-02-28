// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	clientTests "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests"
	clientTestCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"testing"
)

func initAria(t *testing.T, ctx context.Context) {
	ariaController := NewAriaController(clientTestCommon.GetAriaClient(), clientTestCommon.GetAriaAdminClient(), clientTestCommon.GetAriaCredentials())
	err := ariaController.InitAria(ctx)
	if err != nil {
		t.Fatalf("failed to init Aria: %v", err)
	}
}

func CreateCloudAccount(t *testing.T, ctx context.Context, user string, acctType pb.AccountType) *pb.CloudAccount {
	acct := &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  acctType,
	}

	id, err := AriaService.cloudAccountClient.Create(ctx, acct)
	if err != nil {
		t.Fatalf("failed to create cloud account: %v", err)
	}

	acctOut, err := AriaService.cloudAccountClient.GetById(context.Background(), &pb.CloudAccountId{Id: id.GetId()})
	if err != nil {
		t.Fatalf("failed to read cloud account: %v", err)
	}

	return acctOut
}

func CreateBillingAccount(t *testing.T, ctx context.Context, user string, acctType pb.AccountType) *pb.BillingAccount {
	initAria(t, ctx)
	cloudAcct := CreateCloudAccount(t, ctx, user, acctType)
	_, err := clientTestCommon.GetAriaAccountClient().CreateAriaAccount(ctx, client.GetAccountClientId(cloudAcct.Id), client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	return &pb.BillingAccount{CloudAccountId: cloudAcct.Id}
}

func GetCloudAccounts(t *testing.T, ctx context.Context) []*pb.CloudAccount {
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	entUser := "enterprise_" + uuid.NewString() + "@example.com"
	premiumCloudAcct := CreateCloudAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	entCloudAcct := CreateCloudAccount(t, ctx, entUser, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE)
	return []*pb.CloudAccount{premiumCloudAcct, entCloudAcct}
}

func GetBillingAccounts(t *testing.T, ctx context.Context) []*pb.BillingAccount {
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	entUser := "enterprise_" + uuid.NewString() + "@example.com"
	premiumBillingAcct := CreateBillingAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	entBillingAcct := CreateBillingAccount(t, ctx, entUser, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE)
	return []*pb.BillingAccount{premiumBillingAcct, entBillingAcct}
}

func AddTestPaymentMethod(t *testing.T, ctx context.Context, cloudAcctId string) {
	acctDetails, err := clientTestCommon.GetAriaAccountClient().GetAriaAccountDetailsAllForClientId(ctx, client.GetAccountClientId(cloudAcctId))
	if err != nil {
		t.Fatalf("failed to get account details: %v", err)
	}
	payMethodType := 1
	clientBillingGroupId := acctDetails.MasterPlansInfo[0].ClientBillingGroupId
	creditCardDetails := client.CreditCardDetails{
		CCNumber:      4111111111111111,
		CCExpireMonth: 12,
		CCExpireYear:  2025,
		CCV:           987,
	}
	clientPaymentMethodId := uuid.New().String()
	_, err = clientTestCommon.GetAriaPaymentClient().AddAccountPaymentMethod(ctx, client.GetAccountClientId(cloudAcctId), clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("failed to add test payment method: %v", err)
	}
}

func PostUsageForAccount(t *testing.T, ctx context.Context, billingAcct *pb.BillingAccount, amount float64) {
	initAria(t, ctx)
	product := clientTests.GetProduct()
	productFamily := clientTests.GetProductFamily()
	usageType, err := clientTestCommon.GetAriaUsageTypeClient().GetMinutesUsageType(context.Background())
	if err != nil {
		t.Fatalf("failed to get usage type for posting usage: %v", err)
	}
	clientAcctId := client.GetAccountClientId(billingAcct.CloudAccountId)
	_, err = clientTestCommon.GetAriaPlanClient().CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("failed to create plan for posting usage: %v", err)
	}
	clientPlanId := client.GetPlanClientId(product.GetId())
	clientMasterPlanInstanceId := client.GetClientMasterPlanInstanceId(billingAcct.CloudAccountId, product.Id)
	clientRateScheduleId := client.GetRateScheduleClientId(product.Id, "premium")
	_, err = clientTestCommon.GetAriaAccountClient().AssignPlanToAccountWithBillingAndDunningGroup(ctx, clientAcctId, clientPlanId, clientMasterPlanInstanceId, clientRateScheduleId, client.BILL_LAG_DAYS, 0)
	if err != nil {
		t.Fatalf("failed to assign plan to account for posting usage: %v", err)
	}
	usages := []*client.BillingUsage{
		{
			CloudAccountId: clientAcctId,
			ProductId:      product.Id,
			TransactionId:  uuid.New().String(),
			ResourceId:     uuid.New().String(),
			Amount:         amount,
			// todo: this method is always using the same record id.
			RecordId:      "1",
			UsageUnitType: "RATE_UNIT_DOLLARS_PER_MINUTE",
		},
	}
	usageResp, err := clientTestCommon.GetUsageClient().CreateBulkUsageRecords(ctx, usages)
	if client.IsPayloadEmpty(usageResp) {
		t.Fatalf("failed to create usage record")
	}
	if err != nil {
		t.Fatalf("failed to create test usage record for posting usage: %v", err)
	}
}
