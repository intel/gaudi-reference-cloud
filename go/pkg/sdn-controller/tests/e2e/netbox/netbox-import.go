//go:build ginkgo_only

package e2e

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/helper"
	testutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/test_utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	eAPISecretPath     = "/vault/secrets/eapi"
	configTemplatePath = "netbox/local-mock-netbox.yaml.gotmpl"
)

var _ = Describe("netbox Tests", Ordered, Label("sdn", "tsdn", "netbox"), func() {
	testHelper := helper.New(helper.Config{EAPISecretDir: eAPISecretPath})

	BeforeAll(func() {

	})

	AfterAll(func() {

	})

	BeforeEach(func() {
		err := testHelper.DeleteAllK8sResources()
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
	})

	/*
		Purpose of the Test: Test connectivity / parsing of switches from Netbox

		Test Steps:
			- Bring up mock netbox server that serves up a static list of switches
			- Start SDN Controller, set to import from the mock netbox
			- Wait for Netbox_Controller to import the switches and create Switch CRs.

		Test Coverage:
			 - Netbox controller / netbox client & switch creation
	*/
	It("test case 1 - import switches from netbox", func() {
		fmt.Printf("netbox-import case 1 starts... \n")
		ctx := context.Background()

		mux := http.NewServeMux()
		mux.HandleFunc("/api/dcim/devices/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			bytes, err := os.ReadFile("netbox/mock-devices-list-response.json")
			Expect(err).ShouldNot(HaveOccurred())
			w.Write(bytes)
		})
		mux.HandleFunc("/api/status/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			bytes, err := os.ReadFile("netbox/mock-api-status-responsev3.json")
			Expect(err).ShouldNot(HaveOccurred())
			w.Write(bytes)
		})

		mockNetboxServer := httptest.NewUnstartedServer(mux)
		lstnr, err := net.Listen("tcp", "0.0.0.0:0") // Listen on all interfaces, random unused port
		Expect(err).ShouldNot(HaveOccurred())
		mockNetboxServer.Listener = lstnr
		mockServerPort := lstnr.Addr().(*net.TCPAddr).Port
		mockNetboxServer.Start()
		defer mockNetboxServer.Close()

		//Point the SDN controller (running inside Kind) to the mockNetboxServer (running as part of this test on the host).
		err, mockServerIP := testutils.GetOutboundIP()
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(mockNetboxServer.URL)
		fmt.Println(mockServerIP)
		fmt.Println(mockServerPort)
		configYaml := fmt.Sprintf(`
tls:
  client:
    rootCa: kind-singlecluster-root-ca

regions:
  us-dev-1:
    availabilityZones:
      us-dev-1a:
        sdnController:
          managerConfig:
            controllerManagerConfigYaml:
              controllerConfig:
                switchImportSource: netbox
                netboxServer: "%s:%d" # TODO: Dynamically set based on mockServer
                netboxProtocol: "http"
                netboxSwitchFQDNDomainName: "fakeinternal-placeholder.com"
`, mockServerIP, mockServerPort)

		err = testHelper.DeploySDNWithConfig(configYaml, false)
		Expect(err).ShouldNot(HaveOccurred())

		// Check switch has been created
		found := false
		for i := 0; i < 10; i++ {
			switchCR := &idcnetworkv1alpha1.Switch{}
			key := types.NamespacedName{Name: "fxhb3p3r-zal0118a.fakeinternal-placeholder.com", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err = testHelper.K8sClient.Get(ctx, key, switchCR)
			if err != nil {
				time.Sleep(time.Second)
				continue
			} else {
				found = true
				Expect(switchCR.Spec.FQDN).To(Equal("fxhb3p3r-zal0118a.fakeinternal-placeholder.com"))
				Expect(int(switchCR.Spec.BGP.BGPCommunity)).To(Equal(-1))
			}
		}
		if !found {
			Fail(fmt.Sprintf("Timed out waiting for switch CR from mock Netbox to be created, %v", err))
		}

		fmt.Printf("netbox-import case 1 completed \n")
	})

	/*
		Purpose of the Test: Test that SDN doesn't crash if it can't talk to Netbox.

		Test Steps:
			- Start SDN Controller, set to import from an invalid netbox URL
			- Check that the SDN pod hasn't crashed / restarted

		Test Coverage:
			 - Error handling in netbox controller / client
	*/
	It("test case 2 - incorrect netbox url handled correctly (should not cause crash)", func() {
		fmt.Printf("netbox-import case 2 starts... \n")
		ctx := context.Background()

		//Point the SDN controller to the mockNetboxServer.
		err := testHelper.DeploySDNWithConfigFile("netbox/invalid-netbox.yaml.gotmpl", false)
		Expect(err).ShouldNot(HaveOccurred())

		time.Sleep(30 * time.Second)
		// Check still running.
		err, sdnpod := testHelper.GetPod(ctx, "sdn-controller")
		Expect(string(sdnpod.Status.Phase)).To(Equal("Running"))

		// Check SDN controller has not restarted (Crashed) after 30 seconds
		err, restarts := testHelper.GetContainerRestarts(ctx, "sdn-controller", "sdn-controller")
		Expect(restarts).To(Equal(int32(0)))

		fmt.Printf("netbox-import case 2 completed \n")
	})

	testNetboxListDevicesForDiffVersions := func(deviceFile string, versionFile string, caseName string) {
		fmt.Printf("netbox-import case %s starts... \n", caseName)
		ctx := context.Background()

		mux := http.NewServeMux()
		mux.HandleFunc("/api/dcim/devices/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			bytes, err := os.ReadFile(deviceFile)
			Expect(err).ShouldNot(HaveOccurred())
			w.Write(bytes)
		})
		mux.HandleFunc("/api/status/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			bytes, err := os.ReadFile(versionFile)
			Expect(err).ShouldNot(HaveOccurred())
			w.Write(bytes)
		})

		mockNetboxServer := httptest.NewUnstartedServer(mux)
		lstnr, err := net.Listen("tcp", "0.0.0.0:0") // Listen on all interfaces, random unused port
		Expect(err).ShouldNot(HaveOccurred())
		mockNetboxServer.Listener = lstnr
		mockServerPort := lstnr.Addr().(*net.TCPAddr).Port
		mockNetboxServer.Start()
		defer mockNetboxServer.Close()

		//Point the SDN controller (running inside Kind) to the mockNetboxServer (running as part of this test on the host).
		err, mockServerIP := testutils.GetOutboundIP()
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(mockNetboxServer.URL)
		fmt.Println(mockServerIP)
		fmt.Println(mockServerPort)
		configYaml := fmt.Sprintf(`
tls:
  client:
    rootCa: kind-singlecluster-root-ca

regions:
  us-dev-1:
    availabilityZones:
      us-dev-1a:
        sdnController:
          managerConfig:
            controllerManagerConfigYaml:
              controllerConfig:
                dataCenter: fxhb3p3r:fxhb3p3s:clab:phx02
                switchImportSource: netbox
                netboxServer: "%s:%d" # TODO: Dynamically set based on mockServer
                netboxProtocol: "http"
                netboxSwitchFQDNDomainName: "us-dev-1.fakecloud.intel.com"
`, mockServerIP, mockServerPort)

		err = testHelper.DeploySDNWithConfig(configYaml, false)
		Expect(err).ShouldNot(HaveOccurred())

		switchCRsList := &idcnetworkv1alpha1.SwitchList{}

		for i := 0; i < 10; i++ {
			listOpts := &client.ListOptions{
				Namespace: idcnetworkv1alpha1.SDNControllerNamespace,
			}
			err := testHelper.K8sClient.List(ctx, switchCRsList, listOpts)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
		}

		switchCRsExpectedToBeCreated := map[string]struct{}{
			"phx02-a01-acsw001.us-dev-1.fakecloud.intel.com": struct{}{},
			"phx02-a01-acsw002.us-dev-1.fakecloud.intel.com": struct{}{},
			"phx02-c01-acsw001.us-dev-1.fakecloud.intel.com": struct{}{},
		}
		switchCRsActualCreated := map[string]struct{}{}
		for _, sw := range switchCRsList.Items {
			switchCRsActualCreated[sw.Name] = struct{}{}
		}

		// check if the length are the same
		Expect(len(switchCRsExpectedToBeCreated) == len(switchCRsActualCreated)).Should(BeTrue())

		// check if the expected switches have been created
		for sw := range switchCRsExpectedToBeCreated {
			_, found := switchCRsActualCreated[sw]
			Expect(found).Should(BeTrue())
		}

		fmt.Printf("netbox-import case %v completed \n", caseName)
	}

	/*
		Purpose of the Test: Test importing switches from Netbox for both v3 and v4 Netbox server data format and Netbox clients.

		Test Steps:
			- Make Netbox ListDevices API return v3 or v4 format data based on the test cases.
			- Start SDN Controller.
			- Check if the Switch CRs are created correctly.

		Test Coverage:
			 - v3 and v4 Netbox devices data parsing in Netbox client
	*/
	It("test case 3 - import switches from netbox with v3 version of data format", Label("sdn", "netbox", "switches"), func() {
		testNetboxListDevicesForDiffVersions("netbox/list_devices_v3.json", "netbox/mock-api-status-responsev3.json", "3")
	})

	It("test case 4 - import switches from netbox with v4 version of data format", Label("sdn", "netbox", "switches"), func() {
		testNetboxListDevicesForDiffVersions("netbox/list_devices_v4.json", "netbox/mock-api-status-responsev4.json", "4")
	})

	testNetboxGetProviderInterfacesForDiffVersions := func(deviceFile string, interfacesFile string, versionFile string, caseName string) {
		fmt.Printf("netbox-import case %s starts... \n", caseName)
		ctx := context.Background()

		mux := http.NewServeMux()
		mux.HandleFunc("/api/dcim/devices/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			bytes, err := os.ReadFile(deviceFile)
			Expect(err).ShouldNot(HaveOccurred())
			w.Write(bytes)
		})
		mux.HandleFunc("/api/dcim/interfaces/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			bytes, err := os.ReadFile(interfacesFile)
			Expect(err).ShouldNot(HaveOccurred())
			w.Write(bytes)
		})
		mux.HandleFunc("/api/status/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			bytes, err := os.ReadFile(versionFile)
			Expect(err).ShouldNot(HaveOccurred())
			w.Write(bytes)
		})

		mockNetboxServer := httptest.NewUnstartedServer(mux)
		lstnr, err := net.Listen("tcp", "0.0.0.0:0") // Listen on all interfaces, random unused port
		Expect(err).ShouldNot(HaveOccurred())
		mockNetboxServer.Listener = lstnr
		mockServerPort := lstnr.Addr().(*net.TCPAddr).Port
		mockNetboxServer.Start()
		defer mockNetboxServer.Close()

		//Point the SDN controller (running inside Kind) to the mockNetboxServer (running as part of this test on the host).
		err, mockServerIP := testutils.GetOutboundIP()
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(mockNetboxServer.URL)
		fmt.Println(mockServerIP)
		fmt.Println(mockServerPort)
		configYaml := fmt.Sprintf(`
tls:
  client:
    rootCa: kind-singlecluster-root-ca

regions:
  us-dev-1:
    availabilityZones:
      us-dev-1a:
        sdnController:
          managerConfig:
            controllerManagerConfigYaml:
              controllerConfig:
                dataCenter: fxhb3p3r:fxhb3p3s:clab:phx02
                switchImportSource: netbox
                switchPortImportSource: netbox
                netboxServer: "%s:%d" # TODO: Dynamically set based on mockServer
                netboxProtocol: "http"
                netboxSwitchFQDNDomainName: "us-dev-1.fakecloud.intel.com"
          netbox:
            providerServersFilter:
              site:
                - phx02
              role:
                - ikvm
                - akvm
                - rkvm
                - swds
                - hvst
                - spar
            providerInterfacesFilter:
              interfaceName:
                - net0/0
                - net0/1
                - net1/0
                - net1/1
                - net2/0
                - net2/1
                - ocp0/0
                - ocp0/1
                - ocp1/0
                - ocp1/1
            switchesFilter:
              status:
                - active
              site:
                - phx02
              role:
                - cluster-network-leaf-switch
                - cluster-ply-network-switch
                - leaf-switch
                - storage-ply-leaf-switch
                - storage-network-leaf-switch
                - management-switch    
`, mockServerIP, mockServerPort)

		err = testHelper.DeploySDNWithConfig(configYaml, false)
		Expect(err).ShouldNot(HaveOccurred())

		switchCRsList := &idcnetworkv1alpha1.SwitchList{}

		for i := 0; i < 10; i++ {
			listOpts := &client.ListOptions{
				Namespace: idcnetworkv1alpha1.SDNControllerNamespace,
			}
			err := testHelper.K8sClient.List(ctx, switchCRsList, listOpts)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
		}

		// get SP list
		switchPortCRsList := &idcnetworkv1alpha1.SwitchPortList{}
		for i := 0; i < 10; i++ {
			listOpts := &client.ListOptions{
				Namespace: idcnetworkv1alpha1.SDNControllerNamespace,
			}
			err := testHelper.K8sClient.List(ctx, switchPortCRsList, listOpts)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
		}

		// check interfaces
		switchPortCRsExpectedToBeCreated := map[string]struct{}{
			"ethernet1-1.phx02-c01-acsw009.us-dev-1.fakecloud.intel.com": struct{}{},
			"ethernet2-1.phx02-c01-acsw011.us-dev-1.fakecloud.intel.com": struct{}{},
			"ethernet4-1.phx02-c01-acsw009.us-dev-1.fakecloud.intel.com": struct{}{},
		}
		switchPortCRsActualCreated := map[string]struct{}{}
		for _, sp := range switchPortCRsList.Items {
			switchPortCRsActualCreated[sp.Name] = struct{}{}
		}

		Expect(len(switchPortCRsExpectedToBeCreated) == len(switchPortCRsActualCreated)).Should(BeTrue())
		// check if the expected switches have been created
		for sp := range switchPortCRsExpectedToBeCreated {
			_, found := switchPortCRsActualCreated[sp]
			Expect(found).Should(BeTrue())
		}

		fmt.Printf("netbox-import case %v completed \n", caseName)
	}

	/*
		Purpose of the Test: Test importing interfaces from Netbox for both v3 and v4 Netbox server data format and Netbox clients.

		Test Steps:
			- Make Netbox ListDevices and ListInterfaces APIs return v3 or v4 format data based on the test cases.
			- Start SDN Controller.
			- Check if the SwitchPort CRs are created correctly.

		Test Coverage:
			 - v3 and v4 Netbox devices and interfaces data parsing in Netbox client
	*/
	It("test case 5 - import interfaces from netbox with v3 version of data format", Label("sdn", "netbox", "interfaces"), func() {
		testNetboxGetProviderInterfacesForDiffVersions("netbox/list_devices_for_interfaces_test_case_v3.json", "netbox/list_interfaces_v3.json", "netbox/mock-api-status-responsev3.json", "5")
	})
	It("test case 6 - import interfaces from netbox with v4 version of data format", Label("sdn", "netbox", "interfaces"), func() {
		testNetboxGetProviderInterfacesForDiffVersions("netbox/list_devices_for_interfaces_test_case_v4.json", "netbox/list_interfaces_v4.json", "netbox/mock-api-status-responsev4.json", "6")
	})
})
