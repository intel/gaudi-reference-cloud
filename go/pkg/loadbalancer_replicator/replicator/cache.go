// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package replicator

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"k8s.io/apimachinery/pkg/runtime/schema"
	toolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provides a simple implementation of cache.Cache (https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/cache#Cache)
// with a single SharedIndexInformer.
// Intended to be used with ListerWatcher to cache objects read using GRPC.
// Based on:
//   - https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/cache/informer_cache.go
//   - https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/cache/internal/deleg_map.go
type Cache struct {
	Informer toolscache.SharedIndexInformer
}

// Get implements Reader.
func (c *Cache) Get(ctx context.Context, key client.ObjectKey, out client.Object, opts ...client.GetOption) error {
	return fmt.Errorf("Cache.Get not implemented")
}

// List implements Reader.
func (c *Cache) List(ctx context.Context, out client.ObjectList, opts ...client.ListOption) error {
	return fmt.Errorf("Cache.List not implemented")
}

// GetInformerForKind returns the informer for the GroupVersionKind.
func (c *Cache) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind, opts ...cache.InformerGetOption) (cache.Informer, error) {
	return c.Informer, nil
}

// GetInformer returns the informer for the obj.
func (c *Cache) GetInformer(ctx context.Context, obj client.Object, opts ...cache.InformerGetOption) (cache.Informer, error) {
	return c.Informer, nil
}

// RemoveInformer removes an informer entry and stops it if it was running.
func (c *Cache) RemoveInformer(ctx context.Context, obj client.Object) error {
	return fmt.Errorf("Cache.RemoveInformer not implemented")
}

// NeedLeaderElection implements the LeaderElectionRunnable interface
// to indicate that this can be started without requiring the leader lock.
func (c *Cache) NeedLeaderElection() bool {
	return false
}

func (c *Cache) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return fmt.Errorf("Cache.IndexField not implemented")
}

// Start runs all the informers known to this cache until the context is closed.
// It blocks.
func (c *Cache) Start(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Cache.Start")
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")
	c.Informer.Run(ctx.Done())
	return nil
}

// WaitForCacheSync waits for all the caches to sync.  Returns false if it could not sync a cache.
func (c *Cache) WaitForCacheSync(ctx context.Context) bool {
	log := log.FromContext(ctx).WithName("Cache.WaitForCacheSync")
	log.Info("Waiting for cache synchronization to complete")
	success := toolscache.WaitForCacheSync(ctx.Done(), c.Informer.HasSynced)
	log.Info("WaitForCacheSync returned", logkeys.Success, success)
	return success
}
