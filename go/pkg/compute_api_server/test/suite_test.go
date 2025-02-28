// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/compute_api_server/test/..." make test-custom
package test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	_ "github.com/amacneil/dbmate/pkg/driver/postgres"
	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpc_rest_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	testDb            *manageddb.TestDb
	managedDb         *manageddb.ManagedDb
	sqlDb             *sql.DB
	grpcService       *server.GrpcService
	restService       *grpc_rest_gateway.RestService
	grpcListenPort    = uint16(0)
	restListenPort    = uint16(0)
	openApiClient     *openapi.APIClient
	tracerProvider    *observability.TracerProvider
	cloudAccountId    string
	instanceTypeName1 string
	instanceTypeName2 string
	qmsQuotaMap       map[string]*pb.ServiceQuotaResource
)

const (
	// Used for load balancer tests to verify quotas
	loadBalancerCustomQuotaAccountId = "112233445566"
	DefaultLoadBalancerQuota         = 5
	DefaultLoadBalancerListenerQuota = 10
	DefaultLoadBalancerSourceIPQuota = 50
)

func TestComputeApiServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compute API Server Suite")
}

var _ = BeforeSuite(func() {
	const (
		region = "us-dev-1"
	)

	grpcutil.UseTLS = false
	log.SetDefaultLogger()
	ctx := context.Background()

	By("Initializing tracing")
	obs := observability.New(ctx)
	tracerProvider = obs.InitTracer(ctx)

	By("Starting database")
	testDb = &manageddb.TestDb{}
	var err error
	managedDb, err = testDb.Start(ctx)
	Expect(err).Should(Succeed())
	Expect(managedDb).ShouldNot(BeNil())
	Expect(managedDb.Migrate(ctx, db.MigrationsFs, db.MigrationsDir)).Should(Succeed())
	sqlDb, err = managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	Expect(sqlDb).ShouldNot(BeNil())

	By("Populating QmsQuotaMap")
	populateQmsQuotaMap()

	By("Creating mock VM Instance Scheduling Service")
	vmInstanceSchedulingService := NewMockInstanceSchedulingServiceClient()

	By("Creating mock Billing Deactivate Instances Service")
	billingDeactivateInstancesService := NewMockBillingDeactivateInstancesServiceClient(ctx)

	By("Creating mock CloudAccount Service")
	cloudAccountService := NewMockCloudAccountServiceClient()

	By("Creating mock CloudAccount App Client Service")
	cloudAccountAppClientService := NewMockCloudAccountAppClientServiceClient()

	By("Creating mock Object Storage Service")
	objectStoragePrivateService := NewMockObjectStorageServicePrivateClient()

	By("Creating mock Object for QMS Service")
	qmsClient := NewMockQuotaManagementPrivateServiceClient()

	By("Starting GRPC server")
	grpcServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	grpcListenPort = uint16(grpcServerListener.Addr().(*net.TCPAddr).Port)

	cloudaccounts := make(map[string]config.LaunchQuota)
	instanceQuota := map[string]int{
		"vm-spr-sml":        5,
		"vm-spr-med":        2,
		"vm-spr-lrg":        0,
		"bm-spr":            1,
		"bm-spr-pvc-1100-4": 1,
		"bm-icp-gaudi2":     2,
		"bm-spr-gaudi2":     32,
		instanceTypeName1:   1,
		instanceTypeName2:   1,
	}

	cloudaccounts["STANDARD"] = config.LaunchQuota{
		InstanceQuota: instanceQuota,
	}
	grpcService, err = server.New(ctx, &config.Config{
		Region:                         region,
		ListenPort:                     grpcListenPort,
		PurgeInstanceInterval:          time.Duration(5 * time.Minute),
		PurgeInstanceAge:               time.Duration(5 * time.Minute),
		GetDeactivateInstancesInterval: time.Duration(5 * time.Minute),
		CloudAccountQuota: config.CloudAccountQuota{
			CloudAccounts:                    cloudaccounts,
			DefaultLoadBalancerQuota:         DefaultLoadBalancerQuota, // Set to 5 to allow the LoadBalancerServiceSearch tests to pass
			DefaultLoadBalancerListenerQuota: DefaultLoadBalancerListenerQuota,
			DefaultLoadBalancerSourceIPQuota: DefaultLoadBalancerSourceIPQuota,
			CloudAccountIDQuotas: map[string]config.CloudAccountIDQuota{
				loadBalancerCustomQuotaAccountId: {
					LoadbalancerQuota:         1,
					LoadbalancerListenerQuota: 2,
					LoadbalancerSourceIPQuota: 2,
				},
			},
		},
		AcceleratorInterface: config.AcceleratorInterface{
			EnabledInstanceTypes: []string{
				"bm-virtual",
				"bm-virtual-sc",
				"bm-spr-gaudi2",
			},
		},
	}, managedDb, vmInstanceSchedulingService, billingDeactivateInstancesService, cloudAccountService, cloudAccountAppClientService, objectStoragePrivateService, nil, qmsClient, grpcServerListener)
	Expect(err).Should(Succeed())
	Expect(grpcService.Start(ctx)).Should(Succeed())

	By("Starting GRPC-REST gateway")
	restServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	restListenPort = uint16(restServerListener.Addr().(*net.TCPAddr).Port)
	restService, err = grpc_rest_gateway.New(ctx, &grpc_rest_gateway.Config{
		TargetAddr: fmt.Sprintf("localhost:%d", grpcListenPort),
	}, restServerListener)
	Expect(err).Should(Succeed())
	restService.AddService(pb.RegisterSshPublicKeyServiceHandler)
	restService.AddService(pb.RegisterVNetServiceHandler)
	restService.AddService(pb.RegisterInstanceServiceHandler)
	restService.AddService(pb.RegisterInstanceGroupServiceHandler)
	restService.AddService(pb.RegisterLoadBalancerServiceHandler)
	Expect(restService.Start(ctx)).Should(Succeed())

	By("Creating OpenAPI client")
	clientConfig := openapi.NewConfiguration()
	clientConfig.Scheme = "http"
	clientConfig.Host = fmt.Sprintf("localhost:%d", restListenPort)
	openApiClient = openapi.NewAPIClient(clientConfig)

	By("Pinging service until it comes up")
	Eventually(func() error {
		_, _, err = openApiClient.SshPublicKeyServiceApi.SshPublicKeyServicePing(ctx).Execute()
		return err
	}, time.Second*10, time.Millisecond*500).Should(Succeed())
})

