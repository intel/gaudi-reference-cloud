// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package fleet_admin

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type FleetAdminService struct {
	pb.UnimplementedFleetAdminServiceServer
	db  *sql.DB
	cfg config.Config
}

type NodeWrapper struct {
	nodeId                int
	region                string
	availabilityZone      string
	clusterId             string
	namespace             string
	nodeName              string
	overrideInstanceTypes bool
	overridePools         bool
	instanceTypes         []string
	poolIds               []string
	clusterGroup          string
	networkMode           string
}

func NewFleetAdminService(
	db *sql.DB,
	config config.Config,

) (*FleetAdminService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &FleetAdminService{
		db:  db,
		cfg: config,
	}, nil
}

func (s *FleetAdminService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("FleetAdminService.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (s *FleetAdminService) UpdateComputeNodePoolsForCloudAccount(ctx context.Context, req *pb.UpdateComputeNodePoolsForCloudAccountRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.UpdateComputeNodePoolsForCloudAccount").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*emptypb.Empty, error) {

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelDefault,
		})
		if err != nil {
			return nil, fmt.Errorf("BeginTx: %w", err)
		}
		defer tx.Rollback()

		// Delete any existing rows.
		query := `delete from pool_cloud_account where cloud_account_id = $1`
		_, err = tx.ExecContext(ctx, query, req.CloudAccountId)
		if err != nil {
			return nil, fmt.Errorf("delete from pool_cloud_account: %w", err)
		}

		createAdmin := req.GetCreateAdmin()
		if createAdmin == "" {
			return nil, status.Errorf(codes.InvalidArgument, "create admin cannot be empty. provide a valid email address")
		}

		// Insert new rows.
		for _, pool := range req.ComputeNodePools {
			query := `insert into pool_cloud_account (pool_id, cloud_account_id, create_date, create_admin) values ($1, $2, current_timestamp, $3)`
			_, err := tx.ExecContext(ctx, query, pool.PoolId, req.CloudAccountId, createAdmin)
			if err != nil {
				return nil, fmt.Errorf("insert into pool_cloud_account: %w", err)
			}
		}

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}

		resp := &emptypb.Empty{}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *FleetAdminService) SearchComputeNodePoolsForInstanceScheduling(ctx context.Context, req *pb.SearchComputeNodePoolsForInstanceSchedulingRequest) (*pb.SearchComputeNodePoolsForInstanceSchedulingResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.SearchComputeNodePoolsForInstanceScheduling").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*pb.SearchComputeNodePoolsForInstanceSchedulingResponse, error) {
		query := `select pool_id from pool_cloud_account where cloud_account_id = $1 order by pool_id`
		rows, err := s.db.QueryContext(ctx, query, req.CloudAccountId)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var pools []*pb.ComputeNodePoolForInstanceScheduling
		for rows.Next() {
			pool := &pb.ComputeNodePoolForInstanceScheduling{}
			if err := rows.Scan(&pool.PoolId); err != nil {
				return nil, err
			}
			pools = append(pools, pool)
		}
		if len(pools) == 0 && s.cfg.ComputeNodePoolForUnknownCloudAccount != "" {
			logger.Info("No records found for Cloud Account. Returning value of ComputeNodePoolForUnknownCloudAccount.")
			pool := &pb.ComputeNodePoolForInstanceScheduling{
				PoolId: s.cfg.ComputeNodePoolForUnknownCloudAccount,
			}
			pools = append(pools, pool)
		}
		resp := &pb.SearchComputeNodePoolsForInstanceSchedulingResponse{
			ComputeNodePools: pools,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// TODO: Add unit test once Node Reporter is implemented.
func (s *FleetAdminService) GetResourcePatches(ctx context.Context, req *pb.GetResourcePatchesRequest) (*pb.GetResourcePatchesResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.GetResourcePatches").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.GetResourcePatchesResponse, error) {
		query := `select node_id, node_name, namespace from node where cluster_id = $1 and region = $2 and availability_zone = $3`
		rows, err := s.db.QueryContext(ctx, query, req.ClusterId, req.Region, req.AvailabilityZone)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var resourcePatches []*pb.ResourcePatch
		for rows.Next() {
			var nodeId int
			var nodeName string
			var namespace string
			var sourceGroup string
			var sourceVersion string
			var sourceResource string

			if err := rows.Scan(&nodeId, &nodeName, &namespace); err != nil {
				return nil, err
			}

			// For each node_id we will have one corresponding entry in the node_stats table
			var row *sql.Row
			queryNodeStats := `select source_group, source_version, source_resource from node_stats where node_id = $1`
			if row = s.db.QueryRowContext(ctx, queryNodeStats, nodeId); row.Err() != nil {
				return nil, row.Err()
			}
			if err = row.Scan(&sourceGroup, &sourceVersion, &sourceResource); err != nil {
				return nil, err
			}

			instanceTypes, pools, err := s.GetInstanceTypesAndPools(ctx, nodeId)
			if err != nil {
				return nil, err
			}

			// create map of labels for both instance_types and pools with the respective prefix
			labels := make(map[string]string)

			for _, instanceType := range instanceTypes {
				key := util.InstanceTypeLabelPrefix + instanceType
				labels[key] = "true"
			}

			for _, pool := range pools {
				key := util.ComputeNodePoolLabelPrefix + pool
				labels[key] = "true"
			}

			resourcePatch := &pb.ResourcePatch{
				NodeName:  nodeName,
				Namespace: namespace,
				Gvr: &pb.GroupVersionResource{
					Group:    sourceGroup,
					Version:  sourceVersion,
					Resource: sourceResource,
				},
				OwnedLabelsRegex: []string{
					util.InstanceTypeLabelPrefix + ".*",
					util.ComputeNodePoolLabelPrefix + ".*",
				},
				Labels: labels,
			}

			resourcePatches = append(resourcePatches, resourcePatch)
		}
		resp := &pb.GetResourcePatchesResponse{
			ResourcePatches: resourcePatches,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *FleetAdminService) GetInstanceTypesAndPools(ctx context.Context, nodeId int) ([]string, []string, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.GetInstanceTypesAndPools").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	instanceTypes := make([]string, 0)
	queryInstanceTypes := `select instance_type from node_instance_type where node_id = $1`
	rows, err := s.db.QueryContext(ctx, queryInstanceTypes, nodeId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var instanceType string

		if err := rows.Scan(&instanceType); err != nil {
			return nil, nil, err
		}
		instanceTypes = append(instanceTypes, instanceType)
	}

	pools := make([]string, 0)
	queryPools := `select pool_id from node_pool where node_id = $1`
	rows, err = s.db.QueryContext(ctx, queryPools, nodeId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pool string

		if err := rows.Scan(&pool); err != nil {
			return nil, nil, err
		}
		pools = append(pools, pool)
	}

	return instanceTypes, pools, nil
}

func (s *FleetAdminService) ReportNodeStatistics(ctx context.Context, req *pb.ReportNodeStatisticsRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.ReportNodeStatistics").Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("Request", "req", req)
	if err := s.processNodes(ctx, req.SchedulerNodeStatistics); err != nil {
		logger.Error(err, "error reporting node statistics")
		return &emptypb.Empty{}, fmt.Errorf("error reporting node statistics")
	}
	return &emptypb.Empty{}, nil
}

func (s *FleetAdminService) processNodes(ctx context.Context, schedulerNodeStatistics []*pb.SchedulerNodeStatistics) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.processNodes").Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	// Start a single transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "Error starting transaction")
		return err
	}

	// Roll back the transaction if any error
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			logger.Error(rollbackErr, "Error rolling back transaction")
		}
	}()

	for _, stats := range schedulerNodeStatistics {
		logger.Info("SchedulerNodeStatistics", "ClusterId", stats.SchedulerNode.ClusterId, "Region", stats.SchedulerNode.Region,
			"AvailabilityZone", stats.SchedulerNode.AvailabilityZone, "NodeName", stats.SchedulerNode.NodeName)

		var nodeWrapper NodeWrapper
		rows, queryErr := tx.QueryContext(ctx, SelectNodeQuery, stats.SchedulerNode.ClusterId,
			stats.SchedulerNode.Region, stats.SchedulerNode.AvailabilityZone, stats.SchedulerNode.NodeName, stats.SchedulerNode.Namespace)
		if queryErr != nil {
			return queryErr
		}

		// Check if node already exists
		if rows.Next() {
			if scanErr := rows.Scan(&nodeWrapper.nodeId, &nodeWrapper.region, &nodeWrapper.availabilityZone,
				&nodeWrapper.clusterId, &nodeWrapper.namespace, &nodeWrapper.nodeName,
				&nodeWrapper.overrideInstanceTypes, &nodeWrapper.overridePools); scanErr != nil {
				if closeErr := rows.Close(); closeErr != nil {
					logger.Error(closeErr, "Failed to close rows after scan error")
				}
				return scanErr
			}
			logger.Info("Existing node found", "node_id", nodeWrapper.nodeId)
		} else {
			// New node setup
			nodeWrapper = NodeWrapper{
				region:           stats.SchedulerNode.Region,
				availabilityZone: stats.SchedulerNode.AvailabilityZone,
				clusterId:        stats.SchedulerNode.ClusterId,
				namespace:        stats.SchedulerNode.Namespace,
				nodeName:         stats.SchedulerNode.NodeName,
			}

			nodeId, insertErr := s.insertNewNode(ctx, tx, &nodeWrapper)
			if insertErr != nil {
				return insertErr
			}
			nodeWrapper.nodeId = nodeId
			logger.Info("New node inserted", "node_id", nodeId)
		}
		if closeErr := rows.Close(); closeErr != nil {
			logger.Error(closeErr, "Failed to close rows after scan")
		}

		// Update node wrapper and apply policies
		if updateErr := s.updateNodeWrapper(ctx, tx, &nodeWrapper); updateErr != nil {
			return updateErr
		}
		logger.Info("NodeWrapper before policies", "nodeWrapper", nodeWrapper)

		if policyErr := s.applySimpleNodePolicies(ctx, &nodeWrapper, stats); policyErr != nil {
			return policyErr
		}
		logger.Info("NodeWrapper after policies", "nodeWrapper", nodeWrapper)

		// Upsert node pools and instance-type data
		if upsertErr := s.upsertNodePools(ctx, tx, &nodeWrapper); upsertErr != nil {
			return upsertErr
		}

		if upsertErr := s.upsertInstanceTypes(ctx, tx, &nodeWrapper); upsertErr != nil {
			return upsertErr
		}

		// Upsert node statistics data
		if upsertErr := s.upsertNodeStatsData(ctx, tx, &nodeWrapper, stats); upsertErr != nil {
			return upsertErr
		}
	}

	// Handle deletion of unreported nodes
	if deleteErr := s.DeleteUnreportedNodes(ctx, tx, schedulerNodeStatistics); deleteErr != nil {
		return deleteErr
	}

	// Commit the transaction
	logger.Info("Committing the transaction")
	if commitErr := tx.Commit(); commitErr != nil {
		logger.Error(commitErr, "Error committing transaction for nodes")
		return commitErr
	}
	logger.Info("Transaction committed successfully")

	return nil
}

