// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	"github.com/stretchr/testify/assert"
)

func TestCreateAriaAccount(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCreateAriaAccount")
	logger.Info("testing create aria account")
	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	clientPlanId := client.GetDefaultPlanClientId()
	ariaPlanClient := common.GetAriaPlanClient()
	err := ariaPlanClient.CreateDefaultPlan(context.Background())
	if err != nil && !strings.Contains(err.Error(), "error code:1001") {
		t.Fatalf("failed to create default plan: %v", err)
	}
	resp, err := ariaAccountClient.CreateAriaAccount(context.Background(), id, clientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(resp))

	// Aria should prevent us from creating duplicate accounts
	resp, err = ariaAccountClient.CreateAriaAccount(context.Background(), id, client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_PREMIUM)
	if err == nil {
		t.Fatalf("created a duplicate account!")
	}
	const DUP_CLIENT_ACCT_ID = 5036
	if resp.ErrorCode != DUP_CLIENT_ACCT_ID {
		t.Errorf("account creation failed with the wrong error: %v", err)
	}

	details, err := ariaAccountClient.GetAriaAccountDetailsAllForClientId(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get aria account details: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(details))
	respSetGrpId, err := ariaAccountClient.SetAccountNotifyTemplateGroup(context.Background(), id, client.GetDefaultNotificationTemplateGroupId())
	if err != nil {
		t.Fatalf("failed to set notify template group id: %v", err)
	}

	assert.Equal(t, int64(ERROR_CODE_OK), respSetGrpId.GetErrorCode())
	assert.Equal(t, ERROR_MESSAGE_OK, respSetGrpId.GetErrorMsg())
	respGetGrpId, err := ariaAccountClient.GetAccountNotificationDetails(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get notify template group id: %v", err)
	}

	assert.Equal(t, int64(ERROR_CODE_OK), respGetGrpId.GetErrorCode())
	assert.Equal(t, ERROR_MESSAGE_OK, respGetGrpId.GetErrorMsg())
	assert.Equal(t, true, client.ContainsNotificationTemplateGroupId(respGetGrpId.AccountNotificationDetails, client.GetDefaultNotificationTemplateGroupId()))
}

