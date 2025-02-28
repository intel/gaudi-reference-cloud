// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests"
	clientTestsCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"gotest.tools/assert"
)

func TestGetMinutesUsageTypeDetails(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestGetMinutesUsageTypeDetails")
	logger.Info("BEGIN")
	defer logger.Info("End")

	ariaClient := clientTestsCommon.GetAriaClient()
	ariaAdminClient := clientTestsCommon.GetAriaAdminClient()
	ariaCredentials := clientTestsCommon.GetAriaCredentials()
	ariaController := NewAriaController(ariaClient, ariaAdminClient, ariaCredentials)
	usageTypeDetails, err := ariaController.getMinutesUsageTypeDetails(ctx)
	if err != nil {
		logger.Error(err, "got error for getting the minutes usage type")
	}
	if usageTypeDetails != nil {
		logger.Info("got usage unit type of the type mins", "code", usageTypeDetails.UsageTypeCode)
	}
	err = ariaController.ensureMinutesUsageType(ctx)
	if err != nil {
		t.Fatalf("failed to ensure minutes usage type: %v", err)
	}
	usageTypeDetailsAfterEnsuring, err := ariaController.getMinutesUsageTypeDetails(ctx)
	if err != nil {
		t.Fatalf("failed to ensure minutes usage type: %v", err)
	}
	if usageTypeDetailsAfterEnsuring == nil {
		t.Fatalf("failed to ensure minutes usage type: %v", err)
	}
}

func TestGetStorageUsageTypeDetails(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestGetStorageUsageTypeDetails")
	logger.Info("BEGIN")
	defer logger.Info("End")

	ariaClient := clientTestsCommon.GetAriaClient()
	ariaAdminClient := clientTestsCommon.GetAriaAdminClient()
	ariaCredentials := clientTestsCommon.GetAriaCredentials()
	ariaController := NewAriaController(ariaClient, ariaAdminClient, ariaCredentials)
	usageTypeDetails, err := ariaController.getStorageUsageTypeDetails(ctx)
	if err != nil {
		logger.Error(err, "got error for getting the storage usage type")
	}
	if usageTypeDetails != nil {
		logger.Info("got usage unit type of the type storage", "code", usageTypeDetails.UsageTypeCode)
	}
	err = ariaController.ensureStorageUsageType(ctx)
	if err != nil {
		t.Fatalf("failed to ensure storage usage type: %v", err)
	}
	usageTypeDetailsAfterEnsuring, err := ariaController.getStorageUsageTypeDetails(ctx)
	if err != nil {
		t.Fatalf("failed to ensure storage usage type: %v", err)
	}
	if usageTypeDetailsAfterEnsuring == nil {
		t.Fatalf("failed to ensure storage usage type: %v", err)
	}
}

func TestCreateAriaAcct(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreateAriaAcct")
	logger.Info("BEGIN")
	defer logger.Info("End")

	ariaController := NewAriaController(clientTestsCommon.GetAriaClient(), clientTestsCommon.GetAriaAdminClient(), clientTestsCommon.GetAriaCredentials())
	err := ariaController.InitAria(ctx)
	if err != nil {
		t.Fatalf("failed to initialize aria: %v", err)
	}
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	cloudAccountClient := AriaService.cloudAccountClient
	premiumCloudAccountId, err := cloudAccountClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  premiumUser,
			Owner: premiumUser,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("failed to create premium cloud account: %v", err)
	}

	err = ariaController.CreateAriaAccount(ctx, premiumCloudAccountId.Id)
	if err != nil {
		t.Fatalf("failed to create premium aria account: %v", err)
	}

	//verifyDefaultCloudCreditsForPremium(t, ctx, ariaController, premiumCloudAccountId.Id)

	enterprisePendingUser := "enterprise_pending" + uuid.NewString() + "@example.com"
	enterprisePendingCloudAccountId, err := cloudAccountClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  enterprisePendingUser,
			Owner: enterprisePendingUser,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING,
		})
	if err != nil {
		t.Fatalf("failed to create enterprise pending cloud account: %v", err)
	}

	err = ariaController.CreateAriaAccount(ctx, enterprisePendingCloudAccountId.Id)
	if err != nil {
		t.Fatalf("failed to create enterprise pending aria account: %v", err)
	}

	enterpriseUser := "enterprise_" + uuid.NewString() + "@example.com"
	enterpriseCloudAccountId, err := cloudAccountClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  enterpriseUser,
			Owner: enterpriseUser,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_ENTERPRISE,
		})
	if err != nil {
		t.Fatalf("failed to create enterprise cloud account: %v", err)
	}

	err = ariaController.CreateAriaAccount(ctx, enterpriseCloudAccountId.Id)
	if err != nil {
		t.Fatalf("failed to create enterprise aria account: %v", err)
	}

	//verifyDefaultCloudCreditsForEnterprise(t, ctx, ariaController, enterpriseCloudAccountId.Id)

	err = ariaController.CreateAriaAccount(ctx, premiumCloudAccountId.Id)
	if err != nil {
		t.Fatalf("failed to retry create premium aria account: %v", err)
	}

	//verifyDefaultCloudCreditsForPremium(t, ctx, ariaController, premiumCloudAccountId.Id)

	logger.Info("verified works as expected for the creation of premium and enterprise cloud credits")
}

