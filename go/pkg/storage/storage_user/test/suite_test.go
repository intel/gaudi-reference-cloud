// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	server "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_user/pkg/server"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	mockUserClient      *mocks.MockUserServiceClient
	mockS3ServiceClient *mocks.MockS3ServiceClient
	strclient           *sc.StorageControllerClient
	fsResponse          *weka.Filesystem
	fsUser              *server.StorageUserServiceServer
	lcServer            *server.BucketLifecycleServiceServer
	mockUserService     *server.UserService
	//mockBucketAPIClient pb.ObjectStorageServicePrivateClient
	nsResponse         *api.Namespace
	mockCtrl           *gomock.Controller
	mockClient         *mocks.MockFilesystemServiceClient
	nsMockClient       *mocks.MockNamespaceServiceClient
	namespaceID        string
	userResponse       *api.User
	ctx                context.Context
	createPrincipalReq *pb.CreateBucketUserParams
	updatePolicyReq    *pb.UpdateBucketUserPolicyParams
	updateCredReq      *pb.UpdateBucketUserCredsParams
	lcCreate           *pb.CreateOrUpdateLifecycleRuleRequest
	lcRule             *api.LifecycleRule
	bucket             *api.Bucket
	err                error
)

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage API Server Suite")
}

var _ = BeforeSuite(func() {
	grpcutil.UseTLS = false
	log.SetDefaultLogger()
	ctx = context.Background()
	kms := NewMockStorageKMSPrivateServiceClient()
	mockCtrl = gomock.NewController(GinkgoT())
	mockFSClient := mocks.NewMockFilesystemServiceClient(mockCtrl)
	nsMockClient = mocks.NewMockNamespaceServiceClient(mockCtrl)
	mockUserClient = mocks.NewMockUserServiceClient(mockCtrl)
	mockS3ServiceClient = mocks.NewMockS3ServiceClient(mockCtrl)
	//mockBucketAPIClient := NewMockObjectStorageServicePrivateClient()

	strclient = &sc.StorageControllerClient{
		WekaFilesystemSvcClient: mockFSClient,
		NamespaceSvcClient:      nsMockClient,
		UserSvcClient:           mockUserClient,
		S3ServiceClient:         mockS3ServiceClient,
	}

	fsUser, err = server.NewStorageUserServiceServer(kms, strclient)
	Expect(err).To(BeNil())
	Expect(fsUser).NotTo(BeNil())

	lcServer, err = server.NewBucketLifecycleServiceServer(strclient)
	Expect(err).To(BeNil())
	Expect(lcServer).NotTo(BeNil())

	namespaceID = "ns_id"
	userID := "user_id"
	nsResponse = &api.Namespace{
		Id: &api.NamespaceIdentifier{
			ClusterId: &api.ClusterIdentifier{
				Uuid: clusterUUID,
			},
			Id: namespaceID,
		},
		Name: "test-namespace",
		Quota: &api.Namespace_Quota{
			TotalBytes: 50000000,
		},
	}
	userResponse = &api.User{
		Id: &api.UserIdentifier{
			NamespaceId: &api.NamespaceIdentifier{
				ClusterId: &api.ClusterIdentifier{
					Uuid: clusterUUID,
				},
				Id: namespaceID,
			},
			Id: userID,
		},
		Name: "test-user",
		Role: api.User_ROLE_ADMIN,
	}
	policy := &pb.ObjectUserPermissionSpec{
		BucketId:   "bucket-1",
		Prefix:     "/",
		Permission: []pb.BucketPermission{pb.BucketPermission_DeleteBucket, pb.BucketPermission_WriteBucket, pb.BucketPermission_ReadBucket},
		Actions: []pb.ObjectBucketActions{
			pb.ObjectBucketActions_GetBucketLocation,
			pb.ObjectBucketActions_GetBucketPolicy,
			pb.ObjectBucketActions_GetBucketTagging,
			pb.ObjectBucketActions_ListBucket,
			pb.ObjectBucketActions_ListBucketMultipartUploads,
			pb.ObjectBucketActions_ListMultipartUploadParts,
		},
	}
	createPrincipalReq = &pb.CreateBucketUserParams{
		CloudAccountId: "123456789012",
		CreateParams: &pb.BucketUserParams{
			Name:          "tester",
			ClusterUUID:   clusterid,
			UserId:        "8670ccaa-704e-4839-bc72-9a89daa20111",
			Password:      "password",
			Spec:          []*pb.ObjectUserPermissionSpec{policy},
			SecurityGroup: &pb.BucketSecurityGroup{},
		},
	}
	subnet := &pb.BucketNetworkGroup{
		Subnet:       "0.0.0.0",
		PrefixLength: 27,
		Gateway:      "0.0.0.0",
	}
	securityGroup := &pb.BucketSecurityGroup{
		NetworkFilterAllow: []*pb.BucketNetworkGroup{subnet},
		NetworkFilterDeny:  []*pb.BucketNetworkGroup{subnet},
	}
	updatePolicyReq = &pb.UpdateBucketUserPolicyParams{
		CloudAccountId: "123456789012",
		UpdateParams: &pb.BucketUpdateUserPolicyParams{
			PrincipalId:   "tester",
			ClusterUUID:   clusterid,
			Spec:          []*pb.ObjectUserPermissionSpec{policy},
			SecurityGroup: securityGroup,
		},
	}
	updateCredReq = &pb.UpdateBucketUserCredsParams{
		CloudAccountId: "123456789012",
		UpdateParams: &pb.BucketUpdateUserCredsParams{
			PrincipalId: "tester",
			ClusterUUID: clusterid,
			UserId:      "username",
			Password:    "password",
		},
	}
	bucket = &api.Bucket{
		Id: &api.BucketIdentifier{
			ClusterId: &api.ClusterIdentifier{
				Uuid: clusterid,
			},
			Id: clusterid,
		},
		Name:      "test-bucket",
		Versioned: false,
		Capacity: &api.Bucket_Capacity{
			TotalBytes:     1000000000,
			AvailableBytes: 1000000000,
		},
		EndpointUrl: "some-url",
	}
	lcCreate = &pb.CreateOrUpdateLifecycleRuleRequest{
		CloudAccountId: cloudaccountId,
		ClusterId:      clusterid,
		BucketId:       "test-bucket",
		Spec: []*pb.BucketLifecycleRuleSpec{
			{
				Prefix:               "/",
				ExpireDays:           5,
				NoncurrentExpireDays: 5,
				DeleteMarker:         false,
			},
		},
	}
	lcRule = &api.LifecycleRule{
		Id: &api.LifecycleRuleIdentifier{
			Id: clusterid,
		},
		Prefix:               "/",
		ExpireDays:           5,
		NoncurrentExpireDays: 5,
		DeleteMarker:         false,
	}
})

// Helper function to mock KMS service client
func NewMockStorageKMSPrivateServiceClient() pb.StorageKMSPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	kms := pb.NewMockStorageKMSPrivateServiceClient(mockController)
	// TODO: Create mock return object for kms, refer to NewMockCloudAccountServiceClient() for example
	kmsClient := &pb.GetSecretResponse{}
	// Mock the Put call for KMSClient
	kms.EXPECT().Get(gomock.Any(), gomock.Any()).Return(kmsClient, nil).AnyTimes()
	kms.EXPECT().Put(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).AnyTimes()

	return kms
}

func NewMockObjectStorageServicePrivateClient() pb.ObjectStorageServicePrivateClient {
	mockController := gomock.NewController(GinkgoT())
	objectStorageServicePrivateClient := pb.NewMockObjectStorageServicePrivateClient(mockController)

	objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	objectStorageServicePrivateClient.EXPECT().RemoveBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	return objectStorageServicePrivateClient
}
