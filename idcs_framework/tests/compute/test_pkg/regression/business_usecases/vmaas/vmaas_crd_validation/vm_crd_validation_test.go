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

var _ = Describe("Compute Instance CR Validation", Label("compute", "vmaas", "vmaas_crd_validation"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType       string
		createStatusCode   int
		createRespBody     string
		vmName             string
		vmPayload          string
		instanceResourceId string
		isInstanceCreated  = false
		instanceCR         *cloudv1alpha1.Instance
	)

	BeforeAll(func() {
		// load instance details to be created
		vmName = "automation-vm-" + utils.GetRandomString()
		vmPayload = utils.GetJsonValue("instancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")

		// instance creation
		logInstance.Println("Starting the Instance Creation flow via Instance API...")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, vmName,
			instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), "assertion failed on response body")
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 20*time.Second).Should(BeTrue())
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
			instanceCR, _ = client.GetInstanceCustomResource(instanceResourceId, cloudAccount, kubeConfigPath)
		})
	})

	When("validate the namespace and labels", func() {
		It("validate the namespace and labels", func() {
			logInstance.Println("validate the namespace and labels")
			Expect(instanceCR.Namespace).To(Equal(cloudAccount))
			Expect(instanceCR.Labels["cloud-account-id"]).To(Equal(cloudAccount))
			Expect(instanceCR.Labels["instance-category"]).To(Equal("VirtualMachine"))
			//TODO - validate the zone and region details
			//Expect(instanceCR.Labels["region"]).To(Equal(""), instanceCR.Labels)
		})
	})

	When("validate the metadata", func() {
		It("validate the metadata", func() {
			logInstance.Println("validate the metadata")
			Expect(strings.Contains(instanceCR.GetFinalizers()[0], `private.cloud.intel.com/instancefinalizer`)).To(BeTrue())
			Expect(strings.Contains(instanceCR.GetFinalizers()[1], `private.cloud.intel.com/instancemeteringmonitorfinalizer`)).To(BeTrue())
			//Expect(string(instanceCR.UID)).To(Equal(instanceResourceId))
			//TODO - validate the creation timestamp
		})
	})

	When("validate the spec details", func() {
		It("validate the spec details", func() {
			logInstance.Println("validate the spec details")
			Expect(instanceCR.Spec.InstanceType).To(Equal(instanceType))
			// TODO validate the CPU, Memory and Disk details
			// TODO validate the ssh public key present
		})
	})

	When("validate the status details", func() {
		It("validate the status details", func() {
			logInstance.Println("validate the status details")
			Expect(instanceCR.Status.Message).To(Equal("Instance is running and has completed running startup scripts. "))
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

			Expect(instanceCR.Status.UserName).To(Equal("ubuntu"))
		})
	})

	/*
		It("validate the status conditions", func() {
			// TODO validate the status conditions required for BM
		})

		It("validate the instance events ", func() {
			// TODO validate the VM events if any at ready state
		})
	*/

	When("Instance Deletion and Validation via API", func() {
		It("should be successful", func() {
			logInstance.Println("Ensure the created instance is deleted")
		})
	})

	AfterAll(func() {
		// instance deletion
		logInstance.Println("Remove the instance via DELETE api using resource id")
		instanceResourceId := gjson.Get(createRespBody, "metadata.resourceId").String()
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 20*time.Second).Should(BeTrue())
	})
})
