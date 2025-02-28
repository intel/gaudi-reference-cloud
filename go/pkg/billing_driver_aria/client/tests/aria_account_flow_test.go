// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create Account Flow", Ordered, Label("account"), func() {
	BeforeEach(func() {
		if !ShouldRunTest("account") {
			Skip("Create Account Flow Test")
		}
	})
	ctx := context.Background()
	clientAccountId := GetClientAccountId()
	ariaAccountClient := common.GetAriaAccountClient()
	var usageType *data.UsageType
	var err error
	product := GetProduct()
	productFamily := GetProductFamily()
	ariaPlanClient := common.GetAriaPlanClient()
	It("Create and Get Usage Type", func() {
		ctx := context.Background()
		usageTypeClient := common.GetAriaUsageTypeClient()
		usageType, err = usageTypeClient.GetMinutesUsageType(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(usageType.UsageUnitType).To(Equal("minute"), "Usage Type")
	})

	It("Create Master Plan", func() {
		createRespBody, err := ariaPlanClient.CreatePlan(ctx, product, productFamily, usageType)
		Expect(err).NotTo(HaveOccurred())
		Expect(createRespBody.GetErrorCode()).To(Equal(int64(0)), "Master plan created")
	})
	It("Ensure promo plan set and add client plan", func() {
		ariaPromoClient := common.GetPromoClient()
		err = ariaPromoClient.EnsurePlanSet(ctx)
		Expect(err).NotTo(HaveOccurred())
		err = ariaPromoClient.AddPlansToPromo(ctx, []string{GetTestClientPlanId(product.Id)})
		Expect(err).NotTo(HaveOccurred())
	})

	It("Create Account", func() {
		resp, err := ariaAccountClient.CreateAriaAccount(ctx, clientAccountId, client.GetPlanClientId(product.GetId()), client.ACCOUNT_TYPE_PREMIUM)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.GetErrorCode()).To(Equal(int64(0)), "Account created")
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

	It("Account details is valid", func() {
		details, err := ariaAccountClient.GetAriaAccountDetailsAllForClientId(ctx, clientAccountId)
		Expect(err).NotTo(HaveOccurred())
		Expect(client.IsPayloadEmpty(details)).To(Equal(false), "account detail payload check")
		Expect(client.IsPayloadEmpty(details.MasterPlansInfo)).To(Equal(false), "Master Plan Instance")
		Expect(client.IsPayloadEmpty(details.MasterPlansInfo[0].MasterPlansServices)).To(Equal(false), "Master Plan Services")
		Expect(details.MasterPlansInfo[0].ClientBillingGroupId).To(Equal(client.GetBillingGroupId(clientAccountId)), "Billing Group id")
		Expect(details.MasterPlansInfo[0].ClientDunningGroupId).To(Equal(client.GetDunningGroupId(clientAccountId)), "Dunning Group id")
		Expect(client.IsPayloadEmpty(details.SuppField)).To(Equal(false), "account detail supp field")
	})

	It("Account Master Plan and Service is valid", func() {
		clientPlanId := GetTestClientPlanId(product.GetId())
		clientPlanResponse, err := ariaPlanClient.GetAriaPlanDetailsAllForClientPlanId(ctx, clientPlanId)
		Expect(err).NotTo(HaveOccurred())
		Expect(client.IsPayloadEmpty(clientPlanResponse)).To(Equal(false), "account master plan")
		Expect(clientPlanResponse.AllClientPlanDtls[0].ClientPlanId).To(Equal(clientPlanId), "Client Plan Id")
		Expect(clientPlanResponse.AllClientPlanDtls[0].PlanServices[0].UsageType).To(Equal(int64(usageType.UsageTypeNo)), "Usages Type")
		Expect(clientPlanResponse.AllClientPlanDtls[0].PlanServices[0].IsUsageBasedInd).To(Equal(int64(1)), "Usages Based Service")
		Expect(clientPlanResponse.AllClientPlanDtls[0].PromotionalPlanSets[0].ClientPromoSetId).To(Equal(client.GetPlanSetId()), "Client Promo Set Id")
	})
	It("Rate is valid", func() {
		clientPlanId := GetTestClientPlanId(product.GetId())
		clientServiceId := client.GetServiceClientId(productFamily.GetId())
		rateResp, err := ariaPlanClient.GetClientPlanServiceRates(ctx, clientPlanId, clientServiceId)
		Expect(err).NotTo(HaveOccurred())
		Expect(rateResp.PlanServiceRates[0].ClientRateScheduleId).To(Equal(clientPlanId+".premium"), "Client Rate Schedule Id")
		Expect(rateResp.PlanServiceRates[0].FromUnit).To(Equal(float32(1)), "from_unit")
	})

	//TODO: assign plan and test
})
