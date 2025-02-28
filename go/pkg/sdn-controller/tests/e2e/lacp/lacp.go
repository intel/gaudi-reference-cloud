//go:build ginkgo_only

package e2e

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	switchclients "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
	testutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/test_utils"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/clab"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	topologyDir        = "../../../../../networking/containerlab"
	topologylacp       = "../../../../../networking/containerlab/lacp/lacp.clab.yml"
	configTemplatePath = "lacp/lacp.yaml.gotmpl"
	eAPISecretPath     = "/vault/secrets/eapi"
	restAPIBaseURL     = "https://us-dev-1-provider-sdn-controller-rest.idcs-system.svc.cluster.local:443"
	//restAPIBaseURL     = "https://internal-placeholder.com:443"
)

var _ = Describe("Topology LACP Tests", Ordered, Label("sdn", "sdn-rest", "psdn", "lacp"), func() {
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
			err = testHelper.DeploySDNWithConfigFile(configTemplatePath, true)
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
			err = testHelper.ContainerLabManager.Deploy(topologylacp)
			Expect(err).ShouldNot(HaveOccurred())
			topology, err = testHelper.ContainerLabManager.Connect(topologylacp)
			Expect(err).ShouldNot(HaveOccurred())
		}()

		// Wait for both goroutines to finish
		wg.Wait()
	})

	BeforeEach(func() {
		// remove the K8s resources
		err = testHelper.DeleteAllK8sResources()
		Expect(err).ShouldNot(HaveOccurred())
		err = testHelper.DeleteK8sResourcesFromFile(topologyDir + "/lacp/crds/switches.yaml")
		Expect(err).ShouldNot(HaveOccurred())
		err = testHelper.DeleteK8sResourcesFromFile(topologyDir + "/lacp/crds/switchports.yaml")
		Expect(err).ShouldNot(HaveOccurred())

		// recreate the K8s resource
		err = testHelper.CreateK8sResourcesFromFile(topologyDir + "/lacp/crds/switches.yaml")
		Expect(err).ShouldNot(HaveOccurred())
		err = testHelper.CreateK8sResourcesFromFile(topologyDir + "/lacp/crds/switchports.yaml")
		Expect(err).ShouldNot(HaveOccurred())

		// reset all the switches for this topology
		err := testHelper.ContainerLabManager.ResetAllSwitches(topologylacp)
		Expect(err).ShouldNot(HaveOccurred())

		// fmt.Printf("lacp BeforeSuite completed \n")
	})

	It("test case 1 - do all GETs and check they return 200", Label("case1"), func() {
		fmt.Printf("lacp case 1 starts... \n")

		urls := []string{
			restAPIBaseURL + "/devcloud/v4/list/mac_address_table?switch_fqdn=clab-lacp-frontend-leaf1a",
			restAPIBaseURL + "/devcloud/v4/list/ip_mac_info?switch_fqdn=clab-lacp-frontend-leaf1a",
			restAPIBaseURL + "/devcloud/v4/list/lldp_neighbors?switch_fqdn=clab-lacp-frontend-leaf1a",
			restAPIBaseURL + "/devcloud/v4/list/lldp_neighbors?switch_fqdn=clab-lacp-frontend-leaf1a&switch_port=5",
			restAPIBaseURL + "/devcloud/v4/port/running_config?switch_fqdn=clab-lacp-frontend-leaf1a&switch_port=5",
			restAPIBaseURL + "/devcloud/v4/port/running_config?switch_fqdn=clab-lacp-frontend-leaf1a&port_channel=51",
			restAPIBaseURL + "/devcloud/v4/port/details?switch_fqdn=clab-lacp-frontend-leaf1a&switch_port=5",
			restAPIBaseURL + "/devcloud/v4/port/details?switch_fqdn=clab-lacp-frontend-leaf1a&port_channel=51",
			restAPIBaseURL + "/devcloud/v4/list/ports?switch_fqdn=clab-lacp-frontend-leaf1a",
			restAPIBaseURL + "/devcloud/v4/list/vlans?switch_fqdn=clab-lacp-frontend-leaf1a",
		}

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		for _, requestURL := range urls {
			fmt.Printf("GET %s \n", requestURL)
			req, err := http.NewRequest(http.MethodGet, requestURL, nil)
			Expect(err).ShouldNot(HaveOccurred())
			res, err := http.DefaultClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			body, err := io.ReadAll(res.Body)
			Expect(err).ShouldNot(HaveOccurred())
			if res.StatusCode != http.StatusOK {
				fmt.Printf("Response body: %s \n", body)
			}
			Expect(res.StatusCode).To(Equal(http.StatusOK))
		}

		// Ports in portchannels should return mode "portchannel" and the portchannel they are in
		req, err := http.NewRequest(http.MethodGet, restAPIBaseURL+"/devcloud/v4/port/details?switch_fqdn=clab-lacp-frontend-leaf1a&switch_port=5", nil)
		res, err := http.DefaultClient.Do(req)
		Expect(res.Body).NotTo(BeNil())
		Expect(res.StatusCode).To(Equal(http.StatusOK))
		body, err := io.ReadAll(res.Body)
		Expect(string(body)).To(ContainSubstring(`"mode":"portchannel"`))
		Expect(string(body)).To(ContainSubstring(`"port_channel":"51"`))

		// Portchannels set to "access" should contain the mode it is set to & vlan. (Note that this is leaf1b, it was 1a above).
		req, err = http.NewRequest(http.MethodGet, restAPIBaseURL+"/devcloud/v4/port/details?switch_fqdn=clab-lacp-frontend-leaf1b&port_channel=51", nil)
		res, err = http.DefaultClient.Do(req)
		Expect(res.StatusCode).To(Equal(http.StatusOK))
		body, err = io.ReadAll(res.Body)
		Expect(string(body)).To(ContainSubstring(`"mode":"access"`))
		Expect(string(body)).To(ContainSubstring(`"vlan_tag":100`))
		Expect(string(body)).To(ContainSubstring(`"untagged_vlan":100`))

		req, err = http.NewRequest(http.MethodGet, restAPIBaseURL+"/devcloud/v4/port/details?switch_fqdn=clab-lacp-frontend-leaf1b&port_channel=61", nil)
		res, err = http.DefaultClient.Do(req)
		Expect(res.StatusCode).To(Equal(http.StatusOK))
		body, err = io.ReadAll(res.Body)
		Expect(string(body)).To(ContainSubstring(`"mode":"access"`))
		Expect(string(body)).To(ContainSubstring(`"vlan_tag":100`))
		Expect(string(body)).To(ContainSubstring(`"untagged_vlan":100`))

		// Portchannels set to "trunk" should contain the mode it is set to & trunk groups.
		// We can GET the data for a spine-link portchannel, even though we can't modify it.
		req, err = http.NewRequest(http.MethodGet, restAPIBaseURL+"/devcloud/v4/port/details?switch_fqdn=clab-lacp-frontend-leaf1b&port_channel=271", nil)
		res, err = http.DefaultClient.Do(req)
		Expect(res.StatusCode).To(Equal(http.StatusOK))
		body, err = io.ReadAll(res.Body)
		Expect(string(body)).To(ContainSubstring(`"mode":"trunk"`))
		Expect(string(body)).To(ContainSubstring(`"trunk_groups":`))
		Expect(string(body)).To(ContainSubstring(`"native_vlan"`))

		//////////////////////
		// get the switches running-config after the changes - nothing should have changed.
		//////////////////////
		// get running config
		swClient1, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1a")
		Expect(err).ShouldNot(HaveOccurred())

		leaf1ActualConfig, err := swClient1.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// read the expected config file
		leaf1ExpectedConfig, err := os.ReadFile("./lacp/case1/expected/clab-lacp-frontend-leaf1a.txt")
		Expect(err).ShouldNot(HaveOccurred())

		// get clab-lacp-frontend-leaf2 running config
		swClient2, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		leaf2ActualConfig, err := swClient2.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// read the expected config file
		leaf2ExpectedConfig, err := os.ReadFile("./lacp/case1/expected/clab-lacp-frontend-leaf1b.txt")
		Expect(err).ShouldNot(HaveOccurred())
		// Expect(err).To(BeNil())
		// compare

		//////////////////////
		// compare the latest running-config with the expected one
		//////////////////////
		Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf1ExpectedConfig)), strings.TrimSpace(leaf1ActualConfig))).To(BeTrue())
		Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf2ExpectedConfig)), strings.TrimSpace(leaf2ActualConfig))).To(BeTrue())

		fmt.Printf("lacp case 1 completed \n")
	})

	It("test case 2 - Change port access vlan", Label("case2"), func() {
		fmt.Printf("lacp case 2 starts... \n")

		// Wait for the switchport status to be updated / read from the switch.
		for i := 0; i < 120; i++ {
			switchPortCR := &idcnetworkv1alpha1.SwitchPort{}
			key := types.NamespacedName{Name: "ethernet5.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
			Expect(err).ShouldNot(HaveOccurred())
			if switchPortCR.Status.LinkStatus == "connected" {
				break
			}
			time.Sleep(1 * time.Second)
		}
		fmt.Printf("Switchport status is up to date \n")

		makeSDNAPICallAndExpectOk(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/access", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "switch_port": "5",
		  "vlan_tag": 50
		}`)

		//////////////////////
		// get the switches running-config after the changes
		//////////////////////
		// get clab-lacp-frontend-leaf1b running config
		swClient1, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		leaf1ActualConfig, err := swClient1.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// read the expected config file
		leaf1ExpectedConfig, err := os.ReadFile("./lacp/case2/expected/clab-lacp-frontend-leaf1b.txt")
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// compare the latest running-config with the expected one
		//////////////////////
		Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf1ExpectedConfig)), strings.TrimSpace(leaf1ActualConfig))).To(BeTrue())
		//Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf2ExpectedConfig)), strings.TrimSpace(leaf2ActualConfig))).To(BeTrue())

		fmt.Printf("lacp case 2 completed \n")
	})

	It("test case 3 - Change port trunk vlan", Label("case3"), func() {
		fmt.Printf("lacp case 3 starts... \n")

		// Wait for the switchport status to be updated / read from the switch.
		for i := 0; i < 120; i++ {
			switchPortCR := &idcnetworkv1alpha1.SwitchPort{}
			key := types.NamespacedName{Name: "ethernet5.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
			Expect(err).ShouldNot(HaveOccurred())
			if switchPortCR.Status.LinkStatus == "connected" {
				break
			}
			time.Sleep(1 * time.Second)
		}
		fmt.Printf("Switchport status is up to date \n")

		makeSDNAPICallAndExpectOk(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/trunk", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "switch_port": "5",
		  "native_vlan": 58,
		  "trunk_group": "Some_Trunk_Group, Another_Group"
		}`)

		//////////////////////
		// get the switches running-config after the changes
		//////////////////////
		// get clab-lacp-frontend-leaf1b running config
		swClient1, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		leaf1ActualConfig, err := swClient1.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// read the expected config file
		leaf1ExpectedConfig, err := os.ReadFile("./lacp/case3/expected/clab-lacp-frontend-leaf1b.txt")
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// compare the latest running-config with the expected one
		//////////////////////
		Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf1ExpectedConfig)), strings.TrimSpace(leaf1ActualConfig))).To(BeTrue())

		fmt.Printf("lacp case 3 completed \n")
	})

	It("test case 4 - Check portchannel is autodiscovered from switch", Label("case4"), func() {
		fmt.Printf("lacp case 4 starts... \n")

		// Wait for the switchport status to be updated / read from the switch.
		objectExists := false
		portChannelCR := &idcnetworkv1alpha1.PortChannel{}
		for i := 0; i < 120 && !objectExists; i++ {
			key := types.NamespacedName{Name: "po61.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, portChannelCR)
			if err != nil {
				if k8sclient.IgnoreNotFound(err) != nil {
					Expect(err).ShouldNot(HaveOccurred())
				} else {
					// Object not found
					objectExists = false
					time.Sleep(1 * time.Second)
					continue
				}
			}
			// Wait for the portchannel status to be updated / read from the switch.
			if portChannelCR.Status.Name != "" {
				objectExists = true
			} else {
				time.Sleep(1 * time.Second)
			}
		}

		Expect(objectExists).To(BeTrue())
		Expect(portChannelCR.Status.VlanId).To(Equal(int64(100)))

		// CR for PortChannel271 (a spine link port-channel) should NOT be created
		key := types.NamespacedName{Name: "po271.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = testHelper.K8sClient.Get(ctx, key, portChannelCR)
		if err != nil {
			if k8sclient.IgnoreNotFound(err) != nil {
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				// Object not found
				objectExists = false
			}
		} else {
			objectExists = true
		}

		Expect(objectExists).To(BeFalse())

		// Delete the portchannel - deleting an autodiscovered CR (not through the API) should NOT remove the portchannel from the switch.
		testHelper.K8sClient.Delete(ctx, portChannelCR)

		// get running config
		swClient1b, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		leaf1bActualConfig, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// read the expected config file
		leaf1bExpectedConfig, err := os.ReadFile("./lacp/case4/expected/clab-lacp-frontend-leaf1b.txt")
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// compare the latest running-config with the expected one
		//////////////////////
		Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf1bExpectedConfig)), strings.TrimSpace(leaf1bActualConfig))).To(BeTrue())

		fmt.Printf("lacp case 4 completed \n")
	})

	It("test case 4b - Delete autodiscovered portchannel via API ", Label("case4b"), func() {
		fmt.Printf("lacp case 4b starts... \n")

		// Wait for the switchport status to be updated / read from the switch.
		objectExists := false
		portChannelCR := &idcnetworkv1alpha1.PortChannel{}
		for i := 0; i < 120 && !objectExists; i++ {
			key := types.NamespacedName{Name: "po61.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, portChannelCR)
			if err != nil {
				if k8sclient.IgnoreNotFound(err) != nil {
					Expect(err).ShouldNot(HaveOccurred())
				} else {
					// Object not found
					objectExists = false
					time.Sleep(1 * time.Second)
					continue
				}
			}
			// Wait for the portchannel status to be updated / read from the switch.
			if portChannelCR.Status.Name != "" {
				objectExists = true
			} else {
				time.Sleep(1 * time.Second)
			}
		}

		Expect(objectExists).To(BeTrue())
		Expect(portChannelCR.Status.VlanId).To(Equal(int64(100)))

		// Deleting autodiscovered portchannel via API
		makeSDNAPICallAndExpectOk(http.MethodDelete, restAPIBaseURL+"/devcloud/v4/delete/portchannel", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "port_channel": 61
		}`)

		// get running config
		swClient1b, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		leaf1bActualConfig, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// read the expected config file
		leaf1bExpectedConfig, err := os.ReadFile("./lacp/case4b/expected/clab-lacp-frontend-leaf1b.txt")
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// compare the latest running-config with the expected one
		//////////////////////
		Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf1bExpectedConfig)), strings.TrimSpace(leaf1bActualConfig))).To(BeTrue())

		fmt.Printf("lacp case 4b completed \n")
	})

	It("test case 5 - Create new portchannel, configure and add 2 switchports, then delete portchannel", Label("case5"), func() {
		fmt.Printf("lacp case 5 starts... \n")

		// Create portchannel
		makeSDNAPICallAndExpectStatusCode(http.MethodPut, restAPIBaseURL+"/devcloud/v4/create/portchannel", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "port_channel": 88
		}`, 201)

		// Create SAME portchannel AGAIN (should give 200 status code this time)
		makeSDNAPICallAndExpectStatusCode(http.MethodPut, restAPIBaseURL+"/devcloud/v4/create/portchannel", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "port_channel": 88
		}`, 200)

		// Move 2 existing switchports between leaf1b and server1-2 to the new portchannel
		makeSDNAPICallAndExpectOk(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/portchannel", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "switch_port": "6",
		  "port_channel": 88
		}`)
		makeSDNAPICallAndExpectOk(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/portchannel", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "switch_port": "7",
		  "port_channel": 88
		}`)

		// Configure the new portchannel
		makeSDNAPICallAndExpectOk(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/portchannel/access", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "port_channel": 88,
		  "vlan_tag": 98
		}`)

		// get running config
		swClient1b, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		leaf1bActualConfig, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// read the expected config file
		leaf1bExpectedConfig, err := os.ReadFile("./lacp/case5/1-pchan-created/clab-lacp-frontend-leaf1b.txt")
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// compare the latest running-config with the expected one
		//////////////////////
		Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf1bExpectedConfig)), strings.TrimSpace(leaf1bActualConfig))).To(BeTrue())

		// Delete the portchannel
		makeSDNAPICallAndExpectOk(http.MethodDelete, restAPIBaseURL+"/devcloud/v4/delete/portchannel", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "port_channel": 88
		}`)

		// get running config
		swClient1b, err = testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		leaf1bActualConfig, err = swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// read the expected config file
		leaf1bExpectedConfig, err = os.ReadFile("./lacp/case5/2-pchan-deleted/clab-lacp-frontend-leaf1b.txt")
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// compare the latest running-config with the expected one
		//////////////////////
		Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf1bExpectedConfig)), strings.TrimSpace(leaf1bActualConfig))).To(BeTrue())

		// Other switch should not have changed
		swClient1a, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1a")
		Expect(err).ShouldNot(HaveOccurred())

		leaf1aActualConfig, err := swClient1a.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// read the expected config file
		leaf1aExpectedConfig, err := os.ReadFile("./lacp/case5/2-pchan-deleted/clab-lacp-frontend-leaf1a.txt")
		Expect(err).ShouldNot(HaveOccurred())

		Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf1aExpectedConfig)), strings.TrimSpace(leaf1aActualConfig))).To(BeTrue())

		fmt.Printf("lacp case 5 completed \n")
	})

	It("test case 6 - Calls should fail Validation", Label("case6"), func() {
		fmt.Printf("lacp case 6 starts... \n")

		// Try to change VLAN to a tenant value
		makeSDNAPICallAndExpectStatusCode(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/access", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "switch_port": 5,
          "vlan_tag": 105
		}`, http.StatusBadRequest)

		// Try to connect to an unknown switch
		makeSDNAPICallAndExpectStatusCode(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/access", `{
		  "switch_fqdn": "not-a-real-switch",
		  "switch_port": "6",
          "vlan_tag": 100
		}`, http.StatusBadRequest)
		makeSDNAPICallAndExpectStatusCode(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/access", `{
		  "switch_fqdn": "clab-not-a-real-switch",
		  "switch_port": "6",
          "vlan_tag": 100
		}`, http.StatusNotFound)

		// Try to change an unknown port (a spine-link port that exists, but which there is no switchport CR for)
		makeSDNAPICallAndExpectErrorCode(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/access", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "switch_port": "27",
		  "vlan_tag": 100
		}`)

		// Try to delete a spine-link port-channel
		makeSDNAPICallAndExpectErrorCode(http.MethodPut, restAPIBaseURL+"/devcloud/v4/delete/portchannel", `{
		  "switch_fqdn": "clab-lacp-frontend-leaf1b",
		  "port_channel": 271
		}`)

		// TODO: Should not be able to CREATE over the top of a spine-link-port-channel or change its config

		// get running config
		swClient1b, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		leaf1bActualConfig, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// read the expected config file
		leaf1bExpectedConfig, err := os.ReadFile("./lacp/case6/expected/clab-lacp-frontend-leaf1b.txt")
		Expect(err).ShouldNot(HaveOccurred())

		//////////////////////
		// compare the latest running-config with the expected one
		//////////////////////
		Expect(testHelper.CompareConfigs(strings.TrimSpace(string(leaf1bExpectedConfig)), strings.TrimSpace(leaf1bActualConfig))).To(BeTrue())

		fmt.Printf("lacp case 6 completed \n")
	})

	It("test case 7 - list/ports returns data", Label("case7"), func() {
		fmt.Printf("lacp case 7 starts... \n")

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		req, err := http.NewRequest(http.MethodGet, restAPIBaseURL+"/devcloud/v4/list/ports?switch_fqdn=clab-lacp-frontend-leaf1a", nil)
		Expect(err).ShouldNot(HaveOccurred())
		res, err := http.DefaultClient.Do(req)
		Expect(err).ShouldNot(HaveOccurred())
		body, err := io.ReadAll(res.Body)
		Expect(err).ShouldNot(HaveOccurred())

		type ListPortsResponse struct {
			PortList []switchclients.ResPortInfo `json:"port_list"`
		}
		respParsed := ListPortsResponse{}
		err = json.Unmarshal(body, &respParsed)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(len(respParsed.PortList)).To(BeNumerically(">", 5))

		// Check Ethernet5, a port in a portchannel
		found := false
		for _, portInfo := range respParsed.PortList {
			if portInfo.Port == "5" {
				found = true
				Expect(portInfo.PortChannel).To(Equal("51"))
				Expect(portInfo.Mode).To(Equal("portchannel"))
				Expect(portInfo.InterfaceName).To(Equal("Ethernet5"))
				Expect(portInfo.Description).To(Equal("server1-1_8"))
				Expect(portInfo.LinkStatus).To(Equal("connected"))
				Expect(portInfo.VlanId).To(Equal(0))
				Expect(portInfo.NativeVlan).To(Equal(0))
				Expect(portInfo.UntaggedVlan).To(Equal(0))
			}
		}
		Expect(found).To(BeTrue(), "Ethernet5 not found in response")

		// Check PortChannel51, a port-channel in access mode
		found = false
		for _, portInfo := range respParsed.PortList {
			if portInfo.Port == "Port-Channel51" {
				found = true
				Expect(portInfo.PortChannel).To(Equal(""))
				Expect(portInfo.Mode).To(Equal("access"))
				Expect(portInfo.VlanId).To(Equal(100))
				Expect(portInfo.UntaggedVlan).To(Equal(100))
				Expect(portInfo.Description).To(Equal("server1-1"))
				Expect(portInfo.LinkStatus).To(Equal("connected"))
			}
		}
		Expect(found).To(BeTrue(), "Port-Channel51 not found in response")

		fmt.Printf("lacp case 7 completed \n")
	})

	It("test case 8 - list/vlans returns data", Label("case8"), func() {
		fmt.Printf("lacp case 8 starts... \n")

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		req, err := http.NewRequest(http.MethodGet, restAPIBaseURL+"/devcloud/v4/list/vlans?switch_fqdn=clab-lacp-frontend-leaf1a", nil)
		Expect(err).ShouldNot(HaveOccurred())
		res, err := http.DefaultClient.Do(req)
		Expect(err).ShouldNot(HaveOccurred())
		body, err := io.ReadAll(res.Body)
		Expect(err).ShouldNot(HaveOccurred())

		type ListVlansResponse struct {
			Vlans []switchclients.VlanWithTrunkGroups `json:"vlans"`
		}
		respParsed := ListVlansResponse{}
		err = json.Unmarshal(body, &respParsed)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(len(respParsed.Vlans)).To(BeNumerically(">", 30))

		// Check a vlan from the switch config is in the list
		found := false
		for _, vlanInfo := range respParsed.Vlans {
			if vlanInfo.VlanId == 50 {
				found = true
				Expect(vlanInfo.Name).To(Equal("100.64.16.0/26_P_Harvester_dev"))
				Expect(vlanInfo.Status).To(Equal("active"))
				Expect(vlanInfo.InterfaceNames).To(Equal([]string{"Cpu", "Port-Channel271", "Vxlan1"}))
				Expect(vlanInfo.TrunkGroups).To(Equal([]string{"Dev_Nets", "ISC"}))
			}
		}
		Expect(found).To(BeTrue(), "Vlan 50 not found in response")

		fmt.Printf("lacp case 8 completed \n")
	})

	It("test case 9 - Set fields on port to -1 or empty via API port/unmaintained", Label("case9"), func() {
		fmt.Printf("lacp case 9 starts... \n")

		// Get the sw client
		swClient1b, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		// Wait for the switchport status to be updated or read from the switch.
		switchPortCR := &idcnetworkv1alpha1.SwitchPort{}
		for i := 0; i < 120; i++ {
			// Get the switchport
			key := types.NamespacedName{Name: "ethernet5.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
			Expect(err).ShouldNot(HaveOccurred())
			if switchPortCR.Status.LinkStatus == "connected" {
				break
			}
			time.Sleep(1 * time.Second)
		}

		// Set and update switchport with new values
		switchPortCR.Spec.Mode = "access"
		switchPortCR.Spec.NativeVlan = 10
		switchPortCR.Spec.VlanId = 55
		switchPortCR.Spec.PortChannel = 0 // Must remove from portChannel for VlanId to get updated.
		switchPortCR.Spec.Description = "Test access"
		err = testHelper.K8sClient.Update(ctx, switchPortCR)
		Expect(err).ShouldNot(HaveOccurred())

		// Wait for changes to be applied
		startTime := time.Now()
		for time.Now().Before(startTime.Add(60 * time.Second)) {
			key := types.NamespacedName{Name: "ethernet5.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err := testHelper.K8sClient.Get(ctx, key, switchPortCR)
			Expect(err).ShouldNot(HaveOccurred())
			if switchPortCR.Status.Mode == switchPortCR.Spec.Mode && switchPortCR.Status.NativeVlan == switchPortCR.Spec.NativeVlan && switchPortCR.Status.PortChannel == switchPortCR.Spec.PortChannel && switchPortCR.Status.VlanId == switchPortCR.Spec.VlanId {
				break
			}
			time.Sleep(1 * time.Second)
		}
		Expect(switchPortCR.Status.Mode).To(Equal("access"))
		Expect(switchPortCR.Status.NativeVlan).To(Equal(int64(10)))
		Expect(switchPortCR.Status.VlanId).To(Equal(int64(55)))
		Expect(switchPortCR.Status.Description).To(Equal("Test access"))

		// Get the initial switch config
		leaf1bConfigAtStart, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// Making call to rest api for port unmaintained
		makeSDNAPICallAndExpectOk(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/unmaintained", `{
                   "switch_fqdn": "clab-lacp-frontend-leaf1b",
                   "switch_port": "5"
		}`)

		key := types.NamespacedName{Name: "ethernet5.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
		Expect(err).ShouldNot(HaveOccurred())

		// Check for expected values
		Expect(switchPortCR.Spec.Mode).To(Equal(""))
		Expect(switchPortCR.Spec.VlanId).To(Equal(int64(-1)))
		Expect(switchPortCR.Spec.TrunkGroups).To(BeNil())
		Expect(switchPortCR.Spec.NativeVlan).To(Equal(int64(-1)))
		Expect(switchPortCR.Spec.PortChannel).To(Equal(int64(-1)))
		Expect(switchPortCR.Spec.Description).To(Equal(""))

		// Get the final switch config
		leaf1bConfigAtEnd, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// Get the diff of intial & final switch configs
		leaf1bDiff, err := testutils.Diff(leaf1bConfigAtStart, leaf1bConfigAtEnd, 2)
		Expect(err).ShouldNot(HaveOccurred())

		// Expect intial & final switch config to be the same
		Expect(leaf1bDiff).Should(Equal(""))

		fmt.Printf("lacp case 9 completed \n")
	})

	It("test case 10 - Set fields on portChannel to -1 or empty via API portchannel/unmaintained", Label("case10"), func() {
		fmt.Printf("lacp case 10 starts... \n")

		// Get the sw client
		swClient1b, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		objectExists := false
		portChannelCR := &idcnetworkv1alpha1.PortChannel{}

		// Wait for the portChannel status to be updated or read from the switch.
		for i := 0; i < 120 && !objectExists; i++ {
			key := types.NamespacedName{Name: "po61.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, portChannelCR)
			if err != nil {
				if k8sclient.IgnoreNotFound(err) != nil {
					Expect(err).ShouldNot(HaveOccurred())
				} else {
					objectExists = false
					time.Sleep(1 * time.Second)
					continue
				}
			}
			if portChannelCR.Status.Name != "" {
				objectExists = true
			} else {
				time.Sleep(1 * time.Second)
			}
		}
		Expect(objectExists).To(BeTrue())

		// Set and update portchannel with new values
		portChannelCR.Spec.Mode = "access"
		portChannelCR.Spec.VlanId = 95
		portChannelCR.Spec.Description = "Test access"
		err = testHelper.K8sClient.Update(ctx, portChannelCR)
		Expect(err).ShouldNot(HaveOccurred())

		// Wait for changes to be applied
		startTime := time.Now()
		for time.Now().Before(startTime.Add(60 * time.Second)) {
			key := types.NamespacedName{Name: "po61.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err := testHelper.K8sClient.Get(ctx, key, portChannelCR)
			Expect(err).ShouldNot(HaveOccurred())
			if portChannelCR.Status.Mode == portChannelCR.Spec.Mode && portChannelCR.Status.VlanId == portChannelCR.Spec.VlanId {
				break
			}
			time.Sleep(1 * time.Second)
		}

		// Expect status to be updated as per spec
		Expect(portChannelCR.Status.Mode).To(Equal("access"))
		Expect(portChannelCR.Status.VlanId).To(Equal(int64(95)))
		Expect(portChannelCR.Status.Description).To(Equal("Test access"))

		// Get the initial switch config
		leaf1bConfigAtStart, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// Making call to rest api for unmaintained portChannel
		makeSDNAPICallAndExpectOk(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/portchannel/unmaintained", `{
                    "switch_fqdn": "clab-lacp-frontend-leaf1b",
                    "port_channel": 61
		}`)

		key := types.NamespacedName{Name: "po61.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = testHelper.K8sClient.Get(ctx, key, portChannelCR)
		Expect(err).ShouldNot(HaveOccurred())

		// Check for expected values
		Expect(portChannelCR.Spec.Mode).To(Equal(""))
		Expect(portChannelCR.Spec.VlanId).To(Equal(int64(-1)))
		Expect(portChannelCR.Spec.TrunkGroups).To(BeNil())
		Expect(portChannelCR.Spec.NativeVlan).To(Equal(int64(-1)))
		Expect(portChannelCR.Spec.Description).To(Equal(""))

		// Get the final switch config
		leaf1bConfigAtEnd, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// Get the diff of intial & final switch config
		leaf1bDiff, err := testutils.Diff(leaf1bConfigAtStart, leaf1bConfigAtEnd, 2)
		Expect(err).ShouldNot(HaveOccurred())

		// Expect intial & final switch config to be the same
		Expect(leaf1bDiff).Should(Equal(""))

		fmt.Printf("lacp case 10 completed \n")
	})

	It("test case 11 - Test to update description field on port through Rest API", Label("case11"), func() {
		fmt.Printf("lacp case 11 starts... \n")

		// Get the sw client
		swClient1b, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		// Wait for the switchport status to be updated or read from the switch.
		switchPortCR := &idcnetworkv1alpha1.SwitchPort{}
		for i := 0; i < 120; i++ {
			// Get the switchport
			key := types.NamespacedName{Name: "ethernet5.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
			Expect(err).ShouldNot(HaveOccurred())
			if switchPortCR.Status.LinkStatus == "connected" {
				break
			}
			time.Sleep(1 * time.Second)
		}

		// Get the initial switch config
		leaf1bConfigAtStart, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// Making call to rest api to update only description
		makeSDNAPICallAndExpectOk(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/access", `{
                   "switch_fqdn": "clab-lacp-frontend-leaf1b",
                   "switch_port": "5",
                   "vlan_tag": 95,
		   "description": "updated test access"
                }`)

		key := types.NamespacedName{Name: "ethernet5.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
		Expect(err).ShouldNot(HaveOccurred())

		// Check for expected values
		Expect(switchPortCR.Spec.Mode).To(Equal("access"))
		Expect(switchPortCR.Spec.Description).To(Equal("updated test access"))
		Expect(switchPortCR.Status.Mode).To(Equal("access"))
		Expect(switchPortCR.Status.Description).To(Equal("updated test access"))

		// Get the final switch config
		leaf1bConfigAtEnd, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// Get the diff of intial & final switch configs
		leaf1bDiff, err := testutils.Diff(leaf1bConfigAtStart, leaf1bConfigAtEnd, 2)
		Expect(err).ShouldNot(HaveOccurred())

		leaf1bDiffExpected := `@@ @@
 !
 interface Ethernet5
-   description server1-1_9
-   switchport access vlan 4008
+   description updated test access
+   switchport access vlan 95
    switchport
    no snmp trap link-change
-   channel-group 51 mode active
    lacp timer fast
+   no lldp transmit
    spanning-tree portfast
    spanning-tree bpduguard enable
`
		// Expect intial & final switch config to be the different
		Expect(leaf1bDiff).Should(Equal(leaf1bDiffExpected), "Actual leaf1bDiff: %s  Expected: %s", leaf1bDiff, leaf1bDiffExpected)

		fmt.Printf("lacp case 11 completed \n")
	})

	It("test case 12- Test to update description field on portChannel through Rest API", Label("case12"), func() {
		fmt.Printf("lacp case 12 starts... \n")

		// Get the sw client
		swClient1b, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		objectExists := false
		portChannelCR := &idcnetworkv1alpha1.PortChannel{}

		// Wait for the portChannel status to be updated or read from the switch.
		for i := 0; i < 120 && !objectExists; i++ {
			key := types.NamespacedName{Name: "po61.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, portChannelCR)
			if err != nil {
				if k8sclient.IgnoreNotFound(err) != nil {
					Expect(err).ShouldNot(HaveOccurred())
				} else {
					objectExists = false
					time.Sleep(1 * time.Second)
					continue
				}
			}
			if portChannelCR.Status.Name != "" {
				objectExists = true
			} else {
				time.Sleep(1 * time.Second)
			}
		}
		Expect(objectExists).To(BeTrue())

		// Get the initial switch config
		leaf1bConfigAtStart, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// Making call to rest api to update just decription
		makeSDNAPICallAndExpectOk(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/portchannel/trunk", `{
                    "switch_fqdn": "clab-lacp-frontend-leaf1b",
                    "port_channel": 61,
		    "description" : "updated test trunk",
		    "native_vlan": 1
                }`)

		key := types.NamespacedName{Name: "po61.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = testHelper.K8sClient.Get(ctx, key, portChannelCR)
		Expect(err).ShouldNot(HaveOccurred())

		// Check for expected values
		Expect(portChannelCR.Spec.Mode).To(Equal("trunk"))
		Expect(portChannelCR.Spec.Description).To(Equal("updated test trunk"))
		Expect(portChannelCR.Status.Mode).To(Equal("trunk"))
		Expect(portChannelCR.Status.Description).To(Equal("updated test trunk"))

		// Get the final switch config
		leaf1bConfigAtEnd, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// Get the diff of intial & final switch config
		leaf1bDiff, err := testutils.Diff(leaf1bConfigAtStart, leaf1bConfigAtEnd, 2)
		Expect(err).ShouldNot(HaveOccurred())

		leaf1bDiffExpected := `@@ @@
 !
 interface Port-Channel61
-   description server1-2
-   switchport access vlan 100
+   description updated test trunk
+   switchport mode trunk
    switchport
    port-channel lacp fallback individual
`
		// Expect intial & final switch config to be the different
		Expect(leaf1bDiff).Should(Equal(leaf1bDiffExpected), "Actual leaf1bDiff: %s  Expected: %s", leaf1bDiff, leaf1bDiffExpected)
		fmt.Printf("lacp case 12 completed \n")
	})

	It("test case 13- Test to update description field on port which is in portChannel through Rest API", Label("case13"), func() {
		fmt.Printf("lacp case 13 starts... \n")

		// Get the sw client
		swClient1b, err := testHelper.ContainerLabManager.GetSwitchClient(topology.Name, "clab-lacp-frontend-leaf1b")
		Expect(err).ShouldNot(HaveOccurred())

		// Wait for the switchport status to be updated or read from the switch.
		switchPortCR := &idcnetworkv1alpha1.SwitchPort{}
		for i := 0; i < 120; i++ {
			// Get the switchport
			key := types.NamespacedName{Name: "ethernet5.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
			Expect(err).ShouldNot(HaveOccurred())
			if switchPortCR.Status.LinkStatus == "connected" {
				break
			}
			time.Sleep(1 * time.Second)
		}

		// Get the initial switch config
		leaf1bConfigAtStart, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// Making call to rest api to update just decription on port
		makeSDNAPICallAndExpectOk(http.MethodPut, restAPIBaseURL+"/devcloud/v4/configure/port/portchannel", `{
			"switch_fqdn": "clab-lacp-frontend-leaf1b",
			"port_channel": 61,
			"switch_port": "5",
			"description" : "updated description"
                }`)

		key := types.NamespacedName{Name: "ethernet5.clab-lacp-frontend-leaf1b", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = testHelper.K8sClient.Get(ctx, key, switchPortCR)
		Expect(err).ShouldNot(HaveOccurred())

		// Check for expected values
		Expect(switchPortCR.Spec.Description).To(Equal("updated description"))
		Expect(switchPortCR.Status.Description).To(Equal("updated description"))

		// Get the final switch config
		leaf1bConfigAtEnd, err := swClient1b.GetRunningConfig(ctx)
		Expect(err).ShouldNot(HaveOccurred())

		// Get the diff of intial & final switch config
		leaf1bDiff, err := testutils.Diff(leaf1bConfigAtStart, leaf1bConfigAtEnd, 2)
		Expect(err).ShouldNot(HaveOccurred())

		leaf1bDiffExpected := `@@ @@
 !
 interface Ethernet5
-   description server1-1_9
+   description updated description
    switchport access vlan 4008
    switchport
    no snmp trap link-change
-   channel-group 51 mode active
+   channel-group 61 mode active
    lacp timer fast
    spanning-tree portfast
`
		// Expect intial & final switch config to be the different
		Expect(leaf1bDiff).Should(Equal(leaf1bDiffExpected), "Actual leaf1bDiff: %s  Expected: %s", leaf1bDiff, leaf1bDiffExpected)
		fmt.Printf("lacp case 13 completed \n")
	})
})

func makeSDNAPICallAndExpectOk(method string, requestURL string, requestBody string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	fmt.Printf("%s %s \n", method, requestURL)
	req, err := http.NewRequest(method, requestURL, strings.NewReader(requestBody))
	Expect(err).ShouldNot(HaveOccurred())
	res, err := http.DefaultClient.Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	body, err := io.ReadAll(res.Body)
	Expect(err).ShouldNot(HaveOccurred())
	if res.StatusCode != http.StatusOK {
		fmt.Printf("Response body: %s \n", body)
	}
	Expect(res.StatusCode).To(Equal(http.StatusOK))
}

func makeSDNAPICallAndExpectStatusCode(method string, requestURL string, requestBody string, expectStatusCode int) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	fmt.Printf("%s %s \n", method, requestURL)
	req, err := http.NewRequest(method, requestURL, strings.NewReader(requestBody))
	Expect(err).ShouldNot(HaveOccurred())
	res, err := http.DefaultClient.Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	body, err := io.ReadAll(res.Body)
	Expect(err).ShouldNot(HaveOccurred())
	if res.StatusCode != expectStatusCode {
		fmt.Printf("Response body: %s \n", body)
	}
	Expect(res.StatusCode).To(Equal(expectStatusCode))
}

func makeSDNAPICallAndExpectErrorCode(method string, requestURL string, requestBody string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	fmt.Printf("%s %s \n", method, requestURL)
	req, err := http.NewRequest(method, requestURL, strings.NewReader(requestBody))
	Expect(err).ShouldNot(HaveOccurred())
	res, err := http.DefaultClient.Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	body, err := io.ReadAll(res.Body)
	Expect(err).ShouldNot(HaveOccurred())
	is400Error := res.StatusCode >= 400 && res.StatusCode <= 499
	if !is400Error {
		fmt.Printf("Status Code: %d \n Response body: %s \n", res.StatusCode, body)
	}
	Expect(is400Error).To(BeTrue())
}