func (s *FleetAdminService) deleteAndLogTx(ctx context.Context, tx *sql.Tx, query string, nodeID string, tableName string) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.deleteAndLogTx").Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	result, err := tx.ExecContext(ctx, query, nodeID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(err, "Failed to retrieve rows affected", "table", tableName, "NodeID", nodeID)
		return err
	}

	logger.Info(fmt.Sprintf("Deleted from %s", tableName), "NodeID", nodeID, "RowsAffected", rowsAffected)

	return nil
}

type NodeKey struct {
	ClusterID        string
	Region           string
	AvailabilityZone string
	NodeName         string
}

func (s *FleetAdminService) DeleteUnreportedNodes(ctx context.Context, tx *sql.Tx, schedulerNodeStatistics []*pb.SchedulerNodeStatistics) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.DeleteUnreportedNodes").Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	// Create a set of reported node identifiers
	reportedNodes := make(map[NodeKey]struct{})
	for _, stats := range schedulerNodeStatistics {
		nodeKey := NodeKey{
			ClusterID:        stats.SchedulerNode.ClusterId,
			Region:           stats.SchedulerNode.Region,
			AvailabilityZone: stats.SchedulerNode.AvailabilityZone,
			NodeName:         stats.SchedulerNode.NodeName,
		}
		reportedNodes[nodeKey] = struct{}{}
	}

	logger.Info("ReportedNodeKeys", "reportedNodeKeys", reportedNodes)

	// Collect IDs of unreported nodes
	var unreportedNodeIDs []string
	rows, err := tx.QueryContext(ctx, "SELECT node_id, cluster_id, region, availability_zone, node_name FROM node")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, clusterID, region, availabilityZone, nodeName string
		if err := rows.Scan(&nodeID, &clusterID, &region, &availabilityZone, &nodeName); err != nil {
			return err
		}

		nodeKey := NodeKey{
			ClusterID:        clusterID,
			Region:           region,
			AvailabilityZone: availabilityZone,
			NodeName:         nodeName,
		}

		if _, exists := reportedNodes[nodeKey]; !exists {
			unreportedNodeIDs = append(unreportedNodeIDs, nodeID)
		}
	}

	// Ensure rows are closed before starting deletion operations
	if err := rows.Close(); err != nil {
		return err
	}

	// Delete data for unreported nodes
	for _, nodeID := range unreportedNodeIDs {
		logger.Info("Deleting unreported node and its data", "NodeID", nodeID)

		tables := []struct {
			query     string
			tableName string
		}{
			{"DELETE FROM node_pool WHERE node_id = $1", "node_pool"},
			{"DELETE FROM node_instance_type WHERE node_id = $1", "node_instance_type"},
			{"DELETE FROM node_stats WHERE node_id = $1", "node_stats"},
			{"DELETE FROM node_instance_type_stats WHERE node_id = $1", "node_instance_type_stats"},
			{"DELETE FROM node WHERE node_id = $1", "node"},
		}

		for _, table := range tables {
			if err := s.deleteAndLogTx(ctx, tx, table.query, nodeID, table.tableName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *FleetAdminService) updateNodeWrapper(ctx context.Context, tx *sql.Tx, nodeWrapper *NodeWrapper) error {
	var err error
	nodeWrapper.instanceTypes, err = s.getNodeInstanceTypesFromDB(ctx, tx, nodeWrapper.nodeId)
	if err != nil {
		return err
	}

	nodeWrapper.poolIds, err = s.getNodePoolIdsFromDB(ctx, tx, nodeWrapper.nodeId)
	if err != nil {
		return err
	}
	return nil
}

func (s *FleetAdminService) insertNewNode(ctx context.Context, tx *sql.Tx, nodeWrapper *NodeWrapper) (int, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.insertNewNode").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("Inserting the node in the Fleet Admin DB", "nodeName", nodeWrapper.nodeName)

	var nodeId int
	err := tx.QueryRowContext(ctx, InsertNodeQuery, nodeWrapper.region, nodeWrapper.availabilityZone,
		nodeWrapper.clusterId, nodeWrapper.namespace, nodeWrapper.nodeName, nodeWrapper.overrideInstanceTypes, nodeWrapper.overridePools).Scan(&nodeId)

	if err != nil {
		return -1, err
	}

	logger.Info("Inserted node", "node_id", nodeId)
	nodeWrapper.nodeId = nodeId
	return nodeId, nil
}

func (s *FleetAdminService) upsertNodePools(ctx context.Context, tx *sql.Tx, nodeWrapper *NodeWrapper) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.upsertNodePools").Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	// Handle node pools
	if len(nodeWrapper.poolIds) == 0 {
		logger.Info("Deleting all poolIds for node", "node_id", nodeWrapper.nodeId)
		if _, err := tx.ExecContext(ctx, DeleteAllNodePoolQuery, nodeWrapper.nodeId); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, DeleteNodePoolQuery, nodeWrapper.nodeId, pq.Array(nodeWrapper.poolIds)); err != nil {
			return err
		}
		for _, poolId := range nodeWrapper.poolIds {
			// Handling the case where effective NodePools are determined based on node statistics.
			// Ensure a new pool is not created if one already exists.
			logger.Info("UpsertPoolFromNodeStatsQuery", "poolId", poolId)
			_, err := tx.ExecContext(ctx, UpsertPoolFromNodeStatsQuery, poolId, poolId, "")
			if err != nil {
				return err
			}

			logger.Info("UpsertNodePoolQuery", "node_id", nodeWrapper.nodeId, "poolId", poolId)
			if _, err := tx.ExecContext(ctx, UpsertNodePoolQuery, nodeWrapper.nodeId, poolId); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *FleetAdminService) upsertInstanceTypes(ctx context.Context, tx *sql.Tx, nodeWrapper *NodeWrapper) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.upsertInstanceTypes").Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	// Handle instance types
	if len(nodeWrapper.instanceTypes) == 0 {
		logger.Info("Deleting all instance types for node", "node_id", nodeWrapper.nodeId)
		if _, err := tx.ExecContext(ctx, DeleteAllNodeInstanceTypeQuery, nodeWrapper.nodeId); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, DeleteNodeInstanceTypeQuery, nodeWrapper.nodeId, pq.Array(nodeWrapper.instanceTypes)); err != nil {
			return err
		}
		for _, instanceType := range nodeWrapper.instanceTypes {
			logger.Info("UpsertNodeInstanceTypeQuery", "node_id", nodeWrapper.nodeId, "instanceType", instanceType)
			if _, err := tx.ExecContext(ctx, UpsertNodeInstanceTypeQuery, nodeWrapper.nodeId, instanceType); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *FleetAdminService) upsertNodeStatsData(ctx context.Context, tx *sql.Tx, nodeWrapper *NodeWrapper, stats *pb.SchedulerNodeStatistics) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.upsertNodeStatsData").Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")
	var instanceCategory string
	if len(stats.InstanceTypeStatistics) > 0 {
		instanceCategory = stats.InstanceTypeStatistics[0].InstanceCategory
	} else {
		logger.Info("instanceCategory is nil", "NodeName", stats.SchedulerNode.NodeName)
	}

	// Proceed with database operation
	if _, err := tx.ExecContext(ctx, UpsertNodeStatisticsQuery,
		nodeWrapper.nodeId, time.Now(),
		stats.SchedulerNode.SourceGvr.Group,
		stats.SchedulerNode.SourceGvr.Version,
		stats.SchedulerNode.SourceGvr.Resource,
		instanceCategory,
		stats.SchedulerNode.Partition,
		stats.SchedulerNode.ClusterGroup,
		stats.SchedulerNode.NetworkMode,
		stats.NodeResources.FreeMilliCPU,
		stats.NodeResources.UsedMilliCPU,
		stats.NodeResources.FreeMemoryBytes,
		stats.NodeResources.UsedMemoryBytes,
		stats.NodeResources.FreeGPU,
		stats.NodeResources.UsedGPU); err != nil {
		return err
	}

	logger.Info("UpsertNodeInstanceTypeStatsQuery in the Fleet Admin DB")
	for _, instanceTypeStats := range stats.InstanceTypeStatistics {
		logger.Info("UpsertNodeInstanceTypeStatsQuery", "instanceTypeStats", instanceTypeStats)
		if _, err := tx.ExecContext(ctx, UpsertNodeInstanceTypeStatsQuery,
			nodeWrapper.nodeId,
			instanceTypeStats.InstanceType,
			instanceTypeStats.RunningInstances,
			instanceTypeStats.MaxNewInstances); err != nil {
			return err
		}
	}
	return nil
}

