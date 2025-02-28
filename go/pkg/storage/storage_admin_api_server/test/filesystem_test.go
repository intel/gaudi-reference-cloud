package test

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	server "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_admin_api_server/pkg/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ = Describe("StorageAdminServiceClient", func() {
	Context("InsertStorageQuotaByAccount", func() {
		It("Should insert storage quota by account", func() {
			By("Creating a new storage quota")
			ctx := context.Background()
			req := &pb.InsertStorageQuotaByAccountRequest{
				CloudAccountId:    "111111111112",
				Reason:            "test",
				FilesizeQuotaInTB: 3,
				FilevolumesQuota:  3,
				BucketsQuota:      3,
			}
			// Call the InsertStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.InsertStorageQuotaByAccount(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
		})

		It("Should fail to insert storage quota by account", func() {
			By("Creating a new storage quota with invalid cloudaccount")
			ctx := context.Background()
			req := &pb.InsertStorageQuotaByAccountRequest{
				CloudAccountId:    "1111111111123",
				Reason:            "test",
				FilesizeQuotaInTB: 1,
				FilevolumesQuota:  1,
				BucketsQuota:      1,
			}
			// Call the InsertStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.InsertStorageQuotaByAccount(ctx, req)
			Expect(err).NotTo(BeNil())
			Expect(resp).To(BeNil())
		})

		It("Should fail to insert storage quota by account", func() {
			By("Creating a new storage quota with more than max allowed filesize")
			ctx := context.Background()
			req := &pb.InsertStorageQuotaByAccountRequest{
				CloudAccountId:    "111111111112",
				Reason:            "test",
				FilesizeQuotaInTB: 201,
				FilevolumesQuota:  1000,
				BucketsQuota:      3,
			}
			// Call the InsertStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.InsertStorageQuotaByAccount(ctx, req)
			Expect(err).NotTo(BeNil())
			Expect(resp).To(BeNil())
		})

		It("Should fail to insert storage quota by account", func() {
			By("Creating a new storage quota with more than max allowed volumes")
			ctx := context.Background()
			req := &pb.InsertStorageQuotaByAccountRequest{
				CloudAccountId:    "111111111112",
				Reason:            "test",
				FilesizeQuotaInTB: 3,
				FilevolumesQuota:  1000,
				BucketsQuota:      3,
			}
			// Call the InsertStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.InsertStorageQuotaByAccount(ctx, req)
			Expect(err).NotTo(BeNil())
			Expect(resp).To(BeNil())
		})

		It("Should fail cloudaccount validation when inserting storage quota by account", func() {
			By("Creating a new storage quota with more than max allowed buckets")
			ctx := context.Background()
			req := &pb.InsertStorageQuotaByAccountRequest{
				CloudAccountId:    "111111111112",
				Reason:            "test",
				FilesizeQuotaInTB: 1,
				FilevolumesQuota:  1,
				BucketsQuota:      11,
			}
			// Call the InsertStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.InsertStorageQuotaByAccount(ctx, req)
			Expect(err).NotTo(BeNil())
			Expect(resp).To(BeNil())
		})
	})

	Context("UpdateStorageQuotaByAccount", func() {
		It("Should update storage quota by account", func() {
			By("Updating an existing storage quota")
			ctx := context.Background()
			req := &pb.UpdateStorageQuotaByAccountRequest{
				CloudAccountId:    "111111111112",
				Reason:            "test",
				FilesizeQuotaInTB: 4,
				FilevolumesQuota:  4,
				BucketsQuota:      4,
			}
			// Call the UpdateStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.UpdateStorageQuotaByAccount(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
		})

		It("Should not update storage quota by account", func() {
			By("Updating an existing storage quota")
			ctx := context.Background()
			req := &pb.UpdateStorageQuotaByAccountRequest{
				CloudAccountId:    "111111111112",
				Reason:            "test",
				FilesizeQuotaInTB: 1,
				FilevolumesQuota:  1,
				BucketsQuota:      1,
			}
			// Call the UpdateStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.UpdateStorageQuotaByAccount(ctx, req)
			Expect(err).NotTo(BeNil())
			Expect(resp).To(BeNil())
		})
		It("Should fail to delete storage quota by account", func() {
			By("Deleting an existing storage quota")
			ctx := context.Background()
			req := &pb.DeleteStorageQuotaByAccountRequest{
				CloudAccountId: "111111111112",
			}
			// Call the DeleteStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			_, err := storageAdminSvc.DeleteStorageQuotaByAccount(ctx, req)
			// expect error due to function being deprecated
			Expect(err).NotTo(BeNil())
		})
	})

	It("Should fail cloudaccount validation when updating storage quota by account", func() {
		By("Updating an existing storage quota")
		ctx := context.Background()
		req := &pb.UpdateStorageQuotaByAccountRequest{
			CloudAccountId:    "1111111111123",
			Reason:            "test",
			FilesizeQuotaInTB: 1,
			FilevolumesQuota:  1,
			BucketsQuota:      1,
		}
		// Call the InsertStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
		resp, err := storageAdminSvc.UpdateStorageQuotaByAccount(ctx, req)
		Expect(err).NotTo(BeNil())
		Expect(resp).To(BeNil())
	})

})

