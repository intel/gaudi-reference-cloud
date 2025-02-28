// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package db_query_constants

// Cluster Queries
var (
	UpdateClusterStateQuery = `
	 UPDATE public.cluster
	 SET clusterstate_name = $2
	 WHERE cluster_id = $1
 	`
)

// Storage Queries
var (
	UpdateClusterStorageDataQuery = `
	UPDATE public.cluster
	SET storage_enable = $2
	WHERE cluster_id = $1
	`

	UpdateCloudAccountStorageQuery = `
	UPDATE public.cloudaccountextraspec
	SET total_storage_size = total_storage_size + $2
	WHERE cloudaccount_id = $1
	`
)