func (s *FleetAdminService) getNodeInstanceTypesFromDB(ctx context.Context, tx *sql.Tx, nodeId int) ([]string, error) {
	rows, err := tx.QueryContext(ctx, SelectInstanceTypeQuery, nodeId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instanceTypes []string
	for rows.Next() {
		var instanceType string
		if err := rows.Scan(&instanceType); err != nil {
			return nil, err
		}
		instanceTypes = append(instanceTypes, instanceType)
	}
	return instanceTypes, nil
}

func (s *FleetAdminService) getNodePoolIdsFromDB(ctx context.Context, tx *sql.Tx, nodeId int) ([]string, error) {
	rows, err := tx.QueryContext(ctx, SelectPoolIdQuery, nodeId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var poolIds []string
	for rows.Next() {
		var poolId string
		if err := rows.Scan(&poolId); err != nil {
			return nil, err
		}
		poolIds = append(poolIds, poolId)
	}
	return poolIds, nil
}

func (s *FleetAdminService) applySimpleNodePolicies(ctx context.Context, nodeWrapper *NodeWrapper, stats *pb.SchedulerNodeStatistics) error {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminService.applySimpleNodePolicies").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("apply SimpleNode Policies", "nodeWrapper.nodeId", nodeWrapper.nodeId)
	// If override instance types are not set, calculate the effective instance types based on the node statistics.
	// To simplify the migration of regions to Fleet Admin V2, this will simply copy the currently effective instance types from the K8s resource labels
	// to the Fleet Admin DB.
	// Once all regions have been migrated, this logic can be changed
	// to calculate the effective instance types based on other attributes
	// such as the CPU, GPU, memory, etc..
	if !nodeWrapper.overrideInstanceTypes {
		if len(stats.InstanceTypeStatistics) > 0 {
			logger.Info("Calculating the effective instance types based on node statistics")
			nodeWrapper.instanceTypes = make([]string, len(stats.InstanceTypeStatistics))
			for i, instanceTypeStats := range stats.InstanceTypeStatistics {
				nodeWrapper.instanceTypes[i] = instanceTypeStats.InstanceType
			}
		} else {
			logger.Info("Instance types from node statistics are empty")
		}
	}

	if !nodeWrapper.overridePools {
		if len(stats.SchedulerNode.ComputeNodePools) > 0 {
			logger.Info("Calculating the effective nodePools based on node statistics")
			nodeWrapper.poolIds = append([]string{}, stats.SchedulerNode.ComputeNodePools...)
		} else {
			logger.Info("ComputeNodePools from node statistics are empty")
		}
	}

	nodeWrapper.clusterGroup = stats.SchedulerNode.ClusterGroup
	nodeWrapper.networkMode = stats.SchedulerNode.NetworkMode
	return nil
}
