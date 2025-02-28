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
	"context"
	"time"

	scheme "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1 "kubevirt.io/api/core/v1"
)

// KubeVirtsGetter has a method to return a KubeVirtInterface.
// A group's client should implement this interface.
type KubeVirtsGetter interface {
	KubeVirts(namespace string) KubeVirtInterface
}

// KubeVirtInterface has methods to work with KubeVirt resources.
type KubeVirtInterface interface {
	Create(ctx context.Context, kubeVirt *v1.KubeVirt, opts metav1.CreateOptions) (*v1.KubeVirt, error)
	Update(ctx context.Context, kubeVirt *v1.KubeVirt, opts metav1.UpdateOptions) (*v1.KubeVirt, error)
	UpdateStatus(ctx context.Context, kubeVirt *v1.KubeVirt, opts metav1.UpdateOptions) (*v1.KubeVirt, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.KubeVirt, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.KubeVirtList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.KubeVirt, err error)
	KubeVirtExpansion
}

// kubeVirts implements KubeVirtInterface
type kubeVirts struct {
	client rest.Interface
	ns     string
}

// newKubeVirts returns a KubeVirts
func newKubeVirts(c *KubevirtV1Client, namespace string) *kubeVirts {
	return &kubeVirts{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the kubeVirt, and returns the corresponding kubeVirt object, and an error if there is any.
func (c *kubeVirts) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.KubeVirt, err error) {
	result = &v1.KubeVirt{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kubevirts").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KubeVirts that match those selectors.
func (c *kubeVirts) List(ctx context.Context, opts metav1.ListOptions) (result *v1.KubeVirtList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.KubeVirtList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kubevirts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kubeVirts.
func (c *kubeVirts) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("kubevirts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a kubeVirt and creates it.  Returns the server's representation of the kubeVirt, and an error, if there is any.
func (c *kubeVirts) Create(ctx context.Context, kubeVirt *v1.KubeVirt, opts metav1.CreateOptions) (result *v1.KubeVirt, err error) {
	result = &v1.KubeVirt{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("kubevirts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeVirt).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a kubeVirt and updates it. Returns the server's representation of the kubeVirt, and an error, if there is any.
func (c *kubeVirts) Update(ctx context.Context, kubeVirt *v1.KubeVirt, opts metav1.UpdateOptions) (result *v1.KubeVirt, err error) {
	result = &v1.KubeVirt{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kubevirts").
		Name(kubeVirt.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeVirt).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *kubeVirts) UpdateStatus(ctx context.Context, kubeVirt *v1.KubeVirt, opts metav1.UpdateOptions) (result *v1.KubeVirt, err error) {
	result = &v1.KubeVirt{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kubevirts").
		Name(kubeVirt.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kubeVirt).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the kubeVirt and deletes it. Returns an error if one occurs.
func (c *kubeVirts) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kubevirts").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kubeVirts) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kubevirts").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched kubeVirt.
func (c *kubeVirts) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.KubeVirt, err error) {
	result = &v1.KubeVirt{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("kubevirts").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
