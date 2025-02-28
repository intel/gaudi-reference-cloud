// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"io"
	logr "log"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	fs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/api_server/pkg/server"
	server "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/api_server/pkg/server"
	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ctx             context.Context
	testDb          *manageddb.TestDb
	managedDb       *manageddb.ManagedDb
	sqlDb           *sql.DB
	bucketId        string
	ruleId          string
	bkServer        *fs.BucketsServiceServer
	bkUsrClient     pb.BucketUserPrivateServiceClient
	lifecycleClient pb.BucketLifecyclePrivateServiceClient
)

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage API Server Suite")
}

var _ = BeforeSuite(func() {
	grpcutil.UseTLS = false
	log.SetDefaultLogger()
	ctx := context.Background()

	By("Starting database")
	testDb = &manageddb.TestDb{}
	var err error
	managedDb, err = testDb.Start(ctx)
	cfg := &server.Config{
		AuthzEnabled: true,
	}
	vastCfg := &server.Config{
		AuthzEnabled:              true,
		GeneralPurposeVASTEnabled: true,
	}
	// Check Db setup succeeds
	Expect(err).Should(Succeed())
	Expect(managedDb).ShouldNot(BeNil())
	Expect(managedDb.Migrate(ctx, db.Fsys, "migrations")).Should(Succeed())
	sqlDb, err = managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	Expect(sqlDb).ShouldNot(BeNil())

	permission := &pb.ObjectUserPermissionSpec{
		BucketId:   "123456789012-bucket-test",
		Prefix:     "string",
		Permission: []pb.BucketPermission{pb.BucketPermission_ReadBucket},
		Actions:    []pb.ObjectBucketActions{pb.ObjectBucketActions_GetBucketLocation},
	}
	specs = append(specs, permission)

	// Set up mock cloudaccount
	cloudAccountClient := NewMockCloudAccountServiceClient()

	// Set up quota
	cloudaccounts := make(map[string]server.LaunchQuota)
	storageQuota := map[string]int{
		"filesystems":     5,
		"totalSizeGB":     50000,
		"vm-icp-gaudi2-1": 8,
		"buckets":         5,
	}
	cloudaccounts["STANDARD"] = server.LaunchQuota{
		StorageQuota: storageQuota,
	}
	quota := server.CloudAccountQuota{
		CloudAccounts: cloudaccounts,
	}

	// set up mock schedulerClient
	schedularClient := NewMockFilesystemSchedulerPrivateServiceClient()

	//set up mock kmsClient
	kms := NewMockStorageKMSPrivateServiceClient()

	//set up mock productClient
	productClient := NewMockProductCatalogServiceClient()

	//set up mock autzClient
	authzClient := NewMockAuthzServiceClient()

	//set up mock computeAPIClient
	computeAPIClient := NewMockInstancePrivateServiceClient()

	//set up quota management client
	qmsClient := NewMockQuotaManagementPrivateServiceClient()

	//set up quota service
	qService := &server.QuotaService{}
	qService.Init(ctx, sqlDb, quota, qmsClient)

	//set up user client
	userClient := NewMockFilesystemUserServiceClient()

	//set up bucket user client
	bkUsrClient = NewMockBucketUserPrivateServiceClient()

	//set up bucket lifecycle client
	lifecycleClient = NewMockBucketLifecyclePrivateServiceClient()

	//set up weka agent client
	wekaClient := NewMockWekaStatefuleAgentPrivateServiceClient()

	//Initialize filesystemservice with sqlDb, cloudClient and quota
	fsServer, err = server.NewFilesystemService(ctx, sqlDb, cloudAccountClient, productClient, authzClient, schedularClient, kms, qService, userClient, wekaClient, cfg)
	fsServer2, err = server.NewFilesystemService(ctx, sqlDb, cloudAccountClient, productClient, authzClient, schedularClient, kms, qService, userClient, wekaClient, cfg)
	fsServer3, err = server.NewFilesystemService(ctx, sqlDb, cloudAccountClient, productClient, authzClient, schedularClient, kms, qService, userClient, wekaClient, vastCfg)

	Expect(err).Should(Succeed())
	Expect(fsServer).ShouldNot(BeNil())

	bkSize := "10000000000"
	//Initialize bucketservice
	bkServer, err = server.NewObjectService(ctx, sqlDb, bkSize, cloudAccountClient, productClient, authzClient, schedularClient, qService, userClient, bkUsrClient, lifecycleClient, computeAPIClient, cfg)
	Expect(err).To(BeNil())

	//Create testing bucket
	lfmeta := &pb.ObjectBucketCreateMetadata{
		CloudAccountId: "123456789111",
		Name:           "bucket-lf",
	}
	lfspec := &pb.ObjectBucketSpec{
		AvailabilityZone: "az1",
		Request: &pb.StorageCapacityRequest{
			Size: "10000000000",
		},
		Versioned:    false,
		AccessPolicy: pb.BucketAccessPolicy_READ_WRITE,
	}
	lfreq := &pb.ObjectBucketCreateRequest{
		Metadata: lfmeta,
		Spec:     lfspec,
	}
	//create bucket
	bucket, err := bkServer.CreateBucket(ctx, lfreq)
	Expect(err).To(BeNil())
	Expect(bucket).NotTo(BeNil())
	bucketId = bucket.Metadata.ResourceId

})
var _ = AfterSuite(func() {
	//Delete test bucket
	req := &pb.ObjectBucketDeleteRequest{
		Metadata: &pb.ObjectBucketMetadataRef{
			CloudAccountId: "123456789111",
			NameOrId:       &pb.ObjectBucketMetadataRef_BucketName{BucketName: "123456789111-bucket-lf"},
		},
	}
	_, err := bkServer.DeleteBucket(ctx, req)
	Expect(err).To(BeNil())
})

