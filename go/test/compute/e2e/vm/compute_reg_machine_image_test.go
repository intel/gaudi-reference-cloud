// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vm

import (
	"context"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compute machine image endpoint", Label("large", "compute", "vmaas", "compute_machine_image", "vmaas_machine_image"), Ordered, ContinueOnFailure, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	//var token string

	// Retrieval of token can be enabled once we enable authorization in bazel. Keeping it commented for now
	BeforeAll(func() {
		//token = restyclient.GetBearerToken(ctx, oidc_url)
	})

	// positive flows
	When("Listing all the machine images", func() {
		It("should be successful", func() {
			log.Info("Retrieve all the supported machineimages...")
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", machine_image_endpoint, nil)
			Expect(response.StatusCode()).To(Equal(200), "Assertion Failed - while retrieving machine images")
		})
	})

	When("Retrieve VM machine image using name", func() {
		It("should be successful", func() {
			machine_images := []string{"ubuntu-2204-jammy-v20230122"}
			for _, each_image := range machine_images {
				log.Info("Retrieve an machine image via predefined name : " + each_image + "...")
				machine_image_byname_url := machine_image_endpoint + "/" + each_image
				response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", machine_image_byname_url, nil)
				Expect(response.StatusCode()).To(Equal(200), "assertion failed on response code")
				Expect(strings.Contains(response.String(), `"name":"`+each_image+`"`)).To(BeTrue(), "assertion failed on response body")
			}
		})
	})

	// negative flows
	When("Retrieving machine image using invalid name", func() {
		It("should fail with valid error...", func() {
			log.Info("Retrieve an machine image using invalid name...")
			machine_image_byinvalidname := machine_image_endpoint + "/invalidname"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", machine_image_byinvalidname, nil)
			Expect(response.StatusCode()).To(Equal(404), "assertion failed on response code")
			Expect(strings.Contains(response.String(), `"message":"resource not found"`)).To(BeTrue(), "assertion failed on response body")
		})

	})
})
