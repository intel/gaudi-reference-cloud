//go:build ginkgo_only

package e2e

import (
	"context"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"sync"

	testutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/test_utils"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/clab"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	topologyDir              = "../../../../../networking/containerlab"
	topologyFrontendonlyTiny = "../../../../../networking/containerlab/frontendonly-tiny/frontendonly-tiny.clab.yml"
	configTemplatePath       = "test-e2e-sdn.yaml.gotmpl"
	eAPISecretPath           = "/vault/secrets/eapi"
)

var _ = Describe("Topology frontendonly Tests", Ordered, Label("sdn", "tsdn", "frontendonly-tiny"), func() {
	testHelper := helper.New(helper.Config{TopologyDir: topologyDir, EAPISecretDir: eAPISecretPath})
	var topology *clab.Topology
	var err error
	ctx := context.Background()
	logger := log.FromContext(ctx)

	BeforeAll(func() {
		var wg sync.WaitGroup

		// Increment the WaitGroup counter for 2 goroutines
		wg.Add(2)

		// First goroutine
		go func() {
			// Deploy SDN
			defer wg.Done() // Decrement the counter when the goroutine finishes
			defer GinkgoRecover()
			err = testHelper.DeploySDNWithConfigFile(configTemplatePath, false)
			Expect(err).ShouldNot(HaveOccurred())

			// Check prerequisites
			err = testHelper.CheckSDNIsRunningWithTestTenantConfig()
			Expect(err).ShouldNot(HaveOccurred())
		}()

		// Second goroutine
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			// Remove all existing containerlab topologies
			err = testHelper.ContainerLabManager.DestroyAll()
			Expect(err).ShouldNot(HaveOccurred())
			testutils.DeleteClabTmpFolder(".")

			// before running all the cases, we first deploy the containerlab topology
			err = testHelper.ContainerLabManager.Deploy(topologyFrontendonlyTiny)
			Expect(err).ShouldNot(HaveOccurred())
			topology, err = testHelper.ContainerLabManager.Connect(topologyFrontendonlyTiny)
			Expect(err).ShouldNot(HaveOccurred())
		}()

		// Wait for both goroutines to finish
		wg.Wait()
	})

	AfterAll(func() {
	})

	BeforeEach(func() {
		// remove any existing K8s resources
		err = testHelper.DeleteAllK8sResources()
		Expect(err).ShouldNot(HaveOccurred())

		// reset all the switches for this topology
		logger.Info("Resetting config for all switches...")
		err := testHelper.ContainerLabManager.ResetAllSwitches(topologyFrontendonlyTiny)
		Expect(err).ShouldNot(HaveOccurred())

		// recreate the K8s resource
		logger.Info("Creating k8s resources for topology...")
		err = testHelper.CreateK8sResourcesForTopology(topology)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
	})

	/*
		Purpose of the Test:
			The case is similar to the Frontend topology case 1, but with a very simple topology that can run in host with limited resources.

		Test Steps:
			- Deploy the containerlab topology
			- Create the BMH and Switch CRs for this topology
			- Call the SDNClient.UpdateNetworkNodeConfig() to update front end fabric VLAN to 101 for nodes "server1-1"
			- Get the diff for the current running-config and the startup-config. Compare the diff result with the expected result.

		Test Coverage:
			 - BMH Controller can correctly create the NetworkNode CRs from the BMH
			 - NetworkNode Controller can correctly create the SwitchPort CRs
			 - SwitchPort Controller (and the eapi switch client) can correctly update the switch VLAN.
	*/
	It("test case 1 - update frontend vlan", func() {
		fmt.Printf("frontendonly-tiny case 1 starts... \n")
		swClient1, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-tiny-frontend-leaf1")

		//////////////////////
		// fetch the running-config before making any changes
		//////////////////////
		// get clab-frontendonly-frontend-leaf1 running config
		leaf1InitConfig, err := swClient1.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// make the changes
		//////////////////////
		// simulate the BMaaS to reserve a few nodes for a tenant
		err = testHelper.MockBM.ReserveMultipleBMHWithVXX(ctx, []string{"server1-1"}, int64(101))
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// get the switches running-config after the changes
		//////////////////////
		leaf1ActualConfig, err := swClient1.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// get the diffs, and compare with the expected one
		//////////////////////
		leaf1Diff, err := testutils.Diff(leaf1InitConfig, leaf1ActualConfig, 2)
		Expect(err).ShouldNot(HaveOccurred())

		leaf1DiffExpected := `@@ @@
 !
 interface Ethernet5
-   switchport access vlan 100
+   switchport access vlan 101
    switchport
+   no lldp transmit
 !
 interface Ethernet6
`

		Expect(leaf1Diff).Should(Equal(leaf1DiffExpected), "ActualDiff: %s", leaf1Diff)
		fmt.Printf("frontendonly case 1 completed \n")
	})
})
