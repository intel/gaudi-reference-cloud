package test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Private BucketServiceServer APIs", func() {
	Context("Unimplemented API CreateBucketPrivate", func() {
		It("Should fail", func() {
			in := &pb.ObjectBucketCreatePrivateRequest{
				Metadata: &pb.ObjectBucketMetadataPrivate{
					CloudAccountId: "123456789012",
				},
			}
			_, err := bkServer.CreateBucketPrivate(context.Background(), in)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API GetBucketPrivate", func() {
		It("Should fail", func() {
			req := &pb.ObjectBucketGetPrivateRequest{
				Metadata: &pb.ObjectBucketMetadataRef{
					CloudAccountId: "123456789012",
					NameOrId: &pb.ObjectBucketMetadataRef_BucketName{
						BucketName: "bucket404",
					},
				},
			}
			_, err := bkServer.GetBucketPrivate(context.Background(), req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API CreateObjectUserPrivate", func() {
		It("Should fail", func() {
			reqPrivate := &pb.CreateObjectUserPrivateRequest{
				Metadata: &pb.ObjectUserMetadataCreate{
					CloudAccountId: "123456789012",
					Name:           "name",
				},
			}

			res, err := bkServer.CreateObjectUserPrivate(context.Background(), reqPrivate)
			Expect(err).NotTo(BeNil())
			Expect(res).To(BeNil())
		})
	})
	Context("Unimplemented API GetObjectUserPrivate", func() {
		It("Should fail", func() {
			req := &pb.ObjectUserGetPrivateRequest{
				Metadata: &pb.ObjectUserMetadataRef{
					CloudAccountId: "123456789012",
				},
			}

			_, err := bkServer.GetObjectUserPrivate(context.Background(), req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API DeleteObjectUserPrivate", func() {
		It("Should fail", func() {
			req := &pb.ObjectUserDeletePrivateRequest{
				Metadata: &pb.ObjectUserMetadataRef{
					CloudAccountId: "123456789013",
					NameOrId:       &pb.ObjectUserMetadataRef_UserName{UserName: "tester"},
				},
			}

			_, err := bkServer.DeleteObjectUserPrivate(context.Background(), req)
			Expect(err).ToNot(BeNil())
		})
	})
	Context("Unimplemented API DeleteBucketPrivate", func() {
		It("Should fail", func() {
			req3 := &pb.ObjectBucketDeletePrivateRequest{
				Metadata: &pb.ObjectBucketMetadataRef{
					CloudAccountId: "123456789012",
					NameOrId: &pb.ObjectBucketMetadataRef_BucketName{
						BucketName: "test",
					},
				},
			}

			_, err := bkServer.DeleteBucketPrivate(context.Background(), req3)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API CreateBucketLifecycleRulePrivate", func() {
		It("Should fail", func() {
			req := &pb.BucketLifecycleRuleCreatePrivateRequest{
				Metadata: &pb.BucketLifecycleRuleMetadata{
					CloudAccountId: "123456789012",
				},
			}

			_, err := bkServer.CreateBucketLifecycleRulePrivate(context.Background(), req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API GetBucketLifecycleRulePrivate", func() {
		It("Should fail", func() {
			req := &pb.BucketLifecycleRuleGetPrivateRequest{
				Metadata: &pb.BucketLifecycleRuleMetadataRef{
					CloudAccountId: "123456789012",
				},
			}

			_, err := bkServer.GetBucketLifecycleRulePrivate(context.Background(), req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API SearchBucketLifecycleRulePrivate", func() {
		It("Should fail", func() {
			req := &pb.BucketLifecycleRuleSearchPrivateRequest{
				CloudAccountId: "123456789012",
			}

			_, err := bkServer.SearchBucketLifecycleRulePrivate(context.Background(), req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API UpdateBucketLifecycleRulePrivate", func() {
		It("Should fail", func() {
			req := &pb.BucketLifecycleRuleUpdatePrivateRequest{
				Metadata: &pb.BucketLifecycleRuleMetadataRef{
					CloudAccountId: "123456789012",
				},
			}

			_, err := bkServer.UpdateBucketLifecycleRulePrivate(context.Background(), req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API DeleteBucketLifecycleRulePrivate", func() {
		It("Should fail", func() {
			req := &pb.BucketLifecycleRuleDeletePrivateRequest{
				Metadata: &pb.BucketLifecycleRuleMetadataRef{
					CloudAccountId: "123456789012",
				},
			}

			_, err := bkServer.DeleteBucketLifecycleRulePrivate(context.Background(), req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Unimplemented API UpdateObjectUserPrivate", func() {
		It("Should fail", func() {
			req := &pb.ObjectUserUpdatePrivateRequest{
				Metadata: &pb.ObjectUserMetadataRef{
					CloudAccountId: "123456789012",
				},
			}

			_, err := bkServer.UpdateObjectUserPrivate(context.Background(), req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("SearchBucketPrivate", func() {
		It("Should Succeed", func() {
			req := &pb.ObjectBucketSearchPrivateRequest{
				AvailabilityZone: "az1",
				ResourceVersion:  "0",
			}
			rs := pb.NewMockObjectStorageServicePrivate_SearchBucketPrivateServer(gomock.NewController(GinkgoT()))
			rs.EXPECT().Context().Return(context.Background()).AnyTimes()
			rs.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()

			err := bkServer.SearchBucketPrivate(req, rs)
			Expect(err).To(BeNil())
		})
	})
	Context("GetBucketSubnetEvent", func() {
		It("Should Succeed", func() {
			req := &pb.SubnetEventRequest{}
			rs := pb.NewMockObjectStorageServicePrivate_GetBucketSubnetEventServer(gomock.NewController(GinkgoT()))
			rs.EXPECT().Context().Return(context.Background()).AnyTimes()
			rs.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()

			err := bkServer.GetBucketSubnetEvent(req, rs)
			Expect(err).To(BeNil())
		})
	})
})
