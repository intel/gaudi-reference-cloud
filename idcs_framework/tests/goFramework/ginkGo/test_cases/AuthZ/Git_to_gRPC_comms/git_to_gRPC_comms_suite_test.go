package git_to_grpc_comms_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	_ "github.com/amacneil/dbmate/pkg/driver/postgres"
	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpc_rest_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	testDb         *manageddb.TestDb
	managedDb      *manageddb.ManagedDb
	sqlDb          *sql.DB
	grpcService    *GrpcService
	restService    *grpc_rest_gateway.RestService
	grpcListenPort = uint16(0)
	restListenPort = uint16(0)
	openApiClient  *openapi.APIClient
	tracerProvider *observability.TracerProvider
)

func TestGitToGRPCComms(t *testing.T) {
	if os.Getenv("MULTI_RUNNER") != "" {
		t.Skip("Skipping not suitable for multi runner container")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitToGRPCComms Suite")
}

var _ = BeforeSuite(func() {
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

	By("Creating mock VM Instance Scheduling Service")
	vmInstanceSchedulingService := NewMockInstanceSchedulingServiceClient()

	By("Starting GRPC server")
	grpcServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	grpcListenPort = uint16(grpcServerListener.Addr().(*net.TCPAddr).Port)
	grpcService, err = New(ctx, &config.Config{
		ListenPort:            grpcListenPort,
		PurgeInstanceInterval: time.Duration(5 * time.Minute),
		PurgeInstanceAge:      time.Duration(5 * time.Minute),
	}, managedDb, vmInstanceSchedulingService, grpcServerListener)
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
	Expect(restService.Start(ctx)).Should(Succeed())
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
	var err error
	const ROLE = "billing"
	const DOMAIN = "billing.idcs-system.svc.cluster.local"
	var result any

	body := GenerateRequestToGetCertificate(DOMAIN, ROLE)
	jsonerr := json.Unmarshal([]byte(body), &result)
	if jsonerr != nil {
		Fail("Error during Unmarshal(): " + err.Error())
	}

	json := GetFieldInfo(body)
	_, clientTLSConf, err := VaultCertSetup(json["issuing_ca"].(string), json["private_key"].(string), json["certificate"].(string))
	if err != nil {
		fmt.Print(err)
	}

	test_config := credentials.NewTLS(clientTLSConf)
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(test_config))
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
