/*
This software and the related documents are Intel copyrighted materials, and
your use of them is governed by the express license under which they were
provided to you ("License"). Unless the License provides otherwise, you may not
use, modify, copy, publish, distribute, disclose or transmit this software or
the related documents without Intel's prior written permission.

This software and the related documents are provided as is, with no express or
implied warranties, other than those that are expressly stated in the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SshProxyTunnelLister helps list SshProxyTunnels.
// All objects returned here must be treated as read-only.
type SshProxyTunnelLister interface {
	// List lists all SshProxyTunnels in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SshProxyTunnel, err error)
	// SshProxyTunnels returns an object that can list and get SshProxyTunnels.
	SshProxyTunnels(namespace string) SshProxyTunnelNamespaceLister
	SshProxyTunnelListerExpansion
}

// sshProxyTunnelLister implements the SshProxyTunnelLister interface.
type sshProxyTunnelLister struct {
	indexer cache.Indexer
}

// NewSshProxyTunnelLister returns a new SshProxyTunnelLister.
func NewSshProxyTunnelLister(indexer cache.Indexer) SshProxyTunnelLister {
	return &sshProxyTunnelLister{indexer: indexer}
}

// List lists all SshProxyTunnels in the indexer.
func (s *sshProxyTunnelLister) List(selector labels.Selector) (ret []*v1alpha1.SshProxyTunnel, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SshProxyTunnel))
	})
	return ret, err
}

// SshProxyTunnels returns an object that can list and get SshProxyTunnels.
func (s *sshProxyTunnelLister) SshProxyTunnels(namespace string) SshProxyTunnelNamespaceLister {
	return sshProxyTunnelNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SshProxyTunnelNamespaceLister helps list and get SshProxyTunnels.
// All objects returned here must be treated as read-only.
type SshProxyTunnelNamespaceLister interface {
	// List lists all SshProxyTunnels in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SshProxyTunnel, err error)
	// Get retrieves the SshProxyTunnel from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.SshProxyTunnel, error)
	SshProxyTunnelNamespaceListerExpansion
}

// sshProxyTunnelNamespaceLister implements the SshProxyTunnelNamespaceLister
// interface.
type sshProxyTunnelNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SshProxyTunnels in the indexer for a given namespace.
func (s sshProxyTunnelNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.SshProxyTunnel, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SshProxyTunnel))
	})
	return ret, err
}

// Get retrieves the SshProxyTunnel from the indexer for a given namespace and name.
func (s sshProxyTunnelNamespaceLister) Get(name string) (*v1alpha1.SshProxyTunnel, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("sshproxytunnel"), name)
	}
	return obj.(*v1alpha1.SshProxyTunnel), nil
}
