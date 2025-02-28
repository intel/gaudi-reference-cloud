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

var _ = Describe("Test Enterprise Account Flow", Ordered, Label("enterprise"), func() {
	BeforeEach(func() {
		if !ShouldRunTest("enterprise") {
			Skip("Enterprise Flow Test")
		}
	})
	ctx := context.Background()
	parentClientAccountId := GetClientAccountId()
	childClientAccountId := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	var usageType *data.UsageType
	var product *pb.Product
	var productFamily *pb.ProductFamily
	var err error
	var pendingInvoices []data.PendingInvoice
	var invoiceHists []data.InvoiceHist
	var createParentAcctNoResponse *response.CreateAcctCompleteMResponse
	var createChildAcctNoResponse *response.CreateAcctCompleteMResponse
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
	It("Create Parent Account", func() {
		createParentAcctNoResponse, err = ariaAccountClient.CreateAriaAccount(ctx, parentClientAccountId, client.GetPlanClientId(product.GetId()), client.ACCOUNT_TYPE_ENTERPRISE)
		Expect(err).NotTo(HaveOccurred())
		Expect(createParentAcctNoResponse.GetErrorCode()).To(Equal(int64(0)), "Parent Account created")

	})
	It("Set Parent Notification Template Group id", func() {
		clientNotificationTemplateGrpId := client.GetClientNotificationTemplateGroupId(pb.AccountType_ACCOUNT_TYPE_PREMIUM, "US")
		resp, err := ariaAccountClient.SetAccountNotifyTemplateGroup(ctx, parentClientAccountId, clientNotificationTemplateGrpId)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetErrorCode()).To(Equal(int64(0)), "Account Notification Template Group")

	})

	var parentAcctDetails *response.GetAcctDetailsAllMResponse
	It("Get Parent Enterprise Account Details", func() {
		parentAcctDetails, err = ariaAccountClient.GetAriaAccountDetailsAllForClientId(ctx, parentClientAccountId)
		Expect(err).NotTo(HaveOccurred())
		Expect(client.IsPayloadEmpty(parentAcctDetails)).To(Equal(false), "Parent Enterprise Account")
	})

	It("Assigning Master Plan to parent account", func() {
		err := AssignPlanToAccount(ctx, parentClientAccountId, client.GetPlanClientId(product.GetId()), ACCOUNT_TYPE_ENTERPRISE)
		Expect(err).NotTo(HaveOccurred())
	})

	ariaPaymentClient := common.GetAriaPaymentClient()
	It("Assign CollectionsGroup  to Account", func() {
		clientAcctGroupId := "Chase"
		_, err = ariaPaymentClient.AssignCollectionsAccountGroup(ctx, parentClientAccountId, clientAcctGroupId)
		Expect(err).NotTo(HaveOccurred())

	})
	It("Add test payment to account", func() {
		payMethodType := 1
		clientBillingGroupId := parentAcctDetails.MasterPlansInfo[0].ClientBillingGroupId
		creditCardDetails := client.CreditCardDetails{
			CCNumber:      4111111111111111,
			CCExpireMonth: 12,
			CCExpireYear:  2025,
			CCV:           987,
		}
		clientPaymentMethodId := uuid.New().String()
		_, err = ariaPaymentClient.AddAccountPaymentMethod(ctx, parentClientAccountId, clientPaymentMethodId, clientBillingGroupId, payMethodType, creditCardDetails)
		Expect(err).NotTo(HaveOccurred())

	})

	It("Create Child Account", func() {
		createChildAcctNoResponse, err = ariaAccountClient.CreateAriaAccount(ctx, childClientAccountId, client.GetPlanClientId(product.GetId()), client.ACCOUNT_TYPE_ENTERPRISE)
		Expect(err).NotTo(HaveOccurred())
		Expect(createChildAcctNoResponse.GetErrorCode()).To(Equal(int64(0)), "Account created")

	})
	It("Set Child Account Enterprise Notification Template Group id", func() {
		clientNotificationTemplateGrpId := client.GetClientNotificationTemplateGroupId(pb.AccountType_ACCOUNT_TYPE_ENTERPRISE, "US")
		resp, err := ariaAccountClient.SetAccountNotifyTemplateGroup(ctx, childClientAccountId, clientNotificationTemplateGrpId)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetErrorCode()).To(Equal(int64(0)), "Account Notification Template Group")

	})

	It("Assign Default Credit to Child Account", func() {
		currentDate := time.Now()
		newDate := currentDate.AddDate(0, 0, 100)
		expirationDate := fmt.Sprintf("%04d-%02d-%02d", newDate.Year(), newDate.Month(), newDate.Day())
		serviceCreditClient := common.GetServiceCreditClient()
		code := int64(1)
		_, err := serviceCreditClient.CreateServiceCredits(context.Background(), childClientAccountId, 10, code, expirationDate, "testCredit")
		Expect(err).NotTo(HaveOccurred())
	})

	It("Create a Child Account with a Parent Pay Plan and Aligned Billing Dates", func() {
		childAcctNo := createChildAcctNoResponse.OutAcct[0].AcctNo
		parentAcctNo := createParentAcctNoResponse.OutAcct[0].AcctNo
		_, err = ariaAccountClient.UpdateAriaAccount(ctx, childAcctNo, parentAcctNo)
		Expect(err).NotTo(HaveOccurred())
	})
	It("Assigning Master Plan to child account", func() {
		err := AssignPlanToEnterpriseAccount(ctx, childClientAccountId, client.GetPlanClientId(product.GetId()), parentAcctDetails.MasterPlansInfo[0].LastBillThruDate, parentAcctDetails.MasterPlansInfo[0].ClientMasterPlanInstanceId)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Post usages data ", func() {
		ariaUsageClient := common.GetUsageClient()
		pefixLen := len(config.Cfg.ClientIdPrefix)
		currentDate := time.Now()
		newDate := currentDate.AddDate(0, 0, -10)
		usage := client.BillingUsage{CloudAccountId: childClientAccountId[pefixLen+1:], ProductId: product.Id, TransactionId: childClientAccountId, ResourceId: uuid.New().String(), Amount: 500000, UsageDate: newDate}
		_, err = ariaUsageClient.CreateUsageRecord(ctx, usageType.UsageTypeCode, &usage)
		Expect(err).NotTo(HaveOccurred())

	})
	It("Get Usage History", func() {
		ariaUsageClient := common.GetUsageClient()
		layout := "2006-01-02 15:04:05"
		currentTime := time.Now()
		startDateTime := currentTime.AddDate(0, 0, -30).Format(layout)
		endDateTime := currentTime.AddDate(0, 0, 1).Format(layout)
		masterPlanId := client.GetClientMasterPlanInstanceId(GetCloudAcctIdFromClientAcctId(childClientAccountId), product.GetId())
		_, err = ariaUsageClient.GetUsageHistory(ctx, childClientAccountId, masterPlanId, startDateTime, endDateTime)
		Expect(err).NotTo(HaveOccurred())
	})
	ariaInvoiceClient := common.GetAriaInvoiceClient()
	It("Get pending child account invoice number", func() {
		resp, err := ariaInvoiceClient.GetPendingInvoiceNo(ctx, childClientAccountId)
		Expect(err).NotTo(HaveOccurred())
		pendingInvoices = resp.PendingInvoice
	})

	It("Regenerate pending child account invoice to recalcualte credit", func() {
		for _, pendingInvoice := range pendingInvoices {
			_, err = ariaInvoiceClient.ManagePendingInvoiceWithInoviceNo(ctx, childClientAccountId, pendingInvoice.InvoiceNo, 3)
			Expect(err).NotTo(HaveOccurred())
		}

	})
	It("Get pending child account invoice number after recalcualtion of credit", func() {
		resp, err := ariaInvoiceClient.GetPendingInvoiceNo(ctx, childClientAccountId)
		Expect(err).NotTo(HaveOccurred())
		pendingInvoices = resp.PendingInvoice
	})
	It("Approve pending child account invoice", func() {
		for _, pendingInvoice := range pendingInvoices {
			_, err := ariaInvoiceClient.ManagePendingInvoiceWithInoviceNo(ctx, childClientAccountId, pendingInvoice.InvoiceNo, 1)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("Get child account invoice history", func() {
		masterPlanId := client.GetClientMasterPlanInstanceId(GetCloudAcctIdFromClientAcctId(childClientAccountId), product.GetId())
		resp, err := ariaInvoiceClient.GetInvoiceHistory(ctx, childClientAccountId, masterPlanId, client.NO_START_DATE, client.NO_END_DATE)
		invoiceHists = resp.InvoiceHist
		Expect(err).NotTo(HaveOccurred())
	})

	It("Get child account invoice details", func() {
		for _, invoiceHist := range invoiceHists {
			masterPlanId := client.GetClientMasterPlanInstanceId(GetCloudAcctIdFromClientAcctId(childClientAccountId), product.GetId())
			_, err = ariaInvoiceClient.GetInvoiceDetails(ctx, childClientAccountId, invoiceHist.InvoiceNo, masterPlanId)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("Get statement for child account invoice", func() {
		for _, invoiceHist := range invoiceHists {
			_, err = ariaInvoiceClient.GetStatementForInvoice(ctx, childClientAccountId, invoiceHist.InvoiceNo)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("Get pending parent account invoice number", func() {
		resp, err := ariaInvoiceClient.GetPendingInvoiceNo(ctx, parentClientAccountId)
		Expect(err).NotTo(HaveOccurred())
		pendingInvoices = resp.PendingInvoice
	})

	It("Regenerate pending parent account invoice to recalcualte credit", func() {
		for _, pendingInvoice := range pendingInvoices {
			_, err = ariaInvoiceClient.ManagePendingInvoiceWithInoviceNo(ctx, parentClientAccountId, pendingInvoice.InvoiceNo, 3)
			Expect(err).NotTo(HaveOccurred())
		}

	})
	It("Get pending parent account invoice number after recalcualtion of credit", func() {
		resp, err := ariaInvoiceClient.GetPendingInvoiceNo(ctx, parentClientAccountId)
		Expect(err).NotTo(HaveOccurred())
		pendingInvoices = resp.PendingInvoice
	})
	It("Approve pending parent account invoice", func() {
		for _, pendingInvoice := range pendingInvoices {
			_, err := ariaInvoiceClient.ManagePendingInvoiceWithInoviceNo(ctx, parentClientAccountId, pendingInvoice.InvoiceNo, 1)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("Get parent account invoice history", func() {
		masterPlanId := client.GetClientMasterPlanInstanceId(GetCloudAcctIdFromClientAcctId(parentClientAccountId), product.GetId())
		resp, err := ariaInvoiceClient.GetInvoiceHistory(ctx, parentClientAccountId, masterPlanId, client.NO_START_DATE, client.NO_END_DATE)
		invoiceHists = resp.InvoiceHist
		Expect(err).NotTo(HaveOccurred())
	})

	It("Get parent account invoice details", func() {
		for _, invoiceHist := range invoiceHists {
			masterPlanId := client.GetClientMasterPlanInstanceId(GetCloudAcctIdFromClientAcctId(parentClientAccountId), product.GetId())
			_, err = ariaInvoiceClient.GetInvoiceDetails(ctx, parentClientAccountId, invoiceHist.InvoiceNo, masterPlanId)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("Get statement for parent account invoice", func() {
		for _, invoiceHist := range invoiceHists {
			_, err = ariaInvoiceClient.GetStatementForInvoice(ctx, parentClientAccountId, invoiceHist.InvoiceNo)
			Expect(err).NotTo(HaveOccurred())
		}
	})
})
