// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestGetPaymentMethods(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestUpdateAccountBillingGroup")
	logger.Info("testing add account payment method")

	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)
	clientAcctGroupId := "Chase"
	// 1 for external payment type or credit card
	payMethodType := 1
	creditCardDetails := client.CreditCardDetails{
		CCNumber:      4111111111111111,
		CCExpireMonth: 12,
		CCExpireYear:  2025,
		CCV:           987,
	}
	clientPaymentMethodId := uuid.New().String()

	// Create a new plan to map with the account creation
	_, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(context.Background(), GetTestClientPlanId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}

	// Creation of Aria Account
	_, err = ariaAccountClient.CreateAriaAccount(context.Background(), id, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// Get billing group info/details for an account through client_account_id
	getBillingGroupResp, err := ariaAccountClient.GetBillingGroup(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	clientBillingGroupId := getBillingGroupResp.BillingGroupDetails[0].ClientBillingGroupId

	ariaPaymentClient := common.GetAriaPaymentClient()
	res, err := ariaPaymentClient.AssignCollectionsAccountGroup(context.Background(), id, clientAcctGroupId)
	if err != nil {
		t.Fatalf("Failed to assign collections account group: %v", err)
	}
	_, err = ariaPaymentClient.AddAccountPaymentMethod(context.Background(), id, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}
	getPaymentMethodsResp, err := ariaPaymentClient.GetPaymentMethods(context.Background(), id)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}
	assert.Equal(t, false, client.IsPayloadEmpty(getPaymentMethodsResp))
}

func TestUpdateAccountBillingGroup(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestUpdateAccountBillingGroup")
	logger.Info("testing add account payment method")

	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)
	clientAcctGroupId := "Chase"
	// 1 for external payment type or credit card
	payMethodType := 1
	creditCardDetails := client.CreditCardDetails{
		CCNumber:      4111111111111111,
		CCExpireMonth: 12,
		CCExpireYear:  2025,
		CCV:           987,
	}
	clientPaymentMethodId := uuid.New().String()

	// Create a new plan to map with the account creation
	_, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(context.Background(), GetTestClientPlanId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}

	// Creation of Aria Account
	_, err = ariaAccountClient.CreateAriaAccount(context.Background(), id, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// Get billing group info/details for an account through client_account_id
	getBillingGroupResp, err := ariaAccountClient.GetBillingGroup(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	clientBillingGroupId := getBillingGroupResp.BillingGroupDetails[0].ClientBillingGroupId

	ariaPaymentClient := common.GetAriaPaymentClient()
	// As a part of pre payment processing assigning collection account group
	res, err := ariaPaymentClient.AssignCollectionsAccountGroup(context.Background(), id, clientAcctGroupId)
	if err != nil {
		if res != nil && res.ErrorCode == 12004 {
			logger.Info("account already assigned to this group")
		} else {
			t.Fatalf("Failed to assign collections account group: %v", err)
		}
	}

	// Add test card to Aria system
	_, err = ariaPaymentClient.AddAccountPaymentMethod(context.Background(), id, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}
	getPaymentMethodsResp, err := ariaPaymentClient.GetPaymentMethods(context.Background(), id)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}
	paymentMethodNo := getPaymentMethodsResp.AccountPaymentMethods[0].PaymentMethodNo

	//update primary payment method no in account billing group
	_, err = ariaPaymentClient.UpdateAccountBillingGroup(context.Background(), id, clientBillingGroupId, paymentMethodNo)
	if err != nil {
		t.Fatalf("failed to update account billing group: %v", err)
	}

	getBillingGroupRes, err := ariaAccountClient.GetBillingGroup(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	assert.Equal(t, paymentMethodNo, getBillingGroupRes.BillingGroupDetails[0].PrimaryPaymentMethodNo)

}

func TestUpdateNewCreditCardPayment(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestUpdateNewCreditCardPayment")
	logger.Info("testing update account payment method")

	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)
	clientAcctGroupId := "Chase"
	// 1 for external payment type or credit card
	payMethodType := 1
	creditCardDetails := client.CreditCardDetails{
		CCNumber:      4111111111111111,
		CCExpireMonth: 12,
		CCExpireYear:  2025,
		CCV:           987,
	}
	clientPaymentMethodId := uuid.New().String()

	_, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(context.Background(), GetTestClientPlanId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}

	// Creation of Aria Account
	_, err = ariaAccountClient.CreateAriaAccount(context.Background(), id, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// Get billing group info/details for an account through client_account_id
	getBillingGroupResp, err := ariaAccountClient.GetBillingGroup(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}

	clientBillingGroupId := getBillingGroupResp.BillingGroupDetails[0].ClientBillingGroupId

	ariaPaymentClient := common.GetAriaPaymentClient()
	// As a part of pre payment processing assigning collection account group
	res, err := ariaPaymentClient.AssignCollectionsAccountGroup(context.Background(), id, clientAcctGroupId)
	if err != nil {
		if res != nil && res.ErrorCode == 12004 {
			logger.Info("account already assigned to this group")
		} else {
			t.Fatalf("Failed to assign collections account group: %v", err)
		}
	}

	// Add test card to Aria system
	_, err = ariaPaymentClient.AddAccountPaymentMethod(context.Background(), id, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}
	getPaymentMethodsResp, err := ariaPaymentClient.GetPaymentMethods(context.Background(), id)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}
	paymentMethodNo := getPaymentMethodsResp.AccountPaymentMethods[0].PaymentMethodNo

	//update primary payment method no in account billing group
	_, err = ariaPaymentClient.UpdateAccountBillingGroup(context.Background(), id, clientBillingGroupId, paymentMethodNo)
	if err != nil {
		t.Fatalf("failed to update account billing group: %v", err)
	}

	getBillingGroupRes, err := ariaAccountClient.GetBillingGroup(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	assert.Equal(t, paymentMethodNo, getBillingGroupRes.BillingGroupDetails[0].PrimaryPaymentMethodNo)

	// Adding new CreditCard details for payment
	// MasterCard test card
	newCreditCardDetails := client.CreditCardDetails{
		CCNumber:      5431111111111111,
		CCExpireMonth: 10,
		CCExpireYear:  2026,
		CCV:           989,
	}
	newClientPaymentMethodId := uuid.New().String()
	// Add test card to Aria system
	_, err = ariaPaymentClient.AddAccountPaymentMethod(context.Background(), id, newClientPaymentMethodId, clientBillingGroupId, payMethodType, newCreditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}
	getPaymentMethodsResp, err = ariaPaymentClient.GetPaymentMethods(context.Background(), id)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}

	//Fetching new card's PaymentMethodNo
	newPaymentMethodNo := getPaymentMethodsResp.AccountPaymentMethods[1].PaymentMethodNo

	//Update new card's PaymentMethod as primary payment method no in account billing group
	_, err = ariaPaymentClient.UpdateAccountBillingGroup(context.Background(), id, clientBillingGroupId, newPaymentMethodNo)
	if err != nil {
		t.Fatalf("failed to update account billing group: %v", err)
	}

	getBillingGroupRes, err = ariaAccountClient.GetBillingGroup(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}

	// newPaymentMethodNo (new card's payment method no) is '2',
	// paymentMethodNo (old card's payment method no) is '1'
	//
	// Billing group of an account is updated with new card as primary payment
	// Hence, getBillingGroupRes.BillingGroupDetails[0].PrimaryPaymentMethodNo should be equal to newPaymentMethodNo
	assert.Equal(t, newPaymentMethodNo, getBillingGroupRes.BillingGroupDetails[0].PrimaryPaymentMethodNo)

	// Disables the old payment method (old credit card).
	_, err = ariaPaymentClient.RemovePaymentMethod(context.Background(), id, paymentMethodNo)
	if err != nil {
		t.Fatalf("Failed to remove payment method: %v", err)
	}

}

