package compute_api_comms_test

import (
	"context"
	"database/sql"
	"fmt"
	"goFramework/ginkGo/test_cases/testutils"
	"net"
	"os"
	"testing"
	"time"

	_ "github.com/amacneil/dbmate/pkg/driver/postgres"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
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

var (
	testDb            *testutils.SimDB
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
)

func TestComputeApiServer(t *testing.T) {
	if os.Getenv("MULTI_RUNNER") != "" {
		t.Skip("Skipping not suitable for multi runner container")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compute API Server Suite")
}

var _ = BeforeSuite(func() {
	grpcutil.UseTLS = false
	log.SetDefaultLogger()
	ctx := context.Background()

	By("Initializing tracing")
	obs := observability.New(ctx)
	tracerProvider = obs.InitTracer(ctx)

	By("Starting database")
	testDb = &testutils.SimDB{}
	var err error
	managedDb, err = testDb.Start(ctx)
	Expect(err).Should(Succeed())
	Expect(managedDb).ShouldNot(BeNil())
	Expect(managedDb.Migrate(ctx, db.MigrationsFs, db.MigrationsDir)).Should(Succeed())
	sqlDb, err = managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	Expect(sqlDb).ShouldNot(BeNil())

	By("Creating mock VM Instance Scheduling Service")
	vmInstanceSchedulingService := NewMockInstanceSchedulingServiceClient()

	By("Creating mock Billing Deactivate Instances Service")
	billingDeactivateInstancesService := NewMockBillingDeactivateInstancesServiceClient(ctx)

	By("Creating mock CloudAccount Service")
	cloudAccountService := NewMockCloudAccountServiceClient()

	By("Creating mock Object Storage Service")
	objectStorageServicePrivateClient := pb.NewMockObjectStorageServicePrivateClient(gomock.NewController(GinkgoT()))
	Expect(objectStorageServicePrivateClient).ShouldNot(BeNil())

	By("Starting GRPC server")
	grpcServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	grpcListenPort = uint16(grpcServerListener.Addr().(*net.TCPAddr).Port)
	cloudaccounts := make(map[string]config.LaunchQuota)
	instanceQuota := map[string]int{
		"vm-spr-sml":        5,
		"vm-spr-med":        0,
		"vm-spr-lrg":        0,
		"bm-spr":            1,
		"bm-spr-pvc-1100-4": 1,
		"bm-icp-gaudi2":     0,
		instanceTypeName1:   1,
		instanceTypeName2:   1,
	}

	cloudaccounts["STANDARD"] = config.LaunchQuota{
		InstanceQuota: instanceQuota,
	}
	grpcService, err = server.New(ctx, &config.Config{
		ListenPort:                     grpcListenPort,
		PurgeInstanceInterval:          time.Duration(5 * time.Minute),
		PurgeInstanceAge:               time.Duration(5 * time.Minute),
		GetDeactivateInstancesInterval: time.Duration(5 * time.Minute),
		CloudAccountQuota: config.CloudAccountQuota{
			CloudAccounts: cloudaccounts,
		},
	}, managedDb, vmInstanceSchedulingService, billingDeactivateInstancesService, cloudAccountService, objectStorageServicePrivateClient, nil, grpcServerListener)
	Expect(err).Should(Succeed())
	Expect(grpcService.Start(ctx)).Should(Succeed())

	By("Starting GRPC-REST gateway")
	restServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	restListenPort = uint16(restServerListener.Addr().(*net.TCPAddr).Port)
	restService, err = grpc_rest_gateway.New(ctx, &grpc_rest_gateway.Config{
		TargetAddr: testutils.CheckEnvironmentAndGetLocalHost(grpcListenPort),
	}, restServerListener)
	Expect(err).Should(Succeed())
	restService.AddService(pb.RegisterSshPublicKeyServiceHandler)
	restService.AddService(pb.RegisterVNetServiceHandler)
	restService.AddService(pb.RegisterInstanceServiceHandler)
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

func clearDatabase(ctx context.Context) {
	db, err := managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from instance")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from ssh_public_key")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from vnet")
	Expect(err).Should(Succeed())
}

func NewMockInstanceSchedulingServiceClient() pb.InstanceSchedulingServiceClient {
	mockController := gomock.NewController(GinkgoT())
	schedClient := pb.NewMockInstanceSchedulingServiceClient(mockController)
	scheduleResponse := &pb.ScheduleResponse{
		InstanceResults: []*pb.ScheduleInstanceResult{
			{
				ClusterId: "cluster1",
				NodeId:    "node1",
			},
		},
	}
	schedClient.EXPECT().Schedule(gomock.Any(), gomock.Any()).Return(scheduleResponse, nil).AnyTimes()
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

func NewMockBillingDeactivateInstancesServiceClient(ctx context.Context) pb.BillingDeactivateInstancesServiceClient {
	mockController := gomock.NewController(GinkgoT())
	cloudAccountId = cloudaccount.MustNewId()
	instanceTypePrefix := uuid.NewString()
	instanceTypeName1 = fmt.Sprintf("%s-instanceType1", instanceTypePrefix)
	instanceTypeName2 = fmt.Sprintf("%s-instanceType2", instanceTypePrefix)

	billingDeactivateInstancesServiceClient := pb.NewMockBillingDeactivateInstancesServiceClient(mockController)
	deactivateInstancesResponse := &pb.DeactivateInstancesResponse{
		DeactivationList: []*pb.DeactivateInstances{
			{
				CloudAccountId: cloudAccountId,
				Quotas: []*pb.InstanceQuotas{
					{
						InstanceType: instanceTypeName1,
						Limit:        0,
					},
					{
						InstanceType: instanceTypeName2,
						Limit:        0,
					},
				},
			},
		},
	}

	billingDeactivateInstancesServiceClient.EXPECT().GetDeactivateInstances(gomock.Any(), &emptypb.Empty{}).Return(deactivateInstancesResponse, nil).AnyTimes()
	return billingDeactivateInstancesServiceClient
}
