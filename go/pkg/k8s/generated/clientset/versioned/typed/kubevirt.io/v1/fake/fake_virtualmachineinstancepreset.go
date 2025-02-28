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
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1 "kubevirt.io/api/core/v1"
)

// FakeVirtualMachineInstancePresets implements VirtualMachineInstancePresetInterface
type FakeVirtualMachineInstancePresets struct {
	Fake *FakeKubevirtV1
	ns   string
}

var virtualmachineinstancepresetsResource = v1.SchemeGroupVersion.WithResource("virtualmachineinstancepresets")

var virtualmachineinstancepresetsKind = v1.SchemeGroupVersion.WithKind("VirtualMachineInstancePreset")

// Get takes name of the virtualMachineInstancePreset, and returns the corresponding virtualMachineInstancePreset object, and an error if there is any.
func (c *FakeVirtualMachineInstancePresets) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.VirtualMachineInstancePreset, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(virtualmachineinstancepresetsResource, c.ns, name), &v1.VirtualMachineInstancePreset{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.VirtualMachineInstancePreset), err
}

// List takes label and field selectors, and returns the list of VirtualMachineInstancePresets that match those selectors.
func (c *FakeVirtualMachineInstancePresets) List(ctx context.Context, opts metav1.ListOptions) (result *v1.VirtualMachineInstancePresetList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(virtualmachineinstancepresetsResource, virtualmachineinstancepresetsKind, c.ns, opts), &v1.VirtualMachineInstancePresetList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.VirtualMachineInstancePresetList{ListMeta: obj.(*v1.VirtualMachineInstancePresetList).ListMeta}
	for _, item := range obj.(*v1.VirtualMachineInstancePresetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested virtualMachineInstancePresets.
func (c *FakeVirtualMachineInstancePresets) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(virtualmachineinstancepresetsResource, c.ns, opts))

}

// Create takes the representation of a virtualMachineInstancePreset and creates it.  Returns the server's representation of the virtualMachineInstancePreset, and an error, if there is any.
func (c *FakeVirtualMachineInstancePresets) Create(ctx context.Context, virtualMachineInstancePreset *v1.VirtualMachineInstancePreset, opts metav1.CreateOptions) (result *v1.VirtualMachineInstancePreset, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(virtualmachineinstancepresetsResource, c.ns, virtualMachineInstancePreset), &v1.VirtualMachineInstancePreset{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.VirtualMachineInstancePreset), err
}

// Update takes the representation of a virtualMachineInstancePreset and updates it. Returns the server's representation of the virtualMachineInstancePreset, and an error, if there is any.
func (c *FakeVirtualMachineInstancePresets) Update(ctx context.Context, virtualMachineInstancePreset *v1.VirtualMachineInstancePreset, opts metav1.UpdateOptions) (result *v1.VirtualMachineInstancePreset, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(virtualmachineinstancepresetsResource, c.ns, virtualMachineInstancePreset), &v1.VirtualMachineInstancePreset{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.VirtualMachineInstancePreset), err
}

// Delete takes name of the virtualMachineInstancePreset and deletes it. Returns an error if one occurs.
func (c *FakeVirtualMachineInstancePresets) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(virtualmachineinstancepresetsResource, c.ns, name, opts), &v1.VirtualMachineInstancePreset{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVirtualMachineInstancePresets) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(virtualmachineinstancepresetsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1.VirtualMachineInstancePresetList{})
	return err
}

// Patch applies the patch and returns the patched virtualMachineInstancePreset.
func (c *FakeVirtualMachineInstancePresets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.VirtualMachineInstancePreset, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(virtualmachineinstancepresetsResource, c.ns, name, pt, data, subresources...), &v1.VirtualMachineInstancePreset{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.VirtualMachineInstancePreset), err
}