func TestGetAcctNoFromUserId(t *testing.T) {
	clientAccountId := GetClientAccountId()
	clientPlanId := client.GetDefaultPlanClientId()

	ariaPlanClient := common.GetAriaPlanClient()
	logger := log.FromContext(context.Background()).WithName("TestGetAcctNoFromUserId")
	logger.Info("testing get account no from user id aria api")
	err := ariaPlanClient.CreateDefaultPlan(context.Background())
	if err != nil && !strings.Contains(err.Error(), "error code:1001") {
		t.Fatalf("failed to create default plan: %v", err)
	}
	ariaAccountClient := common.GetAriaAccountClient()
	_, err = ariaAccountClient.CreateAriaAccount(context.Background(), clientAccountId, clientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	_, err = ariaAccountClient.GetAccountNoFromUserId(context.Background(), clientAccountId)
	if err != nil {
		t.Fatalf("failed to get account no from user id: %v", err)
	}
}

func TestCreateBillingGroup(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCreateBillingGroup")
	logger.Info("testing create billing group")
	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)
	clientBillingGroupId := client.GetBillingGroupId(id)

	// Create a new plan to map with the account creation
	createRespBody, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	assert.Equal(t, false, client.IsPayloadEmpty(createRespBody))

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(context.Background(), GetTestClientPlanId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(resp))

	// Creation of Aria Account
	ariaAccResp, err := ariaAccountClient.CreateAriaAccount(context.Background(), id, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(ariaAccResp))

	// Appending clientBillingGroupId with "_1", as billing group id is already created during the account creation.
	clientBillingGroupId = clientBillingGroupId + "_1"

	// Creation of the billing group for the account created above
	billingGroupResp, err := ariaAccountClient.CreateBillingGroup(context.Background(), id, clientBillingGroupId)
	if err != nil {
		t.Fatalf("failed to create billing group: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(billingGroupResp))

}

func TestCreateDunningGroup(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCreateDunningGroup")
	logger.Info("testing create dunning group")
	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)
	clientDunningGroupId := client.GetDunningGroupId(id)

	// Create a new plan to map with the account creation
	createRespBody, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	assert.Equal(t, false, client.IsPayloadEmpty(createRespBody))

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(context.Background(), GetTestClientPlanId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(resp))

	// Creation of Aria Account
	ariaAccResp, err := ariaAccountClient.CreateAriaAccount(context.Background(), id, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(ariaAccResp))

	// Appending clientDunningGroupId with "_1", as dunning group id is already created during the account creation.
	clientDunningGroupId = clientDunningGroupId + "_1"

	// Creation of the dunning group for the account created above
	dunningGroupResp, err := ariaAccountClient.CreateDunningGroup(context.Background(), id, clientDunningGroupId)
	if err != nil {
		t.Fatalf("failed to create dunning group: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(dunningGroupResp))

}

func TestGetAccountCredits(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetAccountCredits")
	logger.Info("testing get account credits")
	clientAccountId := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	ariaPlanClient := common.GetAriaPlanClient()
	err := ariaPlanClient.CreateDefaultPlan(context.Background())
	if err != nil && !strings.Contains(err.Error(), "error code:1001") {
		t.Fatalf("failed to create default plan: %v", err)
	}
	acctResp, err := ariaAccountClient.CreateAriaAccount(context.Background(), clientAccountId, client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	serviceCreditClient := common.GetServiceCreditClient()
	code := int64(1)
	currentDate := time.Now()
	newDate := currentDate.AddDate(0, 0, 100)
	expirationDate := fmt.Sprintf("%04d-%02d-%02d", newDate.Year(), newDate.Month(), newDate.Day())

	_, err = serviceCreditClient.CreateServiceCredits(context.Background(), clientAccountId, DefaultCloudCreditAmount, code, expirationDate, "testCredit")
	if err != nil {
		t.Fatalf("failed to create service credits: %v", err)
	}
	getAccountCreditResp, err := ariaAccountClient.GetAccountCredits(context.Background(), acctResp.OutAcct[0].AcctNo)
	if err != nil {
		t.Fatalf("failed to get account credits: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(getAccountCreditResp))
	assert.Equal(t, getAccountCreditResp.AllCredits[0].Amount, float32(DefaultCloudCreditAmount))
	getCreditDetailsResp, err := ariaAccountClient.GetAccountCreditDetails(context.Background(), acctResp.OutAcct[0].ClientAcctId, getAccountCreditResp.AllCredits[0].CreditNo)
	if err != nil {
		t.Fatalf("failed to get credit details: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(getCreditDetailsResp))
}

func TestGetBillingGroup(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCreateBillingGroup")
	logger.Info("testing create billing group")
	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)

	// Create a new plan to map with the account creation
	createRespBody, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	assert.Equal(t, false, client.IsPayloadEmpty(createRespBody))

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(context.Background(), GetTestClientPlanId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(resp))

	// Creation of Aria Account
	ariaAccResp, err := ariaAccountClient.CreateAriaAccount(context.Background(), id, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(ariaAccResp))

	// Get billing group info/details for an account through client_account_id
	getBillingGroupResp, err := ariaAccountClient.GetBillingGroup(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(getBillingGroupResp))

}

func TestGetDunningGroup(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetDunningGroup")
	logger.Info("testing get dunning group")
	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)

	// Create a new plan to map with the account creation
	createRespBody, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	assert.Equal(t, false, client.IsPayloadEmpty(createRespBody))

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(context.Background(), GetTestClientPlanId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(resp))

	// Creation of Aria Account
	ariaAccResp, err := ariaAccountClient.CreateAriaAccount(context.Background(), id, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(ariaAccResp))

	// Get dunning group info/details for an account through client_account_id
	getDunningGroupResp, err := ariaAccountClient.GetDunningGroup(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get dunning group: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(getDunningGroupResp))

}

func TestGetEnterpriseParentAccountPlan(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestGetEnterpriseParentAccountPlan")
	logger.Info("testing get enterprise parent account plan")
	parentClientAccountId := GetClientAccountId()
	childClientAccountId := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	ariaPlanClient := common.GetAriaPlanClient()
	err := ariaPlanClient.CreateDefaultPlan(context.Background())
	if err != nil && !strings.Contains(err.Error(), "error code:1001") {
		t.Fatalf("failed to create default plan: %v", err)
	}
	parentAcctResp, err := ariaAccountClient.CreateAriaAccount(ctx, parentClientAccountId, client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_ENTERPRISE)
	if err != nil {
		t.Fatalf("failed to create parent account: %v", err)
	}
	childAcctResp, err := ariaAccountClient.CreateAriaAccount(ctx, childClientAccountId, client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_ENTERPRISE)
	if err != nil {
		t.Fatalf("failed to create child account: %v", err)
	}
	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)
	_, err = ariaPlan.CreatePlan(context.Background(), product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	childAcctNo := childAcctResp.OutAcct[0].AcctNo
	parentAcctNo := parentAcctResp.OutAcct[0].AcctNo
	_, err = ariaAccountClient.UpdateAriaAccount(ctx, childAcctNo, parentAcctNo)
	if err != nil {
		t.Fatalf("failed to link account: %v", err)
	}

	parentAcctDetails, err := ariaAccountClient.GetAriaAccountDetailsAllForClientId(ctx, parentAcctResp.OutAcct[0].ClientAcctId)
	if err != nil {
		t.Fatalf("Failed to get parent account: %v", err)
	}

	clientPlanId := client.GetPlanClientId(product.Id)
	lastBillThruDate := parentAcctDetails.MasterPlansInfo[0].LastBillThruDate
	parentClientMasterPlanInstanceId := parentAcctDetails.MasterPlansInfo[0].ClientMasterPlanInstanceId
	err = AssignPlanToEnterpriseAccount(ctx, childClientAccountId, clientPlanId, lastBillThruDate, parentClientMasterPlanInstanceId)
	if err != nil {
		t.Fatalf("failed to assign account plan: %v", err)
	}
	masterPlanInfo, err := ariaAccountClient.GetEnterpriseParentAccountPlan(ctx, childClientAccountId)
	if err != nil {
		t.Fatalf("failed to get enterprise account plan: %v", err)
	}
	logger.Info("parent master plan", "masterPlanInfo", masterPlanInfo)
	assert.NotEqual(t, 0, len(masterPlanInfo))
}

func TestUpdateAccountContact(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("UpdateAccountContact")
	logger.Info("BEGIN")
	defer logger.Info("END")
	clientAccountId := GetClientAccountId()
	_, _, err := CreateAcctWithPlanDefaultProd(ctx, clientAccountId)
	if err != nil {
		t.Fatalf("failed to create account with default prod")
	}
	// Update the dunning for the account
	ariaAccountClient := common.GetAriaAccountClient()
	updateAccountContactResp, err := ariaAccountClient.UpdateAccountContact(ctx, clientAccountId, "testpremium@proton.me")
	logger.Info("updated account contact", "updateAccountContactResp", updateAccountContactResp)
	if err != nil {
		t.Fatalf("failed to update dunning group: %v", err)
	}
}
