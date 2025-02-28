//go:build ginkgo_only

package e2e

import (
	"context"
	"fmt"
	"sync"

	testutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/test_utils"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/clab"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	topologyDir        = "../../../../../networking/containerlab"
	topologySmallAccl2 = "../../../../../networking/containerlab/smallaccl2/smallaccl2.clab.yml"
	configTemplatePath = "test-e2e-sdn.yaml.gotmpl"
	eAPISecretPath     = "/vault/secrets/eapi"
)

var _ = Describe("Topology smallaccl2 Tests", Ordered, Label("sdn", "tsdn", "smallaccl2"), func() {
	testHelper := helper.New(helper.Config{TopologyDir: topologyDir, EAPISecretDir: eAPISecretPath})
	var topology *clab.Topology
	var err error

	ctx := context.Background()

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
			err = testHelper.ContainerLabManager.Deploy(topologySmallAccl2)
			Expect(err).ShouldNot(HaveOccurred())
			topology, err = testHelper.ContainerLabManager.Connect(topologySmallAccl2)
			Expect(err).ShouldNot(HaveOccurred())
		}()

		// Wait for both goroutines to finish
		wg.Wait()
	})

	BeforeEach(func() {
		// Delete CRs in case a previous test-run exited / did not clean-up.
		err = testHelper.DeleteAllK8sResources()
		Expect(err).ShouldNot(HaveOccurred())
		err = testHelper.ContainerLabManager.ResetAllSwitches(topologySmallAccl2)
		Expect(err).ShouldNot(HaveOccurred())

		// recreate the K8s resource
		err = testHelper.CreateK8sResourcesForTopology(topology)
		Expect(err).ShouldNot(HaveOccurred())
	})

	/*
			Purpose of the Test: (Describe what the test is intended to verify.)
				This tests the accelerator vlan update using networknode CR in a 2-ply 2-interfaces per ply per-server

			Test Steps: (Outline the steps the test will execute, making it clear what actions are being performed and why.)
				- Deploy the containerlab topology
				- Create the BMH and Switch CRs for this topology
				- Call the SDNClient.UpdateNetworkNodeConfig() to update front end fabric VLAN to 101 & accelerator VLAN to 109 for nodes "server1-1" and "server2-1"
				- Get the diff for the current running-config and the startup-config. Compare the diff result with the expected result.

			Test Coverage: (outlines the specific features being tested and any corner cases or bugs that the test case addresses.)
				 - BMH Controller can correctly create the NetworkNode CRs from the BMH for "small cluster" servers with multiple accelerator interfaces
				 - SwitchPort Controller (and the eapi switch client) can correctly update the switch VLAN for accelerator fabrics.
				 - Multiple fabrics can be updated at the same time (with a single call to sdnclient).
				 - Multiple ports per switch can be updated at the same time.
		         - Multiple plys (switches) per server can be updated at the same time.
	*/
	It("test case 1 - update accelerator vlan & FE", Label("sdn", "smallaccl2", "case1"), func() {
		fmt.Printf("smallaccl2 case 1 starts... \n")
		accPly1SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-smallaccl2-acc-leaf")
		accPly2SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-smallaccl2-accply2-leaf")
		frontendSwClient1, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-smallaccl2-frontend-leaf1")
		//frontendSwClient2, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-smallaccl2-frontend-leaf2")

		//////////////////////
		// fetch the running-config before making any changes
		//////////////////////
		// get clab-frontendonly-frontend-accPly1 running config
		accPly1InitConfig, err := accPly1SwClient.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		// get clab-frontendonly-frontend-accPly2 running config
		accPly2InitConfig, err := accPly2SwClient.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		// get clab-frontendonly-frontend-accPly2 running config
		frontendLeaf1InitConfig, err := frontendSwClient1.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// make the changes
		//////////////////////
		// simulate the BMaaS to reserve 2 (out of 3) nodes for a tenant
		err = testHelper.MockBM.ReserveMultipleBMHWithVVX(ctx, []string{"server1-1", "server2-1"}, int64(101), int64(109))
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// get the switches running-config after the changes
		//////////////////////
		accPly1ActualConfig, err := accPly1SwClient.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		accPly2ActualConfig, err := accPly2SwClient.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		frontendLeaf1ActualConfig, err := frontendSwClient1.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// compare the latest running-config with the expected one
		//////////////////////
		accPly1Diff, err := testutils.Diff(accPly1InitConfig, accPly1ActualConfig, 2)
		Expect(err).ShouldNot(HaveOccurred())

		accPly1DiffExpected := `@@ @@
 interface Ethernet1
    description server1-1
-   switchport access vlan 100
+   switchport access vlan 109
    switchport
+   no lldp transmit
 !
 interface Ethernet2
    description server1-1
-   switchport access vlan 100
+   switchport access vlan 109
    switchport
+   no lldp transmit
 !
 interface Ethernet3
@@ @@
 interface Ethernet5
    description server2-1
-   switchport access vlan 100
+   switchport access vlan 109
    switchport
+   no lldp transmit
 !
 interface Ethernet6
    description server2-1
-   switchport access vlan 100
+   switchport access vlan 109
    switchport
+   no lldp transmit
 !
 interface Management0
`
		Expect(accPly1Diff).Should(Equal(accPly1DiffExpected), fmt.Sprintf("Actual accPly1Diff: %s", accPly1Diff))

		accPly2Diff, err := testutils.Diff(accPly2InitConfig, accPly2ActualConfig, 2)
		Expect(err).ShouldNot(HaveOccurred())
		accPly2DiffExpected := `@@ @@
 interface Ethernet1
    description server1-1
-   switchport access vlan 100
+   switchport access vlan 109
    switchport
+   no lldp transmit
 !
 interface Ethernet2
    description server1-1
-   switchport access vlan 100
+   switchport access vlan 109
    switchport
+   no lldp transmit
 !
 interface Ethernet3
@@ @@
 interface Ethernet5
    description server2-1
-   switchport access vlan 100
+   switchport access vlan 109
    switchport
+   no lldp transmit
 !
 interface Ethernet6
    description server2-1
-   switchport access vlan 100
+   switchport access vlan 109
    switchport
+   no lldp transmit
 !
 interface Management0
`
		Expect(accPly2Diff).Should(Equal(accPly2DiffExpected), fmt.Sprintf("Actual accPly2Diff: %s", accPly2Diff))

		// Check Frontend fabric
		frontendLeaf1Diff, err := testutils.Diff(frontendLeaf1InitConfig, frontendLeaf1ActualConfig, 2)
		Expect(err).ShouldNot(HaveOccurred())
		frontendLeaf1DiffExpected := `@@ @@
 !
 interface Ethernet5
-   switchport access vlan 100
+   switchport access vlan 101
    switchport
+   no lldp transmit
 !
 interface Ethernet6
`
		// Only Ethernet5 should have changed on this frontend-leaf1. The other server that was provisioned is connected to the other frontend-leaf2.

		Expect(frontendLeaf1Diff).Should(Equal(frontendLeaf1DiffExpected), fmt.Sprintf("Actual frontendLeaf1Diff: %s", frontendLeaf1Diff))

		// BUG: There's a difference in this test if the nodeGroup are provisioned "VVX" instead of no nodegroupMapping, then server1-2 is not provisioned, but the frontend port is moved to VLAN 4008.
		// If the nodeGroup has no pool specified (the nodeGroupMapping does not exist) then the networkNode spec.vlanId gets instantiated to -1, and the frontend port on the switch remains unchanged.
		// I think the bug occurs when the move to 101 happens BEFORE the nodeGroup is "ready", and it is then moved into the pool and goes to 4008 (default for the new pool).
		// Not a bug per-se, but it is an unexpected difference between "VVX" and "no nodegroup".

		fmt.Printf("smallaccl2 case 1 completed \n")
	})

})
