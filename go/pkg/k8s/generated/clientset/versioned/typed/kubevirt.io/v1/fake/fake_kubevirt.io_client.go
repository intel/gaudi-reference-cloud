/*
This software and the related documents are Intel copyrighted materials, and
your use of them is governed by the express license under which they were
provided to you ("License"). Unless the License provides otherwise, you may not
use, modify, copy, publish, distribute, disclose or transmit this software or
the related documents without Intel's prior written permission.

This software and the related documents are provided as is, with no express or
implied warranties, other than those that are expressly stated in the License.
*/

// Code generated by main. DO NOT EDIT.

package fake

import (
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/clientset/versioned/typed/kubevirt.io/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeKubevirtV1 struct {
	*testing.Fake
}

func (c *FakeKubevirtV1) KubeVirts(namespace string) v1.KubeVirtInterface {
	return &FakeKubeVirts{c, namespace}
}

func (c *FakeKubevirtV1) VirtualMachines(namespace string) v1.VirtualMachineInterface {
	return &FakeVirtualMachines{c, namespace}
}

func (c *FakeKubevirtV1) VirtualMachineInstances(namespace string) v1.VirtualMachineInstanceInterface {
	return &FakeVirtualMachineInstances{c, namespace}
}

func (c *FakeKubevirtV1) VirtualMachineInstanceMigrations(namespace string) v1.VirtualMachineInstanceMigrationInterface {
	return &FakeVirtualMachineInstanceMigrations{c, namespace}
}

func (c *FakeKubevirtV1) VirtualMachineInstancePresets(namespace string) v1.VirtualMachineInstancePresetInterface {
	return &FakeVirtualMachineInstancePresets{c, namespace}
}

func (c *FakeKubevirtV1) VirtualMachineInstanceReplicaSets(namespace string) v1.VirtualMachineInstanceReplicaSetInterface {
	return &FakeVirtualMachineInstanceReplicaSets{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeKubevirtV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
