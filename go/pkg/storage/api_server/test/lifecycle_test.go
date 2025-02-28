package test

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket lifecycle rule test", func() {
	BeforeEach(func() {
		ctx = CreateContextWithToken("lala@lala.com")
	})
	Context("CreateBucketLifecycleRule", func() {
		It("should create lifecycle rule successfully", func() {
			req := &pb.BucketLifecycleRuleCreateRequest{
				Metadata: &pb.BucketLifecycleRuleMetadata{
					CloudAccountId: "123456789111",
					RuleName:       "test-rule",
					BucketId:       bucketId,
				},
				Spec: &pb.BucketLifecycleRuleSpec{
					Prefix:               "/tmp",
					ExpireDays:           1,
					NoncurrentExpireDays: 5,
					DeleteMarker:         false,
				},
			}
			// make request
			res, err := bkServer.CreateBucketLifecycleRule(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
			//store ruleId
			ruleId = res.Metadata.ResourceId

		})
		It("should fail", func() {
			By("passing invalid/incomplete params")
			req := &pb.BucketLifecycleRuleCreateRequest{
				Metadata: &pb.BucketLifecycleRuleMetadata{CloudAccountId: ""},
			}
			//missing Spec
			_, err := bkServer.CreateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Spec = &pb.BucketLifecycleRuleSpec{
				Prefix:               "/tmp",
				ExpireDays:           1,
				NoncurrentExpireDays: 5,
				DeleteMarker:         false,
			}
			//missing BucketId
			_, err = bkServer.CreateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.BucketId = bucketId
			//missing cloud account
			_, err = bkServer.CreateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.CloudAccountId = "123456789111"
			//missing ruleName
			_, err = bkServer.CreateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.RuleName = "test-rule"
			//creating lcr with name that already exists
			_, err = bkServer.CreateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.RuleName = "test-rule2"
			req.Metadata.BucketId = logkeys.BucketId
			//passing invalid bucketId
			_, err = bkServer.CreateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	}) //context
	Context("GetBucketLifecycleRule", func() {
		It("should succesfully get a lifecycle rule", func() {
			//form request
			req := &pb.BucketLifecycleRuleGetRequest{
				Metadata: &pb.BucketLifecycleRuleMetadataRef{
					CloudAccountId: "123456789111",
					BucketId:       bucketId,
					RuleId:         ruleId,
				},
			}
			//make request
			res, err := bkServer.GetBucketLifecycleRule(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("should fail", func() {
			By("passing invalid/incomlete params")
			req := &pb.BucketLifecycleRuleGetRequest{
				Metadata: &pb.BucketLifecycleRuleMetadataRef{
					CloudAccountId: "",
					BucketId:       "invalid-bucketId",
					RuleId:         "invalid-RuleId",
				},
			}
			//empty inputs
			_, err := bkServer.GetBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//invalid bucketId
			req.Metadata.CloudAccountId = "123456789111"
			_, err = bkServer.GetBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//invalid ruleId
			req.Metadata.BucketId = bucketId
			_, err = bkServer.GetBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//invalid could account
			req.Metadata.RuleId = ruleId
			req.Metadata.CloudAccountId = "111111111111"
			_, err = bkServer.GetBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//supply ruleId for resource that does not exist
			req.Metadata.RuleId = "755f4478-5a42-4efc-8d45-90a55437aa68"
			_, err = bkServer.GetBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
		})
		//fail case
	}) //context
	Context("SearchBucketLifecycleRule", func() {
		It("should successfully list all lifecycle rules for cloudaccount", func() {
			//form request
			req := &pb.BucketLifecycleRuleSearchRequest{
				CloudAccountId: "123456789111",
				BucketId:       bucketId,
			}
			//make request
			res, err := bkServer.SearchBucketLifecycleRule(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("should fail", func() {
			By("passing invalid/incomplete params")
			req := &pb.BucketLifecycleRuleSearchRequest{
				CloudAccountId: "123456789111",
				BucketId:       "123456789111-bucket-lf",
			}
			//supply bucketName instead of bucket resourceId
			_, err := bkServer.SearchBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//supply invalid
			req.CloudAccountId = "111111111111"
			_, err = bkServer.SearchBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//supply empty cloud account
			req.CloudAccountId = ""
			_, err = bkServer.SearchBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//supply empty bucketId
			req.BucketId = ""
			_, err = bkServer.SearchBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("UpdateBucketLifecycleRule", func() {
		It("should successfully update lifecycle rule", func() {
			//form request
			req := &pb.BucketLifecycleRuleUpdateRequest{
				Metadata: &pb.BucketLifecycleRuleMetadataRef{
					CloudAccountId: "123456789111",
					BucketId:       bucketId,
					RuleId:         ruleId,
				},
				Spec: &pb.BucketLifecycleRuleSpec{
					Prefix:               "/tmp",
					ExpireDays:           2,
					NoncurrentExpireDays: 5,
					DeleteMarker:         false,
				},
			}
			//make request
			res, err := bkServer.UpdateBucketLifecycleRule(ctx, req)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("should fail", func() {
			By("passing invalid/incomplete params")
			req := &pb.BucketLifecycleRuleUpdateRequest{
				Metadata: &pb.BucketLifecycleRuleMetadataRef{
					CloudAccountId: "",
					BucketId:       bucketId,
					RuleId:         ruleId,
				},
				Spec: &pb.BucketLifecycleRuleSpec{
					Prefix:               "/tmp",
					ExpireDays:           2,
					NoncurrentExpireDays: 5,
					DeleteMarker:         false,
				},
			}
			//missing cloud account
			_, err := bkServer.UpdateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//invalid/incorrect cloud account
			req.Metadata.CloudAccountId = "111111111111"
			_, err = bkServer.UpdateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.CloudAccountId = "123456789111"
			//missing bucketId
			req.Metadata.BucketId = ""
			_, err = bkServer.UpdateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//pass bucketName for bucketId
			req.Metadata.BucketId = "123456789111-bucket-lf"
			_, err = bkServer.UpdateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.BucketId = bucketId
			//empty ruleId
			req.Metadata.RuleId = ""
			_, err = bkServer.UpdateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//incorrect ruleId
			req.Metadata.RuleId = "755f4478-5a42-4efc-8d45-90a55437aa68"
			_, err = bkServer.UpdateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//empty sepc
			req.Spec = &pb.BucketLifecycleRuleSpec{}
			_, err = bkServer.UpdateBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("DeleteBucketLifecycleRule", func() {
		It("should succesfully delete a lifecycle rule", func() {
			//form request
			req := &pb.BucketLifecycleRuleDeleteRequest{
				Metadata: &pb.BucketLifecycleRuleMetadataRef{
					CloudAccountId: "123456789111",
					BucketId:       bucketId,
					RuleId:         ruleId,
				},
			}
			//make request
			_, err := bkServer.DeleteBucketLifecycleRule(ctx, req)
			Expect(err).To(BeNil())
		})
		It("should fail", func() {
			//form request
			req := &pb.BucketLifecycleRuleDeleteRequest{
				Metadata: &pb.BucketLifecycleRuleMetadataRef{
					CloudAccountId: "",
					BucketId:       bucketId,
					RuleId:         ruleId,
				},
			}
			//missing cloudaccount
			_, err := bkServer.DeleteBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//invalid/incorrect cloud account
			req.Metadata.CloudAccountId = "111111111111"
			_, err = bkServer.DeleteBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.CloudAccountId = "123456789111"
			//empty bucketId
			req.Metadata.BucketId = ""
			_, err = bkServer.DeleteBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//pass bucketName for bucketId
			req.Metadata.BucketId = "123456789111-bucket-lf"
			_, err = bkServer.DeleteBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			req.Metadata.BucketId = bucketId
			//empty ruleId
			req.Metadata.RuleId = ""
			_, err = bkServer.DeleteBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
			//pass incorrect/invalid ruleId
			req.Metadata.BucketId = "755f4478-5a42-4efc-8d45-90a55437aa68"
			_, err = bkServer.DeleteBucketLifecycleRule(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	}) //context
})
