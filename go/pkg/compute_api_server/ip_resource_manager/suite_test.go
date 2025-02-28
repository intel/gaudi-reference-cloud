// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/compute_api_server/ip_resource_manager/..." make test-custom
package ip_resource_manager

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	testDb                  *manageddb.TestDb
	managedDb               *manageddb.ManagedDb
	sqlDb                   *sql.DB
	gprcServerStoppable     *stoppable.Stoppable
	grpcListenPort          = uint16(0)
	ipResourceManagerClient pb.IpResourceManagerServiceClient
)

func TestIpResourceManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IP Resource Manager Suite")
}

var _ = BeforeSuite(func() {
	log.SetDefaultLogger()
	ctx := context.Background()

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

	grpcServer := grpc.NewServer()
	Expect(grpcServer).ShouldNot(BeNil())
	ipResourceManagerService, err := NewIpResourceManagerService(sqlDb)
	Expect(err).Should(Succeed())
	Expect(ipResourceManagerService).ShouldNot(BeNil())

	pb.RegisterIpResourceManagerServiceServer(grpcServer, ipResourceManagerService)

	Expect(err).Should(Succeed())
	gprcServerStoppable = stoppable.New(func(ctx context.Context) error {
		go func() {
			<-ctx.Done()
			grpcServer.Stop()
		}()
		return grpcServer.Serve(grpcServerListener)
	})
	gprcServerStoppable.Start(ctx)

	ipResourceManagerClient = getIpResourceManagerGrpcClient()
})

var _ = AfterSuite(func() {
	ctx := context.Background()
	By("Stopping GRPC server")
	Expect(gprcServerStoppable.Stop(ctx)).Should(Succeed())
	By("Stopping database")
	Expect(testDb.Stop(ctx)).Should(Succeed())
})

func getIpResourceManagerGrpcClient() pb.IpResourceManagerServiceClient {
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	return pb.NewIpResourceManagerServiceClient(clientConn)
}

func clearDatabase(ctx context.Context) {
	db, err := managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from address")
	Expect(err).Should(Succeed())
	_, err = db.ExecContext(ctx, "delete from subnet")
	Expect(err).Should(Succeed())
}
