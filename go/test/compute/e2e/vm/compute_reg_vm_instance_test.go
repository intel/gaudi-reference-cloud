// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vm

import (
	"context"
	"strings"

	"github.com/google/uuid"
	computeopenapi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/tidwall/gjson"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compute ssh-key endpoint", Label("large", "compute", "vmaas", "compute_vm"), Ordered, ContinueOnFailure, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	var (
		instanceType                             string
		machineImage                             string
		vNetName                                 string
		availabilityZone                         string
		sshPublicKeyName, publicKey              string
		instanceResourceId, instanceResourceName string
		createInstanceResp                       *computeopenapi.ProtoInstance
		err                                      error
	)

	BeforeAll(func() {
		instanceType = "vm-spr-sml"
		machineImage = "ubuntu-2204-jammy-v20230122"
		availabilityZone = "us-dev-1a"
		vNetName, _ = computeTestHelper.CreateVNet(ctx, cloudAccount, "us-dev-1a-"+uuid.NewString(), availabilityZone)
		log.Info("Created VNet", "vNetName", vNetName)

		sshPublicKeyName, _, publicKey, _ = computeTestHelper.CreateSshPublicKey(ctx, cloudAccount)
		log.Info("Created SSH Key Pair", "sshPublicKeyName", sshPublicKeyName, "publicKey", publicKey)

		// create instance
		createInstanceReq := computeTestHelper.NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName, availabilityZone)
		createInstanceResp, err = computeTestHelper.CreateInstance(ctx, cloudAccount, createInstanceReq)
		Expect(err).Should(Succeed())
		instanceResourceId = createInstanceResp.Metadata.GetResourceId()
		instanceResourceName = createInstanceResp.Metadata.GetName()

		By("Waiting for instance to have SSH proxy address")
		Eventually(func(g Gomega) {
			instance, err := computeTestHelper.GetInstance(ctx, cloudAccount, instanceResourceId)
			log.Info("Get Instance", "instance", instance)
			g.Expect(err).Should(Succeed())
			g.Expect(instance.Status).ShouldNot(BeNil())
			g.Expect(instance.Status.SshProxy).ShouldNot(BeNil())
			g.Expect(*instance.Status.SshProxy.ProxyAddress).ShouldNot(BeEmpty())
		}, "30s", "2s").Should(Succeed())

		By("Waiting for VM to start")
		skipWaitForInstanceReady := false
		Eventually(func(g Gomega) {
			instance, err := computeTestHelper.GetInstance(ctx, cloudAccount, instanceResourceId)
			log.Info("Get Instance", "instance", instance)
			g.Expect(err).Should(Succeed())
			if skipWaitForInstanceReady {
				g.Expect(*instance.Status.Message).Should(ContainSubstring("Instance specification has been accepted and is being provisioned."))
			} else {
				g.Expect(string(*instance.Status.Phase)).Should(Equal("Ready"))
			}
		}, "4m", "2s").Should(Succeed())
	})

	// Positive flows
	When("Retrieving an instance created using resource id", func() {
		It("should be successful", func() {
			log.Info("Retrieve the instance via GET method using id")
			get_instance_by_id := instance_endpoint + "/id/" + instanceResourceId
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", get_instance_by_id, nil)
			Expect(response.StatusCode()).To(Equal(200))
		})
	})

	When("Retrieving an instance created using resource name", func() {
		It("should be successful", func() {
			log.Info("Retrieve the instance via GET method using name")
			get_instance_by_name := instance_endpoint + "/name/" + instanceResourceName
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", get_instance_by_name, nil)
			Expect(response.StatusCode()).To(Equal(200))
		})
	})

	When("Retrieving already created instance and validate creation timestamp, username and DNS", func() {
		It("validate creation time in instance creation response", func() {
			get_instance_by_id := instance_endpoint + "/id/" + instanceResourceId
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", get_instance_by_id, nil)
			Expect(response.StatusCode()).To(Equal(200))

			// validate the instance creation timestamp is not null
			creation_time_stamp_in_response := createInstanceResp.GetMetadata().CreationTimestamp
			Expect(creation_time_stamp_in_response).To(Not(Equal("null")))

			// validate the instance creation timestamp
			creation_time_stamp_in_response_unix := computeTestHelper.GetUnixTime(creation_time_stamp_in_response.String())
			Expect(computeTestHelper.ValidateTimeStamp(creation_time_stamp_in_response_unix, creation_time_stamp_in_response_unix-300000, creation_time_stamp_in_response_unix+300000)).Should(BeTrue())
		})

		It("validate username in instance creation response", func() {
			get_instance_by_id := instance_endpoint + "/id/" + instanceResourceId
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", get_instance_by_id, nil)
			Expect(response.StatusCode()).To(Equal(200))
			// validate the username
			instance_username := gjson.Get(response.String(), "status.userName").String()
			Expect(instance_username).To(Not(Equal("")), "user name shouldn't be empty")
		})

		It("validate DNS in instance creation response", func() {
			get_instance_by_id := instance_endpoint + "/id/" + instanceResourceId
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", get_instance_by_id, nil)
			Expect(response.StatusCode()).To(Equal(200))
			// validate DNS
			instance_dnsname := gjson.Get(response.String(), "status.interfaces.0.dnsName").String()
			Expect(strings.Contains(instance_dnsname, "idcservice.net")).To(BeTrue(), "dns name is not in the format of instanceName.cloudAccount.region.idcservice.net")
		})
	})

	// Negative flows
	When("Attempt to create an instance using invalid instance type", func() {
		It("should fail with client error", func() {
			log.Info("Attempt to create an instance with invalid instance type")
			createInstanceReq := computeTestHelper.NewCreateInstanceRequest([]string{sshPublicKeyName}, "invalid-instance-type", machineImage, vNetName, availabilityZone)
			_, err := computeTestHelper.CreateInstance(ctx, cloudAccount, createInstanceReq)
			Expect(err).ShouldNot(Succeed())
		})
	})

	When("Attempt to create an instance using invalid machine image", func() {
		It("should fail with client error", func() {
			log.Info("Attempt to create an instance with invalid machine image")
			createInstanceReq := computeTestHelper.NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, "invalid-machine-image", vNetName, availabilityZone)
			_, err := computeTestHelper.CreateInstance(ctx, cloudAccount, createInstanceReq)
			Expect(err).ShouldNot(Succeed())
		})
	})

	When("Attempt to create an instance using invalid SSH public key", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to create an instance with invalid SSH public key")
			createInstanceReq := computeTestHelper.NewCreateInstanceRequest([]string{"invalid-ssh-key"}, instanceType, machineImage, vNetName, availabilityZone)
			_, err := computeTestHelper.CreateInstance(ctx, cloudAccount, createInstanceReq)
			Expect(err).ShouldNot(Succeed())
		})
	})

	When("Retrieving an instance using invalid resource id", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to get an instance with invalid id")
			instance_by_invalid_id := instance_endpoint + "/id/invalidid"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", instance_by_invalid_id, nil)
			Expect(response.StatusCode()).To(Equal(500))
			Expect(strings.Contains(response.String(), `"message":"internal server error"`)).To(BeTrue())
		})
	})

	When("Retrieving an instance using invalid resource name", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to get an instance with invalid name")
			instance_by_invalid_name := instance_endpoint + "/name/invalidname"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", instance_by_invalid_name, nil)
			Expect(response.StatusCode()).To(Equal(404))
			Expect(strings.Contains(response.String(), `"message":"resource not found"`)).To(BeTrue())
		})
	})

	When("Updating an instance with invalid SSH key using resource id", func() {
		It("should fail with valid error...", func() {
			// update instance with invalid ssh key
			log.Info("Attempt to Update an instance with invalid field in payload via resource id")
			put_by_invalid_id_ep := instance_endpoint + "/id/" + instanceResourceId
			put_by_invalid_payload := `{"spec": {"sshPublicKeyNames": ["invalid-key"]}}`
			response_for_invalid_id, _ := computeTestHelper.RestyClient.Request(ctx, "PUT", put_by_invalid_id_ep, []byte(put_by_invalid_payload))
			Expect(response_for_invalid_id.StatusCode()).To(Equal(404))
			Expect(strings.Contains(response_for_invalid_id.String(), `"message":"resource not found"`)).To(BeTrue())

			// update instance with invalid ssh key
			log.Info("Attempt to Update an instance with invalid field in payload via resource id")
			put_by_invalid_name_ep := instance_endpoint + "/name/" + instanceResourceName
			response_for_invalid_name, _ := computeTestHelper.RestyClient.Request(ctx, "PUT", put_by_invalid_name_ep, []byte(put_by_invalid_payload))
			Expect(response_for_invalid_name.StatusCode()).To(Equal(404))
			Expect(strings.Contains(response_for_invalid_name.String(), `"message":"resource not found"`)).To(BeTrue())

		})
	})

	When("Attempt to delete an instance using invalid resource id", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to delete an instance with invalid id")
			delete_by_invalid_id := instance_endpoint + "/id/invalidid"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "DELETE", delete_by_invalid_id, nil)
			Expect(response.StatusCode()).To(Equal(500))
			Expect(strings.Contains(response.String(), `"message":"internal server error"`)).To(BeTrue())
		})
	})
	When("Attempt to delete an instance using invalid resource name", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to delete an instance with invalid name")
			delete_by_invalid_name := instance_endpoint + "/name/invalidname"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "DELETE", delete_by_invalid_name, nil)
			Expect(response.StatusCode()).To(Equal(404))
			Expect(strings.Contains(response.String(), `"message":"resource not found"`)).To(BeTrue())
		})
	})

	When("Creating an instance with too many char's in name", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to create an instance with invalid instance name with too many characters")
			createInstanceReq := computeTestHelper.NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, "invalid-machine-image", vNetName, availabilityZone)
			createInstanceReq.Metadata.SetName("instance-name-to-validate-the-character-length-for-testing-purpose-attempt" + uuid.NewString())
			_, err := computeTestHelper.CreateInstance(ctx, cloudAccount, createInstanceReq)
			Expect(err).ShouldNot(Succeed())
		})
	})

	AfterAll(func() {
		// delete instance
		response := computeTestHelper.DeleteInstanceViaResty(ctx, instance_endpoint, cloudAccount, instanceResourceId)
		Expect(response.StatusCode()).To(Equal(200))

		By("Waiting for instance to be not found")
		Eventually(func(g Gomega) {
			g.Expect(computeTestHelper.CheckInstanceNotFound(ctx, cloudAccount, instanceResourceId)).Should(Succeed())
		}, "3m", "1s").Should(Succeed())
	})
})
