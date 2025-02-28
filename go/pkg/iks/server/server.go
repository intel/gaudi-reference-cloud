// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	mathrand "math/rand"
	"strings"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/config"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"golang.org/x/crypto/ssh"
)

// Server is a cluster
type Server struct {
	pb.UnimplementedIksServer
	computeClient               pb.InstanceTypeServiceClient
	computeInstanceSvsClient    pb.InstanceServiceClient
	sshkeyClient                pb.SshPublicKeyServiceClient
	productcatalogServiceClient pb.ProductCatalogServiceClient
	vnetClient                  pb.VNetServiceClient
	session                     *sql.DB
	cfg                         config.Config
}

// NewIksService Initializes DB connection
func NewIksService(session *sql.DB,
	computegrpcClient pb.InstanceTypeServiceClient,
	sshkeyClient pb.SshPublicKeyServiceClient,
	vnetClient pb.VNetServiceClient,
	productcatalogServiceClient v1.ProductCatalogServiceClient,
	computeInstanceSvcClient pb.InstanceServiceClient,
	cfg config.Config) (*Server, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}
	return &Server{
		session:                     session,
		cfg:                         cfg,
		computeClient:               computegrpcClient,
		sshkeyClient:                sshkeyClient,
		vnetClient:                  vnetClient,
		productcatalogServiceClient: productcatalogServiceClient,
		computeInstanceSvsClient:    computeInstanceSvcClient,
	}, nil
}

// CreateNewCluster will create new cluster entries
func (c *Server) CreateNewCluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterCreateResponseForm, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, req.Name, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("no database connection found")
	}

	if err := validateClusterRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc create new cluster Request")
		return &pb.ClusterCreateResponseForm{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	// Get default cloudaccount
	defaultvalues, err := utils.GetDefaultValues(ctx, dbSession)
	if err != nil {
		logger.Error(err, "Error in get default values")
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("get default cloud account failed %s ", err)
	}

	// Generate SSH keys and upload to compute
	// generateSshKeyForCP()
	key, keypub, err := generateSshKeyForCP(ctx)
	if err != nil {
		logger.Error(err, "Error in generate ssh key for control plane")
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("Ssh key generation for control plane failed %s ", err)
	}

	sshpubkey, err := c.UploadSshkeysForCP(ctx, keypub, defaultvalues["cp_cloudaccountid"], req.Name)
	if err != nil {
		logger.Error(err, "Error in cluster create request")
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("Ssh key upload for control plane failed %s ", err)
	}

	res, err := query.CreateClusterRecord(ctx, dbSession, req, key, keypub, sshpubkey, c.cfg.EncryptionKeys)
	logger.Info("Clusterresponse", logkeys.Response, res)
	if err != nil {
		return &pb.ClusterCreateResponseForm{}, err
	}

	return res, nil
}