// Test to UpdatePaymentMethod and removes non-primary cards/paymentMethod
// This is tested via AddPaymentPostProcessing
func TestAddPaymentPostProcessing(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestAddPaymentPostProcessing")
	logger.Info("testing update account payment method and removing old payment method")
	product := tests.GetProduct()
	productFamily := tests.GetProductFamily()

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
	ariaController := NewAriaController(clientTestsCommon.GetAriaClient(), clientTestsCommon.GetAriaAdminClient(), clientTestsCommon.GetAriaCredentials())
	err := ariaController.InitAria(ctx)
	if err != nil {
		t.Fatalf("failed to initialize aria: %v", err)
	}
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	cloudAccountClient := AriaService.cloudAccountClient
	premiumCloudAccountId, err := cloudAccountClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  premiumUser,
			Owner: premiumUser,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("failed to create premium cloud account: %v", err)
	}

	acctClientId := client.GetAccountClientId(premiumCloudAccountId.Id)

	usageType, err := ariaController.ariaUsageTypeClient.GetMinutesUsageType(ctx)
	if err != nil {
		t.Fatalf("failed to get Usagetype: %v", err)
	}

	// Create a new plan to map with the account creation
	_, err = ariaController.ariaPlanClient.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaController.ariaPlanClient.GetAriaPlanDetailsAllForClientPlanId(context.Background(), clientTestsCommon.GetTestClientId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}

	// Creation of Aria Account
	_, err = ariaController.ariaAccountClient.CreateAriaAccount(ctx, acctClientId, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create premium aria account: %v", err)
	}

	// Get billing group info/details for an account through client_account_id
	getBillingGroupResp, err := ariaController.ariaAccountClient.GetBillingGroup(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	clientBillingGroupId := getBillingGroupResp.BillingGroupDetails[0].ClientBillingGroupId

	// As a part of pre payment processing assigning collection account group
	res, err := ariaController.ariaPaymentClient.AssignCollectionsAccountGroup(context.Background(), acctClientId, clientAcctGroupId)
	if err != nil {
		if res != nil && res.ErrorCode == 12004 {
			logger.Info("account already assigned to this group")
		} else {
			t.Fatalf("Failed to assign collections account group: %v", err)
		}
	}

	// Add test card to Aria system
	_, err = ariaController.ariaPaymentClient.AddAccountPaymentMethod(context.Background(), acctClientId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}

	getPaymentMethodsResp, err := ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}

	paymentMethodNo := getPaymentMethodsResp.AccountPaymentMethods[0].PaymentMethodNo

	//update primary payment method no in account billing group
	_, err = ariaController.ariaPaymentClient.UpdateAccountBillingGroup(context.Background(), acctClientId, clientBillingGroupId, paymentMethodNo)
	if err != nil {
		t.Fatalf("failed to update account billing group: %v", err)
	}

	getBillingGroupRes, err := ariaController.ariaAccountClient.GetBillingGroup(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	assert.Equal(t, paymentMethodNo, getBillingGroupRes.BillingGroupDetails[0].PrimaryPaymentMethodNo)

	// Updating the existing credit card details.
	// Updated CCExpireMonth and CCExpireYear
	creditCardDetails.CCExpireMonth = 10
	creditCardDetails.CCExpireYear = 2050
	clientPaymentMethodId = uuid.New().String()

	_, err = ariaController.ariaPaymentClient.AddAccountPaymentMethod(context.Background(), acctClientId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}
	getPaymentMethodsResp, err = ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}

	//Fetching new PaymentMethodNo
	updatedPrimaryPaymentMethodNo := getPaymentMethodsResp.AccountPaymentMethods[1].PaymentMethodNo

	// Updated the new payment method by calling AddPaymentPostProcessing.
	//
	// This updates the account billing group with new primaryPaymentMethod and removes the old payment method
	err = ariaController.AddPaymentPostProcessing(ctx, premiumCloudAccountId.Id, updatedPrimaryPaymentMethodNo)
	if err != nil {
		t.Fatalf("failed to add payment post processing: %v", err)
	}
}

