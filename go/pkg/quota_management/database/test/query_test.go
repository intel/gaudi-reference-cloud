// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quota_management/database/query"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	ctx = context.Background()
	err error
)
var _ = Describe("Query", func() {
	BeforeEach(func() {
		txDb, err = sqlDb.Begin()
		Expect(err).To(BeNil())
		Expect(txDb).NotTo(BeNil())

	})
	AfterEach(func() {
		txDb.Commit()
		// txDb.Rollback()
	})

	Context("Insert Service for QMS", func() {
		It("Should succeed", func() {
			serviceId := uuid.NewString()
			serviceName := "test-service-1"
			region := "us-dev-1"

			By("Inserting new service in DB")
			registerService, err := query.InsertService(ctx, txDb, serviceId, serviceName, region)
			Expect(err).To(BeNil())
			Expect(registerService.ServiceId).To(BeEquivalentTo(serviceId))
			Expect(registerService.ServiceName).To(BeEquivalentTo(serviceName))

			By("Get registered service from DB")
			registeredService, err := query.GetRegisteredService(ctx, txDb, serviceId)
			Expect(err).To(BeNil())
			Expect(registeredService.ServiceId).To(BeEquivalentTo(serviceId))
			Expect(registeredService.ServiceName).To(BeEquivalentTo(serviceName))
			Expect(registeredService.Region).To(BeEquivalentTo(region))

			By("Inserting service resources and limits for a new service in DB")
			resourceName := "test-resource-1"
			quotaUnit := "COUNT"
			maxLimit := int64(10)
			registerServiceResource, err := query.InsertServiceResource(ctx, txDb, serviceId, resourceName, quotaUnit, maxLimit)
			Expect(err).To(BeNil())
			Expect(registerServiceResource.Name).To(BeEquivalentTo(resourceName))
			Expect(registerServiceResource.QuotaUnit).To(BeEquivalentTo(quotaUnit))
			Expect(registerServiceResource.MaxLimit).To(BeEquivalentTo(maxLimit))

			By("Getting service resources and limits for a new service in DB")
			registeredServiceResource, err := query.GetServiceResource(ctx, txDb, serviceId, resourceName)
			Expect(err).To(BeNil())
			Expect(registeredServiceResource.ResourceName).To(BeEquivalentTo(resourceName))
			Expect(registeredServiceResource.QuotaUnit).To(BeEquivalentTo(quotaUnit))
			Expect(registeredServiceResource.MaxLimit).To(BeEquivalentTo(maxLimit))

			By("Inserting service resource quotas for new service in DB")
			ruleId := uuid.NewString()
			scopeType := string(query.ScopeAccountType)
			scopeValue := string("PREMIUM")
			reason := "unit testing"
			serviceResourceQuota, err := query.InsertServiceResourceQuota(
				ctx, txDb, serviceId, resourceName, ruleId,
				maxLimit, quotaUnit,
				scopeType, scopeValue, reason)
			Expect(err).To(BeNil())
			Expect(serviceResourceQuota.ServiceId).To(BeEquivalentTo(serviceId))
			Expect(serviceResourceQuota.ResourceName).To(BeEquivalentTo(resourceName))
			Expect(serviceResourceQuota.Limits).To(BeEquivalentTo(maxLimit))
			Expect(serviceResourceQuota.QuotaScope).To(BeEquivalentTo(scopeType))
			Expect(serviceResourceQuota.QuotaScopeValue).To(BeEquivalentTo(scopeValue))
			Expect(serviceResourceQuota.Reason).To(BeEquivalentTo(reason))

		})
	})
})
