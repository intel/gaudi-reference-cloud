// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"database/sql"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	fs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/api_server/pkg/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	testDB    *sql.DB
	fsServer  *fs.FilesystemServiceServer
	fsServer2 *fs.FilesystemServiceServer
	fsServer3 *fs.FilesystemServiceServer
)

var _ = Describe("FilesystemServiceServer", func() {
	BeforeEach(func() {
		ctx = CreateContextWithToken("test@user.com")
	})
	Context("Get User", func() {
		It("Should get user credentials succesfully", func() {
			By("Creating a filesystem and user")
			meta := &pb.FilesystemMetadataCreate{
				CloudAccountId: "123456789099",
				Name:           "test-volume",
			}
			request := &pb.FilesystemCapacity{
				Storage: "5000GB",
			}
			spec := &pb.FilesystemSpec{
				AvailabilityZone: "az1",
				Request:          request,
			}
			req := &pb.FilesystemCreateRequest{
				Metadata: meta,
				Spec:     spec,
			}
			// Call the Create method of fsServer and capture the response
			resp, err := fsServer.Create(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

			By("Requesting new user credentials by name")
			metaRef := &pb.FilesystemGetUserRequest{
				Metadata: &pb.FilesystemMetadataReference{
					CloudAccountId: "123456789099",
					NameOrId: &pb.FilesystemMetadataReference_Name{
						Name: "test-volume",
					},
				},
			}
			creds, err2 := fsServer.GetUser(ctx, metaRef)
			Expect(err2).To(BeNil())
			Expect(creds).NotTo(BeNil())

			//get user by resource id
			By("Requesting new user credentials by id")
			metaRef2 := &pb.FilesystemGetUserRequest{
				Metadata: &pb.FilesystemMetadataReference{
					CloudAccountId: "123456789099",
					NameOrId: &pb.FilesystemMetadataReference_ResourceId{
						ResourceId: resp.Metadata.ResourceId,
					},
				},
			}
			creds2, err3 := fsServer.GetUser(ctx, metaRef2)
			Expect(err3).To(BeNil())
			Expect(creds2).NotTo(BeNil())

			By("Requesting new user credentials via private api")
			credP, err4 := fsServer.GetUserPrivate(ctx, &pb.FilesystemGetUserRequestPrivate{Metadata: metaRef.Metadata})
			Expect(err4).To(BeNil())
			Expect(credP).NotTo(BeNil())

		})
	})

	//Clean up volume to free quota for subsequent test cases
	Context("Delete", func() {
		It("should delete a filesystem successfully", func() {
			// Get
			name := &pb.FilesystemMetadataReference_Name{
				Name: "test-volume",
			}
			meta := &pb.FilesystemMetadataReference{
				CloudAccountId: "123456789099",
				NameOrId:       name,
			}
			req := &pb.FilesystemGetRequest{
				Metadata: meta,
			}

			// Call the Delete method of fsServer and capture the response
			resp, err := fsServer.Get(ctx, req)

			// Expectations
			Expect(err).To(BeNil())     // Ensure no error
			Expect(resp).NotTo(BeNil()) // Ensure response is not nil
		})
	})

	Context("Create, Update Status, Update", func() {
		It("should create a filesystem successfully", func() {
			meta := &pb.FilesystemMetadataCreate{
				CloudAccountId: "123456789012",
				Name:           "fs1-test36",
			}
			request := &pb.FilesystemCapacity{
				Storage: "5TB",
			}
			spec := &pb.FilesystemSpec{
				AvailabilityZone: "az1",
				Request:          request,
			}
			req := &pb.FilesystemCreateRequest{
				Metadata: meta,
				Spec:     spec,
			}

			// Call the Create method of fsServer and capture the response
			resp, err := fsServer.Create(ctx, req)

			// Expectations
			Expect(err).To(BeNil())     // Ensure no error
			Expect(resp).NotTo(BeNil()) // Ensure response is not nil

			//Happy case for UpdateStatus
			meta2 := &pb.FilesystemIdReference{
				CloudAccountId:  "123456789012",
				ResourceVersion: "1",
				ResourceId:      resp.Metadata.Name,
			}
			req2 := &pb.FilesystemUpdateStatusRequest{
				Metadata: meta2,
				Status: &pb.FilesystemStatusPrivate{
					Message: "az1",
				},
			}
			//Call the UpdateStatus method of fsServer and capture the response
			resp2, err := fsServer.UpdateStatus(ctx, req2)
			Expect(err).To(BeNil())
			Expect(resp2).NotTo(BeNil())

			//UpdateStatus for phase Failed
			meta2 = &pb.FilesystemIdReference{
				CloudAccountId:  "123456789012",
				ResourceVersion: "1",
				ResourceId:      resp.Metadata.Name,
			}
			req2 = &pb.FilesystemUpdateStatusRequest{
				Metadata: meta2,
				Status: &pb.FilesystemStatusPrivate{
					Message: "az1",
					Phase:   pb.FilesystemPhase_FSFailed,
				},
			}
			//Call the UpdateStatus method of fsServer and capture the response
			resp2, err = fsServer.UpdateStatus(ctx, req2)
			Expect(err).To(BeNil())
			Expect(resp2).NotTo(BeNil())

			//Failed case for Update
			meta3 := &pb.FilesystemMetadataUpdate{
				CloudAccountId:  "123456789012",
				ResourceVersion: "1",
				NameOrId: &pb.FilesystemMetadataUpdate_Name{
					Name: "fs1-test56",
				},
			}
			spec3 := &pb.FilesystemSpec{
				Request: &pb.FilesystemCapacity{
					Storage: "10GB",
				},
			}
			req3 := &pb.FilesystemUpdateRequest{
				Metadata: meta3,
				Spec:     spec3,
			}
			//Call the Update method of fsServer and capture the response
			resp3, err := fsServer.Update(ctx, req3)
			Expect(err).NotTo(BeNil())
			Expect(resp3).To(BeNil())

		})

		It("should return an error for repeated input", func() {

			meta := &pb.FilesystemMetadataCreate{
				CloudAccountId: "123456789012",
				Name:           "fs1-test46",
			}
			request := &pb.FilesystemCapacity{
				Storage: "5000GB",
			}
			spec := &pb.FilesystemSpec{
				AvailabilityZone: "az1",
				Request:          request,
			}
			req := &pb.FilesystemCreateRequest{
				Metadata: meta,
				Spec:     spec,
			}
			// Call the Create method and capture the error
			resp, err1 := fsServer2.Create(ctx, req)
			_, err2 := fsServer2.Create(ctx, req)
			// Expectations
			Expect(err1).NotTo(HaveOccurred()) // Expect Nil
			Expect(err2).To(HaveOccurred())

			meta4 := &pb.FilesystemIdReference{
				CloudAccountId:  "123456789012",
				ResourceVersion: "1",
				ResourceId:      resp.Metadata.Name,
			}

			req4 := &pb.FilesystemRemoveFinalizerRequest{
				Metadata: meta4,
			}

			resp4, err := fsServer.RemoveFinalizer(ctx, req4)

			Expect(err).To(BeNil())
			Expect(resp4).NotTo(BeNil())

			meta5 := &pb.FilesystemIdReference{
				CloudAccountId:  "123456789012",
				ResourceVersion: "1",
				ResourceId:      resp.Metadata.ResourceId,
			}

			req5 := &pb.FilesystemRemoveFinalizerRequest{
				Metadata: meta5,
			}

			resp5, err := fsServer.RemoveFinalizer(ctx, req5)

			Expect(err).NotTo(BeNil())
			Expect(resp5).To(BeNil())

		})

		It("should fail to create file system for invalid request", func() {

			meta := &pb.FilesystemMetadataCreate{
				CloudAccountId: "123456789012",
				Name:           "fs1-test46",
			}
			request := &pb.FilesystemCapacity{
				Storage: "10GB",
			}
			spec := &pb.FilesystemSpec{
				AvailabilityZone: "az1",
				Request:          request,
			}
			req := &pb.FilesystemCreateRequest{
				Metadata: meta,
				Spec:     spec,
			}
			// Call the Create method and capture the error
			resp, err := fsServer2.Create(ctx, req)
			// Expectations
			Expect(err).To(HaveOccurred()) // Expect Nil
			Expect(resp).To(BeNil())
		})

	})

	Context("Search", func() {
		It("should search a filesystem successfully", func() {
			meta := &pb.FilesystemMetadataSearch{
				CloudAccountId: "123456789012",
			}
			req := &pb.FilesystemSearchRequest{
				Metadata: meta,
			}

			// Call the Search method of fsServer and capture the response
			resp, err := fsServer.Search(ctx, req)

			// Expectations
			Expect(err).To(BeNil())     // Ensure no error
			Expect(resp).NotTo(BeNil()) // Ensure response is not nil
		})
	})
	Context("Get", func() {
		It("should get a filesystem successfully", func() {
			// Get
			name := &pb.FilesystemMetadataReference_Name{
				Name: "fs1-test36",
			}
			meta := &pb.FilesystemMetadataReference{
				CloudAccountId: "123456789012",
				NameOrId:       name,
			}
			req := &pb.FilesystemGetRequest{
				Metadata: meta,
			}

			// Call the Delete method of fsServer and capture the response
			resp, err := fsServer.Get(ctx, req)

			// Expectations
			Expect(err).To(BeNil())     // Ensure no error
			Expect(resp).NotTo(BeNil()) // Ensure response is not nil
		})
	})
	Context("Update", func() {
		It("should update a filesystem successfully", func() {
			// Get
			name := &pb.FilesystemMetadataUpdate_Name{
				Name: "fs1-test36",
			}
			meta := &pb.FilesystemMetadataUpdate{
				CloudAccountId: "123456789012",
				NameOrId:       name,
			}
			spec := &pb.FilesystemSpec{
				AvailabilityZone: "az1",
				Request: &pb.FilesystemCapacity{
					Storage: "2TB",
				},
			}
			req := &pb.FilesystemUpdateRequest{
				Metadata: meta,
				Spec:     spec,
			}

			By("updating with smaller size should fail")
			_, err := fsServer.Update(ctx, req)
			Expect(err).NotTo(BeNil()) // Ensure no error

			By("updating with larger size should succeed")
			req.Spec.Request.Storage = "6TB"
			_, err = fsServer.Update(ctx, req)
			Expect(err).To(BeNil()) // Ensure no error

		})
	})
	Context("Delete Filesystem", func() {
		It("should Delete filesystems successfully", func() {
			// Delete payload
			name := &pb.FilesystemMetadataReference_Name{
				Name: "fs1-test36",
			}
			meta := &pb.FilesystemMetadataReference{
				CloudAccountId: "123456789012",
				NameOrId:       name,
			}
			req := &pb.FilesystemDeleteRequest{
				Metadata: meta,
			}

			// Call the Delete method of fsServer and capture the response
			resp, err := fsServer.Delete(ctx, req)

			// Expectations
			Expect(err).To(BeNil())     // Ensure no error
			Expect(resp).NotTo(BeNil()) // Ensure response is not nil
		})
	})
	Context("Agent registration/deregistration", func() {
		It("agent operations", func() {
			cid := "934b5026-d346-78c8-fcd3-899852346509"
			ctx = context.Background()
			agent, err := fsServer.RegisterAgent(ctx, &pb.RegisterAgentRequest{ClusterId: cid})
			Expect(err).To(BeNil())
			Expect(agent).NotTo(BeNil())
			agent, err = fsServer.GetRegisteredAgent(ctx, &pb.GetRegisterAgentRequest{ClusterId: cid})
			Expect(err).To(BeNil())
			Expect(agent).NotTo(BeNil())
			rs := pb.NewMockFilesystemStorageClusterPrivateService_ListRegisteredAgentsServer(gomock.NewController(GinkgoT()))
			rs.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
			rs.EXPECT().Context().Return(context.Background()).AnyTimes()
			err = fsServer.ListRegisteredAgents(&pb.ListRegisteredAgentRequest{ClusterId: cid}, rs)
			Expect(err).To(BeNil())
			_, err = fsServer.DeRegisterAgent(ctx, &pb.DeRegisterAgentRequest{ClusterId: cid})
			Expect(err).To(BeNil())
			req := pb.NewMockFilesystemStorageClusterPrivateService_ListClustersServer(gomock.NewController(GinkgoT()))
			req.EXPECT().Context().Return(context.Background()).AnyTimes()
			req.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
			err = fsServer.ListClusters(&pb.ListClusterRequest{}, req)
			Expect(err).To(BeNil())
		})
	})
	Context("Private funcs", func() {
		It("should return error", func() {

			By("Supplying incomplete arguments")
			res3, err3 := fsServer.CreatePrivate(ctx, &pb.FilesystemCreateRequestPrivate{
				Metadata: &pb.FilesystemMetadataPrivate{
					CloudAccountId: "123456789012",
				},
			})
			Expect(err3).NotTo(BeNil())
			Expect(res3).To(BeNil())

			res4, err4 := fsServer.GetPrivate(ctx, &pb.FilesystemGetRequestPrivate{
				Metadata: &pb.FilesystemMetadataReference{
					CloudAccountId: "123456789012",
				},
			})
			Expect(err4).NotTo(BeNil())
			Expect(res4).To(BeNil())

			res5, err5 := fsServer.DeletePrivate(ctx, &pb.FilesystemDeleteRequestPrivate{
				Metadata: &pb.FilesystemMetadataReference{
					CloudAccountId: "123456789012",
				},
			})
			Expect(err5).NotTo(BeNil())
			Expect(res5).To(BeNil())

		})
	})

	Context("Create vast filesystem successfully", func() {
		It("Should create fs succesfully", func() {
			By("Creating a filesystem")
			meta := &pb.FilesystemMetadataCreate{
				CloudAccountId: "123456789099",
				Name:           "test-volume1",
			}
			request := &pb.FilesystemCapacity{
				Storage: "5000GB",
			}
			spec := &pb.FilesystemSpec{
				AvailabilityZone: "az1",
				Request:          request,
				StorageClass:     pb.FilesystemStorageClass_GeneralPurposeStd,
			}
			req := &pb.FilesystemCreateRequest{
				Metadata: meta,
				Spec:     spec,
			}

			// Call the Create method of fsServer3 and capture the response
			resp, err := fsServer3.Create(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

			By("Creating or getting user credentials via private API")
			metaRef := &pb.FilesystemGetUserRequestPrivate{
				Metadata: &pb.FilesystemMetadataReference{
					CloudAccountId: "123456789099",
					NameOrId: &pb.FilesystemMetadataReference_Name{
						Name: "test-volume1",
					},
				},
			}
			_, err = fsServer.CreateorGetUserPrivate(ctx, metaRef)
			Expect(err).NotTo(BeNil())

		})
	})

	Context("Search Filesystem Requests", func() {
		It("Search should succeed", func() {
			req := &pb.FilesystemSearchStreamPrivateRequest{
				AvailabilityZone: "az1",
				ResourceVersion:  "0",
			}
			rs := pb.NewMockFilesystemPrivateService_SearchFilesystemRequestsServer(gomock.NewController(GinkgoT()))
			rs.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
			rs.EXPECT().Context().Return(context.Background()).AnyTimes()
			err := fsServer.SearchFilesystemRequests(req, rs)
			Expect(err).To(BeNil())
		})
	})

	Context("Ping", func() {
		It("should succeed", func() {
			ctx := context.Background()
			in := &empty.Empty{}
			res, err := fsServer.Ping(ctx, in)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())

			res2, err2 := fsServer.PingPrivate(ctx, in)
			Expect(err2).To(BeNil())
			Expect(res2).NotTo(BeNil())

			_, err = fsServer.PingFilesystemClusterPrivate(ctx, nil)
			Expect(err).To(BeNil())
		})
	})
})
