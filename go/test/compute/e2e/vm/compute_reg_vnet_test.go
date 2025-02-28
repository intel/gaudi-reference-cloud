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

var _ = Describe("Compute vnet endpoint", Label("large", "compute", "vmaas", "compute_vnet", "vmaas_vnet_positive"), Ordered, ContinueOnFailure, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	var (
		vnetName1          string
		vnetName2          string
		vnetResp1          *computeopenapi.ProtoVNet
		vnetResp2          *computeopenapi.ProtoVNet
		vnet1_id_created   string
		vnet2_name_created string
		availabilityZone   string
	)

	BeforeAll(func() {
		availabilityZone = "us-dev-1a"
		vnetName1, vnetResp1 = computeTestHelper.CreateVNet(ctx, cloudAccount, "us-dev-1a-"+uuid.NewString(), availabilityZone)
		log.Info("Created VNET-1 Details", "VNET Name", vnetName1)
		vnet1_id_created = vnetResp1.Metadata.GetResourceId()

		vnetName2, vnetResp2 = computeTestHelper.CreateVNet(ctx, cloudAccount, "us-dev-1a-"+uuid.NewString(), availabilityZone)
		log.Info("Created VNET-2 Details", "VNET Name", vnetName2)
		vnet2_name_created = vnetResp2.Metadata.GetName()
	})

	When("GET the VNET created and validate creation status, name and subnet details", func() {
		It("should be successful", func() {
			// Get VNET via resource id
			log.Info("Retrieve the vnet via GET method using id")
			vnet_by_id := vnet_endpoint + "/id/" + vnet1_id_created
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", vnet_by_id, nil)
			Expect(response.StatusCode()).To(Equal(200))
			Expect(strings.Contains(response.String(), vnet1_id_created)).To(BeTrue())

			// validate the subnet details
			vnet_prefix := gjson.Get(response.String(), "spec.prefixLength").Int()
			Expect(22).To(Equal(int(vnet_prefix)))
		})

	})

	When("Listing all available VNET", func() {
		It("should be successful", func() {
			// Get all the vnet's created
			log.Info("Retrieve all the vnet's available via GET method")
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", vnet_endpoint, nil)
			Expect(response.StatusCode()).To(Equal(200), "assertion failed on response code")
			numberOfVnets := gjson.Get(response.String(), "items").Array()
			log.Info("Debug details : ", "list of vnets", numberOfVnets)
			Expect(len(numberOfVnets)).Should(BeNumerically(">=", 2))
		})
	})

	When("Retrieving VNET using valid resource name", func() {
		It("should be successful", func() {
			log.Info("Retrieve the vnet via GET method using name")
			vnet_by_name := vnet_endpoint + "/name/" + vnet2_name_created
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", vnet_by_name, nil)
			Expect(response.StatusCode()).To(Equal(200))
			Expect(strings.Contains(response.String(), vnet2_name_created)).To(BeTrue())
		})
	})

	AfterAll(func() {
		// Delete by id (subnet has consumed address due to network/dev limitation - handling via 200 or 400)
		log.Info("Remove the vnet via Delete method using id")
		vnet_delete_by_id := vnet_endpoint + "/id/" + vnet1_id_created
		delete_by_id_response, _ := computeTestHelper.RestyClient.Request(ctx, "DELETE", vnet_delete_by_id, nil)
		Expect(delete_by_id_response.StatusCode()).To(Or(Equal(200), Equal(400)))

		get_vnet_after_deletion_byid, _ := computeTestHelper.RestyClient.Request(ctx, "GET", vnet_delete_by_id, nil)
		Expect(get_vnet_after_deletion_byid.StatusCode()).To(Or(Equal(404), Equal(200)))

		// Delete by name
		log.Info("Remove the ssh-public-key via Delete method using resource name")
		vnet_delete_by_name := vnet_endpoint + "/name/" + vnet2_name_created
		delete_by_name_response, _ := computeTestHelper.RestyClient.Request(ctx, "DELETE", vnet_delete_by_name, nil)
		Expect(delete_by_name_response.StatusCode()).To(Or(Equal(200), Equal(400)))

		get_vnet_after_deletion_byname, _ := computeTestHelper.RestyClient.Request(ctx, "GET", vnet_delete_by_name, nil)
		Expect(get_vnet_after_deletion_byname.StatusCode()).To(Or(Equal(404), Equal(200)))
	})
})

var _ = Describe("Compute vnet endpoint", Label("large", "compute", "vmaas", "compute_vnet", "vmaas_vnet_negative"), Ordered, ContinueOnFailure, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	var availabilityZone string

	BeforeAll(func() {
		availabilityZone = "us-dev-1a"
	})

	// negative flows
	When("Creating VNET with invalid char length on name", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to create vnet creation with invalid char length")
			vnet_payload, _ := computeTestHelper.NewPutVNetRequest("vnet-name-to-validate-the-character-length-for-testing-purpose-attempt1", availabilityZone).MarshalJSON()
			response, _ := computeTestHelper.RestyClient.Request(ctx, "POST", vnet_endpoint, vnet_payload)
			Expect(response.StatusCode()).To(Equal(500))
			Expect(strings.Contains(response.String(), `"message":"an unknown error occurred"`)).To(BeTrue())
		})
	})

	When("Creating VNET without name", func() {
		It("should fail with valid error...", func() {
			log.Info("VNet creation without name...")
			vnet_payload, _ := computeTestHelper.NewPutVNetRequest("", availabilityZone).MarshalJSON()
			response, _ := computeTestHelper.RestyClient.Request(ctx, "POST", vnet_endpoint, vnet_payload)
			Expect(response.StatusCode()).To(Equal(400))
			Expect(strings.Contains(response.String(), `missing metadata.name`)).To(BeTrue())
		})
	})

	When("Retrieving VNET using invalid ID", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to get vnet with invalid id")
			vnet_by_id := vnet_endpoint + "/id/invalidid"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", vnet_by_id, nil)
			Expect(response.StatusCode()).To(Equal(500))
			Expect(strings.Contains(response.String(), `"message":"an unknown error occurred"`)).To(BeTrue())
		})
	})

	When("Retrieving VNET using invalid name", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to get vnet with invalid name")
			vnet_by_name := vnet_endpoint + "/name/invalidname"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", vnet_by_name, nil)
			Expect(response.StatusCode()).To(Equal(404))
			Expect(strings.Contains(response.String(), `"message":"resource not found"`)).To(BeTrue())
		})
	})

	When("Deleting VNET using invalid ID", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to delete vnet with invalid id")
			vnet_by_id := vnet_endpoint + "/id/invalidid"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "DELETE", vnet_by_id, nil)
			Expect(response.StatusCode()).To(Equal(500), "assertion failed on response code")
			Expect(strings.Contains(response.String(), `"message":"an unknown error occurred"`)).To(BeTrue())
		})
	})

	When("Deleting VNET using invalid name", func() {
		It("should fail with valid error...", func() {
			log.Info("Attempt to delete vnet with invalid name")
			vnet_by_name := vnet_endpoint + "/name/invalidname"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "DELETE", vnet_by_name, nil)
			Expect(response.StatusCode()).To(Equal(404))
			Expect(strings.Contains(response.String(), `"message":"resource not found"`)).To(BeTrue())
		})
	})

})
