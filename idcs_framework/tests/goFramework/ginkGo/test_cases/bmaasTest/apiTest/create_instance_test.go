//go:build create_instance
// +build create_instance

package apiTest_test

import (
	"log"
	"time"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	utils "goFramework/utils/vmaasutils" // TODO: rename the package to a more general one (?)
	"goFramework/framework/library/bmaas/serviceapi"
)

var _ = Describe("Create Instance", Ordered, Label("large"), func() {

	var instanceAPIInput = "../../../test_config/bmaas_resources/bmaas_api_input.json"
	utils.LoadVMaaSAPIConfig(instanceAPIInput)

	var sshkeyAPIBaseUrl = utils.GetBaseUrl() + "/" + utils.GetCloudAccount() + "/" + "sshpublickeys"
	var instanceAPIBaseUrl = utils.GetBaseUrl() + "/" + utils.GetCloudAccount() + "/" + "instances"
	var sshkeyName = "user2@acme.com"
	var instanceName = "my-metal-instance-1"
	var instanceType = "bm-virtual"

	It("Create SSH public key", func() {
		log.Printf("sshkeyAPIBaseUrl: %s", sshkeyAPIBaseUrl)
		log.Printf("instanceAPIBaseUrl: %s", instanceAPIBaseUrl)
		payload := utils.GetSSHPayload()
		payload = strings.Replace(payload, "<<ssh_key_name>>", sshkeyName, 1)
		body , status := serviceapi.CreateSSHKey(sshkeyAPIBaseUrl, payload)
		log.Printf("Create SSH Key body response: %s", body)
		Expect(status).To(Equal(200), "Failed to create ssh public key")
	})

	It("Verify that the instance does not exist yet", func() {
		body , status := serviceapi.GetInstanceByName(instanceAPIBaseUrl, instanceName)
		log.Printf("Get instance body response: %s", body)
		Expect(status).To(Equal(404), "Failed to create BM instance, already exists")
	})

	It("Create instance by name", func() {
		payload := utils.GetInstancePayload()
		payload = strings.Replace(payload, "<<instance_name>>", instanceName, 1)
		payload = strings.Replace(payload, "<<ssh_public_key_name>>", sshkeyName, 1)
		payload = strings.Replace(payload, "<<instance_type>>", instanceType, 1)
		payload = strings.Replace(payload, "<<machine_image>>", "ubuntu-22.04-server-cloudimg-amd64-latest", 1)
		body , status := serviceapi.CreateInstance(instanceAPIBaseUrl, payload)
		log.Printf("Create instance body response: %s", body)
		Expect(status).To(Equal(200), "Failed to create BM instance")	
		
	})

	It("Verify the instance exists now", func() {
		// Adding a sleep because it seems to take some time to reflect the creation 
		time.Sleep(30 * time.Second)
		body , status := serviceapi.GetInstanceByName(instanceAPIBaseUrl, instanceName)
		log.Printf("Get instance body response: %s", body)
		Expect(status).To(Equal(200), "Failed to create BM instance")
	})

})
