// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package idc_storage

// This file provides a gRPC interface with storage as a service (filesystem.go)

import (
	"context"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	retry "github.com/sethvargo/go-retry"
)

func (idcSvc *IDCStorageServiceClient) CreateStorage(ctx context.Context, createRequest FilesystemCreateRequest) (*v1.FilesystemPrivate, error) {
	log := log.FromContext(ctx).WithName("IDCStorageServiceClient.CreateStorage")

	// Define storage metadata and spec
	storageRequest := v1.FilesystemCreateRequestPrivate{
		Metadata: &v1.FilesystemMetadataPrivate{
			Name:             createRequest.Name,
			CloudAccountId:   createRequest.CloudAccountId,
			Description:      createRequest.Description,
			SkipQuotaCheck:   true,
			SkipProductCheck: true,
		},
		Spec: &v1.FilesystemSpecPrivate{
			AvailabilityZone: createRequest.AvailabilityZone,
			Request: &v1.FilesystemCapacity{
				Storage: createRequest.Capacity,
			},
			AccessModes:   createRequest.AccessModes,
			MountProtocol: createRequest.MountProtocol,
			Encrypted:     true,
		},
	}

	// Create storage filesystem via API
	storage, err := v1.NewFilesystemPrivateServiceClient(idcSvc.StorageAPIConn).CreatePrivate(ctx, &storageRequest)
	if err != nil {
		log.Error(err, "error creating storage")
		return nil, fmt.Errorf("error creating storage")
	}

	log.Info("DEBUG", "storage create response", storage)

	return storage, nil
}

func (idcSvc *IDCStorageServiceClient) CreateUser(ctx context.Context, createRequest FilesystemCreateRequest, storageRequest *v1.FilesystemPrivate) (*v1.FilesystemGetUserResponsePrivate, error) {
	log := log.FromContext(ctx).WithName("IDCStorageServiceClient.CreateUser")

	// after creating the filesystem, must create filesystem user for STaaS to manage
	// this is a consequence of using STaaS's private apis
	// define storage user request
	storageUserReq := v1.FilesystemGetUserRequestPrivate{
		Metadata: &v1.FilesystemMetadataReference{
			CloudAccountId: createRequest.CloudAccountId,
			NameOrId: &v1.FilesystemMetadataReference_ResourceId{
				ResourceId: storageRequest.Metadata.ResourceId,
			},
		},
	}

	// call GetUser to create the user for first filesystem, then update for subsequent filesystems
	user, err := v1.NewFilesystemPrivateServiceClient(idcSvc.StorageAPIConn).GetUserPrivate(ctx, &storageUserReq)
	if err != nil {
		log.Error(err, "error creating storage user")
		return nil, fmt.Errorf("error creating storage user")
	}

	log.Info("DEBUG", "storage user credentials", user)

	return user, nil
}

func (idcSvc *IDCStorageServiceClient) WaitForStorageStateReadygRPC(ctx context.Context, storageId, cloudAccount string, timeout time.Duration) error {
	log := log.FromContext(ctx).WithName("IDCStorageServiceClient.WaitForStorageStateReadygRPC")
	log.Info("get and wait instace state ", "instanceId", storageId)

	backoffTimer := retry.NewConstant(5 * time.Second)
	backoffTimer = retry.WithMaxDuration(timeout*time.Second, backoffTimer)

	storageGetRequest := &v1.FilesystemGetRequestPrivate{
		Metadata: &v1.FilesystemMetadataReference{
			CloudAccountId: cloudAccount,
			NameOrId: &v1.FilesystemMetadataReference_ResourceId{
				ResourceId: storageId,
			},
		},
	}

	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		storage, err := v1.NewFilesystemPrivateServiceClient(idcSvc.StorageAPIConn).GetPrivate(ctx, storageGetRequest)
		if err != nil {
			return fmt.Errorf("error reading storage state")
		}

		if storage == nil {
			return fmt.Errorf("error reading storage state, nil storage")
		}

		log.Info("DEBUG", "storage-phase", storage.Status.Phase)
		if storage.Status.Phase.String() != "FSReady" {
			return retry.RetryableError(fmt.Errorf("storage state not ready, retry again"))
		}

		log.Info("DEBUG", "storage state ready", storageId, "CloudAccountId", cloudAccount)
		return nil
	}); err != nil {
		return fmt.Errorf("storage state not ready after retries")
	}

	return nil
}

func (idcSvc *IDCStorageServiceClient) GetStorageByName(ctx context.Context, storageName, cloudAccountId string) (*v1.FilesystemPrivate, error) {
	log := log.FromContext(ctx).WithName("IDCStorageServiceClient.GetStorageByNamegRPC")

	storageGetRequest := &v1.FilesystemGetRequestPrivate{
		Metadata: &v1.FilesystemMetadataReference{
			CloudAccountId: cloudAccountId,
			NameOrId: &v1.FilesystemMetadataReference_Name{
				Name: storageName,
			},
		},
	}

	storage, err := v1.NewFilesystemPrivateServiceClient(idcSvc.StorageAPIConn).GetPrivate(ctx, storageGetRequest)
	if err != nil {
		log.Error(err, "error getting storage via gRPC")
		return nil, fmt.Errorf("error getting storage request by name")
	}

	return storage, nil
}

func (idcSvc *IDCStorageServiceClient) DeleteStorage(ctx context.Context, storageName, cloudAccountId string) error {
	log := log.FromContext(ctx).WithName("IDCStorageServiceClient.DeleteStorage")

	storageDeleteRequest := &v1.FilesystemDeleteRequestPrivate{
		Metadata: &v1.FilesystemMetadataReference{
			CloudAccountId: cloudAccountId,
			NameOrId: &v1.FilesystemMetadataReference_Name{
				Name: storageName,
			},
		},
	}

	_, err := v1.NewFilesystemPrivateServiceClient(idcSvc.StorageAPIConn).DeletePrivate(ctx, storageDeleteRequest)
	if err != nil {
		log.Error(err, "error deleting storage via gRPC endpoint")
		return err
	}

	return nil
}
