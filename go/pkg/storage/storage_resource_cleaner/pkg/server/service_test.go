package server

import (
	"context"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	serviceClient *StorageResourceCleanerService
	emailClient   pb.EmailNotificationServiceClient
	mockCtrl      *gomock.Controller
	ctx           context.Context
	fsClient      pb.FilesystemPrivateServiceClient
	bkClient      pb.ObjectStorageServicePrivateClient
	billingClient pb.BillingDeactivateInstancesServiceClient
	cfg           *Config
	err           error
	fsMap         map[string][]string
	bkMap         map[string][]string
	cloudaccount  string
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Resource Cleaner Suite")
}

var _ = BeforeSuite(func() {
	ctx = context.Background()
	mockCtrl = gomock.NewController(GinkgoT())
	fsClient = NewMockFSClient()
	bkClient = NewMockBKClient()
	billingClient = NewMockBillingClient()

	cloudaccount = "1123456789012"

	// Test config
	cfg = &Config{
		Interval:             1,
		Threshold:            1,
		StorageAPIServerAddr: "",
		BillingServerAddr:    "",
		ServiceEnabled:       true,
		SenderEmail:          "",
		ConsoleUrl:           "",
		PaymentUrl:           "",
	}
	// Create client
	serviceClient, err = NewStorageResourceCleaner(fsClient, bkClient, billingClient, nil, cfg)
	Expect(err).To(BeNil())
	Expect(serviceClient).NotTo(BeNil())

})

var _ = Describe("Storage Resource Cleaner", func() {
	Context("fetchAccounts", func() {
		It("Should retrieve account and names successfully", func() {
			files, buckets := serviceClient.fetchAccounts(ctx)
			Expect(files).NotTo(BeNil())
			Expect(buckets).NotTo(BeNil())

			fsMap = files
			bkMap = buckets
		})
	})
	Context("cleanResources", func() {
		It("Should clean resources successfully", func() {
			err := serviceClient.cleanResources(ctx, fsMap, bkMap)
			Expect(err).To(BeNil())
		})
	})
	Context("deleteFilesystems", func() {
		It("Should delete volumes successfully", func() {
			err := serviceClient.deleteFilesystems(ctx, cloudaccount, []string{"test-volume"})
			Expect(err).To(BeNil())
		})
	})
	Context("deleteBuckets", func() {
		It("Should delete buckets successfully", func() {
			err := serviceClient.deleteBuckets(ctx, cloudaccount, []string{"123456789012-bucket"})
			Expect(err).To(BeNil())
		})
	})
	Context("sendNotification", func() {
		It("Should send email", func() {
			err := serviceClient.sendNotification(ctx, "tester123@gamil.com")
			Expect(err).To(BeNil())
		})
	})
})

func NewMockFSClient() pb.FilesystemPrivateServiceClient {
	client := pb.NewMockFilesystemPrivateServiceClient(mockCtrl)
	resp := pb.NewMockFilesystemPrivateService_SearchFilesystemRequestsClient(mockCtrl)
	resp.EXPECT().Recv().Return(&pb.FilesystemRequestResponse{
		Filesystem: &pb.FilesystemPrivate{
			Metadata: &pb.FilesystemMetadataPrivate{
				CloudAccountId: "123456789012",
				Name:           "test-volume",
			},
			Spec: &pb.FilesystemSpecPrivate{
				Prefix: "",
			},
		},
	}, nil).Times(1)
	resp.EXPECT().Recv().Return(nil, io.EOF).Times(1)
	client.EXPECT().SearchFilesystemRequests(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()
	client.EXPECT().DeletePrivate(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	return client
}

func NewMockBKClient() pb.ObjectStorageServicePrivateClient {
	client := pb.NewMockObjectStorageServicePrivateClient(mockCtrl)
	resp := pb.NewMockObjectStorageServicePrivate_SearchBucketPrivateClient(mockCtrl)
	resp.EXPECT().Recv().Return(&pb.ObjectBucketSearchPrivateResponse{
		Bucket: &pb.ObjectBucketPrivate{
			Metadata: &pb.ObjectBucketMetadataPrivate{
				CloudAccountId: "123456789012",
				Name:           "123456789012-bucket",
			},
		},
	}, nil).Times(1)
	resp.EXPECT().Recv().Return(nil, io.EOF).Times(1)
	client.EXPECT().SearchBucketPrivate(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()
	client.EXPECT().DeleteBucketPrivate(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	return client
}

func NewMockBillingClient() pb.BillingDeactivateInstancesServiceClient {
	client := pb.NewMockBillingDeactivateInstancesServiceClient(mockCtrl)
	resp := pb.NewMockBillingDeactivateInstancesService_GetDeactivatedServiceAccountsClient(mockCtrl)
	client.EXPECT().GetDeactivatedServiceAccounts(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()
	resp.EXPECT().Recv().Return(&pb.DeactivateAccounts{
		CloudAccountId:  "123456789012",
		Email:           "tester123@gamil.com",
		CreditsDepleted: timestamppb.Now(),
	}, nil).Times(1)
	resp.EXPECT().Recv().Return(nil, io.EOF).Times(1)
	return client
}
