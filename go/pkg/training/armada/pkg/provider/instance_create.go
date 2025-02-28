// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package provider

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/idc_compute"

	"golang.org/x/sync/errgroup"
)

func (clusterSchd *ClusterProvisionScheduler) CreateClusterInstances(ctx context.Context, req []idc_compute.InstanceCreateRequest) error {
	log := log.FromContext(ctx).WithName("ClusterProvisionScheduler.CreateClusterInstances")
	log.Info("entering a instance creation")

	group, ctx := errgroup.WithContext(ctx)

	for idx, compute := range req {
		// Capture closure for the goroutine task
		currIdx := idx
		currCompute := compute

		group.Go(func() error {
			log.Info("instance create worker", "worker-id", currIdx, "instance-id", currCompute.Name)
			if err := clusterSchd.ComputeClient.CreateIDCComputeInstance(ctx, currCompute); err != nil {
				log.Error(err, "error provisioning instance")
				return err
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}
