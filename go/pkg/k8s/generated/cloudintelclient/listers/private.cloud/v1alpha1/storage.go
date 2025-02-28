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

// StorageLister helps list Storages.
// All objects returned here must be treated as read-only.
type StorageLister interface {
	// List lists all Storages in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Storage, err error)
	// Storages returns an object that can list and get Storages.
	Storages(namespace string) StorageNamespaceLister
	StorageListerExpansion
}

// storageLister implements the StorageLister interface.
type storageLister struct {
	indexer cache.Indexer
}

// NewStorageLister returns a new StorageLister.
func NewStorageLister(indexer cache.Indexer) StorageLister {
	return &storageLister{indexer: indexer}
}

// List lists all Storages in the indexer.
func (s *storageLister) List(selector labels.Selector) (ret []*v1alpha1.Storage, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Storage))
	})
	return ret, err
}

// Storages returns an object that can list and get Storages.
func (s *storageLister) Storages(namespace string) StorageNamespaceLister {
	return storageNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// StorageNamespaceLister helps list and get Storages.
// All objects returned here must be treated as read-only.
type StorageNamespaceLister interface {
	// List lists all Storages in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Storage, err error)
	// Get retrieves the Storage from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Storage, error)
	StorageNamespaceListerExpansion
}

// storageNamespaceLister implements the StorageNamespaceLister
// interface.
type storageNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Storages in the indexer for a given namespace.
func (s storageNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Storage, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Storage))
	})
	return ret, err
}

// Get retrieves the Storage from the indexer for a given namespace and name.
func (s storageNamespaceLister) Get(name string) (*v1alpha1.Storage, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("storage"), name)
	}
	return obj.(*v1alpha1.Storage), nil
}
