// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vm

import (
	"context"
	"os"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	testcomputehelper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/compute/helper"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/kindtestenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	computeTestHelper      *testcomputehelper.ComputeTestHelper
	kindTestEnv            *kindtestenv.KindTestEnv
	instance_endpoint      string
	ssh_endpoint           string
	vnet_endpoint          string
	instance_type_endpoint string
	machine_image_endpoint string
	cloudAccount           string
	tempDir                string
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compute Integration K8s Test Suite")
}

var _ = BeforeSuite(func() {
	log.SetDefaultLogger()
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("BeforeSuite")
	logger.Info("BEGIN")
	defer logger.Info("END")

	var err error
	tempDir, err = os.MkdirTemp("", "test_compute_e2e_vm_")
	Expect(err).Should(Succeed())

	kindTestEnvOptions := kindtestenv.KindTestEnvOptions{
		IdcEnv:  "test-e2e-compute-vm",
		TempDir: tempDir,
	}
	kindTestEnv, err = kindtestenv.NewKindTestEnv(ctx, kindTestEnvOptions)
	Expect(err).Should(Succeed())
	Expect(kindTestEnv.Start(ctx)).Should(Succeed())
	computeTestHelper = testcomputehelper.NewComputeTestHelperFromKindTestEnv(ctx, kindTestEnv)
	Eventually(func() error {
		return computeTestHelper.PingComputeApiServer(ctx)
	}, "120s", "1s").Should(Succeed())
	Eventually(func() error {
		return computeTestHelper.PingCloudAccountServer(ctx)
	}, "180s", "1s").Should(Succeed())
	Eventually(func() error {
		return computeTestHelper.PingInstanceScheduler(ctx)
	}, "180s", "1s").Should(Succeed())
	Eventually(func() error {
		return computeTestHelper.PingFleetAdminService(ctx)
	}, "180s", "1s").Should(Succeed())
	// load the compute endpoints to use through out suite
	instance_endpoint, _, ssh_endpoint, vnet_endpoint, instance_type_endpoint, machine_image_endpoint, cloudAccount = computeTestHelper.LoadSuiteLevelTestData(ctx, kindTestEnv)

})

var _ = AfterSuite(func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	log.Info("\n" +
		`~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~` + "\n" +
		`~~~ IF ANY TESTS FAILED, DETAILS WILL BE LOGGED ABOVE THIS LINE. ~~~` + "\n" +
		`~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~`)

	if computeTestHelper != nil {
		computeTestHelper.Cleanup(ctx)
	}
	if kindTestEnv != nil {
		Expect(kindTestEnv.Stop(ctx)).Should(Succeed())
	}
	if tempDir != "" {
		Expect(os.RemoveAll(tempDir)).Should(Succeed())
	}

	log.Info("\n" +
		`~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~` + "\n" +
		`~~~ IF ANY TESTS FAILED, SEARCH ABOVE FOR "~~~".                  ~~~` + "\n" +
		`~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~`)
})
