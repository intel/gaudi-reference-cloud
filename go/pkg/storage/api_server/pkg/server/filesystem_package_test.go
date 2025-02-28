// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filesystem functions", func() {
	Context("Validate requests", func() {
		var version = new(string)
		*version = "v2"
		ctx := context.Background()
		req := &pb.FilesystemCreateRequestPrivate{
			Metadata: &pb.FilesystemMetadataPrivate{
				CloudAccountId: "123456789012",
				Name:           "test1",
			},
			Spec: &pb.FilesystemSpecPrivate{
				AvailabilityZone: "az",
				Request: &pb.FilesystemCapacity{
					Storage: "500GB",
				},
				StorageClass:  pb.FilesystemStorageClass_AIOptimized,
				AccessModes:   pb.FilesystemAccessModes_ReadWrite,
				MountProtocol: pb.FilesystemMountProtocols_Weka,
				Encrypted:     true,
				Scheduler: &pb.FilesystemSchedule{
					FilesystemName: "test",
					Cluster: &pb.AssignedCluster{
						ClusterName:    "1",
						ClusterAddr:    "1",
						ClusterUUID:    "1",
						ClusterVersion: version,
					},
					Namespace: &pb.AssignedNamespace{
						Name:            "123456789012",
						CredentialsPath: "/path/to/secret",
					},
				},
			},
		}
		product := fileProductInfo{
			MinSize:          5,
			MaxSize:          2000,
			UpdatedTimestamp: time.Now(),
		}
		It("Validate delete request should fail", func() {
			By("Passing request with missing fields")
			// Define payload
			ctx := context.Background()
			req := &pb.FilesystemMetadataReference{
				CloudAccountId: "123456789012",
			}
			err := isValidFilesystemDeleteRequest(ctx, req)
			Expect(err).NotTo(BeNil())
		})
		It("Validate create request should succeed", func() {
			By("Passing in valid request")
			err2 := isValidFilesystemCreateRequest(ctx, req, product, false)
			Expect(err2).To(BeNil())
		}) //It
		It("Validate create request should fail", func() {
			By("Supplying req with missing metadata")
			req2 := req
			req2.Metadata = nil
			err3 := isValidFilesystemCreateRequest(ctx, req2, product, false)
			Expect(err3).NotTo(BeNil())

			By("Supplying missing spec")
			req3 := req
			req3.Spec = nil
			err4 := isValidFilesystemCreateRequest(ctx, req3, product, false)
			Expect(err4).NotTo(BeNil())

			By("Not providing a name in Metadata")
			req4 := req
			req4.Metadata.Name = ""
			err5 := isValidFilesystemCreateRequest(ctx, req4, product, false)
			Expect(err5).NotTo(BeNil())
		})

		It("should search filesystem", func() {
			searchReq := &pb.FilesystemSearchRequest{
				Metadata: &pb.FilesystemMetadataSearch{
					CloudAccountId: "0123456789012",
				},
			}
			valid := isValidFilesystemSearchRequest(ctx, searchReq)
			Expect(valid).Should(Equal(true))

			searchReq2 := &pb.FilesystemSearchRequest{
				Metadata: &pb.FilesystemMetadataSearch{
					CloudAccountId: "",
				},
			}
			valid = isValidFilesystemSearchRequest(ctx, searchReq2)
			Expect(valid).Should(Equal(false))
		})

	})
	Context("Utility functions", func() {
		It("should get corresponding type", func() {
			res := getAccountTypeStr(pb.AccountType_ACCOUNT_TYPE_ENTERPRISE)
			Expect(res).To(Equal("ENTERPRISE"))
			res = getAccountTypeStr(pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING)
			Expect(res).To(Equal("ENTERPRISE_PENDING"))
			res = getAccountTypeStr(pb.AccountType_ACCOUNT_TYPE_INTEL)
			Expect(res).To(Equal("INTEL"))
			res = getAccountTypeStr(pb.AccountType_ACCOUNT_TYPE_PREMIUM)
			Expect(res).To(Equal("PREMIUM"))
			res = getAccountTypeStr(pb.AccountType_ACCOUNT_TYPE_STANDARD)
			Expect(res).To(Equal("STANDARD"))
		})
	})

})
