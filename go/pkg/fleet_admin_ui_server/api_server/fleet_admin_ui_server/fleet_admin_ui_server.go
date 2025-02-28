// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package fleet_admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin_ui_server/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/jackc/pgconn"

	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type FleetAdminUIService struct {
	pb.UnimplementedFleetAdminUIServiceServer
	db  *sql.DB
	cfg config.Config
}

type NodeWrapper struct {
	nodeId        int
	instanceTypes []string
	poolIds       []string
}

func NewFleetAdminUIService(
	db *sql.DB,
	config config.Config,

) (*FleetAdminUIService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &FleetAdminUIService{
		db:  db,
		cfg: config,
	}, nil
}

func (s *FleetAdminUIService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("FleetAdminUIService.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (s *FleetAdminUIService) AddCloudAccountToComputeNodePool(ctx context.Context, req *pb.AddCloudAccountToComputeNodePoolRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.AddCloudAccountToComputeNodePool").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		// Validate input.
		if err := cloudaccount.CheckValidId(req.CloudAccountId); err != nil {
			return nil, err
		}
		if err := s.CheckValidPoolId(req.PoolId); err != nil {
			return nil, err
		}
		if err := s.ValidateCreateAdmin(req.CreateAdmin); err != nil {
			return nil, err
		}

		queryAdd := `insert into pool_cloud_account (pool_id, cloud_account_id, create_date, create_admin) values ($1, $2, current_timestamp, $3)`
		_, err := s.db.ExecContext(ctx, queryAdd, req.PoolId, req.CloudAccountId, req.CreateAdmin)
		if err != nil {

			// Check if the error is a primary key violation (unique constraint error)
			// 23505 is the error code for unique constraint violation in PostgreSQL
			if pkError, ok := err.(*pgconn.PgError); ok && pkError.Code == manageddb.KErrUniqueViolation {
				return nil, status.Errorf(codes.AlreadyExists, "cloudaccount %s has already been assigned to pool %s", req.CloudAccountId, req.PoolId)
			}
			return nil, fmt.Errorf("insert into pool_cloud_account: %w", err)
		}
		resp := &emptypb.Empty{}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *FleetAdminUIService) DeleteCloudAccountFromComputeNodePool(ctx context.Context, req *pb.DeleteCloudAccountFromComputeNodePoolRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.DeleteCloudAccountFromComputeNodePool").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		// Validate input.
		if err := cloudaccount.CheckValidId(req.CloudAccountId); err != nil {
			return nil, err
		}
		if err := s.CheckValidPoolId(req.PoolId); err != nil {
			return nil, err
		}

		queryDelete := `delete from pool_cloud_account where cloud_account_id = $1 and pool_id = $2`
		_, err := s.db.ExecContext(ctx, queryDelete, req.CloudAccountId, req.PoolId)
		if err != nil {
			return nil, fmt.Errorf("delete from pool_cloud_account: %w", err)
		}

		resp := &emptypb.Empty{}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *FleetAdminUIService) SearchNodes(ctx context.Context, req *pb.SearchNodesRequest) (*pb.SearchNodesResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.SearchNodes").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.SearchNodesResponse, error) {
		var rows *sql.Rows
		var err error
		query := `
				select
				node.node_id, node.node_name, node.region, node.availability_zone, node.cluster_id, node.namespace,
				node_stats.instance_category, node_stats.free_millicpu, node_stats.used_millicpu, 
				node_stats.free_memory_bytes, node_stats.used_memory_bytes, node_stats.free_gpu, node_stats.used_gpu
				from node
				left join node_stats on node_stats.node_id = node.node_id
			`
		// sorting to ensure consistent ordering
		orderBy := ` order by node.region, node.availability_zone, node.cluster_id, node.namespace, node.node_name`

		if req.PoolId != nil && *req.PoolId != "" {
			query += ` where node.node_id in (select node_id from node_pool where pool_id = $1)` + orderBy
			rows, err = s.db.QueryContext(ctx, query, req.PoolId)
		} else {
			query += orderBy
			rows, err = s.db.QueryContext(ctx, query)
		}

		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var computeNodes []*pb.SearchNodesResponse_ComputeNode
		for rows.Next() {
			var nodeId int
			var nodeName string
			var region string
			var clusterId string
			var namespace string
			var availabilityZone string
			var instanceCategory string
			var freeMillicpu int32
			var usedMillicpu int32
			var freeMemoryBytes int64
			var usedMemoryBytes int64
			var freeGpu int32
			var usedGpu int32

			if err := rows.Scan(&nodeId, &nodeName, &region, &availabilityZone, &clusterId, &namespace,
				&instanceCategory, &freeMillicpu, &usedMillicpu, &freeMemoryBytes, &usedMemoryBytes, &freeGpu, &usedGpu); err != nil {
				return nil, err
			}
			instanceTypes, pools, err := s.GetInstanceTypesAndPools(ctx, nodeId)
			if err != nil {
				return nil, err
			}
			percentageResourcesUsed, err := PercentageResourcesUsed(ctx, instanceCategory, freeMillicpu, usedMillicpu, freeMemoryBytes, usedMemoryBytes, freeGpu, usedGpu)
			if err != nil {
				return nil, err
			}

			computeNode := &pb.SearchNodesResponse_ComputeNode{
				NodeName:                nodeName,
				AvailabilityZone:        availabilityZone,
				Region:                  region,
				ClusterId:               clusterId,
				Namespace:               namespace,
				InstanceTypes:           instanceTypes,
				PoolIds:                 pools,
				PercentageResourcesUsed: percentageResourcesUsed,
				NodeId:                  int32(nodeId),
			}
			computeNodes = append(computeNodes, computeNode)
		}
		resp := &pb.SearchNodesResponse{
			ComputeNodes: computeNodes,
		}
		return resp, nil
	}()

	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *FleetAdminUIService) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.UpdateNode").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {

		err := s.ValidateUpdateNodeRequest(req)
		if err != nil {
			return nil, err
		}

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelDefault,
		})
		if err != nil {
			return nil, fmt.Errorf("BeginTx: %w", err)
		}
		defer tx.Rollback()

		nodeWrapper := NodeWrapper{
			nodeId: int(req.NodeId),
		}
		//Instance Type
		queryUpdateNodeInstanceType := `update node set override_instance_types = $1 where node_id = $2`
		if req.InstanceTypesOverride != nil && req.InstanceTypesOverride.OverridePolicies {
			_, err = tx.ExecContext(ctx, queryUpdateNodeInstanceType, true, req.NodeId)
			nodeWrapper.instanceTypes = req.InstanceTypesOverride.OverrideValues
			if upsertErr := s.upsertInstanceTypes(ctx, tx, &nodeWrapper); upsertErr != nil {
				return nil, upsertErr
			}
		} else {
			_, err = tx.ExecContext(ctx, queryUpdateNodeInstanceType, false, req.NodeId)
		}
		if err != nil {
			return nil, fmt.Errorf("update override_instance_types into node: %w", err)
		}

		//Node Pool
		queryUpdateNodePool := `update node set override_pools = $1 where node_id = $2`
		if req.ComputeNodePoolsOverride != nil && req.ComputeNodePoolsOverride.OverridePolicies {
			_, err = tx.ExecContext(ctx, queryUpdateNodePool, true, req.NodeId)
			nodeWrapper.poolIds = req.ComputeNodePoolsOverride.OverrideValues
			if upsertErr := s.upsertNodePools(ctx, tx, &nodeWrapper); upsertErr != nil {
				return nil, upsertErr
			}
		} else {
			_, err = tx.ExecContext(ctx, queryUpdateNodePool, false, req.NodeId)
		}
		if err != nil {
			return nil, fmt.Errorf("update override_pools into node: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
		return &emptypb.Empty{}, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	if err != nil {
		return nil, fmt.Errorf("UpdateNode failed")
	}
	return resp, nil
}

func (s *FleetAdminUIService) SearchComputeNodePoolsForNodeAdmin(ctx context.Context, req *emptypb.Empty) (*pb.SearchComputeNodePoolsForNodeAdminResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.SearchComputeNodePoolsForNodeAdmin").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.SearchComputeNodePoolsForNodeAdminResponse, error) {
		query := `
			select
				p.pool_id,
				p.pool_name,
				p.pool_account_manager_ags_role,
				COUNT(np.node_id) as number_of_nodes
			FROM pool p
			left join node_pool np on p.pool_id = np.pool_id
			group by p.pool_id, p.pool_name, p.pool_account_manager_ags_role
		`
		rows, err := s.db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		var computeNodePools []*pb.ComputeNodePool
		for rows.Next() {
			var poolId string
			var poolName string
			var poolAccountManagerAgsRole string
			var numberOfNodes int32

			if err := rows.Scan(&poolId, &poolName, &poolAccountManagerAgsRole, &numberOfNodes); err != nil {
				return nil, err
			}

			computeNodePool := &pb.ComputeNodePool{
				PoolId:                    poolId,
				PoolName:                  poolName,
				PoolAccountManagerAgsRole: poolAccountManagerAgsRole,
				NumberOfNodes:             numberOfNodes,
			}
			computeNodePools = append(computeNodePools, computeNodePool)
		}
		resp := &pb.SearchComputeNodePoolsForNodeAdminResponse{
			ComputeNodePools: computeNodePools,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *FleetAdminUIService) PutComputeNodePool(ctx context.Context, req *pb.PutComputeNodePoolRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.PutComputeNodePool").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		// Validate input
		if err := s.CheckValidPoolId(req.PoolId); err != nil {
			return nil, err
		}

		// When the record is initially created, the poolId is used. For subsequent updates, only the pool name and the Account Manager AGS role can be modified.
		query := `
			insert into pool (pool_id, pool_name, pool_account_manager_ags_role)
			values ($1, $2, $3)
			on conflict (pool_id) 
			do update set
				pool_name = excluded.pool_name,
				pool_account_manager_ags_role = excluded.pool_account_manager_ags_role
		`
		_, err := s.db.ExecContext(ctx, query, req.PoolId, req.PoolName, req.PoolAccountManagerAgsRole)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to upsert record: %v", err)
		}

		resp := &emptypb.Empty{}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *FleetAdminUIService) SearchComputeNodePoolsForPoolAccountManager(ctx context.Context, req *emptypb.Empty) (*pb.SearchComputeNodePoolsForPoolAccountManagerResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.SearchComputeNodePoolsForPoolAccountManager").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.SearchComputeNodePoolsForPoolAccountManagerResponse, error) {
		poolAccountManagerAgsRoles, err := s.ExtractPoolAccountManagerAgsRolesFromToken(ctx)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Failed to get ags roles for pool account manager from JWT token : %v", err)
		}
		query := `select pool_id, pool_name from pool where pool_account_manager_ags_role = any($1::text[])`
		rows, err := s.db.QueryContext(ctx, query, poolAccountManagerAgsRoles)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var computeNodePoolsForPoolAccountManager []*pb.ComputeNodePoolForPoolAccountManager
		for rows.Next() {
			var poolId string
			var poolName string
			if err := rows.Scan(&poolId, &poolName); err != nil {
				return nil, nil
			}
			computeNodePoolForPoolAccountManager := &pb.ComputeNodePoolForPoolAccountManager{
				PoolId:   poolId,
				PoolName: poolName,
			}

			computeNodePoolsForPoolAccountManager = append(computeNodePoolsForPoolAccountManager, computeNodePoolForPoolAccountManager)
		}

		resp := &pb.SearchComputeNodePoolsForPoolAccountManagerResponse{
			ComputeNodePoolsForPoolAccountManager: computeNodePoolsForPoolAccountManager,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *FleetAdminUIService) SearchCloudAccountsForComputeNodePool(ctx context.Context, req *pb.SearchCloudAccountsForComputeNodePoolRequest) (*pb.SearchCloudAccountsForComputeNodePoolResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.SearchCloudAccountsForComputeNodePool").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.SearchCloudAccountsForComputeNodePoolResponse, error) {
		// Validate input
		if err := s.CheckValidPoolId(req.PoolId); err != nil {
			return nil, err
		}
		querySearch := `select pool_id, cloud_account_id, create_admin from pool_cloud_account where pool_id = $1`

		rows, err := s.db.QueryContext(ctx, querySearch, req.PoolId)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var cloudAccountsForComputeNodePool []*pb.CloudAccountForComputeNodePool
		for rows.Next() {
			var poolId string
			var cloudAccountId string
			var createAdmin string

			if err := rows.Scan(&poolId, &cloudAccountId, &createAdmin); err != nil {
				return nil, err
			}
			cloudAccountForComputeNodePool := &pb.CloudAccountForComputeNodePool{
				CloudAccountId: cloudAccountId,
				PoolId:         poolId,
				CreateAdmin:    createAdmin,
			}
			cloudAccountsForComputeNodePool = append(cloudAccountsForComputeNodePool, cloudAccountForComputeNodePool)
		}

		resp := &pb.SearchCloudAccountsForComputeNodePoolResponse{
			CloudAccountsForComputeNodePool: cloudAccountsForComputeNodePool,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *FleetAdminUIService) GetInstanceTypesAndPools(ctx context.Context, nodeId int) ([]string, []string, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.GetInstanceTypesAndPools").Start()
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

func (s *FleetAdminUIService) CheckIfCloudAccountAlreadyAssignedToPool(ctx context.Context, poolId string, cloudAccountId string) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.CheckIfCloudAccountAlreadyAssignedToPool").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	query := `select exists (select 1 from pool_cloud_account where pool_id = $1 and cloud_account_id = $2)`
	var exists bool
	err := s.db.QueryRowContext(ctx, query, poolId, cloudAccountId).Scan(&exists)
	if err != nil {
		return exists, status.Errorf(codes.Internal, "record find failed: %v", err)
	}
	return exists, nil
}

func (s *FleetAdminUIService) ExtractPoolAccountManagerAgsRolesFromToken(ctx context.Context) ([]string, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.ExtractPoolAccountManagerAgsRolesFromToken").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	// Attempt to extract the headers.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unable to read the headers from the grpc request")
	}
	autheader := md.Get("authorization")
	if len(autheader) == 0 {
		return nil, fmt.Errorf("JWT token is not passed along with the request")
	}
	// Token validation is not perfomed here since it is already verified at the ingress.
	jwtToken := strings.Replace(autheader[0], "Bearer ", "", 1)
	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("error while parsing the JWT token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("unable to read claims from the JWT token")
	}

	allAGSRoles, ok := claims["roles"].([]interface{})
	if !ok || allAGSRoles == nil {
		return nil, fmt.Errorf("ags roles are not part of the JWT token")
	}

	poolAccountManagerAgsRoles := make([]string, 0)
	for _, agsRole := range allAGSRoles {
		if strings.Contains(agsRole.(string), "IDC.PoolAccountManager") {
			poolAccountManagerAgsRoles = append(poolAccountManagerAgsRoles, agsRole.(string))
		}
	}

	return poolAccountManagerAgsRoles, nil
}