// GetCluster returns the cluster details
func (c *Server) GetCluster(ctx context.Context, req *pb.ClusterID) (*pb.ClusterResponseForm, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetCluster").WithValues(logkeys.ClusterId, req.Clusteruuid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.ClusterResponseForm{}, fmt.Errorf("no database connection found")
	}

	if err := validateClusterID(req); err != nil {
		logger.Error(err, "Error in validating grpc get cluster Request")
		return &pb.ClusterResponseForm{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.GetClusterRecord(ctx, dbSession, req)
	logger.Info("Clusterresponse", logkeys.Response, res)
	if err != nil {
		return &pb.ClusterResponseForm{}, err
	}
	return res, nil
}

// Getclusters returns list of clusters
func (c *Server) GetClusters(ctx context.Context, req *pb.IksCloudAccountId) (*pb.ClustersResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetClusters").WithValues(logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.ClustersResponse{}, fmt.Errorf("no database connection found")
	}

	if err := validateIksCloudAccountId(req); err != nil {
		logger.Error(err, "Error in validating grpc get clusters Request")
		return &pb.ClustersResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.GetClustersRecord(ctx, dbSession, req)
	if err != nil {
		logger.Error(err, "Error in get clusters request")
		return &pb.ClustersResponse{}, err
	}
	return res, nil
}

// PutCluster returns updated cluster
func (c *Server) PutCluster(ctx context.Context, req *pb.UpdateClusterRequest) (*pb.ClusterCreateResponseForm, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.PutCluster").WithValues(logkeys.ClusterId, req.GetClusteruuid(), logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("no database connection found")
	}

	if err := validateUpdateClusterRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc put cluster Request")
		return &pb.ClusterCreateResponseForm{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.PatchRecord(ctx, dbSession, req)
	logger.Info("Clusterresponse", logkeys.Response, res)
	if err != nil {
		return &pb.ClusterCreateResponseForm{}, err
	}
	return res, nil
}

func (c *Server) UpgradeCluster(ctx context.Context, req *pb.UpgradeClusterRequest) (*pb.ClusterStatus, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.UpgradeCluster").WithValues(logkeys.ClusterId, req.GetClusteruuid(), logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.ClusterStatus{}, fmt.Errorf("no database connection found")
	}

	if err := validateUpgradeClusterRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc upgrade cluster Request")
		return &pb.ClusterStatus{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.UpgradeCluster(ctx, dbSession, req)
	logger.Info("Clusterresponse", logkeys.Response, res)
	if err != nil {
		return &pb.ClusterStatus{}, err
	}
	return res, nil
}

func (c *Server) EnableClusterStorage(ctx context.Context, req *pb.ClusterStorageRequest) (*pb.ClusterStorageStatus, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.EnableClusterStorage").WithValues(logkeys.ClusterId, req.GetClusteruuid(), logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	returnError := &pb.ClusterStorageStatus{}

	if c.session == nil {
		return returnError, fmt.Errorf("no database connection found")
	}

	if err := validateClusterStorageRequest(req); err != nil {
		logger.Error(err, "Error in validating enable cluster storage request")
		return &pb.ClusterStorageStatus{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	clusterStorageStatus := &pb.ClusterStorageStatus{}
	clusterStorageStatus, err := query.EnableClusterStorage(ctx, c.session, req)
	if err != nil {
		logger.Error(err, "Error in enable cluster storage request")
		return returnError, err
	}

	return clusterStorageStatus, nil
}

// Update Cluster Storage to increase size of the storage
func (c *Server) UpdateClusterStorage(ctx context.Context, req *pb.ClusterStorageUpdateRequest) (*pb.ClusterStorageStatus, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.UpdateClusterStorage").WithValues(logkeys.ClusterId, req.GetClusteruuid(), logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	failedFunction := "UpdateClusterStorage."
	returnError := &pb.ClusterStorageStatus{}

	if c.session == nil {
		return returnError, fmt.Errorf("no database connection found")
	}

	if err := validateClusterStorageUpdateRequest(req); err != nil {
		logger.Error(err, "Error in validating update cluster storage request")
		return &pb.ClusterStorageStatus{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	// Get Cluster ID
	clusterId, err := utils.ValidateClusterExistance(ctx, c.session, req.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", "Could validate cluster existance.")
	}
	if clusterId == -1 {
		return returnError, errors.New("Cluster not found: " + req.Clusteruuid)
	}

	// Get Current Size
	currentSize, err := utils.GetStorageSize(ctx, c.session, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetStorageSize", "Cannot Get storage size")
	}
	// parse storage size
	storageSize := utils.ParseFileSize(req.Storagesize)

	if storageSize == -1 {
		return returnError, errors.New("Invalid Storage Size")
	}

	if storageSize == currentSize {
		return returnError, errors.New("Requested storage size is equal to the current storage size. New storage size should be greater than current storage size.")
	} else if storageSize < currentSize {
		return returnError, errors.New("Requested storage size is less than current storage size. New storage size should be greater than current storage size.")
	}

	// Validate if the total storage size for cloudaccount is within the limits by checking it with the product catalog

	// Obtain the current total storage size for the cloudaccount
	totalStorageSize, err := utils.GetCloudAccountStorageSize(ctx, c.session, req.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetCloudAccountStorageSize", "Cannot Get total storage size for CloudAccount")
	}

	// Get diff
	diffSize := storageSize - currentSize
	// new total Storage Size for the cloud account
	totalStorageSize = totalStorageSize + diffSize

	// validate with product catalog
	fileProductName := "storage-file"
	valid, err := utils.ValidateWithProductCatalog(ctx, c.productcatalogServiceClient, fileProductName, totalStorageSize)

	if !valid {
		return returnError, err
	}

	clusterStorageStatus := &pb.ClusterStorageStatus{}
	clusterStorageStatus, err = query.UpdateClusterStorage(ctx, c.session, req)
	if err != nil {
		logger.Error(err, "Error in update cluster storage request")
		return returnError, err
	}

	return clusterStorageStatus, nil
}

// Deletecluster deletes cluster
func (c *Server) DeleteCluster(ctx context.Context, req *pb.ClusterID) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.DeleteCluster").WithValues(logkeys.ClusterId, req.GetClusteruuid(), logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}

	if err := validateClusterID(req); err != nil {
		logger.Error(err, "Error in validating grpc delete cluster Request")
		return &empty.Empty{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	_, err := query.DeleteRecord(ctx, dbSession, req)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

// GetStatus returns status of cluster
func (c *Server) GetClusterStatus(ctx context.Context, req *pb.ClusterID) (*pb.ClusterStatus, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetClusterStatus").WithValues(logkeys.ClusterId, req.GetClusteruuid(), logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.ClusterStatus{}, fmt.Errorf("no database connection found")
	}
	if err := validateClusterID(req); err != nil {
		logger.Error(err, "Error in validating grpc getting cluster status Request")
		return &pb.ClusterStatus{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.GetStatusRecord(ctx, dbSession, req)
	if err != nil {
		return &pb.ClusterStatus{}, err
	}
	return res, nil
}

// CreateNodeGroup creates new nodegroup
func (c *Server) CreateNodeGroup(ctx context.Context, req *pb.CreateNodeGroupRequest) (*pb.NodeGroupResponseForm, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.CreateNodegroup").WithValues(logkeys.NodeGroupName, req.Name, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.NodeGroupResponseForm{}, fmt.Errorf("no database connection found")
	}

	if err := validateCreateNodeGroupRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc Nodegroup create Request")
		return &pb.NodeGroupResponseForm{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	// Validate instance Types from compute
	instanceexists, err := c.GetComputeInstanceTypes(ctx, req.Instancetypeid, c.computeClient)
	if err != nil {
		logger.Error(err, "Error in get compute instance types")
		return &pb.NodeGroupResponseForm{}, fmt.Errorf("Get compute instance types failed")
	}
	if !instanceexists {
		return &pb.NodeGroupResponseForm{}, fmt.Errorf("Instance types does not match to compute instance types")
	}

	// Validate vnets are associated to cloud account
	if len(req.Vnets) == 0 {
		return &pb.NodeGroupResponseForm{}, fmt.Errorf("Need at least one VNet to create a nodegroup")
	}
	for _, vnet := range req.Vnets {
		err := c.CheckCloudAccountVnets(ctx, req.CloudAccountId, vnet.Networkinterfacevnetname, vnet.Availabilityzonename)
		if err != nil {
			return &pb.NodeGroupResponseForm{}, err
		}
	}

	// Validate instance types from IKS db
	res, err := query.CreateNodeGroupRecord(ctx, dbSession, req)
	if err != nil {
		return &pb.NodeGroupResponseForm{}, err
	}
	return res, nil
}

func (c *Server) GetNodeGroups(ctx context.Context, req *pb.GetNodeGroupsRequest) (*pb.NodeGroupResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetNodeGroups").WithValues(logkeys.ClusterId, req.Clusteruuid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.NodeGroupResponse{}, fmt.Errorf("no database connection found")
	}

	if err := validateGetNodeGroupsRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc get nodegroups Request")
		return &pb.NodeGroupResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	res, err := query.GetNodeGroups(ctx, dbSession, req)
	if err != nil {
		return &pb.NodeGroupResponse{}, err
	}
	return res, nil
}

func (c *Server) GetNodeGroup(ctx context.Context, req *pb.GetNodeGroupRequest) (*pb.NodeGroupResponseForm, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetNodeGroup").WithValues(logkeys.NodeGroupId, req.Nodegroupuuid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.NodeGroupResponseForm{}, fmt.Errorf("no database connection found")
	}

	if err := validateGetNodeGroupRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc get nodegroup Request")
		return &pb.NodeGroupResponseForm{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	res, err := query.GetNodeGroup(ctx, dbSession, req)
	if err != nil {
		return &pb.NodeGroupResponseForm{}, err
	}
	return res, nil
}

func (c *Server) PutNodeGroup(ctx context.Context, req *pb.UpdateNodeGroupRequest) (*pb.Nodegroupstatus, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.PutNodeGroup").WithValues(logkeys.NodeGroupId, req.Nodegroupuuid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.Nodegroupstatus{}, fmt.Errorf("no database connection found")
	}

	if err := validateUpdateNodeGroupRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc Nodegroup put Request")
		return &pb.Nodegroupstatus{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	res, err := query.PutNodeGroup(ctx, dbSession, req)
	if err != nil {
		return &pb.Nodegroupstatus{}, err
	}
	return res, nil
}

func (c *Server) UpgradeNodeGroup(ctx context.Context, req *pb.NodeGroupid) (*pb.Nodegroupstatus, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.UpgradeNodeGroup").WithValues(logkeys.NodeGroupId, req.Nodegroupuuid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.Nodegroupstatus{}, fmt.Errorf("no database connection found")
	}

	if err := validateNodeGroupid(req); err != nil {
		logger.Error(err, "Error in validating grpc Nodegroup upgrade Request")
		return &pb.Nodegroupstatus{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.UpgradeNodeGroup(ctx, dbSession, req)
	if err != nil {
		return &pb.Nodegroupstatus{}, err
	}
	return res, nil
}

func (c *Server) DeleteNodeGroup(ctx context.Context, req *pb.NodeGroupid) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.DeleteNodegroup").WithValues(logkeys.NodeGroupId, req.Nodegroupuuid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}

	if err := validateNodeGroupidRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc Nodegroup delete Request")
		return &empty.Empty{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.DeleteNodeGroupRecord(ctx, dbSession, req)
	if err != nil {
		return &empty.Empty{}, err
	}
	return res, nil
}

func (c *Server) DeleteNodeGroupInstance(ctx context.Context, req *pb.DeleteNodeGroupInstanceRequest) (*pb.Nodegroupstatus, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).
		WithName("Server.DeleteNodeGroupInstance").
		WithValues(
			logkeys.NodeGroupId,
			req.Nodegroupuuid,
			logkeys.CloudAccountId,
			req.CloudAccountId,
		).Start()
	logger.Info("BEGIN")
	defer span.End()
	defer logger.Info("END")
	logger.Info("deleting an instance from the nodegroup", "instanceName", req.InstanceName)

	if err := validateDeleteNodeGroupInstanceRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc  delete nodegroup instance Request")
		return &pb.Nodegroupstatus{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	// make all the required validation before executing instance deletion
	if err := c.validateDeleteNodeGroupInstance(ctx, req); err != nil {
		return &pb.Nodegroupstatus{}, err
	}

	// delete the instance
	instanceDelReq := &pb.InstanceDeleteRequest{
		Metadata: &pb.InstanceMetadataReference{
			CloudAccountId: req.CloudAccountId,
			NameOrId: &pb.InstanceMetadataReference_Name{
				Name: req.InstanceName,
			},
		},
	}
	if _, err := c.computeInstanceSvsClient.Delete(ctx, instanceDelReq); err != nil {
		logger.Error(err, "failed to delete an instance from the nodegroup",
			"instanceName", req.InstanceName)
		return &pb.Nodegroupstatus{}, err
	}

	// reduce the nodegroup by 1 if downsize is true
	if req.Downsize != nil && *req.Downsize {

		nodeGroupResp, err := c.GetNodeGroup(
			ctx,
			&pb.GetNodeGroupRequest{
				Clusteruuid:    req.Clusteruuid,
				Nodegroupuuid:  req.Nodegroupuuid,
				CloudAccountId: req.CloudAccountId,
			})
		if err != nil {
			return &pb.Nodegroupstatus{}, err
		}
		newCount := nodeGroupResp.Count - 1
		updateNodegroupReq := &pb.UpdateNodeGroupRequest{
			Clusteruuid:    req.Clusteruuid,
			Nodegroupuuid:  req.Nodegroupuuid,
			CloudAccountId: req.CloudAccountId,
			Count:          &newCount,
		}
		return c.PutNodeGroup(ctx, updateNodegroupReq)
	}

	return query.GetNodeGroupStatusRecord(
		ctx,
		c.session,
		&pb.NodeGroupid{
			CloudAccountId: req.CloudAccountId,
			Nodegroupuuid:  req.Nodegroupuuid,
			Clusteruuid:    req.Clusteruuid,
		})
}

func (c *Server) validateDeleteNodeGroupInstance(ctx context.Context, req *pb.DeleteNodeGroupInstanceRequest) error {
	if c.session == nil {
		return fmt.Errorf("no database connection found")
	}

	clusterId, nodeGroupId, err := utils.ValidateNodeGroupExistance(ctx, c.session, req.Clusteruuid, req.Nodegroupuuid)
	if err != nil {
		return utils.ErrorHandler(
			ctx,
			err,
			"DeleteNodeGroupInstance ValidateNodeGroupExistence",
			"Could not delete an instance from the node group. Please try again.",
		)
	}
	if clusterId == -1 || nodeGroupId == -1 {
		return status.Errorf(codes.NotFound, "NodeGroup not found in Cluster: %s", req.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, c.session, req.Clusteruuid, req.CloudAccountId)
	if err != nil {
		return utils.ErrorHandler(
			ctx,
			err,
			"DeleteNodeGroupInstance ValidateClusterCloudAccount",
			"Could not delete an instance from the node group. Please try again.",
		)
	}
	if !isOwner {
		return status.Errorf(codes.NotFound, "Cluster not found: %s", req.Clusteruuid) // returning not found to avoid leaking information
	}

	actionableState, err := utils.ValidaterClusterActionable(ctx, c.session, req.Clusteruuid)
	if err != nil {
		return utils.ErrorHandler(
			ctx,
			err,
			"DeleteNodeGroupInstance ValidaterClusterActionable",
			"Could not delete an instance from the node group. Please try again.")
	}
	if !actionableState {
		return status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	return nil
}

func (c *Server) GetNodeGroupStatus(ctx context.Context, req *pb.NodeGroupid) (*pb.Nodegroupstatus, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetNodeGroupStatus").WithValues(logkeys.NodeGroupId, req.Nodegroupuuid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.Nodegroupstatus{}, fmt.Errorf("no database connection found")
	}

	if err := validateNodeGroupid(req); err != nil {
		logger.Error(err, "Error in validating grpc Nodegroup status Request")
		return &pb.Nodegroupstatus{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.GetNodeGroupStatusRecord(ctx, dbSession, req)
	if err != nil {
		return &pb.Nodegroupstatus{}, err
	}
	return res, nil
}

func (c *Server) GetPublicK8SVersions(ctx context.Context, req *pb.IksCloudAccountId) (*pb.GetPublicAllK8SversionResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetPublicK8SVersions").WithValues(logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.GetPublicAllK8SversionResponse{}, fmt.Errorf("no database connection found")
	}

	if err := validateIksCloudAccountId(req); err != nil {
		logger.Error(err, "Error in validating grpc get public kubernetes versions Request")
		return &pb.GetPublicAllK8SversionResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.GetPublicAllK8sVersionsRecords(ctx, dbSession, req)
	if err != nil {
		logger.Error(err, "Error get k8s versions")
		return &pb.GetPublicAllK8SversionResponse{}, err
	}
	return res, nil
}

func (c *Server) GetPublicRuntimes(ctx context.Context, req *pb.IksCloudAccountId) (*pb.GetPublicAllRuntimeResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetPublicRuntimes").WithValues(logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.GetPublicAllRuntimeResponse{}, fmt.Errorf("no database connection found")
	}

	if err := validateIksCloudAccountId(req); err != nil {
		logger.Error(err, "Error in validating grpc get public runtimes Request")
		return &pb.GetPublicAllRuntimeResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.GetPublicAllRuntimesRecords(ctx, dbSession, req)
	if err != nil {
		logger.Error(err, "Error get runtimes")
		return &pb.GetPublicAllRuntimeResponse{}, err
	}
	return res, nil
}

func (c *Server) GetPublicInstanceTypes(ctx context.Context, req *pb.IksCloudAccountId) (*pb.GetPublicAllInstancetypeResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetPublicInstanceTypes").WithValues(logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.GetPublicAllInstancetypeResponse{}, fmt.Errorf("no database connection found")
	}

	if err := validateIksCloudAccountId(req); err != nil {
		logger.Error(err, "Error in validating grpc get public all instance types Request")
		return &pb.GetPublicAllInstancetypeResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.GetPublicAllInstancetypeRecords(ctx, dbSession, req)
	if err != nil {
		logger.Error(err, "Error get instance types")
		return &pb.GetPublicAllInstancetypeResponse{}, err
	}
	return res, nil
}

/*
func (c *Server) GetPublicNetworks(ctx context.Context, req *pb.IksCloudAccountId) (*pb.GetPublicAllNetworks, error) {
	log := log.FromContext(ctx).WithName("Server.GetPublicNetworks")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.GetPublicAllNetworks{}, fmt.Errorf("no database connection found")
	}
	res, err := query.GetPublicAllNetworks(ctx, dbSession, req)
	if err != nil {
		return &pb.GetPublicAllNetworks{}, err
	}
	return res, nil
}
*/

// Deprecated: Use RetrieveKubeConfig instead.
func (c *Server) GetKubeConfig(ctx context.Context, req *pb.ClusterID) (*pb.GetKubeconfigResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.Getkubeconfig").WithValues(logkeys.ClusterId, req.Clusteruuid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	if c.session == nil {
		return nil, errors.New("no database connection found")
	}
	res, err := query.GetKubeConfig(ctx, c.session, &pb.GetKubeconfigRequest{
		Clusteruuid:    req.Clusteruuid,
		CloudAccountId: req.CloudAccountId,
	}, c.cfg.EncryptionKeys)
	if err != nil {
		logger.Error(err, "Error get kubeconfig")
		return &pb.GetKubeconfigResponse{}, err
	}
	return res, nil
}

func (c *Server) RetrieveKubeConfig(ctx context.Context, req *pb.GetKubeconfigRequest) (*pb.GetKubeconfigResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.RetrieveKubeConfig").WithValues(logkeys.ClusterId, req.Clusteruuid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	if c.session == nil {
		return nil, errors.New("no database connection found")
	}

	if err := validateGetKubeconfigRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc retrieve kubeconfig Request")
		return &pb.GetKubeconfigResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}
	res, err := query.GetKubeConfig(ctx, c.session, req, c.cfg.EncryptionKeys)
	if err != nil {
		logger.Error(err, "Error get kubeconfig")
		return &pb.GetKubeconfigResponse{}, err
	}
	return res, nil
}

func (c *Server) CreateNewVip(ctx context.Context, req *pb.VipCreateRequest) (*pb.VipResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.CreateNewVip").WithValues(logkeys.VipName, req.Name, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.VipResponse{}, fmt.Errorf("no database connection found")
	}

	if err := validateVipCreateRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc create vip Request")
		return &pb.VipResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.CreateNewVip(ctx, dbSession, req)
	if err != nil {
		return &pb.VipResponse{}, err
	}
	return res, nil
}

func (c *Server) GetVips(ctx context.Context, req *pb.ClusterID) (*pb.GetVipsResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetVips").WithValues(logkeys.ClusterId, req.Clusteruuid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.GetVipsResponse{}, fmt.Errorf("no database connection found")
	}

	if err := validateClusterID(req); err != nil {
		logger.Error(err, "Error in validating grpc get vips Request")
		return &pb.GetVipsResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.GetVips(ctx, dbSession, req)
	if err != nil {
		return &pb.GetVipsResponse{}, err
	}
	return res, nil
}

func (c *Server) GetVip(ctx context.Context, req *pb.VipId) (*pb.GetVipResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetVip").WithValues(logkeys.VipId, req.Vipid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.GetVipResponse{}, fmt.Errorf("no database connection found")
	}

	if err := validateVipId(req); err != nil {
		logger.Error(err, "Error in validating grpc get vip Request")
		return &pb.GetVipResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.GetVip(ctx, dbSession, req)
	if err != nil {
		return &pb.GetVipResponse{}, err
	}
	return res, nil
}

func (c *Server) DeleteVip(ctx context.Context, req *pb.VipId) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.DeleteVip").WithValues(logkeys.VipId, req.Vipid, logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}

	if err := validateVipId(req); err != nil {
		logger.Error(err, "Error in validating grpc delete vip Request")
		return &empty.Empty{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.DeleteVip(ctx, dbSession, req)
	if err != nil {
		return &empty.Empty{}, err
	}
	return res, nil
}

func (c *Server) CheckCloudAccountVnets(ctx context.Context, cloudAccount string, vnetName string, availabilityZoneName string) error {

	keyName := &pb.VNetGetRequest_Metadata_Name{
		Name: vnetName,
	}

	/* CHECK IF CURRENT VNET CONNECTION EXISTS */
	getReq := &pb.VNetGetRequest{
		Metadata: &pb.VNetGetRequest_Metadata{
			NameOrId:       keyName,
			CloudAccountId: cloudAccount,
		},
	}
	vnet, err := c.vnetClient.Get(ctx, getReq)
	if err != nil || vnet == nil {
		return fmt.Errorf("Unable to get VNet with name %q", vnetName) // This is a friendly message. Do we want specific?
	}
	if vnet.Spec.AvailabilityZone != availabilityZoneName {
		return fmt.Errorf("VNet %q is in a different availability zone", vnetName)
	}

	return nil
}

func (c *Server) GetComputeInstanceTypes(ctx context.Context, instancetype string, computeClient pb.InstanceTypeServiceClient) (bool, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.GetComputeInstanceTypes").
		WithValues(logkeys.InstanceType, instancetype).Start()

	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	var instanceexists bool
	identifier := "cluster"

	if strings.Contains(instancetype, identifier) {
		igInstanceTypeSplit := strings.Split(instancetype, "-")
		instancetype = strings.Join(igInstanceTypeSplit[:len(igInstanceTypeSplit)-2], "-")
	}

	instances, err := computeClient.Search(ctx, &pb.InstanceTypeSearchRequest{})
	if err != nil {
		logger.Error(err, "\n .. Get compute instance types failed")
		return false, err
	}

	for i, _ := range instances.Items {
		if instancetype == instances.Items[i].Spec.Name {
			instanceexists = true
			break
		}
	}

	return instanceexists, nil
}

func generateSshKeyForCP(ctx context.Context) ([]byte, []byte, error) {
	bitSize := 4096

	// Generate RSA key.
	privatekey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, nil, err
	}

	// Extract public component.
	pubkey := privatekey.PublicKey

	privatekeyPEM := new(bytes.Buffer)
	if err := pem.Encode(privatekeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privatekey),
	}); err != nil {
		return nil, nil, err
	}

	sshPub, err := ssh.NewPublicKey(&pubkey)
	if err != nil {
		return nil, nil, err
	}

	sshPubBytes := ssh.MarshalAuthorizedKey(sshPub)

	_, _, _, _, err = ssh.ParseAuthorizedKey(sshPubBytes)
	if err != nil {
		return nil, nil, err
	}
	return privatekeyPEM.Bytes(), sshPubBytes, nil
}

func (c *Server) UploadSshkeysForCP(ctx context.Context, pubkey []byte, cloudaccount string, clustername string) (string, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Server.UploadSshkeysForCP").
		WithValues(logkeys.ClusterName, clustername).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	const charset = "abcdefgh12345"

	var seededRand *mathrand.Rand = mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	rndb := make([]byte, 4)

	for i := range rndb {
		rndb[i] = charset[seededRand.Intn(len(charset))]
	}

	var keyname string
	if len(clustername) > 13 {
		keyname = clustername[0:12] + "-" + cloudaccount + "-" + string(rndb)
	} else {
		keyname = clustername + "-" + cloudaccount + "-" + string(rndb)
	}

	getsshkeys, err := c.sshkeyClient.Search(ctx, &pb.SshPublicKeySearchRequest{
		Metadata: &pb.ResourceMetadataSearch{
			CloudAccountId: cloudaccount,
		},
	})
	if err != nil {
		return keyname, err
	}
	if len(getsshkeys.Items) != 0 {
		for i, _ := range getsshkeys.Items {
			if getsshkeys.Items[i].Metadata.Name == keyname {
				logger.Info("Ssh key with the name already exists.Creating cp with existing ssh key", logkeys.SSHKeyName, keyname)
				return keyname, nil
			}
			continue
		}
	}
	sshkey, err := c.sshkeyClient.Create(ctx, &pb.SshPublicKeyCreateRequest{
		Metadata: &pb.ResourceMetadataCreate{
			CloudAccountId: cloudaccount,
			Name:           keyname,
		},
		Spec: &pb.SshPublicKeySpec{
			SshPublicKey: string(pubkey),
		},
	})

	if err != nil {
		logger.Error(err, "\n .. Upload ssh key failed")
		return "", err
	}

	return sshkey.Metadata.Name, nil
}

func (c *Server) UpdateFirewallRule(ctx context.Context, req *pb.UpdateFirewallRuleRequest) (*pb.FirewallRuleResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("IKS.UpdateFirewallRule").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.FirewallRuleResponse{}, fmt.Errorf("no database connection found")
	}

	if err := validateUpdateFirewallRuleRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc update firewall rule Request")
		return &pb.FirewallRuleResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.UpdateFirewallRule(ctx, dbSession, req)
	if err != nil {
		logger.Error(err, "Error updating security rule")
		return &pb.FirewallRuleResponse{}, err
	}
	return res, nil
}

func (c *Server) GetFirewallRule(ctx context.Context, req *pb.ClusterID) (*pb.GetFirewallRuleResponse, error) {
	// Get all public vip types
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("IKS.GetFirewallRule").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &pb.GetFirewallRuleResponse{}, fmt.Errorf("no database connection found")
	}

	if err := validateClusterID(req); err != nil {
		logger.Error(err, "Error in validating grpc get firewall rule Request")
		return &pb.GetFirewallRuleResponse{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.GetFirewallRule(ctx, dbSession, req)
	if err != nil {
		logger.Error(err, "Error getting security rule")
		return &pb.GetFirewallRuleResponse{}, err
	}
	return res, nil
}

func (c *Server) DeleteFirewallRule(ctx context.Context, req *pb.DeleteFirewallRuleRequest) (*empty.Empty, error) {
	// delete entry for source ips from vip table baaed on vip id
	// Get all public vip types
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("IKS.DeleteFirewallRule").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")

	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}
	if err := validateDeleteFirewallRuleRequest(req); err != nil {
		logger.Error(err, "Error in validating grpc delete firewall rule Request")
		return &empty.Empty{}, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	res, err := query.DeleteFirewallRule(ctx, dbSession, req)
	if err != nil {
		logger.Error(err, "Error Deleting Security Rule")
		return &empty.Empty{}, err
	}
	return res, nil
}