var _ = AfterSuite(func() {
	ctx := context.Background()
	By("Stopping GRPC-REST gateway")
	Expect(restService.Stop(ctx)).Should(Succeed())
	By("Stopping GRPC server")
	Expect(grpcService.Stop(ctx)).Should(Succeed())
	By("Stopping database")
	Expect(testDb.Stop(ctx)).Should(Succeed())
	By("Stopping tracing")
	Expect(tracerProvider.Shutdown(ctx)).Should(Succeed())
})

func getInstanceGrpcClient() pb.InstancePrivateServiceClient {
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	return pb.NewInstancePrivateServiceClient(clientConn)
}

func getInstanceGroupGrpcClient() pb.InstanceGroupServiceClient {
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	return pb.NewInstanceGroupServiceClient(clientConn)
}

func getLoadBalancerGrpcClient() pb.LoadBalancerPrivateServiceClient {
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	return pb.NewLoadBalancerPrivateServiceClient(clientConn)
}

func clearDatabase(ctx context.Context) {
	By("Clearing database")
	db, err := managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	defer db.Close()
	_, err = db.ExecContext(ctx, "delete from instance")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from ssh_public_key")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from vnet")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from machine_image")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from loadbalancer")
	Expect(err).Should(Succeed())
}

func NewMockInstanceSchedulingServiceClient() pb.InstanceSchedulingServiceClient {
	mockController := gomock.NewController(GinkgoT())
	schedClient := pb.NewMockInstanceSchedulingServiceClient(mockController)

	schedClient.EXPECT().Schedule(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *pb.ScheduleRequest, opts ...grpc.CallOption) (*pb.ScheduleResponse, error) {
		scheduleResponse := &pb.ScheduleResponse{}
		instanceResults := make([]*pb.ScheduleInstanceResult, len(req.Instances))
		if len(req.Instances) == 0 {
			return scheduleResponse, nil
		}
		if len(req.Instances) == 1 && req.Instances[0].Spec.ClusterGroupId == "" {
			scheduleResponse = &pb.ScheduleResponse{
				InstanceResults: []*pb.ScheduleInstanceResult{
					{
						ClusterId: "cluster1",
						NodeId:    "node1",
					},
				},
			}
		} else {
			for i := 0; i < len(req.Instances); i++ {
				resp := &pb.ScheduleInstanceResult{
					ClusterId: "cluster1",
					NodeId:    "node" + strconv.Itoa(i+1),
					GroupId:   "1",
				}
				instanceResults[i] = resp
			}
			scheduleResponse = &pb.ScheduleResponse{
				InstanceResults: instanceResults,
			}
		}
		return scheduleResponse, nil
	}).AnyTimes()

	return schedClient
}

