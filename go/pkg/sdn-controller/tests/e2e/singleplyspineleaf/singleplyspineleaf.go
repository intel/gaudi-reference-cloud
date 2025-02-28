//go:build ginkgo_only

package e2e

import (
	"context"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"sync"
	"time"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	sdnclient "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/client"
	testutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/test_utils"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/clab"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	topologyDir        = "../../../../../networking/containerlab"
	topologyFile       = "../../../../../networking/containerlab/singleplyspineleaf/singleplyspineleaf.clab.yml"
	configTemplatePath = "test-e2e-sdn.yaml.gotmpl"
	eAPISecretPath     = "/vault/secrets/eapi"
)

var _ = Describe("Topology singleplyspineleaf Tests", Ordered, Label("sdn", "tsdn", "singleplyspineleaf"), func() {
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
			err = testHelper.ContainerLabManager.Deploy(topologyFile)
			Expect(err).ShouldNot(HaveOccurred())
			topology, err = testHelper.ContainerLabManager.Connect(topologyFile)
			Expect(err).ShouldNot(HaveOccurred())
		}()

		// Wait for both goroutines to finish
		wg.Wait()
	})

	BeforeEach(func() {
		// remove the K8s resources
		err = testHelper.DeleteAllK8sResources()
		Expect(err).ShouldNot(HaveOccurred())

		// reset all the switches for this topology
		err := testHelper.ContainerLabManager.ResetAllSwitches(topologyFile)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("with full k8s resources ready", func() {

		BeforeEach(func() {
			// recreate the K8s resource
			err = testHelper.CreateK8sResourcesForTopology(topology)
			Expect(err).ShouldNot(HaveOccurred())
			err = testHelper.WaitForNodeGroupReady("1", 2, true)
			Expect(err).ShouldNot(HaveOccurred())
			err = testHelper.WaitForNodeGroupReady("2", 1, true)
			Expect(err).ShouldNot(HaveOccurred())
		})

		/*
			Purpose of the Test: (Describe what the test is intended to verify.)
				This tests BGP community can be updated on switches in the accelerator fabric.

			Test Steps: (Outline the steps the test will execute, making it clear what actions are being performed and why.)
				- Deploy the containerlab topology
				- Create the BMH and Switch CRs for this topology
				- Call the SDNClient.UpdateNetworkNodeConfig() to update BGP for group1
				- Check the switch config has been updated.
				- Call the SDNClient.UpdateNetworkNodeConfig() again, to update BGP for a second server in group1
				- Check the switch config has not changed further.
				- Call the SDNClient.UpdateNetworkNodeConfig() for the second group, making them into a large-cluster consisting of both nodegroups.
				- Check the switch config has been updated for the second group.
				- Revert the BGP community back to 1000 (simulating deprovisioning) & check switch-config has gone back to what it was initially.

			Test Coverage: (outlines the specific features being tested and any corner cases or bugs that the test case addresses.)
				 - BMH Controller can correctly create the NetworkNodes based on the BMH and move them in to the correct Nodegroup.
				 - Nodegroup is moved into correct pool & comes out of maintenance mode.
				 - SDNClient can update the BGP community on the switches.
		*/
		It("test case 1 - update accelerator BGP Cmty", Label("sdn", "singleplyspineleaf", "case1"), func() {
			fmt.Printf("singleplyspineleaf case 1 starts... \n")
			accLeaf1SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-singleplyspineleaf-acc-leaf1")
			accLeaf2SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-singleplyspineleaf-acc-leaf2")
			//frontendSwClient2, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-singleplyspineleaf-frontend-leaf2")

			//////////////////////
			// fetch the running-config before making any changes
			//////////////////////
			accLeaf1InitConfig, err := accLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			accLeaf2InitConfig, err := accLeaf2SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// make the changes
			//////////////////////
			// simulate the BMaaS to reserve "group1" for a tenant
			err = testHelper.MockBM.ReserveSingleBMHWithXBX(ctx, "server1-1", int64(1234))
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// get the switches running-config after the changes
			//////////////////////
			accLeaf1ActualConfig, err := accLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			accLeaf2ActualConfig, err := accLeaf2SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// compare the latest running-config with the expected one
			//////////////////////
			accLeaf1Diff, err := testutils.Diff(accLeaf1InitConfig, accLeaf1ActualConfig, 2)
			Expect(err).ShouldNot(HaveOccurred())

			accLeaf1DiffExpected := `@@ @@
 !
 ip prefix-list outgoing_group seq 10 permit 10.212.1.0/24
-ip community-list incoming_group permit 101:1000
+ip community-list incoming_group permit 101:1234
 !
 ip route 0.0.0.0/0 172.20.20.1
@@ @@
 route-map adv-set-comm permit 10
    match ip address prefix-list outgoing_group
-   set community 101:1000
+   set community 101:1234
 !
 route-map rcv-from-spine permit 10
`
			Expect(accLeaf1Diff).Should(Equal(accLeaf1DiffExpected), "Actual accLeaf1Diff: %s", accLeaf1Diff)

			accLeaf2Diff, err := testutils.Diff(accLeaf2InitConfig, accLeaf2ActualConfig, 2)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(accLeaf2Diff).Should(Equal(""), "Actual accLeaf2Diff: %s", accLeaf2Diff)

			// Reserving another node in the same nodegroup should not change the switch-config any further.
			err = testHelper.MockBM.ReserveSingleBMHWithXBX(ctx, "server1-2", int64(1234))

			accLeaf1ActualConfig2, err := accLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			accLeaf2ActualConfig2, err := accLeaf2SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			// Accelerator switch config should not have changed further
			accLeaf1Diff2, err := testutils.Diff(accLeaf1ActualConfig, accLeaf1ActualConfig2, 2)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(accLeaf1Diff2).Should(Equal(""), "Actual accLeaf1Diff2: %s", accLeaf1Diff2)

			accLeaf2Diff2, err := testutils.Diff(accLeaf2ActualConfig, accLeaf2ActualConfig2, 2)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(accLeaf2Diff2).Should(Equal(""), "Actual accLeaf2Diff2: %s", accLeaf2Diff2)

			// Now reserve "group2" for the same tenant:
			err = testHelper.MockBM.ReserveSingleBMHWithXBX(ctx, "server2-1", int64(1234))
			Expect(err).ShouldNot(HaveOccurred())

			accLeaf1ActualConfig3, err := accLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			accLeaf2ActualConfig3, err := accLeaf2SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			// Leaf1 should not have changed.
			accLeaf1Diff3, err := testutils.Diff(accLeaf1ActualConfig2, accLeaf1ActualConfig3, 2)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(accLeaf1Diff3).Should(Equal(""), "Actual accLeaf1Diff2: %s", accLeaf1Diff2)

			// Leaf2 should now have the same community ID as leaf1.
			accLeaf2Diff3, err := testutils.Diff(accLeaf2ActualConfig2, accLeaf2ActualConfig3, 2)
			Expect(err).ShouldNot(HaveOccurred())
			accLeaf2DiffExpected := `@@ @@
 !
 ip prefix-list outgoing_group seq 10 permit 10.212.2.0/24
-ip community-list incoming_group permit 101:1000
+ip community-list incoming_group permit 101:1234
 !
 ip route 0.0.0.0/0 172.20.20.1
@@ @@
 route-map adv-set-comm permit 10
    match ip address prefix-list outgoing_group
-   set community 101:1000
+   set community 101:1234
 !
 route-map rcv-from-spine permit 10
`
			Expect(accLeaf2Diff3).Should(Equal(accLeaf2DiffExpected), "Actual accLeaf2Diff3: %s", accLeaf2Diff3)

			// Revert both servers back to BGP community 1000:
			err = testHelper.MockBM.ReserveSingleBMHWithXBX(ctx, "server1-1", int64(1000))
			err = testHelper.MockBM.ReserveSingleBMHWithXBX(ctx, "server2-1", int64(1000))

			accLeaf1ActualConfig4, err := accLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			accLeaf2ActualConfig4, err := accLeaf2SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			// Switch config should be back to what it was initially.
			accLeaf1Diff4, err := testutils.Diff(accLeaf1InitConfig, accLeaf1ActualConfig4, 2)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(accLeaf1Diff4).Should(Equal(""), "Actual accLeaf1Diff3: %s", accLeaf1Diff4)

			accLeaf2Diff4, err := testutils.Diff(accLeaf2InitConfig, accLeaf2ActualConfig4, 2)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(accLeaf2Diff4).Should(Equal(""), "Actual accLeaf2Diff3: %s", accLeaf2Diff4)

			fmt.Printf("singleplyspineleaf case 1 completed \n")
		})

		/*
			Purpose of the Test:
				This tests Frontend-network can be updated (eg. as multiple nodes finish provisioning, or as a spare is provisioned) without affecting the BGP community of the accelerator switches.
		*/
		It("test case 2 - Can change frontend network without affecting accelerator BGP cmty", Label("sdn", "singleplyspineleaf", "case2"), func() {
			fmt.Printf("singleplyspineleaf case 2 starts... \n")
			accLeaf1SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-singleplyspineleaf-acc-leaf1")
			//accLeaf2SwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-singleplyspineleaf-acc-leaf2")
			frontendSwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-singleplyspineleaf-frontend-leaf1")

			//////////////////////
			// fetch the running-config before making any changes
			//////////////////////
			accLeaf1InitConfig, err := accLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf1InitConfig, err := frontendSwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// make the changes
			//////////////////////
			// simulate the BMaaS to reserve a single node for a tenant
			err = testHelper.MockBM.ReserveSingleBMHWithVXX(ctx, "server1-1", int64(101))
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// get the switches running-config after the changes
			//////////////////////
			accLeaf1ActualConfig, err := accLeaf1SwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf1ActualConfig, err := frontendSwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			//////////////////////
			// compare the latest running-config with the expected one
			//////////////////////
			accLeaf1Diff, err := testutils.Diff(accLeaf1InitConfig, accLeaf1ActualConfig, 2)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(accLeaf1Diff).Should(Equal(""), "Actual accLeaf1Diff: %s", accLeaf1Diff)

			feLeaf1Diff, err := testutils.Diff(feLeaf1InitConfig, feLeaf1ActualConfig, 2)
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf1DiffExpected := `@@ @@
 !
 interface Ethernet1
-   switchport access vlan 100
+   switchport access vlan 101
    switchport
+   no lldp transmit
 !
 interface Ethernet2
`
			Expect(feLeaf1Diff).Should(Equal(feLeaf1DiffExpected), "Actual feLeaf1Diff: %s", feLeaf1Diff)

			fmt.Printf("singleplyspineleaf case 2 completed \n")

		})

		It("test case 3 - NodeGroups counts are calculated shortly after enrollment", Label("sdn", "singleplyspineleaf", "case3"), func() {
			fmt.Printf("singleplyspineleaf case 3 started \n")

			// Do the remaining checks using Expects so we get useful error messages / test failures.
			nodeGroup1CR := &idcnetworkv1alpha1.NodeGroup{}
			key := types.NamespacedName{Name: "1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, nodeGroup1CR)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(nodeGroup1CR.Status.NetworkNodesCount).To(Equal(2))
			Expect(nodeGroup1CR.Status.FrontEndSwitchCount).To(Equal(1))
			Expect(nodeGroup1CR.Status.AccSwitchCount).To(Equal(1))
			Expect(nodeGroup1CR.Spec.AcceleratorLeafSwitches[0]).To(Equal("clab-singleplyspineleaf-acc-leaf1"))

			nodeGroup2CR := &idcnetworkv1alpha1.NodeGroup{}
			key = types.NamespacedName{Name: "2", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, nodeGroup2CR)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(nodeGroup2CR.Status.NetworkNodesCount).To(Equal(1))
			Expect(nodeGroup2CR.Status.FrontEndSwitchCount).To(Equal(1))
			Expect(nodeGroup2CR.Status.AccSwitchCount).To(Equal(1))
			Expect(nodeGroup2CR.Spec.AcceleratorLeafSwitches[0]).To(Equal("clab-singleplyspineleaf-acc-leaf2"))

			fmt.Printf("singleplyspineleaf case 3 completed \n")
		})

		It("test case 4 -  test for maintenance flag", Label("sdn", "singleplyspineleaf", "case4"), func() {
			fmt.Printf("singleplyspineleaf case 4 started \n")

			feLeaf1CR := &idcnetworkv1alpha1.Switch{}
			key := types.NamespacedName{Name: "clab-singleplyspineleaf-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, feLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf1InitSwClient, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-singleplyspineleaf-frontend-leaf1")
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf1InitConfig, err := feLeaf1InitSwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			// setting maintenance flag to "true"
			feLeaf1CROrig := feLeaf1CR.DeepCopy()
			feLeaf1CR.Spec.Maintenance = "true"
			err = testHelper.K8sClient.Patch(ctx, feLeaf1CR, client.MergeFrom(feLeaf1CROrig))
			Expect(err).ShouldNot(HaveOccurred())
			time.Sleep(2 * time.Second)

			err = testHelper.MockBM.SDNClient.UpdateNetworkNodeConfig(ctx, sdnclient.NetworkNodeConfUpdateRequest{
				NetworkNodeName:    "server1-1",
				FrontEndFabricVlan: 102,
			})
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(60 * time.Second)

			feLeaf1CRConfig, err := feLeaf1InitSwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf1Diff, err := testutils.Diff(feLeaf1InitConfig, feLeaf1CRConfig, 2)
			Expect(err).ShouldNot(HaveOccurred())

			// There should have been no change on the switch because maintenance flag is set.
			Expect(feLeaf1Diff).Should(Equal(""), "Actual feLeaf1Diff: %s  Expected: ", feLeaf1Diff)

			// setting maintenance flag to "", change seen during maintenance-window should now be applied.
			feLeaf1CR = &idcnetworkv1alpha1.Switch{}
			key = types.NamespacedName{Name: "clab-singleplyspineleaf-frontend-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, feLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf1CROrig = feLeaf1CR.DeepCopy()
			feLeaf1CR.Spec.Maintenance = ""
			err = testHelper.K8sClient.Patch(ctx, feLeaf1CR, client.MergeFrom(feLeaf1CROrig))
			Expect(err).ShouldNot(HaveOccurred())
			time.Sleep(10 * time.Second)

			feLeaf1CRConfig2, err := feLeaf1InitSwClient.GetRunningConfig(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			feLeaf1Diff2, err := testutils.Diff(feLeaf1CRConfig, feLeaf1CRConfig2, 2)
			Expect(err).ShouldNot(HaveOccurred())

			feLeaf1Diff2Expected := `@@ @@
 !
 interface Ethernet1
-   switchport access vlan 100
+   switchport access vlan 102
    switchport
+   no lldp transmit
 !
 interface Ethernet2
`
			Expect(feLeaf1Diff2).Should(Equal(feLeaf1Diff2Expected), "Actual feLeaf1Diff2: %s  Expected: %s", feLeaf1Diff2, feLeaf1Diff2Expected)

			fmt.Printf("singleplyspineleaf case 4 completed \n")
		})
	})

	It("test case 5 - Nodes added to an existing provisioned NodeGroup that was already in an XBX pool get set to the nodegroups' BGPCmty", Label("sdn", "singleplyspineleaf", "case5"), func() {
		fmt.Printf("singleplyspineleaf case 5 started \n")

		// Add only SOME of the networkNodes & nodegroups to k8s (simulating partial enrollment)
		ctx := context.Background()
		bmhs, switches, mappings, err := testHelper.GenerateK8sResourcesFromClabTopology(topology)
		Expect(err).ShouldNot(HaveOccurred())

		// Create k8s CRDs, but only enroll server1-1
		//server1_2BMH := bmhs["server1-2"]
		//Expect(server1_2BMH).ShouldNot(BeNil())
		server2_1BMH := bmhs["server2-1"]
		Expect(server2_1BMH).ShouldNot(BeNil())
		delete(bmhs, "server1-2")
		delete(bmhs, "server2-1")
		delete(mappings, "2")
		err = testHelper.CreateK8sResources(ctx, bmhs, switches, mappings, true)
		Expect(err).ShouldNot(HaveOccurred())

		// Wait for nodeGroups to be moved into the XBX pool, and be reconciled.
		err = testHelper.WaitForNodeGroupReady("1", 1, true)
		Expect(err).ShouldNot(HaveOccurred())

		// Provision nodeGroup.
		err = testHelper.MockBM.ReserveSingleBMHWithXBX(ctx, "server1-1", int64(1234))

		// wait for leaf1 to be updated.
		startTime := time.Now()
		accLeaf1CR := &idcnetworkv1alpha1.Switch{}
		for time.Now().Before(startTime.Add(65 * time.Second)) {
			key := types.NamespacedName{Name: "clab-singleplyspineleaf-acc-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, accLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())

			if accLeaf1CR.Status.SwitchBGPConfigStatus != nil && accLeaf1CR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity == 1234 {
				break
			}
			time.Sleep(1 * time.Second)
		}
		Expect(accLeaf1CR.Spec.BGP.BGPCommunity).To(Equal(int64(1234)))
		Expect(accLeaf1CR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity).To(Equal(int64(1234)))

		// Now simulate another server (connected to a different accelerator switch, so we can test that the new switch gets updated) being enrolled.
		server2_1BMH.Labels[idcnetworkv1alpha1.LabelBMHGroupID] = "1" // Override to group-1 (just for testing - in reality we shouldn't have nodes connected to different acc-leafs in the same group)
		lateEnrollmentBmhs := map[string]*baremetalv1alpha1.BareMetalHost{
			"server2-1": server2_1BMH,
		}
		err = testHelper.CreateK8sResources(ctx, lateEnrollmentBmhs, map[string]*idcnetworkv1alpha1.Switch{}, map[string]*idcnetworkv1alpha1.NodeGroupToPoolMapping{}, true)
		Expect(err).ShouldNot(HaveOccurred())

		// Get the leaf-2 switch that was newly added to the nodeGroup.
		startTime = time.Now()
		accLeaf2CR := &idcnetworkv1alpha1.Switch{}
		for time.Now().Before(startTime.Add(65 * time.Second)) {
			key2 := types.NamespacedName{Name: "clab-singleplyspineleaf-acc-leaf2", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key2, accLeaf2CR)
			Expect(err).ShouldNot(HaveOccurred())

			if accLeaf2CR.Status.SwitchBGPConfigStatus != nil && accLeaf2CR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity == 1234 {
				break
			}
			time.Sleep(1 * time.Second)
		}
		Expect(accLeaf2CR.Spec.BGP.BGPCommunity).To(Equal(int64(1234)))
		Expect(accLeaf2CR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity).To(Equal(int64(1234)))

		fmt.Printf("singleplyspineleaf case 5 completed \n")
	})

	It("test case 6 - Nodes added to an existing empty NodeGroup that was already in an XBX pool get set correctly", Label("sdn", "singleplyspineleaf", "case6"), func() {
		fmt.Printf("singleplyspineleaf case 6 started \n")

		ctx := context.Background()
		bmhs, switches, mappings, err := testHelper.GenerateK8sResourcesFromClabTopology(topology)
		Expect(err).ShouldNot(HaveOccurred())

		// Create k8s CRDs, but only enroll server2-1
		delete(bmhs, "server1-1")
		delete(bmhs, "server1-2")
		delete(mappings, "1")
		err = testHelper.CreateK8sResources(ctx, bmhs, switches, mappings, true)
		Expect(err).ShouldNot(HaveOccurred())

		err = testHelper.WaitForNodeGroupReady("2", 1, true)
		Expect(err).ShouldNot(HaveOccurred())

		// Create another copy of BMHs from topology (avoiding mutation of the original bmhs map)
		bmhs, _, _, err = testHelper.GenerateK8sResourcesFromClabTopology(topology)
		Expect(err).ShouldNot(HaveOccurred())
		// Un-enroll server2-1, leaving nodeGroup2 empty, but still with pool XBX.
		fmt.Printf("un-enrolling server2-1... \n")
		server2_1BMH := bmhs["server2-1"]
		Expect(server2_1BMH).To(Not(BeNil()))

		lateEnrollmentBmhs := map[string]*baremetalv1alpha1.BareMetalHost{
			"server2-1": server2_1BMH,
		}
		testHelper.DeleteK8sResources(ctx, lateEnrollmentBmhs, map[string]*idcnetworkv1alpha1.Switch{}, map[string]*idcnetworkv1alpha1.NodeGroupToPoolMapping{})

		time.Sleep(3 * time.Second) // Not strictly needed, but helps to slow the test down a bit for better watchability in k9s.

		// Wait for nodeGroup2 to be empty.
		fmt.Printf("Waiting for nodegroup 2 to be emptied... \n")
		err = testHelper.WaitForNodeGroupReady("2", 0, true)

		// NodeGroup 2 should still have pool XBX.
		nodeGroup2CR := &idcnetworkv1alpha1.NodeGroup{}
		key := types.NamespacedName{Name: "2", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = testHelper.K8sClient.Get(ctx, key, nodeGroup2CR)
		Expect(nodeGroup2CR.Labels[idcnetworkv1alpha1.LabelPool]).To(Equal("XBX"))

		// Now re-enroll the server.
		fmt.Printf("re-enrolling server2-1... \n")
		err = testHelper.CreateK8sResources(ctx, lateEnrollmentBmhs, map[string]*idcnetworkv1alpha1.Switch{}, map[string]*idcnetworkv1alpha1.NodeGroupToPoolMapping{}, true)
		Expect(err).ShouldNot(HaveOccurred())

		time.Sleep(3 * time.Second) // Not strictly needed, but helps to slow the test down a bit for better watchability in k9s.

		err = testHelper.WaitForNodeGroupReady("2", 1, true)
		Expect(err).ShouldNot(HaveOccurred())

		// Provision the server.
		fmt.Printf("provisioning server2-1... \n")
		err = testHelper.MockBM.ReserveSingleBMHWithXBX(ctx, "server2-1", int64(1234))

		// wait for leaf2 to be updated.
		startTime := time.Now()
		accLeaf2CR := &idcnetworkv1alpha1.Switch{}
		for time.Now().Before(startTime.Add(65 * time.Second)) {
			key := types.NamespacedName{Name: "clab-singleplyspineleaf-acc-leaf2", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, accLeaf2CR)
			Expect(err).ShouldNot(HaveOccurred())

			if accLeaf2CR.Status.SwitchBGPConfigStatus != nil && accLeaf2CR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity == 1234 {
				break
			}
			time.Sleep(1 * time.Second)
		}
		Expect(accLeaf2CR.Spec.BGP.BGPCommunity).To(Equal(int64(1234)))
		Expect(accLeaf2CR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity).To(Equal(int64(1234)))

		fmt.Printf("singleplyspineleaf case 6 completed \n")
	})

	It("test case 7 - Add nodes to nodegroup BEFORE moving it to XBX (creating the mapping AFTER enrollment)", Label("sdn", "singleplyspineleaf", "case7"), func() {
		fmt.Printf("singleplyspineleaf case 7 started \n")

		ctx := context.Background()
		bmhs, switches, mappings, err := testHelper.GenerateK8sResourcesFromClabTopology(topology)
		Expect(err).ShouldNot(HaveOccurred())

		// Create k8s CRDs, and enroll server1-1
		err = testHelper.CreateK8sResources(ctx, bmhs, switches, map[string]*idcnetworkv1alpha1.NodeGroupToPoolMapping{}, true)
		Expect(err).ShouldNot(HaveOccurred())

		err = testHelper.WaitForNodeGroupReady("1", 2, false)
		Expect(err).ShouldNot(HaveOccurred())

		// Create another copy of BMHs from topology (avoiding mutation of the original bmhs map)
		bmhs, _, _, err = testHelper.GenerateK8sResourcesFromClabTopology(topology)
		Expect(err).ShouldNot(HaveOccurred())

		// Create the mappings
		fmt.Printf("Moving nodeGroup to XBX pool... \n")
		err = testHelper.CreateK8sResources(ctx, map[string]*baremetalv1alpha1.BareMetalHost{}, map[string]*idcnetworkv1alpha1.Switch{}, mappings, true)
		Expect(err).ShouldNot(HaveOccurred())

		err = testHelper.WaitForNodeGroupReady("1", 2, true)
		Expect(err).ShouldNot(HaveOccurred())

		// Provision the server.
		fmt.Printf("provisioning nodeGroup1... \n")
		err = testHelper.MockBM.ReserveSingleBMHWithXBX(ctx, "server1-1", int64(1234))

		// wait for leaf1 to be updated.
		startTime := time.Now()
		accLeaf1CR := &idcnetworkv1alpha1.Switch{}
		for time.Now().Before(startTime.Add(65 * time.Second)) {
			key := types.NamespacedName{Name: "clab-singleplyspineleaf-acc-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, accLeaf1CR)
			Expect(err).ShouldNot(HaveOccurred())

			if accLeaf1CR.Status.SwitchBGPConfigStatus != nil && accLeaf1CR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity == 1234 {
				break
			}
			time.Sleep(1 * time.Second)
		}
		Expect(accLeaf1CR.Spec.BGP.BGPCommunity).To(Equal(int64(1234)))
		Expect(accLeaf1CR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity).To(Equal(int64(1234)))

		fmt.Printf("singleplyspineleaf case 7 completed \n")
	})

	It("test case 8 - BGPCommunityUpdate/Get metrics do not excessively increase", Label("sdn", "singleplyspineleaf", "case8"), func() {
		logger.Info("frontendonly case8 started")

		const updateBGPCommunityMetric = "eapiUpdateBGPCommunitycounter{application=\"sdn-controller\"}"
		const getBGPCommunityMetric = "eapiGetBGPCommunitycounter{application=\"sdn-controller\"}"

		bgpMetrics := []string{updateBGPCommunityMetric, getBGPCommunityMetric}

		// Create switches, but not BMHs / NetworkNodes / mappings. Only one switch is needed for this test.
		_, switches, _, err := testHelper.GenerateK8sResourcesFromClabTopology(topology)
		delete(switches, "clab-singleplyspineleaf-acc-leaf2")
		delete(switches, "clab-singleplyspineleaf-frontend-leaf1")
		err = testHelper.CreateK8sResources(ctx, map[string]*baremetalv1alpha1.BareMetalHost{}, switches, map[string]*idcnetworkv1alpha1.NodeGroupToPoolMapping{}, true)

		logger.Info("Waiting for switch status to be read...")
		// Wait for the switch status to be updated / read from the switch.
		for i := 0; i < 120; i++ {
			switchCR := &idcnetworkv1alpha1.Switch{}
			key := types.NamespacedName{Name: "clab-singleplyspineleaf-acc-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, switchCR)
			Expect(err).ShouldNot(HaveOccurred())
			if switchCR.Status.SwitchBGPConfigStatus != nil && switchCR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity != 0 {
				break
			}
			time.Sleep(1 * time.Second)
		}
		logger.Info("Switch status is up to date")

		logger.Info("Getting initial metrics...")
		// Get initial metrics
		err, metricsAtStart := testutils.GetMetrics(bgpMetrics)
		Expect(err).ShouldNot(HaveOccurred())

		logger.Info("Updating desired BGP Cmty to 1234")
		// Set and update switch with new BGP value
		switchCR := &idcnetworkv1alpha1.Switch{}
		key := types.NamespacedName{Name: "clab-singleplyspineleaf-acc-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = testHelper.K8sClient.Get(ctx, key, switchCR)
		Expect(err).ShouldNot(HaveOccurred())
		switchCR.Spec.BGP = &idcnetworkv1alpha1.BGPConfig{
			BGPCommunity: 1234,
		}
		err = testHelper.K8sClient.Update(ctx, switchCR)
		Expect(err).ShouldNot(HaveOccurred())

		logger.Info("Waiting for updated switch BGPCmty to be read...")
		// Wait for changes to be applied
		startTime := time.Now()
		for time.Now().Before(startTime.Add(60 * time.Second)) {
			key := types.NamespacedName{Name: "clab-singleplyspineleaf-acc-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err := testHelper.K8sClient.Get(ctx, key, switchCR)
			Expect(err).ShouldNot(HaveOccurred())
			if switchCR.Status.SwitchBGPConfigStatus != nil && switchCR.Spec.BGP != nil && switchCR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity == switchCR.Spec.BGP.BGPCommunity {
				break
			}
			time.Sleep(1 * time.Second)
		}
		Expect(switchCR.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity).To(Equal(int64(1234)))

		logger.Info("Checking metrics...")
		err, metricsAfterUpdate := testutils.GetMetrics(bgpMetrics)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(metricsAfterUpdate[updateBGPCommunityMetric]).To(Equal(metricsAtStart[updateBGPCommunityMetric] + 1))
		Expect(metricsAfterUpdate[getBGPCommunityMetric]).To(Equal(metricsAtStart[getBGPCommunityMetric] + 1))

		// Wait for > 1 period to make sure it doesn't KEEP increasing.
		logger.Info("Waiting...")
		time.Sleep(70 * time.Second)
		err, metricsAfterWait := testutils.GetMetrics(bgpMetrics)
		Expect(err).ShouldNot(HaveOccurred())

		// There shouldn't have been any further update:
		Expect(metricsAfterWait[updateBGPCommunityMetric]).To(BeNumerically("==", metricsAfterUpdate[updateBGPCommunityMetric]), updateBGPCommunityMetric)

		// During the wait, "GET" metrics should have increased by 2 or 3:
		Expect(metricsAfterWait[getBGPCommunityMetric]).To(BeNumerically("<=", metricsAfterUpdate[getBGPCommunityMetric]+3), getBGPCommunityMetric) // 1 accelerated check after the change above + 1 or 2 per switch for the standard "60sec" update (70 seconds + a bit of time for the metrics jobs to run).
		Expect(metricsAfterWait[getBGPCommunityMetric]).To(BeNumerically(">=", metricsAfterUpdate[getBGPCommunityMetric]+1), getBGPCommunityMetric)

		// Reset switch to "unmanaged" (ie. -1 DesiredBGPCommunity)
		logger.Info("Setting switch desired BGP back to -1")
		switchCR = &idcnetworkv1alpha1.Switch{}
		key = types.NamespacedName{Name: "clab-singleplyspineleaf-acc-leaf1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = testHelper.K8sClient.Get(ctx, key, switchCR)
		Expect(err).ShouldNot(HaveOccurred())
		switchCR.Spec.BGP.BGPCommunity = -1
		err = testHelper.K8sClient.Update(ctx, switchCR)
		Expect(err).ShouldNot(HaveOccurred())

		// Wait for > 1 period to make sure it doesn't KEEP increasing.
		logger.Info("Waiting again...")
		time.Sleep(70 * time.Second)
		err, metricsAfterSecondWait := testutils.GetMetrics(bgpMetrics)
		Expect(err).ShouldNot(HaveOccurred())

		// Should not have updated the BGPCommunity any further (-1 means do not push changes to the switch)
		Expect(metricsAfterSecondWait[updateBGPCommunityMetric]).To(BeNumerically("==", metricsAfterWait[updateBGPCommunityMetric]), updateBGPCommunityMetric)

		// During the wait, "GET" metrics should have increased by 2 or 3:
		Expect(metricsAfterSecondWait[getBGPCommunityMetric]).To(BeNumerically("<=", metricsAfterWait[getBGPCommunityMetric]+3), getBGPCommunityMetric) // 1 accelerated check after the change above + 1 or 2 per switch for the standard "60sec" update (70 seconds + a bit of time for the metrics jobs to run).
		Expect(metricsAfterSecondWait[getBGPCommunityMetric]).To(BeNumerically(">=", metricsAfterWait[getBGPCommunityMetric]+1), getBGPCommunityMetric)

		logger.Info("singleplyspineleaf case8 completed")
	})

	It("test case 9 - BMH not yet enrolled, should not create NetworkNode", Label("sdn", "singleplyspineleaf", "case9"), func() {
		fmt.Printf("singleplyspineleaf case 9 started \n")

		ctx := context.Background()
		bmhs, switches, mappings, err := testHelper.GenerateK8sResourcesFromClabTopology(topology)
		Expect(err).ShouldNot(HaveOccurred())

		// Create k8s CRDs, only enroll server1-1
		delete(mappings, "2")
		server1_1BMH := bmhs["server1-1"]
		Expect(server1_1BMH).ShouldNot(BeNil())
		delete(server1_1BMH.Labels, "instance-type.cloud.intel.com/mock-bmh-sdn-e2e")
		var bmhsFiltered = make(map[string]*baremetalv1alpha1.BareMetalHost)
		bmhsFiltered["server1-1"] = server1_1BMH
		err = testHelper.CreateK8sResources(ctx, bmhsFiltered, switches, mappings, false)
		Expect(err).ShouldNot(HaveOccurred())

		// Test that NetworkNode has not been created.
		networkNodeCR := &idcnetworkv1alpha1.NetworkNode{}
		key := types.NamespacedName{Name: "server1-1", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = testHelper.K8sClient.Get(ctx, key, networkNodeCR)
		Expect(err).To(HaveOccurred())
		time.Sleep(30 * time.Second)
		err = testHelper.K8sClient.Get(ctx, key, networkNodeCR)
		Expect(err).To(HaveOccurred())

		//Now "complete" the BMH enrollment and the NN should be created.
		BmhCR := &baremetalv1alpha1.BareMetalHost{}
		BMHkey := types.NamespacedName{Name: "server1-1", Namespace: server1_1BMH.Namespace}
		err = testHelper.K8sClient.Get(ctx, BMHkey, BmhCR)
		Expect(err).NotTo(HaveOccurred())
		BmhCR.Labels["instance-type.cloud.intel.com/mock-bmh-sdn-e2e"] = "true"
		err = testHelper.K8sClient.Update(ctx, BmhCR)
		Expect(err).NotTo(HaveOccurred())

		// Wait until NN is created.
		startTime := time.Now()
		for time.Now().Before(startTime.Add(65 * time.Second)) {
			err = testHelper.K8sClient.Get(ctx, key, networkNodeCR)
			if err == nil {
				break
			}
			time.Sleep(1 * time.Second)
		}

		// Check if NN has been created.
		err = testHelper.K8sClient.Get(ctx, key, networkNodeCR)
		Expect(err).NotTo(HaveOccurred())

		fmt.Printf("singleplyspineleaf case 9 completed \n")
	})

	// TODO: Test moving a node BETWEEN 2 nodegroups. Check that the old spec & status gets removed from the switch when moved between nodegroups in different pools (see https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/pull/7954#issuecomment-2369826653)

})
