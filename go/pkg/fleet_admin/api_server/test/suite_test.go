// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/fleet_admin/api_server/test/..." make test-custom
package test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"testing"

	_ "github.com/amacneil/dbmate/pkg/driver/postgres"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin/api_server/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin/db"
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
	testDb                  *manageddb.TestDb
	managedDb               *manageddb.ManagedDb
	sqlDb                   *sql.DB
	grpcService             *server.GrpcService
	grpcListenPort          = uint16(0)
	tracerProvider          *observability.TracerProvider
	fleetAdminServiceClient pb.FleetAdminServiceClient
)

func TestFleetAdminApiServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fleet Admin API Server Suite")
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

	By("Starting GRPC server")
	grpcServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	grpcListenPort = uint16(grpcServerListener.Addr().(*net.TCPAddr).Port)

	grpcService, err = server.New(ctx, &config.Config{
		Region: region,
	}, managedDb, grpcServerListener)
	Expect(err).Should(Succeed())
	Expect(grpcService.Start(ctx)).Should(Succeed())

	By("Creating GRPC client")
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	fleetAdminServiceClient = pb.NewFleetAdminServiceClient(clientConn)

	By("Pinging service until it comes up")
	Eventually(func(g Gomega) {
		_, err := fleetAdminServiceClient.Ping(ctx, &emptypb.Empty{})
		g.Expect(err).Should(Succeed())
	}, "10s", "1s").Should(Succeed())
})

var _ = AfterSuite(func() {
	ctx := context.Background()
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
	_, err = db.ExecContext(ctx, "delete from pool_cloud_account")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from pool")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from node")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from node_pool")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from node_instance_type")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from node_stats")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from node_instance_type_stats")
	Expect(err).Should(Succeed())
}
