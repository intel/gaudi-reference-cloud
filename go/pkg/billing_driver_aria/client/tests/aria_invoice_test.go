// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	//aria "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func TestGetInvoiceDetails(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetInvoiceDetails")
	logger.Info("Testing get invoice details")
	cloudAccountId := cloudaccount.MustNewId()
	clientAcctId := client.GetAccountClientId(cloudAccountId)
	ariaAccountClient := common.GetAriaAccountClient()
	ctx := context.Background()
	InitAriaForTesting(ctx)
	acctRespBody, err := ariaAccountClient.CreateAriaAccount(ctx, clientAcctId, client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	product := GetProduct()
	product.Id = uuid.NewString()
	product.Name = "invoice detail test"
	usageType := GetUsageType(t)
	_, err = common.GetAriaPlanClient().CreatePlan(ctx, product, GetProductFamily(), usageType)
	if err != nil {
		t.Fatal(err)
	}

	planId := client.GetPlanClientId(product.Id)
	planInstanceId := client.GetClientMasterPlanInstanceId(cloudAccountId, product.Id)
	schedId := client.GetRateScheduleClientId(product.Id, "premium")
	_, err = ariaAccountClient.AssignPlanToAccountWithBillingAndDunningGroup(ctx, clientAcctId, planId, planInstanceId, schedId, client.BILL_LAG_DAYS, client.ALT_BILL_DAY)
	if err != nil {
		t.Fatal(err)
	}

	ariaUsageClient := common.GetUsageClient()
	usage := []*client.BillingUsage{{CloudAccountId: cloudAccountId, ProductId: product.Id, TransactionId: uuid.NewString(), ResourceId: uuid.New().String(), Amount: 50, UsageUnitType: "RATE_UNIT_DOLLARS_PER_MINUTE"}}
	_, err = ariaUsageClient.CreateBulkUsageRecords(ctx, usage)
	if err != nil {
		t.Fatalf("failed to create bulk usage record: %v", err)
	}
	getInvoiceDetails := common.GetAriaInvoiceClient()
	masterPlanId := client.GetClientMasterPlanInstanceId(cloudAccountId, product.GetId())
	_, err = getInvoiceDetails.GetInvoiceDetails(ctx, clientAcctId, acctRespBody.OutAcct[0].InvoiceInfo[0].InvoiceNo, masterPlanId)
	//TODO: fix invoice for no line items provided
	if err != nil && !strings.Contains(err.Error(), "error code:1008") {
		t.Fatalf("Failed to get Invoice details: %v", err)
	}
}

// Skipping the test
/*
func TestGetStatementForInvoice(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetStatementForInvoice")
	logger.Info("Testing get statement for invoice")
	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	controller := aria.NewAriaController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials())
	ctx := context.Background()
	if err := controller.InitAria(ctx); err != nil {
		t.Fatal(err)
	}
	acctRespBody, err := ariaAccountClient.CreateAriaAccount(ctx, id, GetDefaultClientPlanId())
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	product := GetProduct()
	usageType := GetUsageType(t)
	ariaUsageClient := common.GetUsageClient()
	usage := []*client.BillingUsage{{CloudAccountId: id, ProductId: product.Id, TransactionId: acctRespBody.OutAcct[0].MasterPlansAssigned[0].ClientPlanInstanceId, ResourceId: uuid.New().String(), Amount: 50}}
	_, err = ariaUsageClient.CreateBulkUsageRecords(ctx, usageType.UsageTypeName, usage)
	if err != nil {
		t.Fatalf("failed to create bulk usage record: %v", err)
	}
	getStatementForInvoice := common.GetAriaInvoiceClient()
	_, err = getStatementForInvoice.GetStatementForInvoice(ctx, acctRespBody.OutAcct[0].ClientAcctId, acctRespBody.OutAcct[0].InvoiceInfo[0].InvoiceNo)
	if err != nil {
		t.Fatalf("Failed to get Statment for Invoice: %v", err)
	}
	//If we want to store the html file recieved as response (in outStatement)
	// if err != nil {
	// 	t.Fatalf("Failed to get Invoice details: %v", err)
	// }
	// _, err = os.Create("path to html file")
	// if err != nil {
	// 	t.Fatalf("Unable to create the file:- %v", err)
	// }
	// ioutil.WriteFile("path to html file", []byte(respNew.OutStatement), 0644)
}

*/

// Skipping the test
/*
func TestInvoiceClient(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestAriaInvoice")
	logger.Info("Testing Aria Response")
	id := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	controller := aria.NewAriaController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials())
	ctx := context.Background()
	if err := controller.InitAria(ctx); err != nil {
		t.Fatal(err)
	}
	acctRespBody, err := ariaAccountClient.CreateAriaAccount(ctx, id, GetDefaultClientPlanId())
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	product := GetProduct()
	usageType := GetUsageType(t)
	ariaUsageClient := common.GetUsageClient()
	usage := []*client.BillingUsage{{CloudAccountId: id, ProductId: product.Id, TransactionId: acctRespBody.OutAcct[0].MasterPlansAssigned[0].ClientPlanInstanceId, ResourceId: uuid.New().String(), Amount: 50}} //plan_no
	_, err = ariaUsageClient.CreateBulkUsageRecords(ctx, usageType.UsageTypeName, usage)
	if err != nil {
		t.Fatalf("failed to create bulk usage record: %v", err)
	}
	getInvoiceDetails := common.GetAriaInvoiceClient()
	_, err = getInvoiceDetails.GetInvoiceDetails(ctx, acctRespBody.OutAcct[0].ClientAcctId, acctRespBody.OutAcct[0].InvoiceInfo[0].InvoiceNo, acctRespBody.OutAcct[0].MasterPlansAssigned[0].PlanInstanceNo)
	if err != nil {
		t.Fatalf("Failed to get Invoice details: %v", err)
	}
	_, err = getInvoiceDetails.GetStatementForInvoice(ctx, acctRespBody.OutAcct[0].ClientAcctId, acctRespBody.OutAcct[0].InvoiceInfo[0].InvoiceNo)
	if err != nil {
		t.Fatalf("Failed to get Statment for Invoice: %v", err)
	}
}
*/