// Helper function to mock cloud acc service client
func NewMockCloudAccountServiceClient() pb.CloudAccountServiceClient {
	mockController := gomock.NewController(GinkgoT())
	cloudAccountClient := pb.NewMockCloudAccountServiceClient(mockController)

	// create mock cloudaccount
	cloudAccount := &pb.CloudAccount{
		Type: pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}

	cloudAccountClient.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(cloudAccount, nil).AnyTimes()
	return cloudAccountClient
}

// Helper function to mock KMS service client
func NewMockStorageKMSPrivateServiceClient() pb.StorageKMSPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	kms := pb.NewMockStorageKMSPrivateServiceClient(mockController)
	kmsClient := &pb.GetSecretResponse{}
	// Mock the Put call for KMSClient
	kms.EXPECT().Get(gomock.Any(), gomock.Any()).Return(kmsClient, nil).AnyTimes()
	kms.EXPECT().Put(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).AnyTimes()

	return kms
}

// Helper function to mock KMS service client
func NewMockInstancePrivateServiceClient() pb.InstancePrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	inst := pb.NewMockInstancePrivateServiceClient(mockController)

	sampleInstance := &pb.InstancePrivate{
		Status: &pb.InstanceStatusPrivate{
			Interfaces: []*pb.InstanceInterfaceStatusPrivate{
				&pb.InstanceInterfaceStatusPrivate{
					Subnet:       "0.0.0.0",
					PrefixLength: 27,
					Gateway:      "0.0.0.0",
				},
			},
		},
	}

	response := pb.InstanceSearchPrivateResponse{Items: []*pb.InstancePrivate{sampleInstance}}

	inst.EXPECT().SearchPrivate(gomock.Any(), gomock.Any()).Return(&response, nil).AnyTimes()
	return inst
}

// Helper functions to mock productClient
func NewMockProductCatalogServiceClient() pb.ProductCatalogServiceClient {
	mockController := gomock.NewController(GinkgoT())
	product := pb.NewMockProductCatalogServiceClient(mockController)
	m := make(map[string]string)
	// Below sizes are in TB
	m["volume.size.min"] = "1"
	m["volume.size.max"] = "2000"

	// define product
	p1 := &pb.Product{
		Name:     "storage-file",
		Metadata: m,
	}

	// Set mock behavior
	product.EXPECT().AdminRead(gomock.Any(), gomock.Any()).Return(&pb.ProductResponse{
		Products: []*pb.Product{
			p1,
		},
	}, nil).AnyTimes()
	return product
}

