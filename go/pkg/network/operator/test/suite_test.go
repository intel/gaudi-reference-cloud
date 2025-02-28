// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This test suite utilizes the following components:
//
//   - Kubernetes API Server
//   - etcd (for Kubernetes)
//   - Mock of Network API Server (GRPC)
//   - Network Operator
//
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/network/operator/test/..." make test-custom
package test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/internal/controller/subnet"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/db"
	vpcv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/internal/controller/iprm"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/internal/controller/vpc"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/sdn/mock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	testDb                     *manageddb.TestDb
	managedDb                  *manageddb.ManagedDb
	sqlDb                      *sql.DB
	grpcService                *server.GrpcService
	grpcListenPort             = uint16(0)
	tracerProvider             *observability.TracerProvider
	testEnv                    *envtest.Environment
	k8sRestConfig              *rest.Config
	k8sClient                  client.Client
	scheme                     *runtime.Scheme
	vpcServiceClient           pb.VPCServiceClient
	vpcPrivateServiceClient    pb.VPCPrivateServiceClient
	subnetServiceClient        pb.SubnetServiceClient
	subnetPrivateServiceClient pb.SubnetPrivateServiceClient
	iprmPrivateServiceClient   pb.IPRMPrivateServiceClient
	sdnServiceClient           *v1.MockOvnnetClient
	managerStoppable           *stoppable.Stoppable
	poll                       time.Duration = 10 * time.Millisecond
)

func TestNetworkOperator_VPC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scheduler Suite")
}

var _ = BeforeSuite(func() {
	const (
		region = "us-dev-1"
	)

	var availabilityZones = []string{"us-dev-1a"}

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

	By("Creating mock CloudAccount Service")
	cloudAccountService := NewMockCloudAccountServiceClient()

	By("Starting GRPC server")
	grpcServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	grpcListenPort = uint16(grpcServerListener.Addr().(*net.TCPAddr).Port)

	grpcService, err = server.New(ctx, &config.Config{
		Region: region,
	}, managedDb, grpcServerListener, cloudAccountService, availabilityZones)
	Expect(err).Should(Succeed())
	Expect(grpcService.Start(ctx)).Should(Succeed())

	By("Creating GRPC client")
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	vpcServiceClient = pb.NewVPCServiceClient(clientConn)
	vpcPrivateServiceClient = pb.NewVPCPrivateServiceClient(clientConn)
	subnetServiceClient = pb.NewSubnetServiceClient(clientConn)
	subnetPrivateServiceClient = pb.NewSubnetPrivateServiceClient(clientConn)
	iprmPrivateServiceClient = pb.NewIPRMPrivateServiceClient(clientConn)

	By("Pinging VPC service until it comes up")
	Eventually(func(g Gomega) {
		_, err := vpcServiceClient.Ping(ctx, &emptypb.Empty{})
		g.Expect(err).Should(Succeed())
	}, "10s", "1s").Should(Succeed())
	By("VPC Service is ready")

	By("Pinging Subnet service until it comes up")
	Eventually(func(g Gomega) {
		_, err := subnetServiceClient.Ping(ctx, &emptypb.Empty{})
		g.Expect(err).Should(Succeed())
	}, "10s", "1s").Should(Succeed())
	By("Subnet Service is ready")

	By("Pinging IPRM service until it comes up")
	Eventually(func(g Gomega) {
		_, err := iprmPrivateServiceClient.PingPrivate(ctx, &emptypb.Empty{})
		g.Expect(err).Should(Succeed())
	}, "10s", "1s").Should(Succeed())
	By("IPRM Private Service is ready")

	By("Starting Kubernetes API Server")
	testEnv = &envtest.Environment{
		// When adding CRDS, be sure to add them to the data list in BUILD.bazel.
		CRDDirectoryPaths: []string{
			"../config/crd/bases",
		},
		ErrorIfCRDPathMissing:    true,
		AttachControlPlaneOutput: true,
	}

	k8sRestConfig, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sRestConfig).NotTo(BeNil())

	By("Creating manager")
	k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
		Scheme:  scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	By("Configuring scheme")
	scheme = runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).NotTo(HaveOccurred())
	Expect(vpcv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())

	sdnServiceClient = NewMockOvnnetClient()

	_, err = vpc.NewReconciler(ctx, k8sManager, vpcPrivateServiceClient, sdnServiceClient, region)
	Expect(err).NotTo(HaveOccurred())

	_, err = subnet.NewReconciler(ctx, k8sManager, subnetPrivateServiceClient, sdnServiceClient)
	Expect(err).NotTo(HaveOccurred())

	_, err = iprm.NewReconciler(ctx, k8sManager, iprmPrivateServiceClient, sdnServiceClient)
	Expect(err).NotTo(HaveOccurred())

	By("Starting manager")
	managerStoppable = stoppable.New(k8sManager.Start)
	managerStoppable.Start(ctx)
})

