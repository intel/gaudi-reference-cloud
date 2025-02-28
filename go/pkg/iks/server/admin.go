// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/config"
	admin_query "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/admin_query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// Server is a cluster
type PrivateAdminServer struct {
	pb.UnimplementedIksAdminServer
	session       *sql.DB
	cfg           config.Config
	computeClient pb.InstanceTypeServiceClient
}

// NewIksService Initializes DB connection
func NewIksPrivateAdminService(session *sql.DB, cfg config.Config, computegrpcClient pb.InstanceTypeServiceClient) (*PrivateAdminServer, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}
	return &PrivateAdminServer{
		session:       session,
		cfg:           cfg,
		computeClient: computegrpcClient,
	}, nil
}

// Authenticate IKS Admin User
func (c *PrivateAdminServer) AuthenticateIKSAdminUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.AuthenticateIKSAdminUser")
	log.Info("Request", logkeys.Request, req)

	if req.Iksadminkey == "" {
		return &pb.UserResponse{}, fmt.Errorf("Enter Valid User Request")
	}

	isAuthenticatedUser, err := iks_utils.IsIksAdminAuthenticatedUser(ctx, c.cfg.AdminKey, req.Iksadminkey)
	if err != nil {
		return &pb.UserResponse{}, err
	}

	if isAuthenticatedUser {
		return &pb.UserResponse{
			IsAuthenticatedUser: isAuthenticatedUser,
		}, nil
	}

	return &pb.UserResponse{}, nil
}

