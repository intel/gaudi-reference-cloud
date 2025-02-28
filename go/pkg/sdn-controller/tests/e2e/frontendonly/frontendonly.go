//go:build ginkgo_only

package e2e

import (
	"context"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"time"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	sdnclient "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/client"
	testutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/test_utils"
	"k8s.io/apimachinery/pkg/types"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/clab"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	topologyDir                = "../../../../../networking/containerlab"
	topologyFrontendonly       = "../../../../../networking/containerlab/frontendonly/frontendonly.clab.yml"
	configTemplatePath         = "test-e2e-sdn.yaml.gotmpl"
	readOnlyConfigTemplatePath = "read-only.yaml.gotmpl"
	eAPISecretPath             = "/vault/secrets/eapi"
)

var _ = Describe("Topology frontendonly Tests", Ordered, Label("sdn", "tsdn", "frontendonly"), func() {
	testHelper := helper.New(helper.Config{TopologyDir: topologyDir, EAPISecretDir: eAPISecretPath})
	var topology *clab.Topology
	var err error
	ctx := context.Background()
	logger := log.FromContext(ctx)

	Describe("With default config", Ordered, Label(), func() {
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
				err = testHelper.ContainerLabManager.Deploy(topologyFrontendonly)
				Expect(err).ShouldNot(HaveOccurred())
				topology, err = testHelper.ContainerLabManager.Connect(topologyFrontendonly)
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
			err := testHelper.ContainerLabManager.ResetAllSwitches(topologyFrontendonly)
			Expect(err).ShouldNot(HaveOccurred())

			// recreate the K8s resource
			logger.Info("Creating k8s resources for topology...")
			err = testHelper.CreateK8sResourcesForTopology(topology)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
		})

		/*
			Purpose of the Test: (Describe what the test is intended to verify.)
				This is the basic test case that verify if SDN can successfully executed a VLAN update request from the SDNClient.

			Test Steps: (Outline the steps the test will execute, making it clear what actions are being performed and why.)
				- Deploy the containerlab topology
				- Create the BMH and Switch CRs for this topology
				- Call the SDNClient.UpdateNetworkNodeConfig() to update front end fabric VLAN to 101 for nodes "server1-1", "server1-2" and "server2-1"
				- Get the diff for the current running-config and the startup-config. Compare the diff result with the expected result.

			Test Coverage: (outlines the specific features being tested and any corner cases or bugs that the test case addresses.)
				 - BMH Controller can correctly create the NetworkNode CRs from the BMH
				 - NetworkNode Controller can correctly create the SwitchPort CRs
				 - SwitchPort Controller (and the eapi switch client) can correctly update the switch VLAN.
		*/
		It("test case 1 - update frontend vlan", Label("sdn", "frontendonly", "case1"), func() {
			logger.Info("frontendonly case 1 starts...")
			swClient1, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-frontend-leaf1")
			swClient2, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-frontend-leaf2")

			server11 := topology.TopologySpec.Nodes["server1-1"]
			server12 := topology.TopologySpec.Nodes["server1-2"]
			// server21 := topology.TopologySpec.Nodes["server2-1"]

			//////////////////////
			// fetch the running-config before making any changes
			//////////////////////
			// get clab-frontendonly-frontend-leaf1 running config
			leaf1InitConfig, err := swClient1.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			// get clab-frontendonly-frontend-leaf2 running config
			leaf2InitConfig, err := swClient2.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// make the changes
			//////////////////////
			// simulate the BMaaS to reserve a few nodes for a tenant
			// move "server1-1" and "server1-2" to vlan 101, "server2-1" stays in 100
			err = testHelper.MockBM.ReserveMultipleBMHWithVXX(ctx, []string{"server1-1", "server1-2"}, int64(101))
			Expect(err).ShouldNot(HaveOccurred())
			// TODO: find a way to not hard code the IPs?
			// "server1-1" and "server1-2" should be able to ping each other.
			Expect(testHelper.PingInsideContainer(server11.ContainerID, "100.80.1.12", 5)).Should(BeTrue()) // server11 -> server12
			Expect(testHelper.PingInsideContainer(server12.ContainerID, "100.80.1.11", 5)).Should(BeTrue()) // server12 -> server11
			// "server1-1" and "server1-2" should NOT be able to ping "server2-1".
			Expect(testHelper.PingInsideContainer(server11.ContainerID, "100.80.1.21", 1)).Should(BeFalse()) // server11 -> server21
			Expect(testHelper.PingInsideContainer(server12.ContainerID, "100.80.1.21", 1)).Should(BeFalse()) // server12 -> server21

			// now move "server2-1" to vlan 101
			err = testHelper.MockBM.ReserveMultipleBMHWithVXX(ctx, []string{"server2-1"}, int64(101))
			Expect(err).ShouldNot(HaveOccurred())
			// "server1-1" and "server1-2" should be able to ping "server2-1".
			Expect(testHelper.PingInsideContainer(server11.ContainerID, "100.80.1.21", 5)).Should(BeTrue())
			Expect(testHelper.PingInsideContainer(server12.ContainerID, "100.80.1.21", 5)).Should(BeTrue())

			//////////////////////
			// get the switches running-config after the changes
			//////////////////////
			leaf1ActualConfig, err := swClient1.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			leaf2ActualConfig, err := swClient2.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// compare the latest running-config with the expected one
			//////////////////////
			leaf1Diff, err := testutils.Diff(leaf1InitConfig, leaf1ActualConfig, 2)
			Expect(err).ShouldNot(HaveOccurred())
			leaf2Diff, err := testutils.Diff(leaf2InitConfig, leaf2ActualConfig, 2)
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
-   switchport access vlan 100
+   switchport access vlan 101
    switchport
+   no lldp transmit
 !
 interface Ethernet7
`

			leaf2DiffExpected := `@@ @@
 interface Ethernet5
    description server2-1 TenantBM
-   switchport access vlan 100
+   switchport access vlan 101
    switchport
+   no lldp transmit
 !
 interface Ethernet6
`

			Expect(leaf1Diff).Should(Equal(leaf1DiffExpected), "Actual leaf1Diff: %s", leaf1Diff)
			Expect(leaf2Diff).Should(Equal(leaf2DiffExpected), "Actual leaf2Diff: %s", leaf2Diff)

			logger.Info("frontendonly case 1 completed \n")
		})

		/*
			Purpose of the Test:
				This case tries to make 2 calls to SDNClient to update VLAN for a node, the 2nd call is made before the 1st one is done. We expected the 2nd request will be applied to the switch.

			Test Steps:
				- Deploy the containerlab topology
				- Create the BMH and Switch CRs for this topology
				- Call the SDNClient.UpdateNetworkNodeConfig() to update front end fabric VLAN to 102 for nodes "server1-1"
				- wait for a short period, then call the SDNClient.UpdateNetworkNodeConfig() again to update front end fabric VLAN to 4008 for nodes "server1-1"
				- Get the diff. Check if the VLAN is set to 4008.

			Test Coverage:
				 - This tests the SwitchPort Controller behavior that it will reconcile until reach the desired state.
		*/
		It("test case 2 - cancel Vlan change before it completes", Label("sdn", "frontendonly", "case2"), func() {
			logger.Info("frontendonly case 2 starts...")

			// get clab-frontendonly-frontend-leaf1 running config
			swClient1, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-frontend-leaf1")
			Expect(err).ShouldNot(HaveOccurred())
			// get clab-frontendonly-frontend-leaf2 running config
			swClient2, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-frontend-leaf2")
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// fetch the running-config before making any changes
			//////////////////////
			// get clab-frontendonly-frontend-leaf1 running config
			leaf1InitConfig, err := swClient1.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			// get clab-frontendonly-frontend-leaf2 running config
			leaf2InitConfig, err := swClient2.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// make the changes
			//////////////////////
			// simulate the BMaaS to reserve a node for a tenant
			err = testHelper.MockBM.SDNClient.UpdateNetworkNodeConfig(ctx, sdnclient.NetworkNodeConfUpdateRequest{
				NetworkNodeName:    "server1-1",
				FrontEndFabricVlan: 102,
			})
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(2 * time.Second)

			// Before it completes, cancel it
			err = testHelper.MockBM.SDNClient.UpdateNetworkNodeConfig(ctx, sdnclient.NetworkNodeConfUpdateRequest{
				NetworkNodeName:    "server1-1",
				FrontEndFabricVlan: 4008,
			})
			Expect(err).ShouldNot(HaveOccurred())

			var result sdnclient.NetworkNodeConfStatusCheckResponse
			for result.Status != sdnclient.UpdateCompleted {
				result, err = testHelper.MockBM.SDNClient.CheckNetworkNodeStatus(ctx, sdnclient.NetworkNodeConfStatusCheckRequest{
					NetworkNodeName:           "server1-1",
					DesiredFrontEndFabricVlan: 4008,
				})
				Expect(err).ShouldNot(HaveOccurred())

				// sleep for 5 seconds, simulating the requeue behavior
				time.Sleep(5 * time.Second)
			}

			//////////////////////
			// get the switches running-config after the changes
			//////////////////////
			leaf1ActualConfig, err := swClient1.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			leaf2ActualConfig, err := swClient2.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// compare the latest running-config with the expected one
			//////////////////////
			leaf1Diff, err := testutils.Diff(leaf1InitConfig, leaf1ActualConfig, 2)
			Expect(err).ShouldNot(HaveOccurred())
			leaf2Diff, err := testutils.Diff(leaf2InitConfig, leaf2ActualConfig, 2)
			Expect(err).ShouldNot(HaveOccurred())

			leaf1DiffExpected := `@@ @@
 !
 interface Ethernet5
-   switchport access vlan 100
+   switchport access vlan 4008
    switchport
 !
`

			leaf2DiffExpected := ``

			//////////////////////
			// compare the latest running-config with the expected one
			//////////////////////
			Expect(leaf1Diff).Should(Equal(leaf1DiffExpected), "Actual leaf1Diff: %s", leaf1Diff)
			Expect(leaf2Diff).Should(Equal(leaf2DiffExpected), "Actual leaf2Diff: %s", leaf2Diff)

			logger.Info("frontendonly case 2 completed \n")
		})

		It("test case 3 - test for unknown vlan", Label("sdn", "frontendonly", "case3"), func() {
			logger.Info("frontendonly case 3 starts...")

			feLeaf1SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-frontend-leaf1")
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf1InitConfig, err := feLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			err = testHelper.DeleteAllEvents()
			Expect(err).ShouldNot(HaveOccurred())

			err = testutils.Timeout(30*time.Second, func(ctx context.Context) error {
				return testHelper.MockBM.ReserveSingleBMHWithVXX(ctx, "server1-1", int64(200))
			})
			Expect(err).Should(HaveOccurred())

			feLeaf1ActualConfig, err := feLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf1Diff, err := testutils.Diff(feLeaf1InitConfig, feLeaf1ActualConfig, 2)
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf1DiffExpected := ``
			Expect(feLeaf1Diff).Should(Equal(feLeaf1DiffExpected), "Actual feLeaf1Diff: %s", feLeaf1Diff)

			// Get the events generated by switchPort
			found, err := testHelper.WaitForSwitchportEvent(ctx, "ethernet5.clab-frontendonly-frontend-leaf1", "idcs-system", "Requested Vlan entry not found on the switch", 60)
			Expect(err).ShouldNot(HaveOccurred())

			// Expect event found to be true
			Expect(found).To(BeTrue())

			logger.Info("frontendonly case 3 completed \n")
		})

		It("test case 4 - ipOverride with valid ip address", Label("sdn", "frontendonly", "case4"), func() {
			logger.Info("frontendonly case 4 starts...")

			feLeaf1CR := &idcnetworkv1alpha1.Switch{}
			key := types.NamespacedName{Name: "clab-frontendonly-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, feLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf2CR := &idcnetworkv1alpha1.Switch{}
			key = types.NamespacedName{Name: "clab-frontendonly-frontend-leaf2", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, feLeaf2CR)
			Expect(err).ShouldNot(HaveOccurred())

			// Remove the server 1-2 BMH to prevent access-mode changes to ethernet6 getting pushed to leaf2 (after the IP override)
			// if periodic reconciliation of et6 switchport happens to run during this test and causing an unexpected diff (this test operates on Et5).
			server1_2BMHCR := &baremetalv1alpha1.BareMetalHost{}
			key = types.NamespacedName{Name: "server1-2", Namespace: "metal3-1"}
			err = testHelper.K8sClient.Get(ctx, key, server1_2BMHCR)
			Expect(err).ShouldNot(HaveOccurred())
			bmhsToDelete := map[string]*baremetalv1alpha1.BareMetalHost{
				"server1-2": server1_2BMHCR,
			}
			testHelper.DeleteK8sResources(ctx, bmhsToDelete, map[string]*idcnetworkv1alpha1.Switch{}, map[string]*idcnetworkv1alpha1.NodeGroupToPoolMapping{})

			// Get ip address of leaf2 switch to ipOverride
			ipAddressLeaf2 := feLeaf2CR.Spec.Ip

			feLeaf1SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-frontend-leaf1")
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf1Config, err := feLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf2SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-frontend-leaf2")
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf2Config, err := feLeaf2SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			// Override with leaf2 switch IP
			feLeaf1CR.Spec.IpOverride = ipAddressLeaf2

			err = testHelper.K8sClient.Update(ctx, feLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())

			// After overriding to clab-frontendonly-frontend-leaf2 now leaf2 will be updated
			err = testutils.Timeout(30*time.Second, func(ctx context.Context) error {
				return testHelper.MockBM.ReserveSingleBMHWithVXX(ctx, "server1-1", int64(102))
			})
			Expect(err).ShouldNot(HaveOccurred())
			time.Sleep(2 * time.Second)

			// The ContainerLabManager used in the tests does not use the overridden IP, so it really does connect to leaf1.
			feLeaf1Config2, err := feLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf2Config2, err := feLeaf2SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			// Config diff for leaf1
			feLeaf1Diff, err := testutils.Diff(feLeaf1Config, feLeaf1Config2, 2)
			Expect(err).ShouldNot(HaveOccurred())

			// Vlan update should not happen on leaf1
			feLeaf1DiffExpected := ``

			// Config Diff for leaf2
			feLeaf2Diff, err := testutils.Diff(feLeaf2Config, feLeaf2Config2, 2)
			Expect(err).ShouldNot(HaveOccurred())

			// Vlan update should happen on leaf2
			feLeaf2DiffExpected := `@@ @@
 interface Ethernet5
    description server2-1 TenantBM
-   switchport access vlan 100
+   switchport access vlan 102
    switchport
+   no lldp transmit
 !
 interface Ethernet6
`
			// No config diff for leaf1 expected after overriding leaf1
			Expect(feLeaf1Diff).Should(Equal(feLeaf1DiffExpected), "Actual feLeaf1Diff: %s  Expected: %s", feLeaf1Diff, feLeaf1DiffExpected)

			// Expect config diff for leaf2 after overriding leaf1
			Expect(feLeaf2Diff).Should(Equal(feLeaf2DiffExpected), "Actual feLeaf2Diff: %s  Expected: %s", feLeaf2Diff, feLeaf2DiffExpected)

			logger.Info("frontendonly case 4 completed \n")
		})

		It("test case 5 - ipOverride with invalid hostname", Label("sdn", "frontendonly", "case5"), func() {
			logger.Info("frontendonly case 5 starts...")

			err = testHelper.DeleteAllEvents()
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf1CR := &idcnetworkv1alpha1.Switch{}
			key := types.NamespacedName{Name: "clab-frontendonly-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, feLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())

			// IpOverride with invalid hostname
			feLeaf1CR.Spec.IpOverride = "invalid-clab-frontendonly-frontend-leaf1"
			err = testHelper.K8sClient.Update(ctx, feLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())
			time.Sleep(2 * time.Second)

			found, err := testHelper.WaitForSwitchEvent(ctx, "clab-frontendonly-frontend-leaf1", "idcs-system", "invalid-clab-frontendonly-frontend-leaf1: is not a valid IP address or hostname, value is neither a valid IP address nor a valid hostname: invalid-clab-frontendonly-frontend-leaf1", 60)
			Expect(err).ShouldNot(HaveOccurred())

			// Expect event to be found
			Expect(found).To(BeTrue())

			logger.Info("frontendonly case 5 completed \n")
		})

		It("test case 6 - ipOverride with invalid ip address", Label("sdn", "frontendonly", "case6"), func() {
			logger.Info("frontendonly case 6 started")

			err = testHelper.DeleteAllEvents()
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf1CR := &idcnetworkv1alpha1.Switch{}
			key := types.NamespacedName{Name: "clab-frontendonly-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, feLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())

			// IpOverride with invalid ip
			feLeaf1CR.Spec.IpOverride = "256.20.20.20"
			err = testHelper.K8sClient.Update(ctx, feLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())

			found, err := testHelper.WaitForSwitchEvent(ctx, "clab-frontendonly-frontend-leaf1", "idcs-system", "256.20.20.20: is not a valid IP address or hostname, value is neither a valid IP address nor a valid hostname: 256.20.20.20", 60)
			Expect(err).ShouldNot(HaveOccurred())

			// Expect event to be found
			Expect(found).To(BeTrue())

			logger.Info("frontendonly case 6 completed \n")
		})

		It("test case 7 - ipOverride with valid hostname", Label("sdn", "frontendonly", "case7"), func() {
			logger.Info("frontendonly case7 started")

			feLeaf1CR := &idcnetworkv1alpha1.Switch{}
			key := types.NamespacedName{Name: "clab-frontendonly-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, feLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf1SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-frontend-leaf1")
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf1Config, err := feLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf2SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-frontend-leaf2")
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf2Config, err := feLeaf2SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			// Overriding hostname leaf1 switch with hostname of leaf2 switch
			feLeaf1CR.Spec.IpOverride = "clab-frontendonly-frontend-leaf2"

			err = testHelper.K8sClient.Update(ctx, feLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())

			// After overriding to clab-frontendonly-frontend-leaf2 now leaf2 will be updated
			err = testutils.Timeout(30*time.Second, func(ctx context.Context) error {
				return testHelper.MockBM.ReserveSingleBMHWithVXX(ctx, "server1-1", int64(102))
			})
			Expect(err).ShouldNot(HaveOccurred())
			time.Sleep(2 * time.Second)

			// The ContainerLabManager used in the tests does not use the overridden IP, so it really does connect to leaf1.
			feLeaf1Config2, err := feLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf2Config2, err := feLeaf2SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			// config diff for leaf1
			feLeaf1Diff, err := testutils.Diff(feLeaf1Config, feLeaf1Config2, 2)
			Expect(err).ShouldNot(HaveOccurred())

			// Vlan update should not happen on leaf1
			feLeaf1DiffExpected := ``

			// config Diff for leaf2
			feLeaf2Diff, err := testutils.Diff(feLeaf2Config, feLeaf2Config2, 2)
			Expect(err).ShouldNot(HaveOccurred())

			// vlan update should happen on leaf2
			feLeaf2DiffExpected := `@@ @@
 interface Ethernet5
    description server2-1 TenantBM
-   switchport access vlan 100
+   switchport access vlan 102
    switchport
+   no lldp transmit
 !
 interface Ethernet6
`
			// No config diff for leaf1 expected after overriding leaf1
			Expect(feLeaf1Diff).Should(Equal(feLeaf1DiffExpected), "Actual feLeaf1Diff: %s  Expected: %s", feLeaf1Diff, feLeaf1DiffExpected)

			// expect config diff for leaf2 after overriding leaf1
			Expect(feLeaf2Diff).Should(Equal(feLeaf2DiffExpected), "Actual feLeaf2Diff: %s  Expected: %s", feLeaf2Diff, feLeaf2DiffExpected)

			logger.Info("frontendonly case 7 completed \n")
		})

		It("test case 8 - Test to update description field in switchport", Label("sdn", "frontendonly", "case8"), func() {
			logger.Info("frontendonly case8 started")

			// Wait for the switchport status to be updated or read from the switch.
			switchPortCR := &idcnetworkv1alpha1.SwitchPort{}
			for i := 0; i < 120; i++ {
				// Get the switchport
				key := types.NamespacedName{Name: "ethernet5.clab-frontendonly-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
				err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
				Expect(k8sclient.IgnoreNotFound(err)).ShouldNot(HaveOccurred())
				if switchPortCR.Status.LinkStatus == "connected" {
					break
				}
				time.Sleep(1 * time.Second)
			}

			// Set and update switchport with new values
			switchPortCR.Spec.Description = "test access"
			err = testHelper.K8sClient.Update(ctx, switchPortCR)
			Expect(err).ShouldNot(HaveOccurred())

			// Wait for changes to be applied
			startTime := time.Now()
			for time.Now().Before(startTime.Add(60 * time.Second)) {
				key := types.NamespacedName{Name: "ethernet5.clab-frontendonly-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
				err := testHelper.K8sClient.Get(ctx, key, switchPortCR)
				Expect(err).ShouldNot(HaveOccurred())
				if switchPortCR.Status.Description == switchPortCR.Spec.Description {
					break
				}
				time.Sleep(1 * time.Second)
			}
			Expect(switchPortCR.Status.Description).To(Equal("test access"))
		})

		It("test case 9 - Test metrics are exposed, increase when vlan changed, and are not excessively increasing", Label("sdn", "frontendonly", "case9"), func() {
			logger.Info("frontendonly case9 started")

			const vlanUpdateMetric = "eapiUpdateVlancounter{application=\"sdn-controller\"}"
			const getSwitchportsMetric = "eapiGetSwitchPortscounter{application=\"sdn-controller\"}"
			const updateModeMetric = "eapiUpdateModecounter{application=\"sdn-controller\"}"
			const updateDescriptionMetric = "eapiUpdateDescriptioncounter{application=\"sdn-controller\"}"
			const updateNativeVlanMetric = "eapiUpdateNativeVlancounter{application=\"sdn-controller\"}"
			const updateBGPCommunityMetric = "eapiUpdateBGPCommunitycounter{application=\"sdn-controller\"}"
			const getBGPCommunityMetric = "eapiGetBGPCommunitycounter{application=\"sdn-controller\"}"
			const updateTrunkGroupsMetric = "eapiUpdateTrunkGroupscounter{application=\"sdn-controller\"}"
			const getVlansMetric = "eapiGetVlanscounter{application=\"sdn-controller\"}"

			allMetrics := []string{vlanUpdateMetric, getSwitchportsMetric, updateModeMetric, updateDescriptionMetric, updateNativeVlanMetric, updateBGPCommunityMetric, getBGPCommunityMetric, updateTrunkGroupsMetric, getVlansMetric}

			// Wait for the switchport status to be updated or read from the switch.
			logger.Info("waiting for initial switchport status to be read from switch...")
			switchPortCR := &idcnetworkv1alpha1.SwitchPort{}
			for i := 0; i < 120; i++ {
				// Get the switchport
				key := types.NamespacedName{Name: "ethernet5.clab-frontendonly-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
				err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
				Expect(k8sclient.IgnoreNotFound(err)).ShouldNot(HaveOccurred())
				if switchPortCR.Status.LinkStatus == "connected" {
					break
				}
				time.Sleep(1 * time.Second)
			}

			err, metricsAtStart := testutils.GetMetrics(allMetrics)
			Expect(err).ShouldNot(HaveOccurred())

			// Set and update switchport with new values
			switchPortCR.Spec.VlanId = 105
			err = testHelper.K8sClient.Update(ctx, switchPortCR)
			Expect(err).ShouldNot(HaveOccurred())

			// Wait for changes to be applied
			startTime := time.Now()
			for time.Now().Before(startTime.Add(60 * time.Second)) {
				key := types.NamespacedName{Name: "ethernet5.clab-frontendonly-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
				err := testHelper.K8sClient.Get(ctx, key, switchPortCR)
				Expect(err).ShouldNot(HaveOccurred())
				if switchPortCR.Status.VlanId == switchPortCR.Spec.VlanId {
					break
				}
				time.Sleep(1 * time.Second)
			}
			Expect(switchPortCR.Status.VlanId).To(Equal(int64(105)))

			err, metricsAfterUpdate := testutils.GetMetrics(allMetrics)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(metricsAfterUpdate[vlanUpdateMetric]).To(Equal(metricsAtStart[vlanUpdateMetric] + 1))
			Expect(metricsAfterUpdate[getVlansMetric]).To(Equal(metricsAtStart[getVlansMetric] + 1))                         // Should be called once, during the update, to check the vlan exists on the switch.
			Expect(metricsAfterUpdate[getSwitchportsMetric]).To(BeNumerically("<=", metricsAtStart[getSwitchportsMetric]+4)) // One right after the update, another because something changed, another possible one per switch for the periodic check.

			// Wait for > 1 period to make sure it doesn't KEEP increasing.
			time.Sleep(70 * time.Second)
			err, metricsAfterWait := testutils.GetMetrics(allMetrics)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(metricsAfterWait[vlanUpdateMetric]).To(Equal(metricsAfterUpdate[vlanUpdateMetric]))

			// Other UPDATEs to the switch shouldn't have happened at all:
			Expect(metricsAfterWait[updateModeMetric]).To(BeNumerically("==", metricsAtStart[updateModeMetric]), updateModeMetric)
			Expect(metricsAfterWait[updateDescriptionMetric]).To(BeNumerically("==", metricsAtStart[updateDescriptionMetric]), updateDescriptionMetric)
			Expect(metricsAfterWait[updateNativeVlanMetric]).To(BeNumerically("==", metricsAtStart[updateNativeVlanMetric]), updateNativeVlanMetric)
			Expect(metricsAfterWait[updateBGPCommunityMetric]).To(BeNumerically("==", metricsAtStart[updateBGPCommunityMetric]), updateBGPCommunityMetric)
			Expect(metricsAfterWait[updateTrunkGroupsMetric]).To(BeNumerically("==", metricsAtStart[updateTrunkGroupsMetric]), updateTrunkGroupsMetric)

			// During the wait, "GET" metrics should have increased by 2 or 3:
			Expect(metricsAfterWait[getSwitchportsMetric]).To(BeNumerically("<=", metricsAfterUpdate[getSwitchportsMetric]+5), getSwitchportsMetric) // 1 accelerated check after the change above + 1 or 2 per switch for the standard "60sec" update (70 seconds + a bit of time for the metrics jobs to run).
			Expect(metricsAfterWait[getSwitchportsMetric]).To(BeNumerically(">=", metricsAfterUpdate[getSwitchportsMetric]+2), getSwitchportsMetric) // at least 1 per switch standard "60sec" update.
			Expect(metricsAfterWait[getBGPCommunityMetric]).To(BeNumerically("<=", metricsAfterUpdate[getBGPCommunityMetric]+5), getBGPCommunityMetric)
			Expect(metricsAfterWait[getBGPCommunityMetric]).To(BeNumerically(">=", metricsAfterUpdate[getBGPCommunityMetric]+2), getBGPCommunityMetric)

			// getVlans shouldn't be called since the update (it is only used to check vlan exists on the switch when vlan is changed).
			Expect(metricsAfterWait[getVlansMetric]).To(Equal(metricsAfterUpdate[getVlansMetric]))

			logger.Info("frontendonly case9 completed")
		})
	})

	Describe("With readOnly config", Label("readonly"), func() {
		BeforeAll(func() {
			var wg sync.WaitGroup

			// Increment the WaitGroup counter for 2 goroutines
			wg.Add(2)

			// First goroutine
			go func() {
				// Deploy SDN
				defer wg.Done() // Decrement the counter when the goroutine finishes
				defer GinkgoRecover()
				err = testHelper.DeploySDNWithConfigFile(readOnlyConfigTemplatePath, false)
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
				err = testHelper.ContainerLabManager.Deploy(topologyFrontendonly)
				Expect(err).ShouldNot(HaveOccurred())
				topology, err = testHelper.ContainerLabManager.Connect(topologyFrontendonly)
				Expect(err).ShouldNot(HaveOccurred())
			}()

			// Wait for both goroutines to finish
			wg.Wait()
		})

		BeforeEach(func() {
			// remove any existing K8s resources
			err = testHelper.DeleteAllK8sResources()
			Expect(err).ShouldNot(HaveOccurred())

			// reset all the switches for this topology
			logger.Info("Resetting config for all switches...")
			err := testHelper.ContainerLabManager.ResetAllSwitches(topologyFrontendonly)
			Expect(err).ShouldNot(HaveOccurred())

			// recreate the K8s resource
			logger.Info("Creating k8s resources for topology...")
			err = testHelper.CreateK8sResourcesForTopology(topology)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should not change switch config", Label("sdn", "frontendonly", "rodontchangecfg"), func() {
			logger.Info("frontendonly rodontchangecfg started")

			// get initial config
			feLeaf1SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-frontendonly-frontend-leaf1")
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf1Config, err := feLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(feLeaf1Config).NotTo(BeEmpty())

			// Wait for the switchport status to be updated or read from the switch.
			switchPortCR := &idcnetworkv1alpha1.SwitchPort{}
			for i := 0; i < 120; i++ {
				// Get the switchport
				key := types.NamespacedName{Name: "ethernet5.clab-frontendonly-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
				err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
				Expect(k8sclient.IgnoreNotFound(err)).ShouldNot(HaveOccurred())
				if switchPortCR.Status.LinkStatus == "connected" {
					break
				}
				time.Sleep(1 * time.Second)
			}

			// Set and update switchport with new values (change as many things as possible, so if any of them get the readonly check removed, this test will fail)
			switchPortCR.Spec.Description = "newValue"
			switchPortCR.Spec.VlanId = 101
			switchPortCR.Spec.PortChannel = 7
			switchPortCR.Spec.TrunkGroups = &[]string{"Provider_Nets"}
			switchPortCR.Spec.NativeVlan = 102
			err = testHelper.K8sClient.Update(ctx, switchPortCR)
			Expect(err).ShouldNot(HaveOccurred())

			// Wait for changes (not) to be applied
			logger.Info("waiting, then will check updates were not applied...")
			time.Sleep(70 * time.Second)

			// Check that nothing changed.
			Expect(switchPortCR.Status.Description).NotTo(Equal("newValue"))
			Expect(switchPortCR.Status.VlanId).To(Equal(int64(100)))

			// Check nothing changed on the switch itself
			feLeaf1ConfigAfterWait, err := feLeaf1SwClient.GetRunningConfig(ctx)

			// config diff for leaf1
			feLeaf1Diff, err := testutils.Diff(feLeaf1Config, feLeaf1ConfigAfterWait, 2)
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf1DiffExpected := ""

			// No config diff for leaf1 expected after overriding leaf1
			Expect(feLeaf1Diff).Should(Equal(feLeaf1DiffExpected), "Actual feLeaf1Diff: %s  Expected: %s", feLeaf1Diff, feLeaf1DiffExpected)
		})
	})
})
