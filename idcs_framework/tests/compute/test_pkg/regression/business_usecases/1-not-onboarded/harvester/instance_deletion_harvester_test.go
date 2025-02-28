package business_usecase

/*
Note: Not a end user scenario, Internal/Admin scenario
Steps:
1) Create the instance via Compute API
2) Delete the instance via Harvester kubeconfig
3) Validate the instance phase got updated to Failed state in compute API server.
*/

/*import (
	"os"
	"strings"
	"time"
	"github.com/tidwall/gjson"
	"compute/framework_pkg/auth"
	"compute/framework_pkg/client"
	"compute/framework_pkg/service_apis"
	"compute/test_pkg/utils"
	"compute/framework_pkg/logger"
)

var _ = Describe("VMaaS Regression - Harvester instance deletion via API", Label("business", "business_harvester"), Ordered, func() {
	var base_url string
	var create_response_status int
	var create_response_body string
	var token string
	var vm_name string
	var vm_payload string
	var sshPublicKey string
	var instance_id_created string

	BeforeAll(func() {
		utils.LoadTestConfig("../resources", "vmaas_business_input.json")
		automation_cloud_account = utils.GetCloudAccount()
		base_url = utils.GetBaseUrl()
		instance_endpoint = base_url + "/v1/cloudaccounts/" + automation_cloud_account + "/" + "instances"
		sshkey_endpoint = base_url + "/v1/cloudaccounts/" + automation_cloud_account + "/" + "sshpublickeys"
		harvester_login_url = utils.GetHarvesterLoginUrl()
		if utils.GetTestEnv() == "dev3"{
			token = os.Getenv("token_env")
		} else {
			token = auth.GetBearerTokenViaResty(utils.GetTestEnv())
		}
		sshPublicKey = utils.GetPublicKey()

		// name and payload creation
		vm_name = "harvester-deletion-vm-" + utils.GetRandomString()
		vm_payload = utils.GetInstancePayload()
	})

	// Create the SSH key
	It("Starting the SSH-Public-Key Creation flow via API", func () {
		sshkey_payload := utils.GetSSHPayload()
		sshkey_name = "automation-sshkey-" + utils.GetRandomString()
		sshkey_creation_status_positive, sshkey_creation_body_positive := utils.SSHPublicKeyCreation(sshkey_endpoint, token, sshkey_payload, sshkey_name, sshPublicKey)
		Expect(sshkey_creation_status_positive).To(Equal(200), "assertion failed on response code")
		Expect(strings.Contains(sshkey_creation_body_positive, `"name":"`+sshkey_name+`"`)).To(BeTrue(), "assertion failed on response body")
	 })

	 // Create the instance
	 It("Starting the Small Instance Creation flow via Instance API", func() {
		create_response_status, create_response_body = utils.InstanceCreation(instance_endpoint, token, vm_payload, vm_name, "vm-spr-sml", sshkey_name)
		Expect(create_response_status).To(Equal(200), "assertion failed on response code")
		Expect(strings.Contains(create_response_body, `"name":"`+vm_name+`"`)).To(BeTrue(), "assertion failed on response body")
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
	 })

	 // Validation of Instance phase
	It("Validation of Created VM", func() {
		logger.Log.Info("Checking whether instance is in ready state")
		instanceValidation := utils.CheckProvisionState(instance_endpoint, token, instance_id_created)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
	})

	 // Delete the instance via harvester kubeconfig file
	 It("Delete the instance via harvester API", func() {
		//instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		harvester_delete_endpoint := utils.GetHarvesterDeleteEndpoint() + "/" + automation_cloud_account + "/virtualmachines/" + instance_id_created
		harvester_payload := utils.GetHarvesterPayload()

		response_status, response_body := service_apis.DeleteInstanceViaHarvesterApi(harvester_login_url, harvester_delete_endpoint, harvester_payload)
		Expect(response_status).To(Equal(200), "assert instance deletion via harvester response code")
		Expect(strings.Contains(response_body, `"harvesterhci.io/vmName":"`+instance_id_created+`"`)).To(BeTrue(), "assert instance deletion via harvester response body")
	 })

	 // Get Instance by Id
	 It("Get the instance using resource id after deletion", func() {
		logger.log.Info("Retrieve the instance via GET method using id")
		get_response_byid_status, get_response_byid_body := service_apis.GetInstanceById(instance_endpoint, token, instance_id_created)
		Expect(get_response_byid_status).To(Equal(404), "assert response code - instance retrieval after deletion in harvester")
		Expect(strings.Contains(get_response_byid_body, `"phase":"Failed"`)).To(BeTrue(), "assert response body - instance retrieval after deletion in harvester")
	 })

	 // Delete the CRD in compute server
	 It("Delete the instance CRD using resource id", func() {
		logger.log.Info("Remove the instance CRD  via DELETE api using resource id")
		delete_response_byid_status, _ := service_apis.DeleteInstanceById(instance_endpoint, token, instance_id_created)
		Expect(delete_response_byid_status).To(Equal(200), "assert instance CRD deletion response code")
	 })

	 It("Validation of Instance Deletion", func() {
		logger.Log.Info("Validation of Instance Deletion")
		instanceValidation := utils.CheckInstanceDeletionById(instance_endpoint, token, instance_id_created)
		Eventually(instanceValidation, 2*time.Minute, 5*time.Second).Should(BeTrue())
	})

})*/
