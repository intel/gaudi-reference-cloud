package test

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create filesystem", func() {
	Context("create fs org", func() {
		It("Should create fs org private", func() {
			By("Creating a filesystem namespace")
			ctx := context.Background()
			meta := &pb.FilesystemMetadataPrivate{
				CloudAccountId:   "111111111112",
				Name:             "test",
				SkipQuotaCheck:   true,
				SkipProductCheck: true,
			}
			request := &pb.FilesystemCapacity{
				Storage: "50GB",
			}
			spec := &pb.FilesystemSpecPrivate{
				AvailabilityZone: "az1",
				Request:          request,
				Prefix:           "iks",
				FilesystemType:   pb.FilesystemType_ComputeKubernetes,
			}
			req := &pb.FilesystemOrgCreateRequestPrivate{
				Metadata: meta,
				Spec:     spec,
			}
			// Call the Create method of fsServer and capture the response
			resp, err := fsServer.CreateFilesystemOrgPrivate(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

		})
	})

	Context("get fs org", func() {
		It("Should obtain the fs org private details", func() {
			By("Fetch a filesystem org namespace")
			ctx := context.Background()
			meta := &pb.FilesystemMetadataReference{
				CloudAccountId: "111111111112",
				NameOrId: &pb.FilesystemMetadataReference_Name{
					Name: "test",
				},
			}

			req := &pb.FilesystemOrgGetRequestPrivate{
				Metadata: meta,
				Prefix:   "iks",
			}
			// Call the get method of fsServer and capture the response
			resp, err := fsServer.GetFilesystemOrgPrivate(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

		})
	})

	Context("delete fs org", func() {
		It("Should fail to delete the fs org", func() {
			By("delete a filesystem while still in provisioning state")
			ctx := context.Background()
			meta := &pb.FilesystemMetadataReference{
				CloudAccountId: "111111111112",
				NameOrId: &pb.FilesystemMetadataReference_Name{
					Name: "test",
				},
			}

			req := &pb.FilesystemOrgDeleteRequestPrivate{
				Metadata: meta,
				Prefix:   "iks",
			}
			// Call the get method of fsServer and capture the response
			_, err := fsServer.DeleteFilesystemOrgPrivate(ctx, req)
			Expect(err).NotTo(BeNil())

		})
	})

	Context("get fs orgs", func() {
		It("Should obtain org details", func() {
			By("Fetch orgs in cluster")
			ctx := context.Background()
			meta := &pb.FilesystemMetadataReference{
				CloudAccountId: "111111111112",
				NameOrId: &pb.FilesystemMetadataReference_Name{
					Name: "test",
				},
			}

			req := &pb.FilesystemOrgsListRequestPrivate{
				Metadata:  meta,
				Prefix:    "iks",
				ClusterId: "test-cluster",
			}
			// Call the list method of fsServer and capture the response
			resp, err := fsServer.ListFilesystemOrgsPrivate(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

		})
	})

	Context("get fs in orgs", func() {
		It("Should obtain filesystem in org details", func() {
			By("Fetch filesystems in orgs")
			ctx := context.Background()
			meta := &pb.FilesystemMetadataReference{
				CloudAccountId: "111111111112",
				NameOrId: &pb.FilesystemMetadataReference_Name{
					Name: "test",
				},
			}

			req := &pb.FilesystemsInOrgListRequestPrivate{
				Metadata:  meta,
				Prefix:    "iks",
				ClusterId: "test-cluster",
			}
			// Call the list method of fsServer and capture the response
			resp, err := fsServer.ListFilesystemsInOrgPrivate(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

		})
	})

})
