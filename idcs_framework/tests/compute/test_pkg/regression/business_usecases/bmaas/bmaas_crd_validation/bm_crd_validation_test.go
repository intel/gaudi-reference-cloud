package crdvalidation

import (
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("Compute Instance CR Validation", Label("compute", "compute_bu", "compute_crd_validation", "bmaas_crd_validation"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType       string
		createStatusCode   int
		createRespBody     string
		bmName             string
		bmPayload          string
		instanceResourceId string
		isInstanceCreated  = false
		instanceCR         *cloudv1alpha1.Instance
	)

	BeforeAll(func() {
		// load instance details\
		bmName = "automation-bm-" + utils.GetRandomString()
		bmPayload = utils.GetJsonValue("instancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreatedCRValidation")

		// instance creation
		logInstance.Println("Starting Instance Creation flow via Instance API...")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, bmPayload, bmName,
			instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+bmName+`"`)).To(BeTrue(), createRespBody)
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
		isInstanceCreated = true
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Instance creation and its validation - prerequisite", func() {
		It("Validate the instance creation in before all", func() {
			logInstance.Println("is instance created? " + strconv.FormatBool(isInstanceCreated))
			Expect(isInstanceCreated).Should(BeTrue(), "Instance creation failed with following error "+createRespBody)

			client.SetUpKubeClient(kubeConfigPath, testEnv)
			instanceCR, _ = client.GetInstanceCustomResource(instanceResourceId, cloudAccount, "../../resources/kubeconfig/env-kubeconfig.yaml")
		})
	})

	When("Validate the namespace and labels", func() {
		It("Validate the namespace and labels", func() {
			Expect(instanceCR.Namespace).To(Equal(cloudAccount))
			Expect(instanceCR.Labels["cloud-account-id"]).To(Equal(cloudAccount))
			Expect(instanceCR.Labels["instance-category"]).To(Equal("BareMetalHost"))
			//TODO - validate the zone and region details
			//Expect(instanceCR.Labels["region"]).To(Equal(""), instanceCR.Labels)
		})
	})

	When("Validate the metadata", func() {
		It("Validate the metadata", func() {
			Expect(strings.Contains(instanceCR.GetFinalizers()[0], `private.cloud.intel.com/instancefinalizer`)).To(BeTrue())
			Expect(strings.Contains(instanceCR.GetFinalizers()[1], `private.cloud.intel.com/instancemeteringmonitorfinalizer`)).To(BeTrue())
			//Expect(string(instanceCR.UID)).To(Equal(instanceResourceId))
			//TODO - validate the creation timestamp
		})
	})

	When("Validate the spec details", func() {
		It("Validate the spec details", func() {
			Expect(instanceCR.Spec.InstanceType).To(Equal(instanceType))
			// TODO validate the CPU, Memory and Disk details
			// TODO validate the ssh public key present
		})
	})

	When("Validate the status details", func() {
		It("Validate the status details", func() {
			Expect(instanceCR.Status.Message).To(Equal("The instance has completed startup and is available to use. "))
			Expect(string(instanceCR.Status.Phase)).To(Equal("Ready"))
			// Following details shouldn't be empty
			Expect(instanceCR.Status.Interfaces[0].Addresses).ShouldNot(BeNil())
			Expect(instanceCR.Status.Interfaces[0].DnsName).ShouldNot(BeNil())
			Expect(instanceCR.Status.Interfaces[0].Gateway).ShouldNot(BeNil())
			Expect(instanceCR.Status.Interfaces[0].PrefixLength).ShouldNot(BeNil())
			Expect(instanceCR.Status.Interfaces[0].Subnet).ShouldNot(BeNil())
			Expect(instanceCR.Status.Interfaces[0].VNet).ShouldNot(BeNil())
			Expect(instanceCR.Status.Interfaces[0].VlanId).ShouldNot(BeNil())

			// SSH Proxy details shouldn't be empty
			Expect(instanceCR.Status.SshProxy.ProxyAddress).ShouldNot(BeNil())
			Expect(instanceCR.Status.SshProxy.ProxyUser).ShouldNot(BeNil())
			Expect(instanceCR.Status.SshProxy.ProxyPort).ShouldNot(BeNil())

			Expect(instanceCR.Status.UserName).To(Equal("sdp"))
		})
	})
	/*
		It("validate the status conditions", func() {
			// TODO validate the status conditions required for BM
		})

		It("validate the instance events ", func() {
			// TODO validate the BM events if any at ready state
		})
	*/

	When("Instance Deletion and Validation via API", func() {
		It("should be successful", func() {
			logInstance.Println("Ensure the created instance is deleted")
		})
	})

	AfterAll(func() {
		// instance deletion using resource id
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
	})
})