func TestSetSession(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestSetSession")
	logger.Info("testing set aria session method")

	id := GetClientAccountId()
	_, err := CreateAccountWithDefaultPlan(context.Background(), id)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	ariaPaymentClient := common.GetAriaPaymentClient()
	_, err = ariaPaymentClient.SetSession(context.Background(), id)
	if err != nil {
		t.Fatalf("Failed to set aria session: %v", err)
	}
}

func TestFailedPayment(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestFailedPayment")
	logger.Info("testing payment failed")

	ariaAccountClient := common.GetAriaAccountClient()
	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)

	clientAcctId := GetClientAccountId()

	_, err := CreateAccountWithDefaultPlan(ctx, clientAcctId)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	_, err = ariaPlan.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}
	clientPlanId := GetTestClientPlanId(product.GetId())
	err = AssignPlanToAccount(ctx, clientAcctId, clientPlanId, ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to assign plan to account: %v", err)
	}

	clientAcctGroupId := "Chase"
	// 1 for external payment type or credit card
	payMethodType := 1
	creditCardDetails := client.CreditCardDetails{
		CCNumber:      4111111111111111,
		CCExpireMonth: 12,
		CCExpireYear:  2025,
		CCV:           987,
	}
	clientPaymentMethodId := uuid.New().String()

	// Get billing group info/details for an account through client_account_id
	getBillingGroupResp, err := ariaAccountClient.GetBillingGroup(context.Background(), clientAcctId)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	clientBillingGroupId := getBillingGroupResp.BillingGroupDetails[0].ClientBillingGroupId

	ariaPaymentClient := common.GetAriaPaymentClient()
	// As a part of pre payment processing assigning collection account group
	res, err := ariaPaymentClient.AssignCollectionsAccountGroup(context.Background(), clientAcctId, clientAcctGroupId)
	if err != nil {
		if res != nil && res.ErrorCode == 12004 {
			logger.Info("account already assigned to this group")
		} else {
			t.Fatalf("Failed to assign collections account group: %v", err)
		}
	}

	// Add test card to Aria system
	_, err = ariaPaymentClient.AddAccountPaymentMethod(context.Background(), clientAcctId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}
	getPaymentMethodsResp, err := ariaPaymentClient.GetPaymentMethods(context.Background(), clientAcctId)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}
	paymentMethodNo := getPaymentMethodsResp.AccountPaymentMethods[0].PaymentMethodNo

	//update primary payment method no in account billing group
	_, err = ariaPaymentClient.UpdateAccountBillingGroup(context.Background(), clientAcctId, clientBillingGroupId, paymentMethodNo)
	if err != nil {
		t.Fatalf("failed to update account billing group: %v", err)
	}
	// Error code / premium rate 302.89 / .05 // Invalid Expiration Date(605.78)
	failedWithInvalidCardNo := 6057.8
	err = CreateTestUsageRecord(ctx, clientAcctId, clientPlanId, usageType.UsageTypeCode, failedWithInvalidCardNo)
	if err != nil {
		t.Fatalf("failed to create usage record: %v", err)
	}

	cloudAccountId := GetCloudAcctIdFromClientAcctId(clientAcctId)
	invoiceId := ManageInvoice(ctx, t, cloudAccountId)
	assert.NotEqual(t, 0, invoiceId)
}
