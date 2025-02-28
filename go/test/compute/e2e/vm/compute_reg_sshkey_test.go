// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vm

import (
	"context"
	"strings"

	computeopenapi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/tidwall/gjson"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compute ssh-key endpoint", Label("large", "compute", "vmaas", "compute_ssh_key", "vmaas_ssh_key_positive"), Ordered, ContinueOnFailure, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	var (
		sshPublicKeyName1 string
		publicKey1        string
		sshKeyResp1       *computeopenapi.ProtoSshPublicKey
		sshKeyResp2       *computeopenapi.ProtoSshPublicKey
		ssh1_id_created   string
		ssh2_name_created string
	)

	BeforeAll(func() {
		// Retrieval of token can be enabled once we enable authorization in bazel. Keeping it commented for now
		//token = restyclient.GetBearerToken(ctx, oidc_url)
		sshPublicKeyName1, _, publicKey1, sshKeyResp1 = computeTestHelper.CreateSshPublicKey(ctx, cloudAccount)
		log.Info("Created SSH Key Pair", "sshPublicKeyName", sshPublicKeyName1, "publicKey", publicKey1)
		ssh1_id_created = sshKeyResp1.Metadata.GetResourceId()

		sshPublicKeyName2, _, publicKey2, sshkey2 := computeTestHelper.CreateSshPublicKey(ctx, cloudAccount)
		log.Info("Created SSH Key Pair", "sshPublicKeyName", sshPublicKeyName2, "publicKey", publicKey2)

		sshKeyResp2 = sshkey2
		ssh2_name_created = sshKeyResp2.Metadata.GetName()
	})

	// positive flows
	When("GET the ssh-key created and validate", func() {
		It("should be successful", func() {
			// Get SSH key via resource id
			log.Info("Retrieve the ssh-public-key via GET method using id")
			sshkey_by_id := ssh_endpoint + "/id/" + ssh1_id_created
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", sshkey_by_id, nil)
			Expect(response.StatusCode()).To(Equal(200))
			Expect(strings.Contains(response.String(), ssh1_id_created)).To(BeTrue())

			// validate the creation timestamp
			creation_time_stamp_in_response := gjson.Get(response.String(), "metadata.creationTimestamp").String()
			Expect(creation_time_stamp_in_response).To(Not(Equal("null")), "Creation time stamp shouldn't be null")
			creation_time_stamp_in_response_unix := computeTestHelper.GetUnixTime(creation_time_stamp_in_response)
			Expect(computeTestHelper.ValidateTimeStamp(creation_time_stamp_in_response_unix, creation_time_stamp_in_response_unix-6000, creation_time_stamp_in_response_unix+6000)).Should(BeTrue())

			// validate the ssh-publickey-value
			public_key_created := gjson.Get(response.String(), "spec.sshPublicKey").String()
			Expect(publicKey1).To(Equal(public_key_created), "SSH Public key provided while creation and retrieval shouldn't mismatch")

		})
	})

	When("GET all the ssh keys present in cloud account", func() {
		It("should be successful", func() {
			// Get all the ssh keys created
			log.Info("Retrieve all the ssh-public-key available via GET method")
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", ssh_endpoint, nil)
			Expect(response.StatusCode()).To(Equal(200))
			numberOfSSHKeys := gjson.Get(response.String(), "items").Array()
			log.Info("Debug details : ", "list of ssh keys", numberOfSSHKeys)
			Expect(len(numberOfSSHKeys)).Should(BeNumerically(">=", 2))
		})
	})

	When("GET the ssh key using resource name", func() {
		It("should be successful", func() {
			// Get SSH key via resource name
			log.Info("Retrieve the ssh-public-key via GET method using resource id")
			sshkey_by_name := ssh_endpoint + "/name/" + ssh2_name_created
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", sshkey_by_name, nil)
			Expect(response.StatusCode()).To(Equal(200))
			Expect(strings.Contains(response.String(), ssh2_name_created)).To(BeTrue())
		})
	})

	AfterAll(func() {
		// Delete by id
		log.Info("Remove the ssh-public-key via Delete method using resource id")
		ssh_delete_by_id := ssh_endpoint + "/id/" + ssh1_id_created
		delete_by_id_response, _ := computeTestHelper.RestyClient.Request(ctx, "DELETE", ssh_delete_by_id, nil)
		Expect(delete_by_id_response.StatusCode()).To(Equal(200))

		get_ssh_after_deletion_byid, _ := computeTestHelper.RestyClient.Request(ctx, "GET", ssh_delete_by_id, nil)
		Expect(get_ssh_after_deletion_byid.StatusCode()).To(Equal(404))

		// Delete by name
		log.Info("Remove the ssh-public-key via Delete method using resource name")
		ssh_delete_by_name := ssh_endpoint + "/name/" + ssh2_name_created
		delete_by_name_response, _ := computeTestHelper.RestyClient.Request(ctx, "DELETE", ssh_delete_by_name, nil)
		Expect(delete_by_name_response.StatusCode()).To(Equal(200))

		get_ssh_after_deletion_byname, _ := computeTestHelper.RestyClient.Request(ctx, "GET", ssh_delete_by_name, nil)
		Expect(get_ssh_after_deletion_byname.StatusCode()).To(Equal(404))
	})
})

var _ = Describe("Compute ssh-key endpoint", Label("large", "compute", "vmaas", "compute_ssh_key", "vmaas_ssh_key_negative"), Ordered, ContinueOnFailure, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	// negative flows
	When("Creating an SSH public key using invalid key", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to create sSSH-public-key creation with invalid key")
			sshkey_payload, _ := computeTestHelper.NewCreateSshPublicKeyRequest("automation-invalid-key@intel.com", "ssh-rsa invalid-key").MarshalJSON()
			response, _ := computeTestHelper.RestyClient.Request(ctx, "POST", ssh_endpoint, sshkey_payload)
			Expect(response.StatusCode()).To(Equal(400))
			Expect(strings.Contains(response.String(), `could not decode sshpublickey`)).To(BeTrue())
		})
	})

	When("Retrieving SSH public key using invalid resource id", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to get ssh-public-key with invalid id")
			ssh_get_endpoint := ssh_endpoint + "/id/invalidid"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", ssh_get_endpoint, nil)
			Expect(response.StatusCode()).To(Equal(500))
			Expect(strings.Contains(response.String(), `"message":"an unknown error occurred"`)).To(BeTrue())
		})
	})

	When("Retrieving SSH public key using invalid resource name", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to get ssh-public-key with invalid name")
			ssh_get_endpoint := ssh_endpoint + "/name/invalidname"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", ssh_get_endpoint, nil)
			Expect(response.StatusCode()).To(Equal(404))
			Expect(strings.Contains(response.String(), `"message":"resource not found"`)).To(BeTrue())
		})
	})

	When("Deleting SSH public key using invalid resource id", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to delete ssh-public-key with invalid id")
			ssh_delete_endpoint := ssh_endpoint + "/id/invalidid"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "DELETE", ssh_delete_endpoint, nil)
			Expect(response.StatusCode()).To(Equal(500))
			Expect(strings.Contains(response.String(), `"message":"an unknown error occurred"`)).To(BeTrue())
		})
	})

	When("Deleting SSH public key using invalid resource name", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to delete ssh-public-key with invalid name")
			ssh_delete_endpoint := ssh_endpoint + "/name/invalidname"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "DELETE", ssh_delete_endpoint, nil)
			Expect(response.StatusCode()).To(Equal(404))
			Expect(strings.Contains(response.String(), `"message":"resource not found"`)).To(BeTrue())
		})
	})
})
