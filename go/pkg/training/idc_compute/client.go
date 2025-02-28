// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package idc_compute

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/config"
	"google.golang.org/grpc"
)

type IDCServiceClient struct {
	ComputeAPIConn   *grpc.ClientConn
	Region           string
	AvailabilityZone string
}

type InstanceCreateRequest struct {
	Name           string
	ClusterId      string
	CloudAccountId string
	VNet           string
	MachineType    string
	SshKeyNames    []string
	Labels         map[string]string
	ImageName      string
	UserData       string
	SkipQuotaCheck bool
}

func NewIDCComputeServiceClient(ctx context.Context, cloudConfig *config.IdcConfig) (*IDCServiceClient, error) {
	logger := log.FromContext(ctx).WithName("NewIDCComputeServiceClient")
	computeClientConn, err := grpcutil.NewClient(ctx, cloudConfig.ComputeGrpcAPIEndpoint)
	if err != nil {
		logger.Error(err, "error creating instanceServiceClient")
		return nil, err
	}

	return &IDCServiceClient{
		ComputeAPIConn:   computeClientConn,
		Region:           cloudConfig.Region,
		AvailabilityZone: cloudConfig.AvailabilityZone,
	}, nil
}

func (instMgr *IDCServiceClient) CreateIDCComputeInstance(ctx context.Context, req InstanceCreateRequest) error {
	log := log.FromContext(ctx).WithName("IDCServiceClient.CreateInstance")
	log.Info("entering a instance creation")

	if req.Name == "" {
		log.Info("instance create worker", "invalid job", "skipping..")
		return fmt.Errorf("invalid input arguments")
	}

	log.Info("instance create worker", "instance-id", req.Name)

	currInst, err := instMgr.CreateInstance(ctx, req)
	if err != nil {
		log.Error(err, "error provisioning instance")
		return fmt.Errorf("error creating instance")
	}

	// Wait for compute instance to be "Ready"
	if err := instMgr.WaitForInstanceStateReadygRPC(ctx, currInst.Metadata.ResourceId, req.CloudAccountId, 1800); err != nil {
		log.Error(err, "instance not ready in given timeout")
		return fmt.Errorf("instance not ready, timeout")
	}
	return nil
}

func (instMgr *IDCServiceClient) DeleteIDCComputeInstance(ctx context.Context, req InstanceCreateRequest) error {
	log := log.FromContext(ctx).WithName("IDCServiceClient.DeleteIDCComputeInstance")
	log.Info("entering to delete an instance")

	if req.Name == "" {
		log.Info("instance delete worker", "invalid job", "skipping...")
		return fmt.Errorf("invalid input arguments")
	}

	log.Info("instance delete worker", "instance-id", req.Name)

	if err := instMgr.DeleteInstance(ctx, req.Name, req.CloudAccountId); err != nil {
		log.Error(err, "error deleting instance")
		return fmt.Errorf("error deleting instance")
	}

	return nil
}

func (instMgr *IDCServiceClient) GetInstanceStateByName(ctx context.Context, req InstanceCreateRequest) (*v1.Instance, error) {
	log := log.FromContext(ctx).WithName("ClusterProvisionScheduler.GetInstanceStateByName")

	instanceState, err := instMgr.GetInstanceByName(ctx, req.Name, req.CloudAccountId)
	if err != nil {
		log.Error(err, "error reading instance state")
		return nil, fmt.Errorf("error reading instance state")
	}
	return instanceState, nil
}