// Helper functions to mock authzClient
func NewMockAuthzServiceClient() pb.AuthzServiceClient {
	mockController := gomock.NewController(GinkgoT())
	authz := pb.NewMockAuthzServiceClient(mockController)
	authz.EXPECT().LookupInternal(gomock.Any(), gomock.Any()).Return(&pb.LookupResponse{
		ResourceIds: []string{"8623ccaa-704e-4839-bc72-9a89daa20111"},
	}, nil).AnyTimes()
	authz.EXPECT().RemoveResourceFromCloudAccountRole(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	return authz
}

// Helper function to mock FilesystemSchedulerPrivateServiceClient service client
func NewMockFilesystemSchedulerPrivateServiceClient() pb.FilesystemSchedulerPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	schedular := pb.NewMockFilesystemSchedulerPrivateServiceClient(mockController)
	// Mock the call for Schedular
	schedular.EXPECT().ScheduleFile(gomock.Any(), gomock.Any()).Return(&pb.FilesystemScheduleResponse{
		NewSchedule: false,
		Schedule: &pb.ResourceSchedule{
			ClusterName:    "Weka",
			ClusterAddr:    "",
			ClusterUUID:    "8623ccaa-704e-4839-bc72-9a89daa20111",
			ClusterVersion: "4.2.2", //FIXME: Update cluster versioning
			Namespace:      "test-volume",

			// TODO: Set fields as needed
		},
	}, nil).AnyTimes()
	listResp := &pb.FilesystemStorageClusters{
		ClusterId: "",
		Name:      "",
		Capacity: &pb.StorageClusterCapacity{
			TotalBytes:     1000000000,
			AvailableBytes: 1000000000,
		},
	}
	rs := pb.NewMockFilesystemSchedulerPrivateService_ListClustersClient(mockController)
	gomock.InOrder(
		rs.EXPECT().Recv().Return(listResp, nil),
		rs.EXPECT().Recv().Return(nil, io.EOF),
		rs.EXPECT().Recv().Return(listResp, nil),
		rs.EXPECT().Recv().Return(nil, io.EOF),
		rs.EXPECT().Recv().Return(listResp, nil),
		rs.EXPECT().Recv().Return(nil, io.EOF),
		rs.EXPECT().Recv().Return(listResp, nil),
		rs.EXPECT().Recv().Return(nil, io.EOF),
		rs.EXPECT().Recv().Return(listResp, nil),
		rs.EXPECT().Recv().Return(nil, io.EOF),
		rs.EXPECT().Recv().Return(listResp, nil),
		rs.EXPECT().Recv().Return(nil, io.EOF),
		rs.EXPECT().Recv().Return(listResp, nil),
		rs.EXPECT().Recv().Return(nil, io.EOF),
		rs.EXPECT().Recv().Return(listResp, nil),
		rs.EXPECT().Recv().Return(nil, io.EOF),
		rs.EXPECT().Recv().Return(listResp, nil),
		rs.EXPECT().Recv().Return(nil, io.EOF),
	)
	schedular.EXPECT().ListClusters(gomock.Any(), gomock.Any()).Return(rs, nil).AnyTimes()

	response := &pb.FilesystemOrgsResponsePrivate{
		Org: make([]*pb.FilesystemOrgsPrivate, 0, 1),
	}
	nsPrivate := &pb.FilesystemOrgsPrivate{
		Name: "test-iks",
	}
	response.Org = append(response.Org, nsPrivate)
	exists := &pb.FilesystemOrgsIsExistsResponsePrivate{
		Exists: false,
	}
	schedular.EXPECT().ListFilesystemOrgs(gomock.Any(), gomock.Any()).Return(response, nil).AnyTimes()
	schedular.EXPECT().IsOrgExists(gomock.Any(), gomock.Any()).Return(exists, nil).AnyTimes()

	resp := &pb.FilesystemsInOrgListResponsePrivate{
		Items: make([]*pb.FilesystemPrivate, 0, 1),
	}

	// Convert each filesystem to FilesystemPrivate format.
	fsPrivate := &pb.FilesystemPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			Name:           "test-iks",
			CloudAccountId: "111111111112",
		},
		Spec: &pb.FilesystemSpecPrivate{
			Request: &pb.FilesystemCapacity{
				Storage: "500GB",
			},
		},
	}
	resp.Items = append(resp.Items, fsPrivate)
	schedular.EXPECT().ListFilesystemInOrgs(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()

	schedular.EXPECT().ScheduleBucket(gomock.Any(), gomock.Any()).Return(&pb.BucketScheduleResponse{
		AvailabilityZone: "az1",
		Schedule: &pb.BucketSchedule{
			Cluster: &pb.AssignedCluster{
				ClusterName: "Minio",
				ClusterUUID: "8623ccaa-704e-4839-bc72-9a89daa20111",
				ClusterAddr: "",
			},
		},
	}, nil).AnyTimes()

	return schedular
}