func NewMockCloudAccountServiceClient() pb.CloudAccountServiceClient {
	mockController := gomock.NewController(GinkgoT())
	cloudAccountClient := pb.NewMockCloudAccountServiceClient(mockController)

	cloudAccount := &pb.CloudAccount{
		Type: pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}

	cloudAccountClient.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(cloudAccount, nil).AnyTimes()
	return cloudAccountClient
}

func NewMockCloudAccountAppClientServiceClient() pb.CloudAccountAppClientServiceClient {
	mockController := gomock.NewController(GinkgoT())
	cloudAccountAppClientServiceClient := pb.NewMockCloudAccountAppClientServiceClient(mockController)

	cloudAccount := &pb.CloudAccount{
		Type: pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}

	cloudAccountAppClientServiceClient.EXPECT().GetAppClientCloudAccount(gomock.Any(), gomock.Any()).Return(cloudAccount, nil).AnyTimes()
	return cloudAccountAppClientServiceClient
}

func NewMockBillingDeactivateInstancesServiceClient(ctx context.Context) pb.BillingDeactivateInstancesServiceClient {
	mockController := gomock.NewController(GinkgoT())
	cloudAccountId = cloudaccount.MustNewId()
	instanceTypeName1 = "instanceTypeName1"
	instanceTypeName2 = "instanceTypeName2"

	billingDeactivateInstancesServiceClient := pb.NewMockBillingDeactivateInstancesServiceClient(mockController)

	deactivateInstancesResponses := []*pb.DeactivateInstancesResponse{
		{
			DeactivationList: []*pb.DeactivateInstances{
				{
					CloudAccountId: cloudAccountId,
					Quotas: []*pb.InstanceQuotas{
						{InstanceType: instanceTypeName1, Limit: 0},
					},
				},
			},
		},
		{
			DeactivationList: []*pb.DeactivateInstances{
				{
					CloudAccountId: cloudAccountId,
					Quotas: []*pb.InstanceQuotas{
						{InstanceType: instanceTypeName2, Limit: 0},
					},
				},
			},
		},
	}
	mockStream := pb.NewMockBillingDeactivateInstancesService_GetDeactivateInstancesStreamClient(mockController)
	mockStream.EXPECT().Recv().Return(deactivateInstancesResponses[0], nil).Times(1)
	mockStream.EXPECT().Recv().Return(deactivateInstancesResponses[1], nil).Times(1)
	mockStream.EXPECT().Recv().Return(nil, io.EOF).Times(1)
	billingDeactivateInstancesServiceClient.EXPECT().GetDeactivateInstancesStream(gomock.Any(), &emptypb.Empty{}).Return(mockStream, nil).AnyTimes()
	return billingDeactivateInstancesServiceClient
}

func NewMockObjectStorageServicePrivateClient() pb.ObjectStorageServicePrivateClient {
	mockController := gomock.NewController(GinkgoT())
	objectStorageServicePrivateClient := pb.NewMockObjectStorageServicePrivateClient(mockController)

	objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	objectStorageServicePrivateClient.EXPECT().RemoveBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	return objectStorageServicePrivateClient
}

