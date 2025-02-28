// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	storeSvcUtils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"go.opentelemetry.io/otel/attribute"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/database/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TrainingClusterServiceServer struct {
	pb.UnimplementedTrainingClusterServiceServer
	session *sql.DB
}

func NewTrainingClusterService(session *sql.DB) (*TrainingClusterServiceServer, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}
	return &TrainingClusterServiceServer{
		session: session,
	}, nil
}

func validateSlurmNodesSpec(slurmCluster *pb.Cluster) error {
	nodeCount := map[pb.NodeRole]int{
		pb.NodeRole_JUPYTERHUB_NODE: 0,
		pb.NodeRole_COMPUTE_NODE:    0,
		pb.NodeRole_CONTROLLER_NODE: 0,
		pb.NodeRole_LOGIN_NODE:      0,
	}

	for _, instanceRequest := range slurmCluster.GetNodes() {
		nodeCount[instanceRequest.Role]++
	}

	// Check for a required node instances
	if nodeCount[pb.NodeRole_JUPYTERHUB_NODE] == 0 {
		return status.Errorf(codes.InvalidArgument, "missing jupyterhub-node role")
	}
	if nodeCount[pb.NodeRole_COMPUTE_NODE] == 0 {
		return status.Errorf(codes.InvalidArgument, "missing compute-node role")
	}
	if nodeCount[pb.NodeRole_CONTROLLER_NODE] == 0 {
		return status.Errorf(codes.InvalidArgument, "missing controller-node role")
	}
	if nodeCount[pb.NodeRole_LOGIN_NODE] == 0 {
		return status.Errorf(codes.InvalidArgument, "missing login-node role")
	}

	return nil
}

func validateSlurmStorageSpec(slurmCluster *pb.Cluster) error {
	storageNodes := slurmCluster.GetStorageNodes()

	for idx, storageSpecReq := range storageNodes {
		// Validate each storage node name spec and its length
		if err := storeSvcUtils.ValidateInstanceName(storageSpecReq.Name); err != nil {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid storage name %s. Node index errored on %d", storageSpecReq.Name, idx+1))
		}
		if len(storageSpecReq.Name) <= 0 || len(storageSpecReq.Name) > 32 {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("storage node name #%d exceeds 32 character limit: %s", idx+1, storageSpecReq.Name))
		}

		// Validate each storage node capacity spec
		if storeSvcUtils.ParseFileSizeInGB(storageSpecReq.Capacity) <= 0 {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid storage capacity %s. Node name errored on %s", storageSpecReq.Capacity, storageSpecReq.Name))
		}
	}

	return nil
}

func validateVNetSpec(slurmCluster *pb.Cluster) error {
	vnetSpec := slurmCluster.GetSpec()

	if (vnetSpec.Region) == "" || len(vnetSpec.Region) <= 0 || len(vnetSpec.Region) > 32 {
		return status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid vnet region: %s", vnetSpec.Region))
	}
	if (vnetSpec.AvailabilityZone) == "" || len(vnetSpec.AvailabilityZone) <= 0 || len(vnetSpec.AvailabilityZone) > 32 {
		return status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid vnet availabilityzone: %s", vnetSpec.AvailabilityZone))
	}
	if (vnetSpec.PrefixLength) <= 0 {
		return status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid vnet PrefixLength %d.", vnetSpec.PrefixLength))
	}

	return nil
}

func (svc *TrainingClusterServiceServer) Create(ctx context.Context, in *pb.SlurmClusterCreateRequest) (*pb.SlurmClusterCreateResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TrainingClusterServiceServer.Create").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("Create", "entering cluster record search", in)

	dbSession := svc.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	// Validate inputs
	if in.Cluster == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing cluster")
	}
	if in.Cluster.Nodes == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing slurm cluster node specs")
	}
	if err := validateSlurmNodesSpec(in.Cluster); err != nil {
		return nil, err
	}
	if err := validateSlurmStorageSpec(in.Cluster); err != nil {
		return nil, err
	}
	if err := validateVNetSpec(in.Cluster); err != nil {
		return nil, err
	}

	// Validate cloudAccounntId
	if err := cloudaccount.CheckValidId(in.CloudAccountId); err != nil {
		return nil, err
	}

	clusterId := uuid.NewString()

	span.SetAttributes(attribute.String("clusterId", clusterId))

	if err := query.CreateClusterState(ctx, dbSession, clusterId, in); err != nil {
		logger.Error(err, "error creating cluster state", "clusterId", clusterId)
		return nil, err
	}

	logger.Info("cluster created", "clusterId", clusterId)

	return &pb.SlurmClusterCreateResponse{ClusterId: clusterId}, nil
}

func (svc *TrainingClusterServiceServer) Get(ctx context.Context, in *pb.SlurmClusterRequest) (*pb.Cluster, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TrainingClusterServiceServer.Get").WithValues("cloudAccountId", in.CloudAccountId, "clusterId", in.ClusterId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("Get", "entering cluster record search by ID", in)

	dbSession := svc.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	// Validate clusterID
	if _, err := uuid.Parse(in.ClusterId); err != nil {
		logger.Error(err, "invalid clusterId")
		return nil, status.Errorf(codes.InvalidArgument, "invalid clusterId provided")
	}

	// Validate cloudAccounntId
	if err := cloudaccount.CheckValidId(in.CloudAccountId); err != nil {
		logger.Error(err, "invalid cloudAccountId")
		return nil, err
	}

	clusterInfo, err := query.GetClusterByID(ctx, dbSession, in)
	if err != nil {
		logger.Error(err, "error getting cluster by ID")
		return nil, err
	}

	return clusterInfo, nil
}

func (svc *TrainingClusterServiceServer) List(ctx context.Context, in *pb.ClusterListOption) (*pb.SlurmClusterResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TrainingClusterServiceServer.List").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("List", "entering cluster record search by ID", in)

	dbSession := svc.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	// Validate cloudAccountId
	if err := cloudaccount.CheckValidId(in.CloudAccountId); err != nil {
		logger.Error(err, "invalid cloudAccountId")
		return nil, err
	}

	slurmClustersInfo, err := query.GetClustersByCloudAccount(ctx, dbSession, in)
	if err != nil {
		logger.Error(err, "error getting clusters with cloudAccountId")
		return nil, err
	}

	return slurmClustersInfo, nil
}
