package IKS_WorkerNodes_OutOfCredit_EU_test

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/service_api/financials"
	"goFramework/framework/service_api/iks"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/utils"
	"goFramework/framework/service_api/compute/frisby"


	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)
var _ = Describe("Check IKS WorkerNodes OutOfCredit enterprise user", Ordered, Label("IKS"), func() {

	It("Validate Enterprise cloudAccount", func() {
		logger.Logf.Info("Checking user cloudAccount")
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountByName(url, token, userName)
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		Expect(cloudAccountType).NotTo(BeNil(), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
		logger.Log.Info("CloudAccount Validated")
	})

	It("Get IKS Versions", func() {
		logger.Logf.Info("Getting all IKS versions")
		baseUrl := compute_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/iks/metadata/runtimes"
		fmt.Println("IKS Versions URL...", baseUrl)
		response_status, responseBody := iks.GetVersions(baseUrl, token)
		iks_version = gjson.Get(responseBody, "runtimes.0.k8sversionname.0").String()
		Expect(response_status).To(Equal(200), "Failed to retrieve IKS version")
	})

	It("Create IKS cluster", func() {
		logger.Logf.Info("Creating IKS Cluster")
		baseUrl := compute_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/iks/clusters"
		payload := `{"instanceType": "vm-spr-sml", "description": "test", "k8sversionname": "` + iks_version + `", "name": "test", "runtimename": "", "network": {"region": "us", "enableloadbalancer": true, "clusterdns": "0.0.0.0", "clustercidr": "0.0.0.0"}, "tags": []}`
		logger.Logf.Info("Payload: ", payload)
		response_status, responseBody := iks.CreateCluster(baseUrl, token, payload)
		logger.Logf.Info("Response body", responseBody)
		clusterId = gjson.Get(responseBody, "uuid").String()
		Expect(response_status).NotTo(Equal(200), "Failed create IKS cluster without credits is enabled")
	})

	//CREATE VNET
	It("Create vnet with name", func() {
		// fmt.Println("Starting the VNet Creation via API...")
		// // form the endpoint and payload
		vnet_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
		vnet_name := compute_utils.GetVnetName()
		vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
		// hit the api
		logger.Logf.Info("Vnet end point ", vnet_endpoint)
		vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
		vnet_created = gjson.Get(vnet_creation_body, "metadata.name").String()
		Expect(vnet_creation_status).To(Equal(200), "Failed to create VNet")
		Expect(vnet_name).To(Equal(vnet_created), "Failed to create Vnet, response validation failed")
	})

	// CREATE SSH KEY
	It("Create ssh public key with name", func() {
		logger.Logf.Info("Starting the SSH-Public-Key Creation via API...")
		// form the endpoint and payload
		ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
		sshPublicKey = utils.GetSSHKey()
		logger.Logf.Info("SSH key is" + sshPublicKey)
		sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
		logger.Logf.Info("SSH  end point ", ssh_publickey_endpoint)
		ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
		// hit the api
		sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, token, ssh_publickey_payload)
		Expect(sshkey_creation_status).To(Equal(200), "Failed to create SSH Public key")
		ssh_publickey_name_created = gjson.Get(sshkey_creation_body, "metadata.name").String()
		Expect(sshkey_name).To(Equal(ssh_publickey_name_created), "Failed to create SSH Public key, response validation failed")
	})

	It("Create worker node", func() {
		logger.Logf.Info("Creating worker node")
		baseUrl := base_url + "/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/iks/clusters/" + clusterId + "/nodegroups"
		payload := `{ 
						"count": 1, 
					    "vnet"s: [ {"availabilityzonename": "us-dev-1a", "networkinterfacevnetname": "us-dev-1a-default"} ],
					    "instancetypeid": "vm-spr-sml",
					    "instanceType": "vm-spr-sml",
					    "userdataurl": "www.test.com",
					    "name": "testSmallWorkerNode",
					    "description": "",
					    "tags": [],
					    "sshkeyname": [{ "sshkey": "` + ssh_publickey_name_created + `""}]
					    "upgradestrategy": {
					        "drainnodes": true,
					        "maxunavailablepercentage": 0
					    },
					}`
		response_status, responseBody := iks.CreateWorkerNode(baseUrl, token, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).NotTo(Equal(200), "Failed, creating a worker without credits should be disabled.")
	})
})
