// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Invoice Generation", Ordered, Label("invoice"), func() {
	BeforeEach(func() {
		if !ShouldRunTest("invoice") {
			Skip("Invoicing Flow Test")
		}
	})
	ctx := context.Background()
	clientAccountId := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	var usageType *data.UsageType
	var product *pb.Product
	var productFamily *pb.ProductFamily
	var err error
	var pendingInvoices []data.PendingInvoice
	var invoiceHists []data.InvoiceHist

	It("Create Usage Type", func() {
		usageTypeClient := common.GetAriaUsageTypeClient()
		usageType, err = usageTypeClient.GetMinutesUsageType(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(usageType.UsageUnitType).To(Equal("minute"), "Usage Type created")
	})

	It("Create Master Plan", func() {
		product = GetProduct()
		productFamily = GetProductFamily()
		ariaPlanClient := common.GetAriaPlanClient()
		createRespBody, err := ariaPlanClient.CreatePlan(ctx, product, productFamily, usageType)
		Expect(err).NotTo(HaveOccurred())
		Expect(createRespBody.GetErrorCode()).To(Equal(int64(0)), "Master plan created")

	})
	It("Create Account", func() {
		resp, err := ariaAccountClient.CreateAriaAccount(ctx, clientAccountId, client.GetPlanClientId(product.GetId()), client.ACCOUNT_TYPE_PREMIUM)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetErrorCode()).To(Equal(int64(0)), "Account created")

	})
	It("Set Notification Template Group id", func() {
		clientNotificationTemplateGrpId := client.GetClientNotificationTemplateGroupId(pb.AccountType_ACCOUNT_TYPE_PREMIUM, "US")
		resp, err := ariaAccountClient.SetAccountNotifyTemplateGroup(ctx, clientAccountId, clientNotificationTemplateGrpId)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetErrorCode()).To(Equal(int64(0)), "Account Notification Template Group")

	})

	It("Assign Default Credit to Account", func() {
		currentDate := time.Now()
		newDate := currentDate.AddDate(0, 0, 100)
		expirationDate := fmt.Sprintf("%04d-%02d-%02d", newDate.Year(), newDate.Month(), newDate.Day())
		serviceCreditClient := common.GetServiceCreditClient()
		code := int64(1)
		_, err := serviceCreditClient.CreateServiceCredits(context.Background(), clientAccountId, DefaultCloudCreditAmount, code, expirationDate, "testCredit")
		Expect(err).NotTo(HaveOccurred())
	})

	var details *response.GetAcctDetailsAllMResponse
	It("Get Account Details", func() {
		details, err = ariaAccountClient.GetAriaAccountDetailsAllForClientId(ctx, clientAccountId)
		Expect(err).NotTo(HaveOccurred())
		Expect(client.IsPayloadEmpty(details)).To(Equal(false), "Got Account")
	})

	ariaPaymentClient := common.GetAriaPaymentClient()
	It("Assign CollectionsGroup  to Account", func() {
		clientAcctGroupId := "Chase"
		_, err = ariaPaymentClient.AssignCollectionsAccountGroup(ctx, clientAccountId, clientAcctGroupId)
		Expect(err).NotTo(HaveOccurred())

	})
	It("Add test payment to account", func() {
		payMethodType := 1
		clientBillingGroupId := details.MasterPlansInfo[0].ClientBillingGroupId
		creditCardDetails := client.CreditCardDetails{
			CCNumber:      4111111111111111,
			CCExpireMonth: 12,
			CCExpireYear:  2025,
			CCV:           987,
		}
		clientPaymentMethodId := uuid.New().String()
		_, err = ariaPaymentClient.AddAccountPaymentMethod(ctx, clientAccountId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
		Expect(err).NotTo(HaveOccurred())

	})
	It("Assigning Master Plan for invocing", func() {
		err := AssignPlanToAccount(ctx, clientAccountId, client.GetPlanClientId(product.GetId()), ACCOUNT_TYPE_PREMIUM)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Post usages data ", func() {
		ariaUsageClient := common.GetUsageClient()
		pefixLen := len(config.Cfg.ClientIdPrefix)
		currentDate := time.Now()
		newDate := currentDate.AddDate(0, 0, -10)
		usage := client.BillingUsage{CloudAccountId: clientAccountId[pefixLen+1:], ProductId: product.Id, TransactionId: clientAccountId, ResourceId: uuid.New().String(), Amount: 50, UsageDate: newDate}
		_, err = ariaUsageClient.CreateUsageRecord(ctx, usageType.UsageTypeCode, &usage)
		Expect(err).NotTo(HaveOccurred())

	})
	It("Get Usage History", func() {
		ariaUsageClient := common.GetUsageClient()
		layout := "2006-01-02 15:04:05"
		currentTime := time.Now()
		startDateTime := currentTime.AddDate(0, 0, -30).Format(layout)
		endDateTime := currentTime.AddDate(0, 0, 1).Format(layout)
		masterPlanId := client.GetClientMasterPlanInstanceId(GetCloudAcctIdFromClientAcctId(clientAccountId), product.GetId())
		_, err = ariaUsageClient.GetUsageHistory(ctx, clientAccountId, masterPlanId, startDateTime, endDateTime)
		Expect(err).NotTo(HaveOccurred())
	})
	ariaInvoiceClient := common.GetAriaInvoiceClient()
	It("Get Pending invoice number", func() {
		resp, err := ariaInvoiceClient.GetPendingInvoiceNo(ctx, clientAccountId)
		Expect(err).NotTo(HaveOccurred())
		pendingInvoices = resp.PendingInvoice
	})

	It("Regenerate pending invoice to recalcualte credit", func() {
		for _, pendingInvoice := range pendingInvoices {
			_, err = ariaInvoiceClient.ManagePendingInvoiceWithInoviceNo(ctx, clientAccountId, pendingInvoice.InvoiceNo, 3)
			Expect(err).NotTo(HaveOccurred())
		}

	})
	It("Get pending invoice number after recalcualtion of credit", func() {
		resp, err := ariaInvoiceClient.GetPendingInvoiceNo(ctx, clientAccountId)
		Expect(err).NotTo(HaveOccurred())
		pendingInvoices = resp.PendingInvoice
	})
	It("Approve pending invoice", func() {
		for _, pendingInvoice := range pendingInvoices {
			_, err := ariaInvoiceClient.ManagePendingInvoiceWithInoviceNo(ctx, clientAccountId, pendingInvoice.InvoiceNo, 1)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("Get invoice history", func() {
		masterPlanId := client.GetClientMasterPlanInstanceId(GetCloudAcctIdFromClientAcctId(clientAccountId), product.GetId())
		resp, err := ariaInvoiceClient.GetInvoiceHistory(ctx, clientAccountId, masterPlanId, client.NO_START_DATE, client.NO_END_DATE)
		invoiceHists = resp.InvoiceHist
		Expect(err).NotTo(HaveOccurred())
	})

	It("Get invoice details", func() {
		for _, invoiceHist := range invoiceHists {
			masterPlanId := client.GetClientMasterPlanInstanceId(GetCloudAcctIdFromClientAcctId(clientAccountId), product.GetId())
			_, err = ariaInvoiceClient.GetInvoiceDetails(ctx, clientAccountId, invoiceHist.InvoiceNo, masterPlanId)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("Get statement for invoice", func() {
		for _, invoiceHist := range invoiceHists {
			_, err = ariaInvoiceClient.GetStatementForInvoice(ctx, clientAccountId, invoiceHist.InvoiceNo)
			Expect(err).NotTo(HaveOccurred())
		}
	})
})