// Put IMI
func (c *PrivateAdminServer) PutIMI(ctx context.Context, req *pb.UpdateIMIRequest) (*pb.IMIResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.PutIMI")
	log.Info("Request", logkeys.Request, req)
	if req.Iksadminkey == "" {
		return &pb.IMIResponse{}, fmt.Errorf("enter valid user request")
	}

	isAuthenticatedUser, err := iks_utils.IsIksAdminAuthenticatedUser(ctx, c.cfg.AdminKey, req.Iksadminkey)
	if err != nil {
		return &pb.IMIResponse{}, err
	}
	if !isAuthenticatedUser {
		return &pb.IMIResponse{}, fmt.Errorf("user is not authorized to perform this operation")
	}
	dbSession := c.session
	if dbSession == nil {
		return &pb.IMIResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.PutIMI(ctx, dbSession, req)
	if err != nil {
		return &pb.IMIResponse{}, err
	}
	return res, nil
}

// Get IMI
func (c *PrivateAdminServer) GetIMI(ctx context.Context, req *pb.GetIMIRequest) (*pb.IMIResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetIMI")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.IMIResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetIMI(ctx, dbSession, req)
	if err != nil {
		return &pb.IMIResponse{}, err
	}
	return res, nil
}

// Get IMIs
func (c *PrivateAdminServer) GetIMIs(ctx context.Context, req *empty.Empty) (*pb.GetIMIResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetIMIs")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.GetIMIResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetIMIs(ctx, dbSession)
	if err != nil {
		return &pb.GetIMIResponse{}, err
	}
	return res, nil
}

// Get IMIsInfo
func (c *PrivateAdminServer) GetIMIsInfo(ctx context.Context, req *empty.Empty) (*pb.GetIMIsInfoResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetIMIsInfo")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.GetIMIsInfoResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetIMIsInfo(ctx, dbSession)
	if err != nil {
		return &pb.GetIMIsInfoResponse{}, err
	}
	return res, nil
}

// Delete IMI
func (c *PrivateAdminServer) DeleteIMI(ctx context.Context, req *pb.DeleteIMIRequest) (*empty.Empty, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.DeleteIMI")
	log.Info("Request", logkeys.Request, req)
	if req.Iksadminkey == "" {
		return &empty.Empty{}, fmt.Errorf("enter valid user request")
	}
	isAuthenticatedUser, err := iks_utils.IsIksAdminAuthenticatedUser(ctx, c.cfg.AdminKey, req.Iksadminkey)
	if err != nil {
		return &empty.Empty{}, err
	}
	if !isAuthenticatedUser {
		return &empty.Empty{}, fmt.Errorf("user is not authorized to perform this operation")
	}
	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}
	_, err = admin_query.DeleteIMI(ctx, dbSession, req)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

// Create IMIs
func (c *PrivateAdminServer) CreateIMI(ctx context.Context, record *pb.IMIRequest) (*pb.IMIResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.CreateIMI")
	log.Info("Request", logkeys.Request, record)
	if record.Iksadminkey == "" {
		return &pb.IMIResponse{}, fmt.Errorf("enter valid user request")
	}
	isAuthenticatedUser, err := iks_utils.IsIksAdminAuthenticatedUser(ctx, c.cfg.AdminKey, record.Iksadminkey)
	if err != nil {
		return &pb.IMIResponse{}, err
	}
	if !isAuthenticatedUser {
		return &pb.IMIResponse{}, fmt.Errorf("user is not authorized to perform this operation")
	}
	dbSession := c.session
	if dbSession == nil {
		return &pb.IMIResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.CreateIMI(ctx, dbSession, record)
	if err != nil {
		return &pb.IMIResponse{}, err
	}
	return res, nil
}

// Create IMI InstaceType K8s Compatibility
func (c *PrivateAdminServer) UpdateIMIInstanceTypeToK8SCompatibility(ctx context.Context, record *pb.IMIInstanceTypeK8SRequest) (*pb.IMIInstanceTypeK8SResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.UpdateIMIInstanceTypeToK8SCompatibility")
	log.Info("Request", logkeys.Request, record)
	dbSession := c.session
	if dbSession == nil {
		return &pb.IMIInstanceTypeK8SResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.UpdateIMIToK8sCompatibility(ctx, dbSession, record)
	if err != nil {
		return &pb.IMIInstanceTypeK8SResponse{}, err
	}
	return res, nil
}

// Get Cluster
func (c *PrivateAdminServer) GetCluster(ctx context.Context, req *pb.AdminClusterID) (*pb.GetClusterAdmin, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetCluster")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.GetClusterAdmin{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetCluster(ctx, dbSession, req)
	if err != nil {
		return &pb.GetClusterAdmin{}, err
	}
	return res, nil
}

// Get Clusters
func (c *PrivateAdminServer) GetClusters(ctx context.Context, req *empty.Empty) (*pb.GetClustersAdmin, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetClusters")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.GetClustersAdmin{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetClusters(ctx, dbSession)
	if err != nil {
		return &pb.GetClustersAdmin{}, err
	}
	return res, nil
}

// Get K8s Versions
func (c *PrivateAdminServer) GetK8SVersions(ctx context.Context, req *empty.Empty) (*pb.GetK8SVersionResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetK8SVersions")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.GetK8SVersionResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetK8SVersions(ctx, dbSession)
	if err != nil {
		return &pb.GetK8SVersionResponse{}, err
	}
	return res, nil
}

func (c *PrivateAdminServer) GetK8SVersion(ctx context.Context, req *pb.GetK8SRequest) (*pb.K8SversionResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetK8SVersion")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.K8SversionResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetK8SVersion(ctx, dbSession, req)
	if err != nil {
		return &pb.K8SversionResponse{}, err
	}
	return res, nil
}

// Create K8sVersion
func (c *PrivateAdminServer) CreateK8SVersion(ctx context.Context, record *pb.Createk8SversionRequest) (*pb.K8SversionResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.CreateK8SVersion")
	log.Info("Request", logkeys.Request, record)
	dbSession := c.session
	if dbSession == nil {
		return &pb.K8SversionResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.CreateK8SVersion(ctx, dbSession, record)
	if err != nil {
		return &pb.K8SversionResponse{}, err
	}
	return res, nil
}

// Delete K8sVersion
func (c *PrivateAdminServer) DeleteK8SVersion(ctx context.Context, req *pb.GetK8SRequest) (*empty.Empty, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.DeleteK8SVersion")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}
	_, err := admin_query.DeleteK8SVersion(ctx, dbSession, req)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

// Put K8sVersion
func (c *PrivateAdminServer) PutK8SVersion(ctx context.Context, req *pb.UpdateK8SRequest) (*pb.K8SversionResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.PutK8SVersion")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &pb.K8SversionResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.PutK8SVersion(ctx, dbSession, req)

	if err != nil {
		return &pb.K8SversionResponse{}, err
	}
	return res, nil
}

// Upgrade Cluster
func (c *PrivateAdminServer) UpgradeClusterControlPlane(ctx context.Context, req *pb.UpgradeControlPlaneRequest) (*empty.Empty, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.UpgradeClusterControlPlane")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.UpgradeClusterControlPlane(ctx, dbSession, req)

	if err != nil {
		return &empty.Empty{}, err
	}
	return res, nil
}

// GetcontrolPlaneSshKeys
func (c *PrivateAdminServer) GetControlPlaneSSHKeys(ctx context.Context, req *pb.AdminClusterID) (*pb.ClusterSSHKeys, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetControlPlaneSSHKeys")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &pb.ClusterSSHKeys{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetControlPlaneSSHKeys(ctx, dbSession, req, c.cfg.EncryptionKeys)
	if err != nil {
		return &pb.ClusterSSHKeys{}, err
	}
	return res, nil
}

// Get CloudAccount Approvelist
func (c *PrivateAdminServer) GetCloudAccountApproveList(ctx context.Context, req *empty.Empty) (*pb.CloudAccountApproveListResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetCloudAccountApproveList")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &pb.CloudAccountApproveListResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetCloudAccountApproveList(ctx, dbSession)
	if err != nil {
		return &pb.CloudAccountApproveListResponse{}, err
	}
	return res, nil
}

// Post CloudAccount Approvelist
func (c *PrivateAdminServer) PostCloudAccountApproveList(ctx context.Context, req *pb.CloudAccountApproveListRequest) (*pb.CloudAccountApproveList, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.PostCloudAccountApproveList")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &pb.CloudAccountApproveList{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.CreateCloudAccountApproveList(ctx, dbSession, req)
	if err != nil {
		return &pb.CloudAccountApproveList{}, err
	}
	return res, nil
}

// Put CloudAccount Approvelist
func (c *PrivateAdminServer) PutCloudAccountApproveList(ctx context.Context, req *pb.CloudAccountApproveListRequest) (*pb.CloudAccountApproveList, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.PutCloudAccountApproveList")
	log.Info("Request", logkeys.Request, req)

	err := req.ValidateAll()
	if err != nil {
		return &pb.CloudAccountApproveList{}, fmt.Errorf("input request validation failed")
	}

	if req.Iksadminkey == "" {
		return &pb.CloudAccountApproveList{}, fmt.Errorf("enter valid user request")
	}

	isAuthenticatedUser, err := iks_utils.IsIksAdminAuthenticatedUser(ctx, c.cfg.AdminKey, req.Iksadminkey)
	if err != nil {
		return &pb.CloudAccountApproveList{}, err
	}
	if !isAuthenticatedUser {
		return &pb.CloudAccountApproveList{}, fmt.Errorf("user is not authorized to perform this operation")
	}
	dbSession := c.session
	if dbSession == nil {
		return &pb.CloudAccountApproveList{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.PutCloudAccountApproveList(ctx, dbSession, req)
	if err != nil {
		return &pb.CloudAccountApproveList{}, err
	}
	return res, nil
}

// Get LoadBalancers
func (c *PrivateAdminServer) GetLoadBalancers(ctx context.Context, req *pb.AdminClusterID) (*pb.LoadBalancers, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetLoadBalancers")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &pb.LoadBalancers{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetLoadBalancers(ctx, dbSession, req)
	if err != nil {
		return &pb.LoadBalancers{}, err
	}
	return res, nil
}

// Get LoadBalancer
func (c *PrivateAdminServer) GetLoadBalancer(ctx context.Context, req *pb.GetLbRequest) (*pb.LoadbalancerResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetLoadBalancer")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &pb.LoadbalancerResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetLoadBalancer(ctx, dbSession, req)
	if err != nil {
		return &pb.LoadbalancerResponse{}, err
	}
	return res, nil
}

// Create InstanceType
func (c *PrivateAdminServer) CreateInstanceTypes(ctx context.Context, req *pb.CreateInstanceTypeRequest) (*pb.InstanceTypeResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.CreateInstanceTypes")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.InstanceTypeResponse{}, fmt.Errorf("no database connection found")
	}
	if req.Iksadminkey == "" {
		return &pb.InstanceTypeResponse{}, fmt.Errorf("enter valid user request")
	}
	isAuthenticatedUser, err := iks_utils.IsIksAdminAuthenticatedUser(ctx, c.cfg.AdminKey, req.Iksadminkey)
	if err != nil {
		return &pb.InstanceTypeResponse{}, err
	}
	if !isAuthenticatedUser {
		return &pb.InstanceTypeResponse{}, fmt.Errorf("user is not authorized to perform this operation")
	}
	res, err := admin_query.CreateInstanceTypes(ctx, dbSession, req)
	if err != nil {
		return &pb.InstanceTypeResponse{}, err
	}
	return res, nil
}

// Get InstanceTypes
func (c *PrivateAdminServer) GetInstanceTypes(ctx context.Context, req *empty.Empty) (*pb.GetInstanceTypesResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetInstanceTypes")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.GetInstanceTypesResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetInstanceTypes(ctx, dbSession)
	if err != nil {
		return &pb.GetInstanceTypesResponse{}, err
	}
	return res, nil
}

// Get InstanceType
func (c *PrivateAdminServer) GetInstanceType(ctx context.Context, req *pb.GetInstanceTypeRequest) (*pb.GetInstanceTypeResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetInstanceType")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.GetInstanceTypeResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetInstanceType(ctx, dbSession, req)
	if err != nil {
		return &pb.GetInstanceTypeResponse{}, err
	}
	return &pb.GetInstanceTypeResponse{
		IksInstanceType:     res,
		ComputeInstanceType: c.GetComputeInstanceTypeByName(ctx, res.Instancetypename, c.computeClient),
	}, nil
}

// Get InstanceType Info
func (c *PrivateAdminServer) GetInstanceTypeInfo(ctx context.Context, req *empty.Empty) (*pb.GetInstanceTypeInfoResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetInstanceTypeInfo")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.GetInstanceTypeInfoResponse{}, fmt.Errorf("no database connection found")
	}
	instanceTypeResponse, err := c.GetComputeInstanceTypes(ctx, dbSession, c.computeClient)
	if err != nil {
		log.Error(err, "\n .. Get compute instance types life cycle states failed")
		return &pb.GetInstanceTypeInfoResponse{}, err
	}
	lifeCycleStates, err := iks_utils.GetDefaultLifeCycleStates(ctx, dbSession)
	if err != nil {
		return &pb.GetInstanceTypeInfoResponse{}, err
	}
	nodeProviderNames, err := iks_utils.GetDefaultNodeProviderNames(ctx, dbSession)
	if err != nil {
		return &pb.GetInstanceTypeInfoResponse{}, err
	}
	return &pb.GetInstanceTypeInfoResponse{
		ComputeResponse:  instanceTypeResponse,
		States:           lifeCycleStates,
		Nodeprovidername: nodeProviderNames,
	}, nil
}

// Update Instance Type
func (c *PrivateAdminServer) PutInstanceType(ctx context.Context, req *pb.UpdateInstanceTypeRequest) (*pb.InstanceTypeResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.PutInstanceType")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.InstanceTypeResponse{}, fmt.Errorf("no database connection found")
	}
	if req.Iksadminkey == "" {
		return &pb.InstanceTypeResponse{}, fmt.Errorf("enter valid user request")
	}
	isAuthenticatedUser, err := iks_utils.IsIksAdminAuthenticatedUser(ctx, c.cfg.AdminKey, req.Iksadminkey)
	if err != nil {
		return &pb.InstanceTypeResponse{}, err
	}
	if !isAuthenticatedUser {
		return &pb.InstanceTypeResponse{}, fmt.Errorf("user is not authorized to perform this operation")
	}
	res, err := admin_query.PutInstanceType(ctx, dbSession, req)
	if err != nil {
		return &pb.InstanceTypeResponse{}, err
	}
	return res, nil
}

// Create InstaceType IMI K8s Compatibility
func (c *PrivateAdminServer) UpdateInstanceTypeIMIToK8SCompatibility(ctx context.Context, record *pb.InstanceTypeIMIK8SRequest) (*pb.InstanceTypeIMIK8SResponse, error) {
	log := log.FromContext(ctx).WithName("Iks.UpdateInstanceTypeToK8sCompatibility")
	log.Info("Request", "req", record)
	dbSession := c.session
	if dbSession == nil {
		return &pb.InstanceTypeIMIK8SResponse{}, fmt.Errorf("no database connection found")
	}
	if record.Iksadminkey == "" {
		return &pb.InstanceTypeIMIK8SResponse{}, fmt.Errorf("enter valid user request")
	}
	isAuthenticatedUser, err := iks_utils.IsIksAdminAuthenticatedUser(ctx, c.cfg.AdminKey, record.Iksadminkey)
	if err != nil {
		return &pb.InstanceTypeIMIK8SResponse{}, err
	}
	if !isAuthenticatedUser {
		return &pb.InstanceTypeIMIK8SResponse{}, fmt.Errorf("user is not authorized to perform this operation")
	}
	res, err := admin_query.UpdateInstanceTypeToK8sCompatibility(ctx, dbSession, record)
	if err != nil {
		return &pb.InstanceTypeIMIK8SResponse{}, err
	}
	return res, nil
}

// Delete IMI
func (c *PrivateAdminServer) DeleteInstanceType(ctx context.Context, req *pb.DeleteInstanceTypeRequest) (*empty.Empty, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.DeleteInstanceType")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}
	if req.Iksadminkey == "" {
		return &empty.Empty{}, fmt.Errorf("enter valid user request")
	}
	isAuthenticatedUser, err := iks_utils.IsIksAdminAuthenticatedUser(ctx, c.cfg.AdminKey, req.Iksadminkey)
	if err != nil {
		return &empty.Empty{}, err
	}
	if !isAuthenticatedUser {
		return &empty.Empty{}, fmt.Errorf("user is not authorized to perform this operation")
	}
	_, err = admin_query.DeleteInstanceType(ctx, dbSession, req)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

// Get Compute Instance Types
func (c *PrivateAdminServer) GetComputeInstanceTypes(ctx context.Context, dbconn *sql.DB, computeClient pb.InstanceTypeServiceClient) ([]*pb.InstanceTypeResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetComputeInstanceTypes")

	var instanceexists bool
	var instanceTypeResponse []*pb.InstanceTypeResponse

	instances, err := computeClient.Search(ctx, &pb.InstanceTypeSearchRequest{})
	if err != nil {
		log.Error(err, "\n .. Get compute instance types failed")
		return instanceTypeResponse, err
	}

	for i, _ := range instances.Items {
		instanceexists, err = iks_utils.ValidateInstanceTypeExistance(ctx, dbconn, instances.Items[i].Spec.Name)
		if err != nil {
			log.Error(err, "\n .. Validate Instance Type Existance failed")
		}
		if !instanceexists {
			var memory int32
			_, err = fmt.Sscan(instances.Items[i].Spec.Memory.Size, &memory)
			if err != nil {
				log.Error(err, "\n .. Instance Type Memory Scan failed")
			}
			var storage int32
			_, err = fmt.Sscan(instances.Items[i].Spec.Disks[0].Size, &storage)
			if err != nil {
				log.Error(err, "\n .. Instance Type Storage Disk Scan failed")
			}
			instanceType := &pb.InstanceTypeResponse{
				Instancetypename: instances.Items[i].Spec.Name,
				Memory:           memory,
				Cpu:              instances.Items[i].Spec.Cpu.Cores,
				Displayname:      instances.Items[i].Spec.DisplayName,
				Description:      instances.Items[i].Spec.Description,
				Category:         instances.Items[i].Spec.InstanceCategory.String(),
				Storage:          storage,
				IksDB:            instanceexists,
			}
			instanceTypeResponse = append(instanceTypeResponse, instanceType)
		}
	}

	return instanceTypeResponse, nil
}

// Get Compute Instance By Name
func (c *PrivateAdminServer) GetComputeInstanceTypeByName(ctx context.Context, instancetype string, computeClient pb.InstanceTypeServiceClient) *pb.InstanceTypeResponse {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetComputeInstanceTypeByName")
	var instanceTypeResponse *pb.InstanceTypeResponse

	input := &pb.InstanceTypeGetRequest{
		Metadata: &pb.InstanceTypeGetRequest_Metadata{
			Name: instancetype,
		},
	}

	instance, err := computeClient.Get(ctx, input)
	if err != nil {
		log.Error(err, "\n .. Get compute instance types failed")
		return instanceTypeResponse
	}

	log.Info("GetComputeInstanceTypeByName", logkeys.InstanceType, instance)
	var memory int32
	_, err = fmt.Sscan(instance.Spec.Memory.Size, &memory)
	if err != nil {
		log.Error(err, "\n .. Compute Instance Type Memory Scan failed")
	}
	var storage int32
	_, err = fmt.Sscan(instance.Spec.Disks[0].Size, &storage)
	if err != nil {
		log.Error(err, "\n .. Compute Instance Type Storage Disk Scan failed")
	}
	instanceTypeResponse = &pb.InstanceTypeResponse{
		Instancetypename: instance.Spec.Name,
		Memory:           memory,
		Cpu:              instance.Spec.Cpu.Cores,
		Displayname:      instance.Spec.DisplayName,
		Description:      instance.Spec.Description,
		Category:         instance.Spec.InstanceCategory.String(),
		Storage:          storage,
		IksDB:            true,
	}

	return instanceTypeResponse
}

func (c *PrivateAdminServer) GetFirewallRule(ctx context.Context, req *pb.AdminClusterID) (*pb.GetAdminFirewallRuleResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateAdminServer.GetFirewallRule")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &pb.GetAdminFirewallRuleResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := admin_query.GetFirewallRule(ctx, dbSession, req)
	if err != nil {
		return &pb.GetAdminFirewallRuleResponse{}, err
	}
	return res, nil
}
