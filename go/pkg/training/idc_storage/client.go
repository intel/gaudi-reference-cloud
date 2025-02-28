// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package idc_storage

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	storeSvcUtils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/config"
	"google.golang.org/grpc"
)

type IDCStorageServiceClient struct {
	StorageAPIConn   *grpc.ClientConn
	Region           string
	AvailabilityZone string
}

type FilesystemCreateRequest struct {
	Name             string
	CloudAccountId   string
	Description      string
	AvailabilityZone string
	Capacity         string // Units: GB
	AccessModes      v1.FilesystemAccessModes
	MountProtocol    v1.FilesystemMountProtocols
}

func NewIDCStorageServiceClient(ctx context.Context, cloudConfig *config.IdcConfig) (*IDCStorageServiceClient, error) {
	log := log.FromContext(ctx).WithName("NewIDCStorageServiceClient")

	storageClientConn, err := grpcutil.NewClient(ctx, cloudConfig.StorageGrpcAPIEndpoint)
	if err != nil {
		log.Error(err, "error creating storage service client")
		return nil, err
	}

	return &IDCStorageServiceClient{
		StorageAPIConn:   storageClientConn,
		Region:           cloudConfig.Region,
		AvailabilityZone: cloudConfig.AvailabilityZone,
	}, nil
}

func (storeMgr *IDCStorageServiceClient) CreateIDCStorageFilesystem(ctx context.Context, req FilesystemCreateRequest) error {
	log := log.FromContext(ctx).WithName("IDCStorageServiceClient.CreateIDCStorageFilesysten")
	log.Info("entering storage instance creation")

	// Ensure a name is assigned to the new storage within parameters set by STaaS
	if req.Name == "" {
		log.Info("storage create worker", "invalid job", "skipping..")
		return fmt.Errorf("invalid input arguments - missing Name")
	}
	if err := storeSvcUtils.ValidateInstanceName(req.Name); err != nil {
		log.Info("storage create worker", "invalid job", "skipping...")
		return fmt.Errorf("invalid input arguments for Name - %s", err)
	}

	// Ensure a capacity value is defined within parameters set by STaaS
	if req.Capacity == "" {
		log.Info("storage create worker", "invalid job", "skipping...")
		return fmt.Errorf("invalid input arguments - missing Capacity")
	}
	if storeSvcUtils.ParseFileSizeInGB(req.Capacity) <= 0 {
		log.Info("storage create worker", "invalid job", "skipping...")
		return fmt.Errorf("invalid input arguments. Capacity has to be greater than 1GB")
	}

	log.Info("storage create worker", "storage name/capacity", fmt.Sprintf("%s %s", req.Name, req.Capacity))

	newStorage, err := storeMgr.CreateStorage(ctx, req)
	if err != nil {
		log.Error(err, "error creating storage requested")
		return fmt.Errorf("error creating storage requested")
	}

	// Wait for storage filesystem to be "FSReady"
	if err := storeMgr.WaitForStorageStateReadygRPC(ctx, newStorage.Metadata.ResourceId, req.CloudAccountId, 600); err != nil {
		log.Error(err, "storage not ready in given timeout")
		return fmt.Errorf("storage not ready, timeout")
	}

	newUser, err := storeMgr.CreateUser(ctx, req, newStorage)
	if err != nil {
		log.Error(err, "error creating storage user requested")
		log.Info("DEBUG", "storage user credentials", newUser)
		return fmt.Errorf("error creating storage user requested")
	}

	return nil
}

func (storeMgr *IDCStorageServiceClient) DeleteIDCStorageFilesystem(ctx context.Context, req FilesystemCreateRequest) error {
	log := log.FromContext(ctx).WithName("IDCStorageServiceClient.DeleteIDCStorageFilesystem")
	log.Info("entering to delete storage filesystem")

	if req.Name == "" {
		log.Info("storage delete worker", "invalid job", "skipping..")
		return fmt.Errorf("invalid input arguments - missing Name")
	}

	log.Info("storage delete worker", "storage name", req.Name)

	if err := storeMgr.DeleteStorage(ctx, req.Name, req.CloudAccountId); err != nil {
		log.Error(err, "error deleting storage")
		return fmt.Errorf("error deleting storage")
	}

	return nil
}

func (storeMgr *IDCStorageServiceClient) GetStorageStatusByName(ctx context.Context, req FilesystemCreateRequest) (*v1.FilesystemPrivate, error) {
	log := log.FromContext(ctx).WithName("ClusterProvisionScheduler.GetStorageStatusByName")

	storageStatus, err := storeMgr.GetStorageByName(ctx, req.Name, req.CloudAccountId)
	if err != nil {
		log.Error(err, "error reading storage status")
		return nil, fmt.Errorf("error reading storage state")
	}

	return storageStatus, nil
}
