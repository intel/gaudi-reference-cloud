package test

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	createReq     *pb.ObjectBucketCreateRequest
	meta          *pb.ObjectBucketCreateMetadata
	spec          *pb.ObjectBucketSpec
	bkId          string
	userId        string
	createUserReq *pb.CreateObjectUserRequest
	userMeta      *pb.ObjectUserMetadataCreate
	specs         []*pb.ObjectUserPermissionSpec
	addSubnetReq  *pb.VNetPrivate
)

var _ = Describe("BucketServiceServer", func() {
	Context("Bucket", func() {
		BeforeEach(func() {
			ctx = CreateContextWithToken("test@user.com")
			userMeta = &pb.ObjectUserMetadataCreate{
				CloudAccountId: "123456789012",
				Name:           "tester",
			}
			createUserReq = &pb.CreateObjectUserRequest{
				Metadata: userMeta,
				Spec:     specs,
			}
			meta = &pb.ObjectBucketCreateMetadata{
				CloudAccountId: "123456789012",
				Name:           "bucket-test",
			}
			spec = &pb.ObjectBucketSpec{
				AvailabilityZone: "az1",
				Request: &pb.StorageCapacityRequest{
					Size: "10000000000",
				},
				Versioned:    false,
				AccessPolicy: pb.BucketAccessPolicy_READ_WRITE,
			}
			createReq = &pb.ObjectBucketCreateRequest{
				Metadata: meta,
				Spec:     spec,
			}
			addSubnetReq = &pb.VNetPrivate{
				Metadata: &pb.VNetPrivate_Metadata{
					CloudAccountId: "123456789012",
					Name:           "default",
					ResourceId:     "8623ccaa-704e-4839-bc72-9a89daa20111",
				},
				Spec: &pb.VNetSpecPrivate{
					Region:           "us-region-1",
					AvailabilityZone: "az1",
					Subnet:           "0.0.0.0",
					PrefixLength:     0,
					Gateway:          "0.0.0.0",
				},
			}
		})
		It("Should create bucket successfully", func() {
			//add subnet
			_, err1 := bkServer.AddBucketSubnet(ctx, addSubnetReq)
			Expect(err1).To(BeNil())
			By("Creating a bucket with valid inputs")
			res, err := bkServer.CreateBucket(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
			bkId = res.Metadata.ResourceId
		})
		It("Should fail", func() {
			By("Creating a bucket with same name")
			_, err := bkServer.CreateBucket(ctx, createReq)
			Expect(err).NotTo(BeNil())
			By("Providing invalid input")
			//missing Metadata
			req := &pb.ObjectBucketCreateRequest{Spec: spec, Metadata: &pb.ObjectBucketCreateMetadata{}}
			_, err = bkServer.CreateBucket(ctx, req)
			Expect(err).NotTo(BeNil())
			//missing Spec
			req = &pb.ObjectBucketCreateRequest{Spec: &pb.ObjectBucketSpec{}, Metadata: meta}
			_, err = bkServer.CreateBucket(ctx, req)
			Expect(err).NotTo(BeNil())
			//missing name
			req = &pb.ObjectBucketCreateRequest{Spec: spec, Metadata: &pb.ObjectBucketCreateMetadata{CloudAccountId: "123456789012"}}
			_, err = bkServer.CreateBucket(ctx, req)
			Expect(err).NotTo(BeNil())
			//invalid name, using caps
			req = &pb.ObjectBucketCreateRequest{Spec: spec, Metadata: &pb.ObjectBucketCreateMetadata{CloudAccountId: "123456789012", Name: "ALL_CAPS"}}
			_, err = bkServer.CreateBucket(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Get Bucket", func() {
		It("Should get bucket successfully", func() {
			By("Using valid name")
			req := &pb.ObjectBucketGetRequest{
				Metadata: &pb.ObjectBucketMetadataRef{
					CloudAccountId: "123456789012",
					NameOrId: &pb.ObjectBucketMetadataRef_BucketName{
						BucketName: "123456789012-bucket-test",
					},
				},
			}
			res, err := bkServer.GetBucket(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
			By("Using valid bucket id")
			req.Metadata.NameOrId = &pb.ObjectBucketMetadataRef_BucketId{BucketId: bkId}
			res, err = bkServer.GetBucket(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail", func() {
			By("getting a non-existing bucket")
			req := &pb.ObjectBucketGetRequest{
				Metadata: &pb.ObjectBucketMetadataRef{
					CloudAccountId: "123456789012",
					NameOrId: &pb.ObjectBucketMetadataRef_BucketName{
						BucketName: "bucket404",
					},
				},
			}
			_, err := bkServer.GetBucket(ctx, req)
			Expect(err).NotTo(BeNil())
			By("providing invalid input")
			//missing cloudaccount
			req.Metadata.CloudAccountId = ""
			_, err = bkServer.GetBucket(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.CloudAccountId = "123456789012"
			//invalid bucket name
			req.Metadata.NameOrId = &pb.ObjectBucketMetadataRef_BucketName{BucketName: "BUCKET404"}
			_, err = bkServer.GetBucket(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Search Bucket", func() {
		It("Should search bucket successfully", func() {
			req := &pb.ObjectBucketSearchRequest{
				Metadata: &pb.ObjectBucketSearchMetadata{
					CloudAccountId: "123456789012",
				},
			}
			res, err := bkServer.SearchBucket(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail", func() {
			By("providing not providing cloudaccount")
			req := &pb.ObjectBucketSearchRequest{
				Metadata: &pb.ObjectBucketSearchMetadata{
					CloudAccountId: "",
				},
			}
			_, err := bkServer.SearchBucket(ctx, req)
			Expect(err).NotTo(BeNil())

		})
	})
	Context("Update Bucket Status", func() {
		It("Should fail to update status", func() {
			req := &pb.ObjectBucketStatusUpdateRequest{
				Metadata: &pb.ObjectBucketIdReference{
					CloudAccountId: "123456789012",
					ResourceId:     bkId,
				},
				Status: &pb.ObjectBucketStatus{
					Phase:   pb.BucketPhase_BucketReady,
					Message: "update bucket",
				},
			}
			_, err := bkServer.UpdateBucketStatus(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Create Bucket User", func() {
		It("Should create user successfully", func() {
			By("passing valid input")
			res, err := bkServer.CreateObjectUser(ctx, createUserReq)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())

			//set userId
			userId = res.Metadata.UserId
		})
		It("Should fail", func() {
			By("passing invalid input")
			req := &pb.CreateObjectUserRequest{}
			req.Metadata = &pb.ObjectUserMetadataCreate{
				CloudAccountId: "",
				Name:           "",
			}
			//missing spec
			_, err := bkServer.CreateObjectUser(ctx, req)
			Expect(err).NotTo(BeNil())
			permission := &pb.ObjectUserPermissionSpec{BucketId: ""}
			req.Spec = []*pb.ObjectUserPermissionSpec{
				permission,
			}
			//missing name
			_, err = bkServer.CreateObjectUser(ctx, req)
			Expect(err).NotTo(BeNil())
			//missing cloudaccount
			req.Metadata.Name = "tester"
			_, err = bkServer.CreateObjectUser(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.CloudAccountId = "123456789012"
			By("By creating user with name that already exists")
			_, err = bkServer.CreateObjectUser(ctx, createUserReq)
			Expect(err).NotTo(BeNil())
			By("By creating user for a bucket that does not exist")
			spec := &pb.ObjectUserPermissionSpec{
				BucketId:   "bucket-404",
				Prefix:     "string",
				Permission: []pb.BucketPermission{pb.BucketPermission_ReadBucket},
				Actions:    []pb.ObjectBucketActions{pb.ObjectBucketActions_GetBucketLocation},
			}
			req.Spec = []*pb.ObjectUserPermissionSpec{spec}
			_, err = bkServer.CreateObjectUser(ctx, createUserReq)
			Expect(err).NotTo(BeNil())
			//TODO: make grpc call fail

		})
	})

	Context("Get Bucket user", func() {
		It("Should get user successfully", func() {
			By("passing request with valid params")
			req := &pb.ObjectUserGetRequest{
				Metadata: &pb.ObjectUserMetadataRef{
					CloudAccountId: "123456789012",
					NameOrId:       &pb.ObjectUserMetadataRef_UserName{UserName: "tester"},
				},
			}
			res, err := bkServer.GetObjectUser(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail", func() {
			By("passing invalid input")
			req := &pb.ObjectUserGetRequest{}
			//missing cloudaccount
			req.Metadata = &pb.ObjectUserMetadataRef{CloudAccountId: ""}
			_, err := bkServer.GetObjectUser(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Search Bucket User", func() {
		It("Should search user successfully", func() {
			By("passing request with valid params")
			req := &pb.ObjectUserSearchRequest{CloudAccountId: "123456789012"}
			res, err := bkServer.SearchObjectUser(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("should fail", func() {
			By("passing invalid input")
			//missing cloud account
			req := &pb.ObjectUserSearchRequest{CloudAccountId: ""}
			_, err := bkServer.SearchObjectUser(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Update Bucker User Policy", func() {
		It("Should update policy successfully", func() {
			By("passing request with valid params")
			req := &pb.ObjectUserUpdateRequest{
				Metadata: &pb.ObjectUserMetadataRef{
					CloudAccountId: "123456789012",
					NameOrId:       &pb.ObjectUserMetadataRef_UserName{UserName: "tester"},
				},
				Spec: specs,
			}
			res, err := bkServer.UpdateObjectUserPolicy(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail", func() {
			By("passing invalid input")
			req := &pb.ObjectUserUpdateRequest{
				Metadata: &pb.ObjectUserMetadataRef{
					CloudAccountId: "123456789012",
					NameOrId:       &pb.ObjectUserMetadataRef_UserName{UserName: "404"},
				},
				Spec: specs,
			}
			//nonexisting user
			_, err := bkServer.UpdateObjectUserPolicy(ctx, req)
			Expect(err).NotTo(BeNil())
			//wrong cloud account
			req.Metadata.CloudAccountId = "111111111111"
			req.Metadata.NameOrId = &pb.ObjectUserMetadataRef_UserName{UserName: "tester"}
			_, err = bkServer.UpdateObjectUserPolicy(ctx, req)
			Expect(err).NotTo(BeNil())

		})
	})

	Context("Update Bucket User credential", func() {
		It("Should update user credential successfully", func() {
			By("passing request with valid params")
			req := &pb.ObjectUserUpdateCredsRequest{
				Metadata: &pb.ObjectUserMetadataRef{
					CloudAccountId: "123456789012",
					NameOrId:       &pb.ObjectUserMetadataRef_UserId{UserId: userId},
				},
			}
			res, err := bkServer.UpdateObjectUserCredentials(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail", func() {
			By("passing invalid input")
			req := &pb.ObjectUserUpdateCredsRequest{
				Metadata: &pb.ObjectUserMetadataRef{
					CloudAccountId: "",
					NameOrId:       &pb.ObjectUserMetadataRef_UserId{UserId: userId},
				},
			}
			//missing cloudaccount
			_, err := bkServer.UpdateObjectUserCredentials(ctx, req)
			Expect(err).NotTo(BeNil())
			//Incorrect userID
			req.Metadata.NameOrId = &pb.ObjectUserMetadataRef_UserId{UserId: "dbkDBKFKB"}
			_, err = bkServer.UpdateObjectUserCredentials(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Delete Bucket User", func() {
		It("Should delete user successfully", func() {
			By("passing request with valid params")
			req := &pb.ObjectUserDeleteRequest{
				Metadata: &pb.ObjectUserMetadataRef{
					CloudAccountId: "123456789012",
					NameOrId:       &pb.ObjectUserMetadataRef_UserName{UserName: "tester"},
				},
			}
			_, err := bkServer.DeleteObjectUser(ctx, req)
			Expect(err).To(BeNil())
		})
		It("Should fail", func() {
			By("passing invalid input")
			req := &pb.ObjectUserDeleteRequest{
				Metadata: &pb.ObjectUserMetadataRef{
					CloudAccountId: "",
				},
			}
			//missing cloud account
			_, err := bkServer.DeleteObjectUser(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.CloudAccountId = "123456789012"
			req.Metadata.NameOrId = &pb.ObjectUserMetadataRef_UserName{UserName: "ALLCAPS"}

			//passing invalid username
			_, err = bkServer.DeleteObjectUser(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Add/Remove bucket subnet", func() {
		It("Add/Remove bucket subnet", func() {
			req := &pb.VNetPrivate{
				Metadata: &pb.VNetPrivate_Metadata{
					CloudAccountId: "123456789012",
					Name:           "default",
					ResourceId:     "8623ccaa-704e-4839-bc72-9a89daa20111",
				},
				Spec: &pb.VNetSpecPrivate{
					Region:           "us-region-1",
					AvailabilityZone: "az1",
					Subnet:           "0.0.0.0",
					PrefixLength:     0,
					Gateway:          "0.0.0.0",
				},
			}
			req2 := &pb.BucketSubnetStatusUpdateRequest{
				ResourceId:      "8623ccaa-704e-4839-bc72-9a89daa20111",
				CloudacccountId: "123456789012",
				VNetName:        "default",
				Status:          pb.BucketSubnetEventStatus_E_ADDED,
			}
			req3 := &pb.VNetReleaseSubnetRequest{
				VNetReference: &pb.VNetReference{CloudAccountId: "123456789012", Name: "default"},
			}

			_, err := bkServer.AddBucketSubnet(ctx, req)
			Expect(err).To(BeNil())
			_, err = bkServer.UpdateBucketSubnetStatus(ctx, req2)
			Expect(err).To(BeNil())
			//case for subnet deleted status
			req2.Status = pb.BucketSubnetEventStatus_E_DELETED
			_, err = bkServer.UpdateBucketSubnetStatus(ctx, req2)
			Expect(err).To(BeNil())
			_, err = bkServer.AddBucketSubnet(ctx, req)
			Expect(err).To(BeNil())
			_, err = bkServer.RemoveBucketSubnet(ctx, req3)
			Expect(err).To(BeNil())

		})
	})
	Context("Remove bucket finalizer", func() {
		It("remove bucket finalizer", func() {
			req := &pb.ObjectBucketRemoveFinalizerRequest{Metadata: &pb.ObjectBucketIdReference{CloudAccountId: "123456789012", ResourceId: bkId}}
			_, err := bkServer.RemoveBucketFinalizer(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Delete Bucket", func() {
		It("Should delete bucket successfully", func() {
			By("passing request with valid params")
			req := &pb.ObjectBucketDeleteRequest{
				Metadata: &pb.ObjectBucketMetadataRef{
					CloudAccountId: "123456789012",
					NameOrId:       &pb.ObjectBucketMetadataRef_BucketName{BucketName: "123456789012-bucket-test"},
				},
			}
			res, err := bkServer.DeleteBucket(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail", func() {
			By("passing invalid input")
			req := &pb.ObjectBucketDeleteRequest{
				Metadata: &pb.ObjectBucketMetadataRef{
					CloudAccountId: "",
					NameOrId:       &pb.ObjectBucketMetadataRef_BucketName{BucketName: "bucket-test"},
				},
			}
			//missing cloudaccount
			_, err := bkServer.DeleteBucket(ctx, req)
			Expect(err).NotTo(BeNil())
			//missing bucketId
			req.Metadata.CloudAccountId = "123456789012"
			req.Metadata.NameOrId = nil
			_, err = bkServer.DeleteBucket(ctx, req)
			Expect(err).NotTo(BeNil())
			//invalid bucket name
			req.Metadata.NameOrId = &pb.ObjectBucketMetadataRef_BucketName{BucketName: "ALLCAPS"}
			_, err = bkServer.DeleteBucket(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API UpdateObjectUserStatus", func() {
		It("Should fail", func() {
			req := &pb.ObjectUserStatusUpdateRequest{
				Metadata: &pb.ObjectUserIdReference{
					CloudAccountId: "123456789012",
				},
			}

			_, err := bkServer.UpdateObjectUserStatus(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API RemoveObjectUserFinalizer", func() {
		It("Should fail", func() {
			req := &pb.ObjectUserRemoveFinalizerRequest{
				Metadata: &pb.ObjectUserIdReference{
					CloudAccountId: "123456789012",
				},
			}

			_, err := bkServer.RemoveObjectUserFinalizer(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
})
