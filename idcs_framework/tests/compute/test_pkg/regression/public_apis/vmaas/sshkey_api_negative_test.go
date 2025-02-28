package vmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Compute SSH public key endpoint(VM negative flow)", Label("compute", "vmaas", "vmaas_ssh_key", "vmaas_ssh_key_negative"), Ordered, ContinueOnFailure, func() {
	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Creating an SSH public key using invalid key", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create sSSH-public-key creation with invalid key")
			sshkeyPayload := utils.GetJsonValue("sshPublicKeyPayload")
			statusCode, responseBody := service_apis.CreateSSHKey(sshEndpoint, token, sshkeyPayload, "automation-invalid-key", "ssh-rsa invalid-key")
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, `"could not decode sshpublickey: illegal base64 data at input byte 7"`)).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving SSH public key using invalid resource id", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to get ssh-public-key with invalid id")
			statusCode, responseBody := service_apis.GetSSHKeyById(sshEndpoint, token, "invalid-id")
			Expect(statusCode).To(Equal(500), responseBody)
			Expect(strings.Contains(responseBody, `"message":"an unknown error occurred"`)).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving SSH public key using invalid resource name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to get ssh-public-key with invalid name")
			statusCode, responseBody := service_apis.GetSSHKeyByName(sshEndpoint, token, "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"resource not found"`)).To(BeTrue(), responseBody)
		})
	})

	When("Deleting SSH public key using invalid resource id", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to delete ssh-public-key with invalid id")
			statusCode, responseBody := service_apis.DeleteSSHKeyById(sshEndpoint, token, "invalid-id")
			Expect(statusCode).To(Equal(500), responseBody)
			Expect(strings.Contains(responseBody, `"message":"an unknown error occurred"`)).To(BeTrue(), responseBody)
		})
	})

	When("Deleting SSH public key using invalid resource name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to delete ssh-public-key with invalid name")
			statusCode, responseBody := service_apis.DeleteSSHKeyByName(sshEndpoint, token, "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"resource not found"`)).To(BeTrue(), responseBody)
		})
	})
})
