// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This test suite utilizes the following components:
//
//   - Kubernetes API Server
//   - etcd (for Kubernetes)
//   - Compute Database (Postgres)
//   - Compute API Server (GRPC)
//   - Compute API Gateway (GRPC-REST gateway)
//   - VM Instance Scheduler
//   - Instance Replicator (Compute API Server to K8s Instance)
//   - VM Instance Operator (K8s Instance to Kubevirt VirtualMachine)
//   - Fleet Admin Service
//
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/compute_integration_test/..." make test-custom
package basic

import (
	"context"
	"database/sql"
	"testing"

	computedb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	fleetadmindb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	testEnv                            *envtest.Environment
	k8sRestConfig                      *rest.Config
	scheme                             *runtime.Scheme
	k8sClient                          client.Client
	computeTestDb                      *manageddb.TestDb
	computeManagedDb                   *manageddb.ManagedDb
	fleetAdminTestDb                   *manageddb.TestDb
	fleetAdminManagedDb                *manageddb.ManagedDb
	computeApiServerGrpcListenPort     = uint16(0)
	computeApiServerRestListenPort     = uint16(0)
	instanceSchedulingServerListenPort = uint16(0)
	openApiClient                      *openapi.APIClient
	sshKeyAuthorizedFileTmpDir         string
	billingDeactivateInstancesService  pb.BillingDeactivateInstancesServiceClient
	crdDirectoryPaths                  []string
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compute Integration Basic Test Suite")
}

var _ = BeforeSuite(func() {
	grpcutil.UseTLS = false
	log.SetDefaultLogger()
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	crdDirectoryPaths = []string{
		"../../k8s/config/crd/bases",
		"../../instance_operator/testdata/crd",
		"../../instance_scheduler/vm/testdata/crd",
	}
	By("Starting Kubernetes API Server")
	testEnv = &envtest.Environment{
		// When adding CRDS, be sure to add them to the data list in BUILD.bazel.
		CRDDirectoryPaths:        crdDirectoryPaths,
		ErrorIfCRDPathMissing:    true,
		AttachControlPlaneOutput: true,
	}
	var err error
	k8sRestConfig, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sRestConfig).NotTo(BeNil())

	By("Configuring scheme")
	scheme = runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).NotTo(HaveOccurred())
	Expect(privatecloudv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
	Expect(kubevirtv1.AddToScheme(scheme)).NotTo(HaveOccurred())

	By("Creating Kubernetes client")
	k8sClient, err = client.New(k8sRestConfig, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("Starting Compute Database (Postgres)")
	computeTestDb = &manageddb.TestDb{}
	computeManagedDb, err = computeTestDb.Start(ctx)
	Expect(err).Should(Succeed())
	Expect(computeManagedDb).ShouldNot(BeNil())

	By("Starting Fleet Admin Database (Postgres)")
	fleetAdminTestDb = &manageddb.TestDb{}
	fleetAdminManagedDb, err = fleetAdminTestDb.Start(ctx)
	Expect(err).Should(Succeed())
	Expect(fleetAdminManagedDb).ShouldNot(BeNil())

	By("Migrating Compute Database")
	Expect(computeManagedDb.Migrate(ctx, computedb.MigrationsFs, computedb.MigrationsDir)).Should(Succeed())

	By("Migrating Fleet Admin Database")
	Expect(fleetAdminManagedDb.Migrate(ctx, fleetadmindb.MigrationsFs, fleetadmindb.MigrationsDir)).Should(Succeed())
})

var _ = AfterSuite(func() {
	ctx := context.Background()
	By("Stopping Kubernetes API Server")
	Eventually(func() error {
		return testEnv.Stop()
	}).ShouldNot(HaveOccurred())
	By("Stopping Compute Database (Postgres)")
	Expect(computeTestDb.Stop(ctx)).Should(Succeed())
	By("Stopping Fleet Admin Database (Postgres)")
	Expect(fleetAdminTestDb.Stop(ctx)).Should(Succeed())
})

func clearManagedDb(ctx context.Context, managedDb *manageddb.ManagedDb, tables []string) {
	var db *sql.DB
	Eventually(func() error {
		var err error
		db, err = managedDb.Open(ctx)
		return err
	}).ShouldNot(HaveOccurred())
	for _, t := range tables {
		_, err := db.ExecContext(ctx, "delete from "+t)
		Expect(err).Should(Succeed())
	}
}

func clearDatabase(ctx context.Context) {
	By("Clearing Compute DB")
	clearManagedDb(ctx, computeManagedDb, []string{"instance", "ssh_public_key", "vnet", "address", "subnet"})
	By("Clearing Fleet Admin DB")
	clearManagedDb(ctx, fleetAdminManagedDb, []string{"pool_cloud_account"})
}
