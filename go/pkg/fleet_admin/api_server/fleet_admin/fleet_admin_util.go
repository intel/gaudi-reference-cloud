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
	SelectInstanceTypeQuery = `
		SELECT instance_type FROM node_instance_type WHERE node_id = $1
	`
	SelectPoolIdQuery = `
		SELECT pool_id FROM node_pool WHERE node_id = $1
	`
	InsertNodeQuery = `
		INSERT INTO node (
			region, availability_zone, cluster_id, namespace, node_name, override_instance_types, override_pools
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING node_id
	`

	UpsertNodePoolQuery = `
		INSERT INTO node_pool (
			node_id, pool_id
		) VALUES (
			$1, $2
		)
		ON CONFLICT (node_id, pool_id) DO NOTHING
	`

	UpsertNodeStatisticsQuery = `
		INSERT INTO node_stats (
			node_id, reported_time, source_group, source_version, source_resource, instance_category, 
			partition, cluster_group, network_mode, free_millicpu, used_millicpu, 
			free_memory_bytes, used_memory_bytes, free_gpu, used_gpu
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (node_id) DO UPDATE SET
			reported_time = EXCLUDED.reported_time,
			source_group = EXCLUDED.source_group,
			source_version = EXCLUDED.source_version,
			source_resource = EXCLUDED.source_resource,
			instance_category = EXCLUDED.instance_category,
			partition = EXCLUDED.partition,
			cluster_group = EXCLUDED.cluster_group,
			network_mode = EXCLUDED.network_mode,
			free_millicpu = EXCLUDED.free_millicpu,
			used_millicpu = EXCLUDED.used_millicpu,
			free_memory_bytes = EXCLUDED.free_memory_bytes,
			used_memory_bytes = EXCLUDED.used_memory_bytes,
			free_gpu = EXCLUDED.free_gpu,
			used_gpu = EXCLUDED.used_gpu
	`

	UpsertNodeInstanceTypeStatsQuery = `
		INSERT INTO node_instance_type_stats (
			node_id, instance_type, running_instances, max_new_instances
		) VALUES (
			$1, $2, $3, $4
		)
		ON CONFLICT (node_id, instance_type) DO UPDATE SET
			running_instances = EXCLUDED.running_instances,
			max_new_instances = EXCLUDED.max_new_instances;
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

	SelectNodeQuery = `
		SELECT node_id, region, availability_zone, cluster_id, namespace, node_name, override_instance_types, override_pools 
		FROM node 
		WHERE cluster_id = $1 AND region = $2 AND availability_zone = $3 AND node_name = $4 AND namespace = $5
	`

	UpsertPoolFromNodeStatsQuery = `
		INSERT INTO pool (pool_id, pool_name, pool_account_manager_ags_role)
		VALUES ($1, $2, $3)
		ON CONFLICT (pool_id) DO NOTHING;
	`
)

func PercentageResourcesUsed(ctx context.Context, instanceCategory string, freeMillicpu int32, usedMillicpu int32, freeMemoryBytes int64, usedMemoryBytes int64, freeGpu int32, usedGpu int32) (float32, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.PercentageResourcesUsed").Start()
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
