// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package provider

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/idc_storage"
)

func (clusterSchd *ClusterProvisionScheduler) CreateClusterStorageFilesystems(ctx context.Context, req []idc_storage.FilesystemCreateRequest) error {
	log := log.FromContext(ctx).WithName("ClusterProvisionScheduler.CreateClusterStorageFilesystems")
	log.Info("entering a storage creation")

	// STaaS provisioning process for a new volume does not take a long time, use basic loop instead for creation
	for idx, fsReq := range req {
		if fsReq.Name == "" || fsReq.Capacity == "" {
			return fmt.Errorf("new storage request missing name/capacity")
		}

		// Capture closure for the goroutine task
		currIdx := idx
		currFS := fsReq

		log.Info("storage create request", "index", currIdx, "storage name/capacity", fmt.Sprintf("%s %s", currFS.Name, currFS.Capacity))
		if err := clusterSchd.StorageClient.CreateIDCStorageFilesystem(ctx, currFS); err != nil {
			log.Error(err, "error provisioning storage filesystem")
			return err
		}
	}

	return nil
}
