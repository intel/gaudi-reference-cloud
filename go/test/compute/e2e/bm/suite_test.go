// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	testcommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/common"
	testcomputehelper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/compute/helper"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/compute/helper/netbox_validation"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/kindtestenv"
	. "github.com/onsi/ginkgo/v2"
	ginkgotypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	computeTestHelper      *testcomputehelper.ComputeTestHelper
	kindTestEnv            *kindtestenv.KindTestEnv
	instance_endpoint      string
	instancegroupEndpoint  string
	ssh_endpoint           string
	vnet_endpoint          string
	instance_type_endpoint string
	machine_image_endpoint string
	cloudAccount           string
	tempDir                string
	sshPublicKeyName       string
	vNetName               string
)

// Test timeouts are defined in several places.
// In .bazelrc, `test --test_timeout=-1,-1,-1,7200` increases eternal test timeout to 2 hours.
// The entire Ginkgo test suite has a timeout defined in RunSpecs.
// Each It has a timeout.
func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig := ginkgotypes.NewDefaultSuiteConfig()
	suiteConfig.Timeout = 2 * time.Hour
	RunSpecs(t, "Compute e2e BMaaS Test Suite", suiteConfig)
}

var _ = BeforeSuite(func() {
	log.SetDefaultLogger()
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("BeforeSuite")
	logger.Info("BEGIN")
	defer logger.Info("END")

	var err error
	tempDir, err = os.MkdirTemp("", "test_compute_e2e_bm_")
	Expect(err).Should(Succeed())

	os.Setenv("BAREMETAL_OPERATOR_NAMESPACES", "metal3-1 metal3-2")

	ip_cmd := exec.CommandContext(ctx, "/bin/sh", "-c", `ip -o route get to 10.248.2.1 | sed -n 's/.*src \([0-9.]\+\).*/\1/p'`)
	output, err_ip := ip_cmd.CombinedOutput()
	Expect(err_ip).Should(Succeed())

	trimmedOutput := strings.TrimSpace(string(output))
	// Setting the output as env variable
	os.Setenv("KIND_API_SERVER_ADDRESS", trimmedOutput)
	os.Setenv("HOST_IP", trimmedOutput)

	kindTestEnvOptions := kindtestenv.KindTestEnvOptions{
		ClusterPrefix: "idc",
		IdcEnv:        "test-e2e-compute-bm",
		TempDir:       tempDir,
	}
	kindTestEnv, err = kindtestenv.NewKindTestEnv(ctx, kindTestEnvOptions)
	Expect(err).Should(Succeed())

	ssh_proxy_ip, _ := testcomputehelper.ReadFileAsString(ctx, filepath.Join(kindTestEnv.RunfilesDir, "local/secrets/test-e2e-compute-bm/ssh_proxy_ip"))
	os.Setenv("SSH_PROXY_IP", ssh_proxy_ip)

	bmc_username, _ := testcomputehelper.ReadFileAsString(ctx, filepath.Join(kindTestEnv.RunfilesDir, "local/secrets/test-e2e-compute-bm/DEFAULT_BMC_USERNAME"))
	os.Setenv("DEFAULT_BMC_USERNAME", bmc_username)

	bmc_password, _ := testcomputehelper.ReadFileAsString(ctx, filepath.Join(kindTestEnv.RunfilesDir, "local/secrets/test-e2e-compute-bm/DEFAULT_BMC_PASSWD"))
	os.Setenv("DEFAULT_BMC_PASSWD", bmc_password)

	filePathParent := filepath.Join(kindTestEnv.RunfilesDir, "local/secrets/test-e2e-compute-bm/ssh-proxy-operator")
	testcomputehelper.ScanHostPubKeyfile(ctx, filePathParent)
	filePathParent1 := filepath.Join(kindTestEnv.RunfilesDir, "local/secrets/test-e2e-compute-bm/bm-instance-operator")
	testcomputehelper.ScanHostPubKeyfile(ctx, filePathParent1)

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
	instance_endpoint, instancegroupEndpoint, ssh_endpoint, vnet_endpoint, instance_type_endpoint, machine_image_endpoint, cloudAccount = computeTestHelper.LoadSuiteLevelTestData(ctx, kindTestEnv)

	os.Setenv("VAULT_TOKEN", kindTestEnv.VaultToken())

	vault_addr := kindTestEnv.VaultAddr()
	os.Setenv("VAULT_ADDR", vault_addr)

	netbox_host := fmt.Sprintf("dev.netbox.us-dev-1.api.cloud.intel.com.kind.local:%d", kindTestEnv.IngressHttpsPort())
	os.Setenv("NETBOX_HOST", netbox_host)

	os.Setenv("SECRETS_DIR", kindTestEnv.SecretsDir)

	// Set GUEST_HOST_DEPLOYMENTS to the desired nunber of nodes
	os.Setenv("GUEST_HOST_DEPLOYMENTS", "12")

	cmd := exec.CommandContext(ctx, "deployment/common/netbox/populate_samples.sh")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "NETBOX_API=https://"+netbox_host+"/api")
	cmd.Env = append(cmd.Env, "REGION=us-dev-1")
	err = testcommon.RunCmd(ctx, cmd)
	Expect(err).Should(Succeed())

	// Netbox initial validation and enrollment
	netbox_validation.TestNetboxInitial(ctx)
	netbox_validation.TestNetboxEnrollment(ctx)

	// Add all nodes to general compute node pool.
	pool_cmd := exec.CommandContext(ctx, "kubectl", "label", "--overwrite", "--all-namespaces", "--all", "BareMetalHosts", "pool.cloud.intel.com/general=true")
	err = testcommon.RunCmd(ctx, pool_cmd)
	Expect(err).NotTo(HaveOccurred())

	// Add BGP nodes as a cluster
	nodes_cmd := exec.CommandContext(ctx, "kubectl", "label", "bmh", "-n", "metal3-1", "device-1", "device-3", "device-5", "device-7", "device-9", "device-11",
		"cloud.intel.com/instance-group-id=1", "cloud.intel.com/cluster-size=6", "cloud.intel.com/network-mode=XBX", "--overwrite")
	err = testcommon.RunCmd(ctx, nodes_cmd)
	Expect(err).NotTo(HaveOccurred())

	// Add non-BGP nodes as a cluster
	non_bgp_nodes := exec.CommandContext(ctx, "kubectl", "label", "bmh", "-n", "metal3-2", "device-2", "device-4",
		"cloud.intel.com/instance-group-id=2", "cloud.intel.com/cluster-size=2", "--overwrite")
	err = testcommon.RunCmd(ctx, non_bgp_nodes)
	Expect(err).NotTo(HaveOccurred())

	// SSH key creation
	var publicKey string
	sshPublicKeyName, _, publicKey, _ = computeTestHelper.CreateSshPublicKey(ctx, cloudAccount)
	logger.Info("Created SSH Key Pair", "sshPublicKeyName", sshPublicKeyName, "publicKey", publicKey)

	// VNet creation
	vNetName, _ = computeTestHelper.CreateVNet(ctx, cloudAccount, "us-dev-1b-default", "us-dev-1b")
	logger.Info("Created VNet", "vNetName", vNetName)
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
