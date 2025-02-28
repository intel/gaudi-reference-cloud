// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var serviceId string
var ruleId string

const (
	region = "us-dev-1"
)

var _ = Describe("QuotaManagementServiceClient", func() {
	Context("Bootstrap service", func() {
		It("Should bootstrap quota management service", func() {
			By("registering service resource and creating quotas")
			ctx := context.Background()
			err := quotaManagementSvc.BootstrapQuotaManagementService(ctx, bootstrappedConfig, region)
			Expect(err).To(BeNil())
		})
		It("Should bootstrap quota management service", func() {
			By("skipping service resources with invalid max limits")
			ctx := context.Background()
			err := quotaManagementSvc.BootstrapQuotaManagementService(ctx, bootstrappedConfigInvalid, region)
			Expect(err).To(BeNil())
		})
		It("Should fail to bootstrap quota management service", func() {
			By("skipping service resources with invalid resources")
			ctx := context.Background()
			bootstrappedConfigInvalid[0].QuotaAccountType["ENTERPRISE"].DefaultLimits["storage"]["invalid-resource"] = 8
			err := quotaManagementSvc.BootstrapQuotaManagementService(ctx, bootstrappedConfigInvalid, region)
			Expect(err).ToNot(BeNil())
		})
		It("Should list all service resource quotas for all services", func() {
			ctx := context.Background()
			req := &pb.ListAllServiceQuotaRequest{
				Filters: map[string]string{},
			}
			resp, err := quotaManagementSvc.ListAllServiceQuotas(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
			Expect(len(resp.AllServicesQuotaResponse)).To(BeNumerically(">", 0)) // since 4 quotas in total for 2 services
			Expect(resp.AllServicesQuotaResponse[0].ServiceName).NotTo(BeEmpty())
			Expect(resp.AllServicesQuotaResponse[0].ServiceQuotaResource[0].Scope.ScopeValue).NotTo(BeEmpty())
			Expect(resp.AllServicesQuotaResponse[2].ServiceName).NotTo(BeEmpty())
			Expect(resp.AllServicesQuotaResponse[3].ServiceQuotaResource[0].Scope.ScopeValue).NotTo(BeEmpty())
		})
		It("Should fail to bootstrap quota management service with updates", func() {
			By("by using an existing invalid resource")
			ctx := context.Background()
			err := quotaManagementSvc.BootstrapQuotaManagementService(ctx, bootstrappedConfig, region)
			Expect(err).ToNot(BeNil())
		})
	})
	Context("RegisterService", func() {
		It("Should register a new service and it's resource limits", func() {
			By("Registering a new service")
			ctx := context.Background()
			req := &pb.ServiceQuotaRegistrationRequest{
				ServiceName: "storage-test",
				Region:      region,
				ServiceResources: []*pb.ServiceResource{
					{
						Name:      "filesystems",
						QuotaUnit: "COUNT",
						MaxLimit:  10,
					},
					{
						Name:      "buckets",
						QuotaUnit: "COUNT",
						MaxLimit:  10,
					},
				},
			}
			resp, err := quotaManagementSvc.Register(ctx, req)
			serviceId = resp.ServiceId
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
		})
		It("Should fail to register a new service", func() {
			By("Registering a new service")
			ctx := context.Background()
			req := &pb.ServiceQuotaRegistrationRequest{
				ServiceName: "storage-test",
				Region:      region,
				ServiceResources: []*pb.ServiceResource{
					{
						Name:      "filesystems-invalid",
						QuotaUnit: "COUNTING",
						MaxLimit:  10,
					},
					{
						Name:      "buckets-invalid",
						QuotaUnit: "COUNT",
						MaxLimit:  10,
					},
				},
			}
			resp, err := quotaManagementSvc.Register(ctx, req)
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})
	})
	It("Should list registered services", func() {
		By("Listing all registered services")
		ctx := context.Background()
		req := &pb.ServicesListRequest{
			Filters: map[string]string{},
		}
		resp, err := quotaManagementSvc.ListRegisteredServices(ctx, req)
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())

	})
	It("Should list all service resources", func() {
		By("Listing all resources for given service Id")
		ctx := context.Background()
		req := &pb.ServiceResourcesListRequest{
			ServiceId: serviceId,
		}
		resp, err := quotaManagementSvc.ListServiceResources(ctx, req)
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())
		Expect(len(resp.ServiceResources)).To(BeEquivalentTo(2))

	})
	It("Should fail to list all service resources", func() {
		By("Listing all resources for invalid service Id")
		ctx := context.Background()
		req := &pb.ServiceResourcesListRequest{
			ServiceId: "invalid-service-id",
		}
		resp, err := quotaManagementSvc.ListServiceResources(ctx, req)
		Expect(err).ToNot(BeNil())
		Expect(resp).To(BeNil())
	})
	It("Should create a new service resource quota", func() {
		By("Creating a new service resource quotas")
		ctx := context.Background()
		req := &pb.CreateServiceQuotaRequest{
			ServiceId: serviceId,
			ServiceQuotaResource: &pb.ServiceQuotaResource{
				ResourceType: "filesystems",
				QuotaConfig:  &pb.QuotaConfig{QuotaUnit: "COUNT", Limits: 10},
				Scope:        &pb.QuotaScope{ScopeType: "QUOTA_ACCOUNT_TYPE", ScopeValue: "PREMIUM"},
				RuleId:       "test-rule",
				Reason:       "testing",
				CreatedTime:  timestamppb.New(time.Now()),
				UpdatedTime:  timestamppb.New(time.Now()),
			},
		}
		resp, err := quotaManagementSvc.CreateServiceQuota(ctx, req)
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())
	})
	It("Should fail to create a new service resource quota", func() {
		By("Creating a new service resource quotas with empty reason")
		ctx := context.Background()
		req := &pb.CreateServiceQuotaRequest{
			ServiceId: serviceId,
			ServiceQuotaResource: &pb.ServiceQuotaResource{
				ResourceType: "filesystems",
				QuotaConfig:  &pb.QuotaConfig{QuotaUnit: "COUNT", Limits: 10},
				Scope:        &pb.QuotaScope{ScopeType: "QUOTA_ACCOUNT_TYPE", ScopeValue: "PREMIUM"},
				RuleId:       "test-rule",
				Reason:       "",
				CreatedTime:  timestamppb.New(time.Now()),
				UpdatedTime:  timestamppb.New(time.Now()),
			},
		}
		resp, err := quotaManagementSvc.CreateServiceQuota(ctx, req)
		Expect(err).ToNot(BeNil())
		Expect(resp).To(BeNil())
	})
	It("Should fail to create a new service resource quota", func() {
		By("Creating a new service resource quotas with invalid max limit")
		ctx := context.Background()
		req := &pb.CreateServiceQuotaRequest{
			ServiceId: serviceId,
			ServiceQuotaResource: &pb.ServiceQuotaResource{
				ResourceType: "filesystems",
				QuotaConfig:  &pb.QuotaConfig{QuotaUnit: "COUNT", Limits: 10000},
				Scope:        &pb.QuotaScope{ScopeType: "QUOTA_ACCOUNT_TYPE", ScopeValue: "PREMIUM"},
				RuleId:       "test-rule",
				Reason:       "test reason",
				CreatedTime:  timestamppb.New(time.Now()),
				UpdatedTime:  timestamppb.New(time.Now()),
			},
		}
		resp, err := quotaManagementSvc.CreateServiceQuota(ctx, req)
		Expect(err).ToNot(BeNil())
		Expect(resp).To(BeNil())
	})
	It("Should first update and then do a service resource quota for a given resource type", func() {
		By("Update and get service resource quotas")
		ctx := context.Background()
		req := &pb.ServiceResourceRequest{
			ServiceId:    serviceId,
			ResourceName: "filesystems",
		}
		resp, err := quotaManagementSvc.GetServiceResource(ctx, req)
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())
		Expect(resp.ServiceResources.Name).To(BeEquivalentTo("filesystems"))

		updReq := &pb.UpdateServiceRegistrationRequest{
			ServiceId: serviceId,
			ServiceResources: []*pb.ServiceResource{{
				Name:      "filesystems",
				QuotaUnit: "COUNT",
				MaxLimit:  15,
			}},
		}
		updateResp, err := quotaManagementSvc.UpdateServiceRegistration(ctx, updReq)
		Expect(err).To(BeNil())
		Expect(updateResp).NotTo(BeNil())
		Expect(updateResp.ServiceResources[0].MaxLimit).To(BeEquivalentTo(15))
	})
	It("Should fail to create a new service resource quota", func() {
		By("Creating a new service resource quotas with invalid quota unit")
		ctx := context.Background()
		req := &pb.CreateServiceQuotaRequest{
			ServiceId: serviceId,
			ServiceQuotaResource: &pb.ServiceQuotaResource{
				ResourceType: "filesystems",
				QuotaConfig:  &pb.QuotaConfig{QuotaUnit: "COUNTING", Limits: 10},
				Scope:        &pb.QuotaScope{ScopeType: "QUOTA_ACCOUNT_TYPE", ScopeValue: "PREMIUM"},
				RuleId:       "test-rule",
				Reason:       "testing",
				CreatedTime:  timestamppb.New(time.Now()),
				UpdatedTime:  timestamppb.New(time.Now()),
			},
		}
		resp, err := quotaManagementSvc.CreateServiceQuota(ctx, req)
		Expect(err).ToNot(BeNil())
		Expect(resp).To(BeNil())
	})
	It("Should get service resource quota", func() {
		By("Getting a new service resource quotas")
		ctx := context.Background()
		req := &pb.ServiceQuotaResourceRequest{
			ServiceId:    serviceId,
			ResourceType: "filesystems",
		}
		resp, err := quotaManagementSvc.GetServiceQuotaResource(ctx, req)
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())
		ruleId = resp.ServiceQuotaResource[0].RuleId
	})
	It("Should update service resource quota", func() {
		By("Updating a new service resource quotas")
		ctx := context.Background()
		req := &pb.UpdateQuotaServiceRequest{
			ServiceId:    serviceId,
			ResourceType: "filesystems",
			RuleId:       ruleId,
			QuotaConfig: &pb.QuotaConfig{
				Limits:    12,
				QuotaUnit: "COUNT",
			},
			Reason: "test reason",
		}
		resp, err := quotaManagementSvc.UpdateServiceQuotaResource(ctx, req)
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())
	})
	It("Should fail to update service resource quota", func() {
		By("Updating a service resource quotas with invalid resource")
		ctx := context.Background()
		req := &pb.UpdateQuotaServiceRequest{
			ServiceId:    serviceId,
			ResourceType: "filesystems-invalid",
			RuleId:       ruleId,
			QuotaConfig: &pb.QuotaConfig{
				Limits:    12,
				QuotaUnit: "COUNT",
			},
			Reason: "test reason",
		}
		resp, err := quotaManagementSvc.UpdateServiceQuotaResource(ctx, req)
		Expect(err).ToNot(BeNil())
		Expect(resp).To(BeNil())
	})
	It("Should fail to update service resource quota", func() {
		By("Updating a service resource quotas with invalid limit")
		ctx := context.Background()
		req := &pb.UpdateQuotaServiceRequest{
			ServiceId:    serviceId,
			ResourceType: "filesystems",
			RuleId:       ruleId,
			QuotaConfig: &pb.QuotaConfig{
				Limits:    1000,
				QuotaUnit: "COUNT",
			},
			Reason: "test reason",
		}
		resp, err := quotaManagementSvc.UpdateServiceQuotaResource(ctx, req)
		Expect(err).ToNot(BeNil())
		Expect(resp).To(BeNil())
	})
	It("Should list all service resource quota", func() {
		ctx := context.Background()
		req := &pb.ListServiceQuotaRequest{
			ServiceId: serviceId,
		}
		resp, err := quotaManagementSvc.ListServiceQuota(ctx, req)
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())
	})
	It("Should fail to list all service resource quota", func() {
		ctx := context.Background()
		req := &pb.ListServiceQuotaRequest{
			ServiceId: "invalid-service-id",
		}
		resp, err := quotaManagementSvc.ListServiceQuota(ctx, req)
		Expect(err).To(BeNil())
		Expect(resp).ToNot(BeNil())
	})
	It("Should list all service resource quotas for all services", func() {
		ctx := context.Background()
		req := &pb.ListAllServiceQuotaRequest{
			Filters: map[string]string{},
		}
		resp, err := quotaManagementSvc.ListAllServiceQuotas(ctx, req)
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())
	})
	It("Should delete service resource quota", func() {
		By("Deleting service resource quota")
		ctx := context.Background()
		req := &pb.ListServiceQuotaRequest{
			ServiceId: serviceId,
		}
		respList, err := quotaManagementSvc.ListServiceQuota(ctx, req)
		Expect(err).To(BeNil())
		Expect(respList).NotTo(BeNil())

		for _, quota := range respList.ServiceQuotaAllResources {
			reqDelete := &pb.DeleteServiceQuotaRequest{
				ServiceId:    serviceId,
				ResourceType: quota.ResourceType,
				RuleId:       quota.RuleId,
			}
			_, err := quotaManagementSvc.DeleteServiceQuotaResource(ctx, reqDelete)
			Expect(err).To(BeNil())
		}
	})
	It("Should fail to delete service resource quota", func() {
		By("Deleting service resource quota with invalid resource")
		ctx := context.Background()
		req := &pb.ListServiceQuotaRequest{
			ServiceId: serviceId,
		}
		respList, err := quotaManagementSvc.ListServiceQuota(ctx, req)
		Expect(err).To(BeNil())
		Expect(respList).NotTo(BeNil())

		for _, quota := range respList.ServiceQuotaAllResources {
			reqDelete := &pb.DeleteServiceQuotaRequest{
				ServiceId:    serviceId,
				ResourceType: "filesystems-invalid",
				RuleId:       quota.RuleId,
			}
			_, err := quotaManagementSvc.DeleteServiceQuotaResource(ctx, reqDelete)
			Expect(err).ToNot(BeNil())
		}
	})
	It("Should delete service, resources and quota", func() {
		By("Deleting service and all it's resources and quotas")
		ctx := context.Background()
		// Adding a resource quota since previous one was deleted
		reqCreate := &pb.CreateServiceQuotaRequest{
			ServiceId: serviceId,
			ServiceQuotaResource: &pb.ServiceQuotaResource{
				ResourceType: "filesystems",
				QuotaConfig:  &pb.QuotaConfig{QuotaUnit: "COUNT", Limits: 10},
				Scope:        &pb.QuotaScope{ScopeType: "QUOTA_ACCOUNT_TYPE", ScopeValue: "STANDARD"},
				RuleId:       "test-rule",
				Reason:       "testing",
				CreatedTime:  timestamppb.New(time.Now()),
				UpdatedTime:  timestamppb.New(time.Now()),
			},
		}
		resp, err := quotaManagementSvc.CreateServiceQuota(ctx, reqCreate)
		Expect(err).To(BeNil())
		Expect(resp).ToNot(BeNil())

		req := &pb.DeleteServiceRequest{
			ServiceId: serviceId,
		}
		_, err = quotaManagementSvc.DeleteService(ctx, req)
		Expect(err).To(BeNil())

		reqList := &pb.ServiceResourcesListRequest{
			ServiceId: serviceId,
		}
		respList, err := quotaManagementSvc.ListServiceResources(ctx, reqList)
		Expect(err).ToNot(BeNil())
		Expect(respList).To(BeNil())
	})
	It("Should fail to delete service, resources and quota", func() {
		By("Deleting service that does not exist")
		ctx := context.Background()
		// Adding a resource quota since previous one was deleted
		reqCreate := &pb.CreateServiceQuotaRequest{
			ServiceId: "invalid-service-id",
			ServiceQuotaResource: &pb.ServiceQuotaResource{
				ResourceType: "filesystems",
				QuotaConfig:  &pb.QuotaConfig{QuotaUnit: "COUNT", Limits: 10},
				Scope:        &pb.QuotaScope{ScopeType: "QUOTA_ACCOUNT_TYPE", ScopeValue: "STANDARD"},
				RuleId:       "test-rule",
				Reason:       "testing",
				CreatedTime:  timestamppb.New(time.Now()),
				UpdatedTime:  timestamppb.New(time.Now()),
			},
		}
		resp, err := quotaManagementSvc.CreateServiceQuota(ctx, reqCreate)
		Expect(err).ToNot(BeNil())
		Expect(resp).To(BeNil())

		req := &pb.DeleteServiceRequest{
			ServiceId: serviceId,
		}
		_, err = quotaManagementSvc.DeleteService(ctx, req)
		Expect(err).ToNot(BeNil())

		reqList := &pb.ServiceResourcesListRequest{
			ServiceId: serviceId,
		}
		respList, err := quotaManagementSvc.ListServiceResources(ctx, reqList)
		Expect(err).ToNot(BeNil())
		Expect(respList).To(BeNil())
	})
})