var _ = Describe("StorageAdminServiceClient", func() {
	Context("GetStorageQuotaByAccount", func() {
		It("Should get storage quota by account", func() {
			By("Getting an existing storage quota")
			ctx := context.Background()
			req := &pb.GetStorageQuotaByAccountRequest{
				CloudAccountId: "111111111112",
			}
			// Call the GetStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.GetStorageQuotaByAccount(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
		})
	})

	Context("GetStorageQuotaByAccountFail", func() {
		It("Should fail cloudaccount validation when getting storage quota by account", func() {
			By("Getting an existing storage quota")
			ctx := context.Background()
			req := &pb.GetStorageQuotaByAccountRequest{
				CloudAccountId: "1111111111123",
			}
			// Call the GetStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.GetStorageQuotaByAccount(ctx, req)
			Expect(err).NotTo(BeNil())
			Expect(resp).To(BeNil())
		})
	})

	Context("GetAllStorageQuota", func() {
		It("Should get all storage quotas", func() {
			By("Getting all storage quotas")
			ctx := context.Background()
			req := &emptypb.Empty{}
			// Call the GetAllStorageQuota method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.GetAllStorageQuota(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
		})
	})
})

var _ = Describe("StorageAdminServiceClient", func() {
	Context("GetResourceUsage", func() {
		It("Should get resource usage for filesystems and buckets", func() {
			By("Getting an existing storage quota")
			ctx := context.Background()
			req := &empty.Empty{}
			// Call the GetStorageQuotaByAccount method of StorageAdminServiceClient and capture the response
			resp, err := storageAdminSvc.GetResourceUsage(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
			Expect(resp.FilesystemUsages[0].CloudAccountId).Should(Equal("123456789012"))
			Expect(resp.FilesystemUsages[0].TotalProvisioned).Should(Equal("14GB"))
			Expect(resp.BucketUsages[0].CloudAccountId).Should(Equal("123456789012"))
			Expect(resp.BucketUsages[0].BucketSize).Should(Equal("120GB"))
			Expect(resp.BucketUsages[0].UsedCapacity).Should(Equal("30GB"))
		})
	})

})

var _ = Describe("StorageAdminServiceClient", func() {
	Context("GetAccountType", func() {
		It("Should return corresponding cloudaccount type as string", func() {
			By("Getting an existing storage quota")
			req := pb.AccountType_ACCOUNT_TYPE_STANDARD
			res := server.GetAccountType(req)
			Expect(res).Should(Equal("STANDARD"))

			req = pb.AccountType_ACCOUNT_TYPE_PREMIUM
			res = server.GetAccountType(req)
			Expect(res).Should(Equal("PREMIUM"))

			req = pb.AccountType_ACCOUNT_TYPE_ENTERPRISE
			res = server.GetAccountType(req)
			Expect(res).Should(Equal("ENTERPRISE"))

			req = pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING
			res = server.GetAccountType(req)
			Expect(res).Should(Equal("ENTERPRISE_PENDING"))

			req = pb.AccountType_ACCOUNT_TYPE_INTEL
			res = server.GetAccountType(req)
			Expect(res).Should(Equal("INTEL"))
		})
	})

})
