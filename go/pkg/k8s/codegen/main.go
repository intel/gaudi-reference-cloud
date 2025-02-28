// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Create clients for kubevirtv1 types.
package main

import (
	"log"
	"os"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	controllergen "github.com/rancher/wrangler/pkg/controller-gen"
	"github.com/rancher/wrangler/pkg/controller-gen/args"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

func main() {
	err := os.Unsetenv("GOPATH")
	if err != nil {
		log.Fatalln("error encountered while unsetting env variable GOPATH:", err)
	}
	controllergen.Run(args.Options{
		OutputPackage: "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated",
		Boilerplate:   "hack/boilerplate.go.txt",
		Groups: map[string]args.Group{
			cloudv1alpha1.SchemeGroupVersion.Group: {
				Types: []interface{}{
					cloudv1alpha1.Instance{},
					cloudv1alpha1.SshProxyTunnel{},
				},
				GenerateTypes: false,
			},
			kubevirtv1.SchemeGroupVersion.Group: {
				Types: []interface{}{
					kubevirtv1.VirtualMachine{},
					kubevirtv1.VirtualMachineInstance{},
					kubevirtv1.VirtualMachineInstanceMigration{},
				},
				GenerateTypes:   false,
				GenerateClients: true,
			},
			baremetalv1alpha1.SchemeGroupVersion.Group: {
				Types: []interface{}{
					baremetalv1alpha1.BareMetalHost{},
				},
				GenerateTypes: false,
			},
		},
	})
}