func populateQmsQuotaMap() {
	qmsQuotaMap = map[string]*pb.ServiceQuotaResource{
		"vm-spr-sml": {
			ResourceType: "vm-spr-sml",
			QuotaConfig: &pb.QuotaConfig{
				Limits:    5,
				QuotaUnit: "COUNT",
			},
			Scope: &pb.QuotaScope{
				ScopeType:  "QUOTA_ACCOUNT_TYPE",
				ScopeValue: "STANDARD",
			},
		},
		"vm-spr-med": {
			ResourceType: "vm-spr-med",
			QuotaConfig: &pb.QuotaConfig{
				Limits:    2,
				QuotaUnit: "COUNT",
			},
			Scope: &pb.QuotaScope{
				ScopeType:  "QUOTA_ACCOUNT_TYPE",
				ScopeValue: "STANDARD",
			},
		},
		"vm-spr-lrg": {
			ResourceType: "vm-spr-lrg",
			QuotaConfig: &pb.QuotaConfig{
				Limits:    0,
				QuotaUnit: "COUNT",
			},
			Scope: &pb.QuotaScope{
				ScopeType:  "QUOTA_ACCOUNT_TYPE",
				ScopeValue: "STANDARD",
			},
		},
		"bm-spr": {
			ResourceType: "bm-spr",
			QuotaConfig: &pb.QuotaConfig{
				Limits:    1,
				QuotaUnit: "COUNT",
			},
			Scope: &pb.QuotaScope{
				ScopeType:  "QUOTA_ACCOUNT_TYPE",
				ScopeValue: "STANDARD",
			},
		},
		"bm-spr-pvc-1100-4": {
			ResourceType: "bm-spr-pvc-1100-4",
			QuotaConfig: &pb.QuotaConfig{
				Limits:    1,
				QuotaUnit: "COUNT",
			},
			Scope: &pb.QuotaScope{
				ScopeType:  "QUOTA_ACCOUNT_TYPE",
				ScopeValue: "STANDARD",
			},
		},
		"bm-icp-gaudi2": {
			ResourceType: "bm-icp-gaudi2",
			QuotaConfig: &pb.QuotaConfig{
				Limits:    2,
				QuotaUnit: "COUNT",
			},
			Scope: &pb.QuotaScope{
				ScopeType:  "QUOTA_ACCOUNT_TYPE",
				ScopeValue: "STANDARD",
			},
		},
		"bm-spr-gaudi2": {
			ResourceType: "bm-icp-gaudi2",
			QuotaConfig: &pb.QuotaConfig{
				Limits:    32,
				QuotaUnit: "COUNT",
			},
			Scope: &pb.QuotaScope{
				ScopeType:  "QUOTA_ACCOUNT_TYPE",
				ScopeValue: "STANDARD",
			},
		},
		"instanceTypeName1": {
			ResourceType: "instanceTypeName1",
			QuotaConfig: &pb.QuotaConfig{
				Limits:    1,
				QuotaUnit: "COUNT",
			},
			Scope: &pb.QuotaScope{
				ScopeType:  "QUOTA_ACCOUNT_TYPE",
				ScopeValue: "STANDARD",
			},
		},
		"instanceTypeName2": {
			ResourceType: "instanceTypeName2",
			QuotaConfig: &pb.QuotaConfig{
				Limits:    1,
				QuotaUnit: "COUNT",
			},
			Scope: &pb.QuotaScope{
				ScopeType:  "QUOTA_ACCOUNT_TYPE",
				ScopeValue: "STANDARD",
			},
		},
	}
}

func NewMockQuotaManagementPrivateServiceClient() pb.QuotaManagementPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	client := pb.NewMockQuotaManagementPrivateServiceClient(mockController)
	client.EXPECT().
		GetResourceQuotaPrivate(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *pb.ServiceQuotaResourceRequestPrivate, opts ...grpc.CallOption) (*pb.ServiceQuotasPrivate, error) {
			if req == nil {
				return nil, fmt.Errorf("request is nil")
			}

			resource := qmsQuotaMap[req.ResourceType]

			return &pb.ServiceQuotasPrivate{
				DefaultQuota: &pb.ServiceQuotaPrivate{
					ServiceName: req.ServiceName,
					ServiceResources: []*pb.ServiceQuotaResource{
						resource,
					},
				},
			}, nil
		}).AnyTimes()

	return client
}

func CreateMachineImage(ctx context.Context, instanceCategories []pb.InstanceCategory, instanceTypes []string, machineImageName string, hidden bool) (pb.MachineImageServiceClient, error) {
	By("Creating machine image client")
	computeApiServerAddress := fmt.Sprintf("localhost:%d", grpcListenPort)
	clientConn, err := grpc.Dial(computeApiServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).Should(Succeed())
	machineImageClient := pb.NewMachineImageServiceClient(clientConn)

	By("Creating MachineImage")
	machineImage := &pb.MachineImage{
		Metadata: &pb.MachineImage_Metadata{
			Name: machineImageName,
		},
		Spec: &pb.MachineImageSpec{
			DisplayName:        "Ubuntu 22.04 LTS (Jammy Jellyfish) v20230128",
			Description:        "Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.",
			Icon:               "https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png",
			InstanceCategories: instanceCategories,
			InstanceTypes:      instanceTypes,
			Md5Sum:             "764efa883dda1e11db47671c4a3bbd9e",
			Sha256Sum:          "98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4",
			Labels: map[string]string{
				"architecture": "X86_64",
				"family":       "ubuntu-2204-lts",
			},
			ImageCategories: []string{
				"AI",
			},
			Components: []*pb.MachineImageComponent{
				{
					Name:        "Ubuntu 22.04 LTS",
					Type:        "OS",
					Version:     "22.04",
					Description: "Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.",
					InfoUrl:     "https://discourse.ubuntu.com/t/jammy-jellyfish-release-notes/24668?_ga=2.61253994.716223186.1673128475-452578204.1673128475",
					ImageUrl:    "https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png",
				},
			},
			Hidden: hidden,
		},
	}
	_, err = machineImageClient.Put(ctx, machineImage)
	return machineImageClient, err
}