var _ = AfterSuite(func() {
	ctx := context.Background()
	By("Stopping manager")
	Expect(managerStoppable.Stop(ctx)).Should(Succeed())
	By("Manager stopped")
	By("Stopping Kubernetes API Server")
	Eventually(func() error {
		return testEnv.Stop()
	}, timeout, poll).ShouldNot(HaveOccurred())
	By("Stopping GRPC server")
	Expect(grpcService.Stop(ctx)).Should(Succeed())
	By("Stopping database")
	Expect(testDb.Stop(ctx)).Should(Succeed())
	By("Stopping tracing")
	Expect(tracerProvider.Shutdown(ctx)).Should(Succeed())
})

func clearDatabase(ctx context.Context) {
	db, err := managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from port")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from subnet")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from vpc")
	Expect(err).Should(Succeed())
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

func NewMockOvnnetClient() *v1.MockOvnnetClient {
	mockController := gomock.NewController(GinkgoT())
	mockOvnnetClient := v1.NewMockOvnnetClient(mockController)

	mockOvnnetClient.EXPECT().CreateVPC(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockOvnnetClient.EXPECT().GetVPC(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	mockOvnnetClient.EXPECT().CreateSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockOvnnetClient.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockOvnnetClient.EXPECT().GetSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	return mockOvnnetClient
}

func NewCloudAcctId() (string, error) {
	intId, err := rand.Int(rand.Reader, big.NewInt(maxId))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%012d", intId), nil
}

func createVPCAndSubnet(ctx context.Context, cloudAccountId string, vpcName string) (*pb.VPC, *pb.VPCSubnet, error) {
	// Create VPC
	vpc, err := vpcServiceClient.Create(ctx, &pb.VPCCreateRequest{
		Metadata: &pb.VPCMetadataCreate{
			CloudAccountId: cloudAccountId,
			Name:           vpcName,
		},
		Spec: &pb.VPCSpec{
			CidrBlock: "10.0.0.0/16",
		},
	})
	Expect(err).Should(Succeed())
	By("Waiting for vpc to be created in SDN1")
	Eventually(func(g Gomega) {
		got, err := vpcPrivateServiceClient.GetPrivate(ctx, &pb.VPCGetPrivateRequest{
			Metadata: &pb.VPCMetadataReference{
				CloudAccountId: cloudAccountId,
				NameOrId:       &pb.VPCMetadataReference_ResourceId{ResourceId: vpc.Metadata.ResourceId},
			},
		})
		g.Expect(err).Should(Succeed())
		g.Expect(got.Spec.CidrBlock).Should(Equal("10.0.0.0/16"))
		g.Expect(got.Status.Phase).Should(Equal(pb.VPCPhase_VPCPhase_Ready))
		g.Expect(got.Status.Message).Should(Equal("VPC ready"))
	}, timeout).Should(Succeed())

	// Create Subnet
	subnetName := uuid.NewString()
	subnet, err := subnetServiceClient.Create(ctx, &pb.SubnetCreateRequest{
		Metadata: &pb.SubnetMetadataCreate{
			CloudAccountId: cloudAccountId,
			Name:           subnetName,
		},
		Spec: &pb.SubnetSpec{
			CidrBlock:        "10.0.0.0/16",
			AvailabilityZone: "us-dev-1a",
			VpcId:            vpc.Metadata.ResourceId,
		},
	})
	Expect(err).Should(Succeed())

	By("Waiting for subnet to be created in SDN2")
	Eventually(func(g Gomega) {
		got, err := subnetPrivateServiceClient.GetPrivate(ctx, &pb.SubnetGetPrivateRequest{
			Metadata: &pb.SubnetMetadataReference{
				CloudAccountId: cloudAccountId,
				NameOrId: &pb.SubnetMetadataReference_ResourceId{
					ResourceId: subnet.Metadata.ResourceId,
				},
			},
		})

		g.Expect(err).Should(Succeed())
		g.Expect(got.Spec.CidrBlock).Should(Equal("10.0.0.0/16"))
		g.Expect(got.Status.Phase).Should(Equal(pb.SubnetPhase_SubnetPhase_Ready))
		g.Expect(got.Status.Message).Should(Equal("Subnet ready"))
	}, timeout).Should(Succeed())

	return vpc, subnet, err
}
