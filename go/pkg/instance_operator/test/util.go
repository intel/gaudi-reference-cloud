// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"path/filepath"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
)

func NewTestVmInstanceOperatorConfig(testDataDir string, scheme *runtime.Scheme) *cloudv1alpha1.VmInstanceOperatorConfig {
	configFile := filepath.Join(testDataDir, "operatorconfig.yaml")
	cfg := &cloudv1alpha1.VmInstanceOperatorConfig{}
	options := ctrl.Options{Scheme: scheme}
	var err error
	options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(cfg))
	Expect(err).Should(Succeed())
	return cfg
}

func NewTestBmInstanceOperatorConfig(testDataDir string, scheme *runtime.Scheme) *cloudv1alpha1.BmInstanceOperatorConfig {
	configFile := filepath.Join(testDataDir, "operatorconfig.yaml")
	cfg := &cloudv1alpha1.BmInstanceOperatorConfig{}
	options := ctrl.Options{Scheme: scheme}
	var err error
	options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(cfg))
	Expect(err).Should(Succeed())
	return cfg
}

func NewNamespace(namespace string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func NewInstanceTypeSpec(instanceType string) *cloudv1alpha1.InstanceTypeSpec {
	diskSpecs := []cloudv1alpha1.DiskSpec{
		{Size: "200Gi"},
	}
	cpuSpec := cloudv1alpha1.CpuSpec{
		Cores:     4,
		Sockets:   1,
		Threads:   1,
		ModelName: "Intel Test Server",
		Id:        "0x806F2",
	}
	memorySpec := cloudv1alpha1.MemorySpec{
		Size:      "8Gi",
		DimmSize:  "8Gi",
		DimmCount: 1,
		Speed:     3200,
	}
	return &cloudv1alpha1.InstanceTypeSpec{
		Name:             instanceType,
		DisplayName:      instanceType,
		Description:      "Intel Test Server",
		InstanceCategory: cloudv1alpha1.InstanceCategoryVirtualMachine,
		Disks:            diskSpecs,
		Cpu:              cpuSpec,
		Memory:           memorySpec,
	}
}

func NewSshPublicKeySpecs(sshPublicKeys ...string) []cloudv1alpha1.SshPublicKeySpec {
	var spec []cloudv1alpha1.SshPublicKeySpec
	for _, sshPublicKey := range sshPublicKeys {
		spec = append(spec, cloudv1alpha1.SshPublicKeySpec{
			SshPublicKey: sshPublicKey,
		})
	}
	return spec
}

func NewInterfaceSpecs() []cloudv1alpha1.InterfaceSpec {
	return []cloudv1alpha1.InterfaceSpec{
		{
			Name:        "eth0",
			VNet:        "us-dev-1a-default",
			DnsName:     "my-virtual-machine-tiny-1.03165859732720551183.us-dev-1.cloud.intel.com",
			Nameservers: []string{"1.1.1.1"},
		},
	}
}

func NewInstance(namespace string, instanceName string, availabilityZone string, region string, instanceTypeSpec *cloudv1alpha1.InstanceTypeSpec,
	sshPublicKeys []cloudv1alpha1.SshPublicKeySpec, interfaces []cloudv1alpha1.InterfaceSpec) *cloudv1alpha1.Instance {
	return &cloudv1alpha1.Instance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "private.cloud.intel.com/v1alpha1",
			Kind:       "Instance",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instanceName,
			Namespace: namespace,
		},
		Spec: cloudv1alpha1.InstanceSpec{
			AvailabilityZone: availabilityZone,
			Region:           region,
			RunStrategy:      "RerunOnFailure",
			Interfaces:       interfaces,
			InstanceTypeSpec: *instanceTypeSpec,
			MachineImageSpec: cloudv1alpha1.MachineImageSpec{
				Name: "ubuntu-22.04",
			},
			SshPublicKeySpecs: sshPublicKeys,
			ClusterGroupId:    "test-clustergroup-1a-2",
			ClusterId:         "test-clustergroup-1a-2-4",
		},
	}
}
