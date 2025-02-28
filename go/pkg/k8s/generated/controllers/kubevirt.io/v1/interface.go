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

package v1

import (
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/wrangler/pkg/schemes"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1 "kubevirt.io/api/core/v1"
)

func init() {
	schemes.Register(v1.AddToScheme)
}

type Interface interface {
	VirtualMachine() VirtualMachineController
	VirtualMachineInstance() VirtualMachineInstanceController
	VirtualMachineInstanceMigration() VirtualMachineInstanceMigrationController
}

func New(controllerFactory controller.SharedControllerFactory) Interface {
	return &version{
		controllerFactory: controllerFactory,
	}
}

type version struct {
	controllerFactory controller.SharedControllerFactory
}

func (c *version) VirtualMachine() VirtualMachineController {
	return NewVirtualMachineController(schema.GroupVersionKind{Group: "kubevirt.io", Version: "v1", Kind: "VirtualMachine"}, "virtualmachines", true, c.controllerFactory)
}
func (c *version) VirtualMachineInstance() VirtualMachineInstanceController {
	return NewVirtualMachineInstanceController(schema.GroupVersionKind{Group: "kubevirt.io", Version: "v1", Kind: "VirtualMachineInstance"}, "virtualmachineinstances", true, c.controllerFactory)
}
func (c *version) VirtualMachineInstanceMigration() VirtualMachineInstanceMigrationController {
	return NewVirtualMachineInstanceMigrationController(schema.GroupVersionKind{Group: "kubevirt.io", Version: "v1", Kind: "VirtualMachineInstanceMigration"}, "virtualmachineinstancemigrations", true, c.controllerFactory)
}