func TestAddPaymentPostProcessingMultipleCards(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestAddPaymentPostProcessingMultipleCards")
	logger.Info("testing update account payment method and removing multiple old payment method")
	product := tests.GetProduct()
	productFamily := tests.GetProductFamily()

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
	ariaController := NewAriaController(clientTestsCommon.GetAriaClient(), clientTestsCommon.GetAriaAdminClient(), clientTestsCommon.GetAriaCredentials())
	err := ariaController.InitAria(ctx)
	if err != nil {
		t.Fatalf("failed to initialize aria: %v", err)
	}
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	cloudAccountClient := AriaService.cloudAccountClient
	premiumCloudAccountId, err := cloudAccountClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  premiumUser,
			Owner: premiumUser,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("failed to create premium cloud account: %v", err)
	}

	acctClientId := client.GetAccountClientId(premiumCloudAccountId.Id)

	usageType, err := ariaController.ariaUsageTypeClient.GetMinutesUsageType(ctx)
	if err != nil {
		t.Fatalf("failed to get Usagetype: %v", err)
	}

	// Create a new plan to map with the account creation
	_, err = ariaController.ariaPlanClient.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaController.ariaPlanClient.GetAriaPlanDetailsAllForClientPlanId(context.Background(), clientTestsCommon.GetTestClientId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}

	// Creation of Aria Account
	_, err = ariaController.ariaAccountClient.CreateAriaAccount(ctx, acctClientId, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create premium aria account: %v", err)
	}

	// Get billing group info/details for an account through client_account_id
	getBillingGroupResp, err := ariaController.ariaAccountClient.GetBillingGroup(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	clientBillingGroupId := getBillingGroupResp.BillingGroupDetails[0].ClientBillingGroupId

	// As a part of pre payment processing assigning collection account group
	res, err := ariaController.ariaPaymentClient.AssignCollectionsAccountGroup(context.Background(), acctClientId, clientAcctGroupId)
	if err != nil {
		if res != nil && res.ErrorCode == 12004 {
			logger.Info("account already assigned to this group")
		} else {
			t.Fatalf("Failed to assign collections account group: %v", err)
		}
	}

	// Add test card to Aria system
	_, err = ariaController.ariaPaymentClient.AddAccountPaymentMethod(context.Background(), acctClientId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}

	getPaymentMethodsResp, err := ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}

	paymentMethodNo := getPaymentMethodsResp.AccountPaymentMethods[0].PaymentMethodNo

	//update primary payment method no in account billing group
	_, err = ariaController.ariaPaymentClient.UpdateAccountBillingGroup(context.Background(), acctClientId, clientBillingGroupId, paymentMethodNo)
	if err != nil {
		t.Fatalf("failed to update account billing group: %v", err)
	}

	getBillingGroupRes, err := ariaController.ariaAccountClient.GetBillingGroup(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	assert.Equal(t, paymentMethodNo, getBillingGroupRes.BillingGroupDetails[0].PrimaryPaymentMethodNo)

	// Updating the existing credit card details.
	// Updated CCExpireMonth and CCExpireYear
	// PaymentMethod 2
	creditCardDetails.CCExpireMonth = 10
	creditCardDetails.CCExpireYear = 2050
	clientPaymentMethodId = uuid.New().String()

	_, err = ariaController.ariaPaymentClient.AddAccountPaymentMethod(context.Background(), acctClientId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}
	getPaymentMethodsResp, err = ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}

	//Fetching new PaymentMethodNo
	updatedPrimaryPaymentMethodNo := getPaymentMethodsResp.AccountPaymentMethods[1].PaymentMethodNo

	// Updated the new payment method by calling AddPaymentPostProcessing.
	//
	// This updates the account billing group with new primaryPaymentMethod and removes the old payment method
	err = ariaController.AddPaymentPostProcessing(ctx, premiumCloudAccountId.Id, updatedPrimaryPaymentMethodNo)
	if err != nil {
		t.Fatalf("failed to add payment post processing: %v", err)
	}

	// Updating the existing credit card details.
	// Updated CCExpireMonth and CCExpireYear
	// PaymentMethod 3
	creditCardDetails.CCExpireMonth = 01
	creditCardDetails.CCExpireYear = 2090
	clientPaymentMethodId = uuid.New().String()

	_, err = ariaController.ariaPaymentClient.AddAccountPaymentMethod(context.Background(), acctClientId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}
	getPaymentMethodsResp, err = ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}

	//Fetching new PaymentMethodNo
	updatedPrimaryPaymentMethodNo = getPaymentMethodsResp.AccountPaymentMethods[1].PaymentMethodNo

	err = ariaController.AddPaymentPostProcessing(ctx, premiumCloudAccountId.Id, updatedPrimaryPaymentMethodNo)
	if err != nil {
		t.Fatalf("failed to add payment post processing: %v", err)
	}

}

