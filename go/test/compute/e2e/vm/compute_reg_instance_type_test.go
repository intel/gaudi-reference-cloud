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

var _ = Describe("Compute instance type endpoint", Label("large", "compute", "vmaas", "compute_instance_type", "vmaas_instance_type"), Ordered, ContinueOnFailure, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	//var token string

	// Retrieval of token can be enabled once we enable authorization in bazel. Keeping it commented for now
	BeforeAll(func() {
		//token = restyclient.GetBearerToken(ctx, oidc_url)
	})

	// positive flows
	When("Listing all the instance types", func() {
		It("should be successful", func() {
			log.Info("Retrieve all the supported instance types...")
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", instance_type_endpoint, nil)
			Expect(response.StatusCode()).To(Equal(200), "Assertion Failed - while retrieving instance types")
		})
	})

	When("Retrieve VM instance type using name", func() {
		It("should be successful", func() {
			instance_types := []string{"vm-spr-tny", "vm-spr-sml", "vm-spr-med", "vm-spr-lrg"}
			for _, each_instance := range instance_types {
				log.Info("Retrieve an instance type via predefined instance type - name : " + each_instance + "...")
				instance_type_byname_url := instance_type_endpoint + "/" + each_instance
				response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", instance_type_byname_url, nil)
				Expect(response.StatusCode()).To(Equal(200), "assertion failed on response code")
				Expect(strings.Contains(response.String(), `"name":"`+each_instance+`"`)).To(BeTrue(), "assertion failed on response body")
			}
		})
	})

	// negative flows
	When("Retrieving instance type using invalid name", func() {
		It("should fail with valid error...", func() {
			log.Info("Retrieve an instance type via invalid name...")
			instance_type_byinvalidname := instance_type_endpoint + "/invalidname"
			response, _ := computeTestHelper.RestyClient.Request(ctx, "GET", instance_type_byinvalidname, nil)
			Expect(response.StatusCode()).To(Equal(404), "assertion failed on response code")
			Expect(strings.Contains(response.String(), `"message":"resource not found"`)).To(BeTrue(), "assertion failed on response body")
		})
	})
})
