// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ = Describe("QuotaManagementPrivate", func() {
	Context("GetQuotaForCloudAccount", func() {
		It("Should get quota for a cloudaccount", func() {
			By("Using cloudaccount ID")
			ctx := context.Background()
			svcName := "test-svc-" + time.Now().Format("02030405")
			// register service
			req := &pb.ServiceQuotaRegistrationRequest{
				ServiceName: svcName,
				Region:      "us-region-1",
				ServiceResources: []*pb.ServiceResource{
					{
						Name:      "filesystems",
						QuotaUnit: "COUNT",
						MaxLimit:  12,
					},
					{
						Name:      "buckets",
						QuotaUnit: "COUNT",
						MaxLimit:  15,
					},
				},
			}
			resp, err := quotaManagementSvc.Register(ctx, req)
			serviceId = resp.ServiceId
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

			// create service resource quota
			reqRes := &pb.CreateServiceQuotaRequest{
				ServiceId: serviceId,
				ServiceQuotaResource: &pb.ServiceQuotaResource{
					ResourceType: "filesystems",
					QuotaConfig:  &pb.QuotaConfig{QuotaUnit: "COUNT", Limits: 10},
					Scope:        &pb.QuotaScope{ScopeType: "QUOTA_ACCOUNT_ID", ScopeValue: "123456789012"},
					RuleId:       "test-rule",
					Reason:       "testing",
					CreatedTime:  timestamppb.New(time.Now()),
					UpdatedTime:  timestamppb.New(time.Now()),
				},
			}
			respRes, err := quotaManagementSvc.CreateServiceQuota(ctx, reqRes)
			Expect(err).To(BeNil())
			Expect(respRes).NotTo(BeNil())

			// create service resource quota with account type set
			reqRes = &pb.CreateServiceQuotaRequest{
				ServiceId: serviceId,
				ServiceQuotaResource: &pb.ServiceQuotaResource{
					ResourceType: "filesystems",
					QuotaConfig:  &pb.QuotaConfig{QuotaUnit: "COUNT", Limits: 4},
					Scope:        &pb.QuotaScope{ScopeType: "QUOTA_ACCOUNT_TYPE", ScopeValue: pb.AccountType_ACCOUNT_TYPE_STANDARD.String()},
					RuleId:       "test-rule",
					Reason:       "testing",
					CreatedTime:  timestamppb.New(time.Now()),
					UpdatedTime:  timestamppb.New(time.Now()),
				},
			}
			respRes, err = quotaManagementSvc.CreateServiceQuota(ctx, reqRes)
			Expect(err).To(BeNil())
			Expect(respRes).NotTo(BeNil())

			reqRes = &pb.CreateServiceQuotaRequest{
				ServiceId: serviceId,
				ServiceQuotaResource: &pb.ServiceQuotaResource{
					ResourceType: "buckets",
					QuotaConfig:  &pb.QuotaConfig{QuotaUnit: "COUNT", Limits: 11},
					Scope:        &pb.QuotaScope{ScopeType: "QUOTA_ACCOUNT_ID", ScopeValue: "123456789012"},
					RuleId:       "test-rule",
					Reason:       "testing",
					CreatedTime:  timestamppb.New(time.Now()),
					UpdatedTime:  timestamppb.New(time.Now()),
				},
			}

			respRes, err = quotaManagementSvc.CreateServiceQuota(ctx, reqRes)
			Expect(err).To(BeNil())
			Expect(respRes).NotTo(BeNil())

			// fetch using service name, resource type and cloudaccount ID
			reqPvt := &pb.ServiceQuotaResourceRequestPrivate{
				ServiceName:    svcName,
				ResourceType:   "filesystems",
				CloudAccountId: "123456789012",
			}
			respPvt, err := quotaManagementSvc.GetResourceQuotaPrivate(ctx, reqPvt)
			Expect(err).To(BeNil())
			Expect(respPvt.CustomQuota).NotTo(BeNil())
			Expect(respPvt.CustomQuota.ServiceResources).NotTo(BeNil())
			Expect(len(respPvt.CustomQuota.ServiceResources)).To(Equal(1))
			Expect(respPvt.CustomQuota.ServiceResources[0].QuotaConfig.Limits).To(Equal(int64(10)))
			// validate values for default quota
			Expect(respPvt.DefaultQuota).NotTo(BeNil())
			Expect(respPvt.DefaultQuota.ServiceResources).NotTo(BeNil())
			Expect(len(respPvt.DefaultQuota.ServiceResources)).To(Equal(1))
			Expect(respPvt.DefaultQuota.ServiceResources[0].QuotaConfig.Limits).To(Equal(int64(4)))

			// fetch using service name, resource type only
			reqPvt = &pb.ServiceQuotaResourceRequestPrivate{
				ServiceName:    svcName,
				ResourceType:   "filesystems",
				CloudAccountId: "",
			}
			respPvt, err = quotaManagementSvc.GetResourceQuotaPrivate(ctx, reqPvt)
			Expect(err).To(BeNil())
			Expect(respPvt.CustomQuota).NotTo(BeNil())
			Expect(respPvt.CustomQuota.ServiceResources).NotTo(BeNil())
			Expect(len(respPvt.CustomQuota.ServiceResources)).To(Equal(2))

			// fetch using service name only
			reqPvt = &pb.ServiceQuotaResourceRequestPrivate{
				ServiceName:    svcName,
				ResourceType:   "",
				CloudAccountId: "",
			}
			respPvt, err = quotaManagementSvc.GetResourceQuotaPrivate(ctx, reqPvt)
			Expect(err).To(BeNil())
			Expect(respPvt.CustomQuota).NotTo(BeNil())
			Expect(respPvt.CustomQuota.ServiceResources).NotTo(BeNil())
			Expect(len(respPvt.CustomQuota.ServiceResources)).To(Equal(3))
			Expect(respPvt.DefaultQuota.ServiceResources).NotTo(BeNil())
			Expect(len(respPvt.DefaultQuota.ServiceResources)).To(Equal(3))

			// fetch using service name and cloudaccount ID
			reqPvt = &pb.ServiceQuotaResourceRequestPrivate{
				ServiceName:    svcName,
				ResourceType:   "",
				CloudAccountId: "123456789012",
			}
			respPvt, err = quotaManagementSvc.GetResourceQuotaPrivate(ctx, reqPvt)
			Expect(err).To(BeNil())
			Expect(respPvt.CustomQuota).NotTo(BeNil())
			Expect(len(respPvt.CustomQuota.ServiceResources)).To(Equal(2))
			Expect(respPvt.DefaultQuota).NotTo(BeNil())
			Expect(respPvt.DefaultQuota.ServiceResources[0].QuotaConfig.Limits).To(Equal(int64(4)))

			// fetch using invalid resource type
			reqPvt = &pb.ServiceQuotaResourceRequestPrivate{
				ServiceName:    svcName,
				ResourceType:   "invalid-resource",
				CloudAccountId: "",
			}
			respPvt, err = quotaManagementSvc.GetResourceQuotaPrivate(ctx, reqPvt)
			Expect(err).NotTo(BeNil())
			Expect(respPvt).To(BeNil())

			// Test for fetching quota by cloudaccount type for a valid cloudaccount, if no quota exists for resource type
			// fetch using invalid resource type
			reqPvt = &pb.ServiceQuotaResourceRequestPrivate{
				ServiceName:    svcName,
				ResourceType:   "",
				CloudAccountId: "123456789012",
			}
			respPvt, err = quotaManagementSvc.GetResourceQuotaPrivate(ctx, reqPvt)
			Expect(err).To(BeNil())
			Expect(respPvt.CustomQuota).NotTo(BeNil())
			Expect(respPvt.DefaultQuota).NotTo(BeNil())

			// Test for fetching quota by cloudaccount type for a invalid cloudaccount, if no quota exists for resource type
			// fetch using invalid resource type
			reqPvt = &pb.ServiceQuotaResourceRequestPrivate{
				ServiceName:    svcName,
				ResourceType:   "",
				CloudAccountId: "123456789013",
			}
			respPvt, err = quotaManagementSvc.GetResourceQuotaPrivate(ctx, reqPvt)
			Expect(err).NotTo(BeNil())
			Expect(respPvt).To(BeNil())

			// Ping test
			pingResp, err := quotaManagementSvc.PingPrivate(ctx, &emptypb.Empty{})
			Expect(err).To(BeNil())
			Expect(pingResp).NotTo(BeNil())

			// fetch using invalid service name
			reqPvt = &pb.ServiceQuotaResourceRequestPrivate{
				ServiceName:    svcName + "invalid",
				ResourceType:   "",
				CloudAccountId: "",
			}
			respPvt, err = quotaManagementSvc.GetResourceQuotaPrivate(ctx, reqPvt)
			Expect(err).ToNot(BeNil())
			Expect(respPvt).To(BeNil())
		})
	})
})