func verifyDefaultCloudCreditsForEnterprise(t *testing.T, ctx context.Context, ariaController *AriaController, cloudAccountId string) {
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

func verifyDefaultCloudCreditsForPremium(t *testing.T, ctx context.Context, ariaController *AriaController, cloudAccountId string) {
	acctCredits, err := ariaController.GetAccountCredits(ctx, cloudAccountId)
	if err != nil {
		t.Fatalf("failed to get account credits: %v", err)
	}
	for _, acctCredit := range acctCredits {
		if acctCredit.ReasonCode == kFixMeWeHaventImplementedReasonCodeYet {
			if config.Cfg.PremiumDefaultCreditAmount != float64(acctCredit.Amount) {
				t.Fatalf("incorrect credit amount: %v", err)
			}
			return
		}
	}
	t.Fatalf("did not find default cloud credits assigned")
}

func TestGetUnappliedCredits(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestGetUnappliedCredits")
	logger.Info("BEGIN")
	defer logger.Info("End")
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateCloudAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	_, clientPlanId, err := tests.CreateAcctWithPlanDefaultProd(ctx, client.GetAccountClientId(cloudAcct.Id))
	if err != nil {
		t.Fatalf("failed to create account with default prod")
	}
	_, err = tests.AssignCreditsToAccount(ctx, client.GetAccountClientId(cloudAcct.Id), tests.DefaultCloudCreditAmount)
	if err != nil {
		t.Fatalf("failed to assign default cloud credits to account")
	}
	err = tests.CreateTestUsageRecord(ctx,
		client.GetAccountClientId(cloudAcct.Id),
		clientPlanId, client.GetMinsUsageTypeCode(), tests.DefaultUsageAmount)
	if err != nil {
		t.Fatalf("failed to add default usage record")
	}
	ariaController := NewAriaController(clientTestsCommon.GetAriaClient(), clientTestsCommon.GetAriaAdminClient(), clientTestsCommon.GetAriaCredentials())
	unAppliedCredits, err := ariaController.GetUnAppliedServiceCredits(ctx, cloudAcct.Id)
	if err != nil {
		t.Fatalf("failed to get unapplied credits")
	}
	logger.Info("got unapplied credits", "value", unAppliedCredits)
}

func TestGetUnappliedCreditsHavingUsesMoreThanCredit(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestGetUnappliedCreditsHavingUsesMoreThanCredit")
	logger.Info("BEGIN")
	defer logger.Info("End")
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateCloudAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	_, clientPlanId, err := tests.CreateAcctWithPlanDefaultProd(ctx, client.GetAccountClientId(cloudAcct.Id))
	if err != nil {
		t.Fatalf("failed to create account with default prod")
	}
	_, err = tests.AssignCreditsToAccount(ctx, client.GetAccountClientId(cloudAcct.Id), tests.DefaultCloudCreditAmount)
	if err != nil {
		t.Fatalf("failed to assign default cloud credits to account")
	}
	err = tests.CreateTestUsageRecord(ctx,
		client.GetAccountClientId(cloudAcct.Id),
		clientPlanId, client.GetMinsUsageTypeCode(), (tests.DefaultUsageAmount+tests.DefaultCloudCreditAmount)*100)
	if err != nil {
		t.Fatalf("failed to add default usage record")
	}
	ariaInvoiceClient := clientTestsCommon.GetAriaInvoiceClient()
	clientAccountId := client.GetAccountClientId(cloudAcct.Id)
	getPendingInovice, err := ariaInvoiceClient.GetPendingInvoiceNo(ctx, clientAccountId)
	if err != nil {
		t.Fatalf("failed to get pending invoice: %v", err)
	}
	for _, pendingInvoice := range getPendingInovice.PendingInvoice {
		_, err := ariaInvoiceClient.ManagePendingInvoiceWithInoviceNo(ctx, clientAccountId, pendingInvoice.InvoiceNo, tests.ACTION_DIRECTIVE_REGENERATE)
		if err != nil {
			t.Fatalf("failed to regenerate invoice: %v", err)
		}
	}
	ariaController := NewAriaController(clientTestsCommon.GetAriaClient(), clientTestsCommon.GetAriaAdminClient(), clientTestsCommon.GetAriaCredentials())
	unAppliedCredits, err := ariaController.GetUnAppliedServiceCredits(ctx, cloudAcct.Id)
	if err != nil {
		t.Fatalf("failed to get unapplied credits")
	}
	if unAppliedCredits != 0 {
		t.Fatalf("failed to get expected unapplied credits")
	}
	logger.Info("got unapplied credits", "value", unAppliedCredits)
}

func TestAddPaymentPostProcessingForCloudAccoutUpdate(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestAddPaymentPostProcessingForCloudAccoutUpdate")
	logger.Info("testing update account payment method and removing old payment method")
	product := tests.GetProduct()
	productFamily := tests.GetProductFamily()

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
	ariaController := NewAriaController(clientTestsCommon.GetAriaClient(), clientTestsCommon.GetAriaAdminClient(), clientTestsCommon.GetAriaCredentials())
	err := ariaController.InitAria(ctx)
	if err != nil {
		t.Fatalf("failed to initialize aria: %v", err)
	}
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	cloudAccountClient := AriaService.cloudAccountClient
	premiumCloudAccountId, err := cloudAccountClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  premiumUser,
			Owner: premiumUser,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("failed to create premium cloud account: %v", err)
	}

	acctClientId := client.GetAccountClientId(premiumCloudAccountId.Id)

	usageType, err := ariaController.ariaUsageTypeClient.GetMinutesUsageType(ctx)
	if err != nil {
		t.Fatalf("failed to get Usagetype: %v", err)
	}

	// Create a new plan to map with the account creation
	_, err = ariaController.ariaPlanClient.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaController.ariaPlanClient.GetAriaPlanDetailsAllForClientPlanId(context.Background(), clientTestsCommon.GetTestClientId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}

	// Creation of Aria Account
	_, err = ariaController.ariaAccountClient.CreateAriaAccount(ctx, acctClientId, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create premium aria account: %v", err)
	}

	// Get billing group info/details for an account through client_account_id
	getBillingGroupResp, err := ariaController.ariaAccountClient.GetBillingGroup(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	clientBillingGroupId := getBillingGroupResp.BillingGroupDetails[0].ClientBillingGroupId

	// As a part of pre payment processing assigning collection account group
	res, err := ariaController.ariaPaymentClient.AssignCollectionsAccountGroup(context.Background(), acctClientId, clientAcctGroupId)
	if err != nil {
		if res != nil && res.ErrorCode == 12004 {
			logger.Info("account already assigned to this group")
		} else {
			t.Fatalf("Failed to assign collections account group: %v", err)
		}
	}

	// Add test card to Aria system
	_, err = ariaController.ariaPaymentClient.AddAccountPaymentMethod(context.Background(), acctClientId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}

	getPaymentMethodsResp, err := ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}

	paymentMethodNo := getPaymentMethodsResp.AccountPaymentMethods[0].PaymentMethodNo

	//update primary payment method no in account billing group
	_, err = ariaController.ariaPaymentClient.UpdateAccountBillingGroup(context.Background(), acctClientId, clientBillingGroupId, paymentMethodNo)
	if err != nil {
		t.Fatalf("failed to update account billing group: %v", err)
	}

	getBillingGroupRes, err := ariaController.ariaAccountClient.GetBillingGroup(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	assert.Equal(t, paymentMethodNo, getBillingGroupRes.BillingGroupDetails[0].PrimaryPaymentMethodNo)

	err = ariaController.AddPaymentPostProcessing(ctx, premiumCloudAccountId.Id, paymentMethodNo)
	if err != nil {
		t.Fatalf("failed to add payment post processing: %v", err)
	}
	account, err := cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: premiumCloudAccountId.Id})
	if err != nil {
		t.Fatalf("error in cloud account client response %v", err)
	}
	assert.Equal(t, false, account.GetLowCredits())
	assert.Equal(t, false, account.GetTerminatePaidServices())
	assert.Equal(t, false, account.GetTerminateMessageQueued())
	assert.Equal(t, true, account.GetPaidServicesAllowed())
}

