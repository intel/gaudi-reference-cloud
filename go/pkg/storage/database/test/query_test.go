// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database/query"
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

	Context("Get filesystem", func() {
		It("Should succeed", func() {
			cloudAcc := fsPrivate.Metadata.CloudAccountId

			By("Storing fs in DB")
			err := query.StoreFilesystemRequest(ctx, txDb, fsPrivate)
			Expect(err).To(BeNil())

			By("Storing fs account in DB")
			err = query.StoreFilesystemAccount(ctx, txDb, cloudAcc, fsSchedule)
			Expect(err).To(BeNil())

			By("Getting fs by cloudaccountid")
			dt := "2021-12-16T00:00:00Z"
			fsType := pb.FilesystemType_ComputeGeneral
			res, err := query.GetFilesystemsByCloudaccountId(ctx, txDb, cloudAcc, fsType, dt)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())

			By("Getting quota by cloudccountid")
			quota, err2 := query.GetUsedQuotaByAccountId(ctx, txDb, cloudAcc)
			Expect(err2).To(BeNil())
			Expect(quota).NotTo(BeNil())

			By("Getting fs accounts by cloudccountid")
			out, err3 := query.GetFilesystemAccounts(ctx, txDb, cloudAcc, dt)
			Expect(err3).To(BeNil())
			Expect(out).NotTo(BeNil())

			rs := NewMockFilesystemPrivateService_SearchFilesystemRequestsServer()

			err4 := query.GetFilesystemsRequests(txDb, "1", dt, rs)
			Expect(err4).To(BeNil())

		})
		It("Should not succeed", func() {
			By("Storing fs in DB")
			err := query.StoreFilesystemRequest(ctx, txDb, fsPrivate)
			Expect(err).NotTo(BeNil())

			By("Retrieving fs by resourceId")
			rid := fsPrivate.Metadata.ResourceId
			cloudAcc := fsPrivate.Metadata.CloudAccountId
			dt := "2021-12-16 00:00:00"
			res2, err := query.GetFilesystemByResourceId(ctx, txDb, cloudAcc, rid, dt)
			Expect(err).NotTo(BeNil())
			Expect(res2).To(BeNil())

			By("Getting fs with invalid cloudaccountid")
			dt = ""
			fsType := pb.FilesystemType_ComputeGeneral
			res, err := query.GetFilesystemsByCloudaccountId(ctx, txDb, "123456789000", fsType, dt)
			Expect(err).NotTo(BeNil())
			Expect(res).To(BeNil())

			By("Retrieving fs by name")
			name := "test"
			res3, err := query.GetFilesystemByName(ctx, txDb, cloudAcc, name, dt)
			Expect(err).NotTo(BeNil())
			Expect(res3).To(BeNil())

			cloudAcc = "123"

			By("Getting fs accounts with incorrect deletion time")
			out, err3 := query.GetFilesystemAccounts(ctx, txDb, cloudAcc, "")
			Expect(err3).NotTo(BeNil())
			Expect(out).To(BeNil())
			rs := NewMockFilesystemPrivateService_SearchFilesystemRequestsServer()

			err4 := query.GetFilesystemsRequests(txDb, "", "", rs)
			Expect(err4).NotTo(BeNil())
		})
	})
	Context("Update", func() {
		It("Should succeed", func() {
			cloudAcc := fsPrivate.Metadata.CloudAccountId
			err2 := query.UpdateFilesystemDeletionTime(ctx, txDb, cloudAcc, "test")
			Expect(err2).To(BeNil())

			err3 := query.UpdateFilesystemState(ctx, txDb, fsUpdate)
			Expect(err3).NotTo(BeNil())

		})
		It("Should fail and throw error", func() {
			By("Updating fs that dont exist")
			name := "test"
			cloudAcc := fsPrivate.Metadata.CloudAccountId
			err2 := query.UpdateFilesystemDeletionTime(ctx, txDb, cloudAcc, name)
			Expect(err2).NotTo(BeNil())

			err3 := query.UpdateFilesystemState(ctx, txDb, fsUpdate)
			Expect(err3).NotTo(BeNil())
		})
	})
	Context("Update Filesystem for Deletion", func() {
		It("should call UpdateFilesystemForDeletion", func() {
			By("Storing fs in DB")
			fsPrivate.Metadata.Name = "test2"
			fsPrivate.Metadata.ResourceId = "7787226a-2a55-4d6f-bae9-fa2a2ca2450a"
			err := query.StoreFilesystemRequest(ctx, txDb, fsPrivate)
			Expect(err).To(BeNil())

			//Failed case
			resp2, err := query.UpdateFilesystemForDeletion(ctx, txDb, fsUpdateDel2)
			Expect(err).NotTo(BeNil())
			Expect(resp2).To(BeZero())
		})

	})
	Context("Get Used Quota By Account ID", func() {
		It("should succeed", func() {
			err := query.UpdateUsedFileQuotaForAllAccounts(ctx, sqlDb, &map[string]query.UsedQuotaFile{}, "infinity")
			Expect(err).To(BeNil())
		})

		It("should fail", func() {
			err := query.UpdateUsedFileQuotaForAllAccounts(ctx, sqlDb, &map[string]query.UsedQuotaFile{}, "")
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Update bucket", func() {
		It("should update bucket deletion time", func() {
			err := query.StoreBucketRequest(ctx, txDb, bucket)
			Expect(err).To(BeNil())
			// update the bucket
			_, err2 := query.UpdateBucketForDeletion(ctx, txDb, &pb.ObjectBucketMetadataRef{
				CloudAccountId: "123456789012",
				NameOrId: &pb.ObjectBucketMetadataRef_BucketName{
					BucketName: "bucket1",
				},
			})
			Expect(err2).To(BeNil())
		})
		It("should update bucket state", func() {
			err := query.UpdateBucketState(ctx, txDb, bucketUpdateStatus)
			Expect(err).NotTo(BeNil())
		})
		It("should update bucket subnet", func() {
			err := query.UpdateSubnetBucketRequest(ctx, txDb, bucket)
			Expect(err).To(BeNil())
		})
		It("Get bucket user access", func() {
			_, err := query.GetBucketAccessUsers(ctx, txDb, bucket.Metadata.CloudAccountId, bucket.Metadata.ResourceId)
			Expect(err).To(BeNil())
		})
	})
	Context("Bucket subnet queries", func() {
		It("should store subnet", func() {
			err := query.StoreBucketSubnet(ctx, txDb, vnet)
			Expect(err).To(BeNil())
		})
		It("get bucket subnet by account", func() {
			res, err := query.GetBucketSubnetByAccount(ctx, txDb, "123456789012")
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("get all bucket subnet events", func() {
			res, err := query.GetAllBucketSubnetEvents(ctx, txDb)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("update status for subnet", func() {
			err := query.UpdateStatusForSubnet(ctx, txDb, vnetDelete)
			Expect(err).To(BeNil())
		})
		It("delete bucket subnet from db", func() {
			vnetDelete.Status = pb.BucketSubnetEventStatus_E_DELETED
			err := query.DeleteBucketSubnetFromDB(ctx, txDb, vnetDelete)
			Expect(err).To(BeNil())
		})
	})
	Context("GetBucketsRequests", func() {
		It("should succeed", func() {
			rs := pb.NewMockObjectStorageServicePrivate_SearchBucketPrivateServer(gomock.NewController(GinkgoT()))
			rs.EXPECT().Context().Return(context.Background()).AnyTimes()
			rs.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
			err := query.GetBucketsRequests(txDb, "1", "infinity", rs)
			Expect(err).To(BeNil())
		})
	})
})
