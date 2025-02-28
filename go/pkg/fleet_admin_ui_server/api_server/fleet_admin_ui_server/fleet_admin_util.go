// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package fleet_admin

import (
	"context"
	"math"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	UpsertNodePoolQuery = `
		INSERT INTO node_pool (
			node_id, pool_id
		) VALUES (
			$1, $2
		)
		ON CONFLICT (node_id, pool_id) DO NOTHING
	`

	DeleteAllNodeInstanceTypeQuery = `
		DELETE FROM node_instance_type WHERE node_id = $1
	`

	DeleteNodeInstanceTypeQuery = `
		DELETE FROM node_instance_type
		WHERE node_id = $1
		AND instance_type NOT IN (SELECT UNNEST($2::text[]))
	`

	DeleteAllNodePoolQuery = `
	DELETE FROM node_pool WHERE node_id = $1
	`

	DeleteNodePoolQuery = `
	DELETE FROM node_pool
	WHERE node_id = $1
	AND pool_id NOT IN (SELECT UNNEST($2::text[]))
	`

	UpsertNodeInstanceTypeQuery = `
		INSERT INTO node_instance_type (
			node_id, instance_type
		) VALUES (
			$1, $2
		)
		ON CONFLICT (node_id, instance_type) DO NOTHING;
	`
)

func PercentageResourcesUsed(ctx context.Context, instanceCategory string, freeMillicpu int32, usedMillicpu int32, freeMemoryBytes int64, usedMemoryBytes int64, freeGpu int32, usedGpu int32) (float32, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.PercentageResourcesUsed").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	var percentageGpuUsed float64
	if instanceCategory == pb.InstanceCategory_BareMetalHost.String() {
		if usedMillicpu > 0 {
			return float32(100), nil
		}
		return float32(0), nil
	}

	percentageCpuUsed := (float64(usedMillicpu) / float64(usedMillicpu+freeMillicpu)) * 100
	percentageMemoryUsed := (float64(usedMemoryBytes) / float64(usedMemoryBytes+freeMemoryBytes)) * 100
	// For nodes which does not have any GPU devices free and used GPU will be 0.0
	sumFreeAndUsedGPU := float64(usedGpu + freeGpu)
	if sumFreeAndUsedGPU == 0 {
		percentageGpuUsed = 0
	} else {
		percentageGpuUsed = (float64(usedGpu) / sumFreeAndUsedGPU) * 100
	}

	maxOfCpuMemory := math.Max(percentageCpuUsed, percentageMemoryUsed)
	percentageResourcesUsed := math.Max(maxOfCpuMemory, percentageGpuUsed)

	return float32(percentageResourcesUsed), nil
}