func NewMockFilesystemSchedulerPrivateServiceClientFalse() pb.FilesystemSchedulerPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	schedular := pb.NewMockFilesystemSchedulerPrivateServiceClient(mockController)
	// Mock the call for Schedular
	schedular.EXPECT().ScheduleFile(gomock.Any(), gomock.Any()).Return(&pb.FilesystemScheduleResponse{
		NewSchedule: false,
		Schedule: &pb.ResourceSchedule{
			ClusterName:    "Weka",
			ClusterAddr:    "",
			ClusterUUID:    "8623ccaa-704e-4839-bc72-9a89daa20111",
			ClusterVersion: "4.2.2", //FIXME: Update cluster versioning
			Namespace:      "test-volume",

			// TODO: Set fields as needed
		},
	}, nil).AnyTimes()

	return schedular
}

func NewMockFilesystemUserServiceClient() pb.FilesystemUserPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	user := pb.NewMockFilesystemUserPrivateServiceClient(mockController)
	// set mock behavior if needed
	res := &empty.Empty{}
	user.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any()).Return(res, nil).AnyTimes()
	return user
}

func NewMockBucketLifecyclePrivateServiceClient() pb.BucketLifecyclePrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	client := pb.NewMockBucketLifecyclePrivateServiceClient(mockController)
	// set mock behavior
	client.EXPECT().CreateOrUpdateLifecycleRule(gomock.Any(), gomock.Any()).Return(&pb.LifecycleRulePrivate{
		ClusterId: "918b5026-d516-48c8-bfd3-5998547265b2",
		BucketId:  "123456789111-bucket-lf",
		Spec: []*pb.BucketLifecycleRuleSpec{
			&pb.BucketLifecycleRuleSpec{
				Prefix:               "",
				ExpireDays:           1,
				NoncurrentExpireDays: 5,
				DeleteMarker:         true,
			},
		},
	}, nil).AnyTimes()

	return client
}

func NewMockBucketUserPrivateServiceClient() pb.BucketUserPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	client := pb.NewMockBucketUserPrivateServiceClient(mockController)
	//set mocks if needed
	client.EXPECT().CreateBucketUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(&pb.BucketPrincipal{
		ClusterId:      "123456789086",
		PrincipalId:    "123456789012",
		AccessEndpoint: "",
		Spec:           specs,
	}, nil).Times(1)
	client.EXPECT().DeleteBucketUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).Times(1)
	client.EXPECT().UpdateBucketUserPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(&pb.BucketPrincipal{
		ClusterId:      "123456789086",
		PrincipalId:    "123456789012",
		AccessEndpoint: "",
		Spec:           specs,
	}, nil).Times(1)
	client.EXPECT().UpdateBucketUserCredentials(gomock.Any(), gomock.Any(), gomock.Any()).Return(&pb.BucketPrincipal{
		ClusterId:      "123456789086",
		PrincipalId:    "123456789012",
		AccessEndpoint: "",
		Spec:           specs,
	}, nil).Times(1)
	client.EXPECT().GetBucketCapacity(gomock.Any(), gomock.Any()).Return(&pb.BucketCapacity{
		Id:   "123456789111-bucket-lf",
		Name: "123456789111-bucket-lf",
		Capacity: &pb.Capacity{
			TotalBytes: 10000000000,
		},
	}, nil).AnyTimes()
	return client
}