func TestDowngradePremiumtoStandard(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestDowngradePremiumtoStandard")

	product := tests.GetProduct()
	productFamily := tests.GetProductFamily()

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
	ariaController := NewAriaController(clientTestsCommon.GetAriaClient(), clientTestsCommon.GetAriaAdminClient(), clientTestsCommon.GetAriaCredentials())
	err := ariaController.InitAria(ctx)
	if err != nil {
		t.Fatalf("failed to initialize aria: %v", err)
	}
	premiumUser := "premium_" + uuid.NewString() + "@example.com"
	cloudAccountClient := AriaService.cloudAccountClient
	premiumCloudAccountId, err := cloudAccountClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  premiumUser,
			Owner: premiumUser,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("failed to create premium cloud account: %v", err)
	}

	acctClientId := client.GetAccountClientId(premiumCloudAccountId.Id)

	usageType, err := ariaController.ariaUsageTypeClient.GetMinutesUsageType(ctx)
	if err != nil {
		t.Fatalf("failed to get Usagetype: %v", err)
	}

	// Create a new plan to map with the account creation
	_, err = ariaController.ariaPlanClient.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	// Get Plan details to fetch the client plan id which is needed while Creation of Aria account
	resp, err := ariaController.ariaPlanClient.GetAriaPlanDetailsAllForClientPlanId(context.Background(), clientTestsCommon.GetTestClientId(product.GetId()))
	if err != nil {
		t.Fatalf("Failed to get client plan: %v", err)
	}

	// Creation of Aria Account
	_, err = ariaController.ariaAccountClient.CreateAriaAccount(ctx, acctClientId, resp.AllClientPlanDtls[0].ClientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create premium aria account: %v", err)
	}

	// Get billing group info/details for an account through client_account_id
	getBillingGroupResp, err := ariaController.ariaAccountClient.GetBillingGroup(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	clientBillingGroupId := getBillingGroupResp.BillingGroupDetails[0].ClientBillingGroupId

	// As a part of pre payment processing assigning collection account group
	res, err := ariaController.ariaPaymentClient.AssignCollectionsAccountGroup(context.Background(), acctClientId, clientAcctGroupId)
	if err != nil {
		if res != nil && res.ErrorCode == 12004 {
			logger.Info("account already assigned to this group")
		} else {
			t.Fatalf("Failed to assign collections account group: %v", err)
		}
	}

	// Add test card to Aria system
	_, err = ariaController.ariaPaymentClient.AddAccountPaymentMethod(context.Background(), acctClientId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
	if err != nil {
		t.Fatalf("Failed to add account payment method: %v", res)
	}

	getPaymentMethodsResp, err := ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}

	paymentMethodNo := getPaymentMethodsResp.AccountPaymentMethods[0].PaymentMethodNo
	_, err = ariaController.ariaPaymentClient.UpdateAccountBillingGroup(context.Background(), acctClientId, clientBillingGroupId, paymentMethodNo)
	if err != nil {
		t.Fatalf("failed to update account billing group: %v", err)
	}

	getBillingGroupRes, err := ariaController.ariaAccountClient.GetBillingGroup(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("failed to get billing group: %v", err)
	}
	assert.Equal(t, paymentMethodNo, getBillingGroupRes.BillingGroupDetails[0].PrimaryPaymentMethodNo)

	// This downgrade premium account to standard by deactivate the account and remove credit card
	err = ariaController.DowngradePremiumtoStandard(ctx, premiumCloudAccountId.Id, false)
	if err != nil {
		t.Fatalf("failed to downgrade: %v", err)
	}

	// Checks
	getPaymentMethodsResp, err = ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), acctClientId)
	if err != nil {
		t.Fatalf("Failed to get all payment methods: %v", err)
	}
	assert.Equal(t, len(getPaymentMethodsResp.AccountPaymentMethods), 0)

	accountDetails, err := ariaController.ariaAccountClient.GetAriaAccountDetailsAllForClientId(ctx, acctClientId)
	if err != nil {
		t.Fatalf("Failed to get account details: %v", err)
	}
	assert.Equal(t, accountDetails.StatusCd, int64(0))
}