func (s *FleetAdminUIService) CheckValidPoolId(name string) error {
	return s.CheckValidName(name, "PoolId")
}

func (s *FleetAdminUIService) CheckValidInstanceType(name string) error {
	return s.CheckValidName(name, "InstanceType")
}

func (s *FleetAdminUIService) CheckValidName(name, fieldName string) error {
	// The name (i.e. PoolId, InstanceType) will be used as part of a Kubernetes label in the format “x.cloud.intel.com/<name>”.
	// The complete label has a limit of 63 characters. Therefore, the name is limited to 42 characters.
	if len(name) > 42 {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("the %s must not exceed 42 characters", fieldName))
	}
	// Checking for valid characters. Refer https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set.
	pattern := `^[a-zA-Z0-9][a-zA-Z0-9._-]*[a-zA-Z0-9]$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(name) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid %s: make sure it begins and ends with an alphanumeric character, and only contains dashes, underscores, dots, and alphanumerics in between", fieldName))
	}
	return nil
}

func (s *FleetAdminUIService) ValidateCreateAdmin(createAdmin string) error {
	if createAdmin == "" {
		return status.Error(codes.InvalidArgument, "createAdmin cannot be empty")
	}
	_, err := mail.ParseAddress(createAdmin)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid createAdmin format")
	}
	return nil
}

func (s *FleetAdminUIService) upsertNodePools(ctx context.Context, tx *sql.Tx, nodeWrapper *NodeWrapper) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.upsertNodePools").Start()
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
			logger.Info("UpsertNodePoolQuery", "node_id", nodeWrapper.nodeId, "poolId", poolId)
			if _, err := tx.ExecContext(ctx, UpsertNodePoolQuery, nodeWrapper.nodeId, poolId); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *FleetAdminUIService) upsertInstanceTypes(ctx context.Context, tx *sql.Tx, nodeWrapper *NodeWrapper) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.upsertInstanceTypes").Start()
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

func (s *FleetAdminUIService) ValidateUpdateNodeRequest(updateNodeRequest *pb.UpdateNodeRequest) error {
	// Validate InstanceTypesOverride
	if updateNodeRequest.InstanceTypesOverride != nil {
		if err := s.ValidateOverrideValues(updateNodeRequest.InstanceTypesOverride.OverrideValues, "InstanceType"); err != nil {
			return err
		}
	}

	// Validate ComputeNodePoolsOverride
	if updateNodeRequest.ComputeNodePoolsOverride != nil {
		if err := s.ValidateOverrideValues(updateNodeRequest.ComputeNodePoolsOverride.OverrideValues, "PoolId"); err != nil {
			return err
		}
	}

	return nil
}

func (s *FleetAdminUIService) ValidateOverrideValues(values []string, fieldName string) error {
	for _, value := range values {
		if fieldName == "PoolId" {
			if err := s.CheckValidPoolId(value); err != nil {
				return err
			}
		} else if fieldName == "InstanceType" {
			if err := s.CheckValidInstanceType(value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *FleetAdminUIService) SearchInstanceTypeStatsForNode(ctx context.Context, req *pb.SearchInstanceTypeStatsForNodeRequest) (*pb.SearchInstanceTypeStatsForNodeResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FleetAdminUIService.SearchInstanceTypeStatsForNode").Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*pb.SearchInstanceTypeStatsForNodeResponse, error) {
		var stmt *sql.Stmt
		var rows *sql.Rows
		var err error

		if req.NodeId > 0 {
			stmt, err = s.db.PrepareContext(ctx, `select node_id, instance_type, running_instances, max_new_instances from node_instance_type_stats where node_id=$1 order by node_id`)
			if err != nil {
				return nil, err
			}
			defer stmt.Close()
			rows, err = stmt.QueryContext(ctx, req.NodeId)
		} else {
			stmt, err = s.db.PrepareContext(ctx, `select node_id, instance_type, running_instances, max_new_instances from node_instance_type_stats order by node_id`)
			if err != nil {
				return nil, err
			}
			defer stmt.Close()
			rows, err = stmt.QueryContext(ctx)
		}
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Efficiently map node IDs to their instance type stats
		nodeInstanceTypeStatsMap := make(map[int32]*pb.NodeInstanceTypeStats)

		for rows.Next() {
			var nodeId int32
			var instanceType string
			var runningInstances int32
			var maxNewInstances int32

			if err := rows.Scan(&nodeId, &instanceType, &runningInstances, &maxNewInstances); err != nil {
				return nil, err
			}

			// Get existing NodeInstanceTypeStats or create a new one
			nodeStats, ok := nodeInstanceTypeStatsMap[nodeId]
			if !ok {
				nodeStats = &pb.NodeInstanceTypeStats{
					NodeId: nodeId,
				}
				nodeInstanceTypeStatsMap[nodeId] = nodeStats
			}

			nodeStats.InstanceTypeStats = append(nodeStats.InstanceTypeStats, &pb.InstanceTypeStats{
				InstanceType:     instanceType,
				RunningInstances: runningInstances,
				MaxNewInstances:  maxNewInstances,
			})
		}

		// Convert the map values to a slice
		var nodeInstanceTypeStats []*pb.NodeInstanceTypeStats
		for _, stats := range nodeInstanceTypeStatsMap {
			nodeInstanceTypeStats = append(nodeInstanceTypeStats, stats)
		}

		resp := &pb.SearchInstanceTypeStatsForNodeResponse{
			NodeInstanceTypeStats: nodeInstanceTypeStats,
		}
		return resp, nil
	}()

	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}