func NewMockWekaStatefuleAgentPrivateServiceClient() pb.WekaStatefulAgentPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	wekaClient := pb.NewMockWekaStatefulAgentPrivateServiceClient(mockController)
	//set mocks if needed
	agent := &pb.FilesystemAgent{
		ClusterId:        "934b5026-d346-78c8-fcd3-899852346509",
		ClientId:         "824b5026-d346-12c8-fcd3-123452346509",
		Name:             "default",
		CustomStatus:     "",
		PredefinedStatus: "",
	}
	rs := pb.NewMockWekaStatefulAgentPrivateService_ListRegisteredAgentsClient(mockController)
	rs.EXPECT().Recv().Return(agent, nil).Times(1)
	rs.EXPECT().Recv().Return(nil, io.EOF).Times(1)
	rs.EXPECT().Context().Return(context.Background()).AnyTimes()
	wekaClient.EXPECT().RegisterAgent(gomock.Any(), gomock.Any()).Return(agent, nil).Times(1)
	wekaClient.EXPECT().DeRegisterAgent(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).Times(1)
	wekaClient.EXPECT().GetRegisteredAgent(gomock.Any(), gomock.Any()).Return(agent, nil).Times(1)
	wekaClient.EXPECT().ListRegisteredAgents(gomock.Any(), gomock.Any()).Return(rs, nil).Times(1)
	return wekaClient
}

func NewMockQuotaManagementPrivateServiceClient() pb.QuotaManagementPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	client := pb.NewMockQuotaManagementPrivateServiceClient(mockController)
	ServiceResources := []*pb.ServiceQuotaResource{
		{
			QuotaConfig: &pb.QuotaConfig{
				Limits: 50,
			},
		},
	}
	client.EXPECT().GetResourceQuotaPrivate(gomock.Any(), gomock.Any()).Return(&pb.ServiceQuotasPrivate{
		DefaultQuota: &pb.ServiceQuotaPrivate{
			ServiceName:      "",
			ServiceResources: ServiceResources,
		},
		CustomQuota: &pb.ServiceQuotaPrivate{
			ServiceName:      "",
			ServiceResources: ServiceResources,
		},
	}, nil).AnyTimes()

	return client
}

//Command that generated mock functions: mockgen -destination=go/pkg/pb/mock_storage_private_grpc.pb.go -package=pb -source=go/pkg/pb/storage_private_grpc.pb.go

type CustomClaims struct {
	jwt.Claims
	Email        string `json:"email"`
	EnterpriseId string `json:"enterpriseId"`
	CountryCode  string `json:"countryCode"`
}

func CreateTokenJWT(email string) (error, string) {
	signerOpts := &jose.SignerOptions{}
	signerOpts.WithType("JWT")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logr.Fatalf("error generating the key: %v", err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}, signerOpts)
	if err != nil {
		logr.Fatalf("error creating signer: %v", err)
		return err, ""
	}

	customClaims := CustomClaims{
		Claims: jwt.Claims{
			Issuer:    "your-issuer",
			Subject:   "user-id",
			Audience:  jwt.Audience{"your-audience"},
			Expiry:    jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email:        email,
		EnterpriseId: "enterprise-id",
		CountryCode:  "country-code",
	}

	jwtBuilder := jwt.Signed(signer)
	jwtBuilder = jwtBuilder.Claims(customClaims)
	rawToken, err := jwtBuilder.CompactSerialize()
	if err != nil {
		logr.Fatalf("error on compact serialize: %v", err)
		return err, ""
	}
	return nil, rawToken
}

func CreateContextWithToken(email string) context.Context {
	err, rawToken := CreateTokenJWT(email)
	if err != nil {
		logr.Fatalf("error creating jwt token: %v", err)
	}

	md := metadata.Pairs(
		"Authorization", "Bearer "+rawToken,
	)
	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx
}
