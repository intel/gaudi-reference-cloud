/*
This software and the related documents are Intel copyrighted materials, and
your use of them is governed by the express license under which they were
provided to you ("License"). Unless the License provides otherwise, you may not
use, modify, copy, publish, distribute, disclose or transmit this software or
the related documents without Intel's prior written permission.

This software and the related documents are provided as is, with no express or
implied warranties, other than those that are expressly stated in the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	scheme "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// BareMetalHostsGetter has a method to return a BareMetalHostInterface.
// A group's client should implement this interface.
type BareMetalHostsGetter interface {
	BareMetalHosts(namespace string) BareMetalHostInterface
}

// BareMetalHostInterface has methods to work with BareMetalHost resources.
type BareMetalHostInterface interface {
	Create(ctx context.Context, bareMetalHost *v1alpha1.BareMetalHost, opts v1.CreateOptions) (*v1alpha1.BareMetalHost, error)
	Update(ctx context.Context, bareMetalHost *v1alpha1.BareMetalHost, opts v1.UpdateOptions) (*v1alpha1.BareMetalHost, error)
	UpdateStatus(ctx context.Context, bareMetalHost *v1alpha1.BareMetalHost, opts v1.UpdateOptions) (*v1alpha1.BareMetalHost, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.BareMetalHost, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.BareMetalHostList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.BareMetalHost, err error)
	BareMetalHostExpansion
}

// bareMetalHosts implements BareMetalHostInterface
type bareMetalHosts struct {
	client rest.Interface
	ns     string
}

// newBareMetalHosts returns a BareMetalHosts
func newBareMetalHosts(c *Metal3V1alpha1Client, namespace string) *bareMetalHosts {
	return &bareMetalHosts{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the bareMetalHost, and returns the corresponding bareMetalHost object, and an error if there is any.
func (c *bareMetalHosts) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.BareMetalHost, err error) {
	result = &v1alpha1.BareMetalHost{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("baremetalhosts").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of BareMetalHosts that match those selectors.
func (c *bareMetalHosts) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.BareMetalHostList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.BareMetalHostList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("baremetalhosts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested bareMetalHosts.
func (c *bareMetalHosts) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("baremetalhosts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a bareMetalHost and creates it.  Returns the server's representation of the bareMetalHost, and an error, if there is any.
func (c *bareMetalHosts) Create(ctx context.Context, bareMetalHost *v1alpha1.BareMetalHost, opts v1.CreateOptions) (result *v1alpha1.BareMetalHost, err error) {
	result = &v1alpha1.BareMetalHost{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("baremetalhosts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(bareMetalHost).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a bareMetalHost and updates it. Returns the server's representation of the bareMetalHost, and an error, if there is any.
func (c *bareMetalHosts) Update(ctx context.Context, bareMetalHost *v1alpha1.BareMetalHost, opts v1.UpdateOptions) (result *v1alpha1.BareMetalHost, err error) {
	result = &v1alpha1.BareMetalHost{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("baremetalhosts").
		Name(bareMetalHost.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(bareMetalHost).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *bareMetalHosts) UpdateStatus(ctx context.Context, bareMetalHost *v1alpha1.BareMetalHost, opts v1.UpdateOptions) (result *v1alpha1.BareMetalHost, err error) {
	result = &v1alpha1.BareMetalHost{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("baremetalhosts").
		Name(bareMetalHost.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(bareMetalHost).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the bareMetalHost and deletes it. Returns an error if one occurs.
func (c *bareMetalHosts) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("baremetalhosts").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *bareMetalHosts) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("baremetalhosts").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched bareMetalHost.
func (c *bareMetalHosts) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.BareMetalHost, err error) {
	result = &v1alpha1.BareMetalHost{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("baremetalhosts").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
