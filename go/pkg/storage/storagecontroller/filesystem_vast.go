// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package storagecontroller

import (
	"context"
	"fmt"

	storageControllerApi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	storageControllerVastApi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/vast"
)

// ListFilesystemsParams holds the parameters for listing filesystems
type ListFilesystemsParams struct {
	NamespaceID string
	Names       []string
	ClusterID   string
}

// GetFilesystemParams holds the parameters for getting a filesystem
type GetFilesystemParams struct {
	NamespaceID  string
	FilesystemID string
	ClusterID    string
}

// CreateFilesystemParams holds the parameters for creating a filesystem
type CreateFilesystemParams struct {
	NamespaceID string
	Name        string
	Path        string
	TotalBytes  uint64
	Protocols   []storageControllerVastApi.Filesystem_Protocol
	ClusterID   string
}

// UpdateFilesystemParams holds the parameters for updating a filesystem
type UpdateFilesystemParams struct {
	NamespaceID   string
	FilesystemID  string
	NewName       string
	NewTotalBytes uint64
	ClusterID     string
}

// DeleteFilesystemParams holds the parameters for deleting a filesystem
type DeleteFilesystemParams struct {
	NamespaceID  string
	FilesystemID string
	ClusterID    string
}

func (c *StorageControllerClient) ListVastFilesystems(ctx context.Context, params *ListFilesystemsParams) ([]*storageControllerVastApi.Filesystem, error) {
	req := &storageControllerVastApi.ListFilesystemsRequest{
		NamespaceId: &storageControllerApi.NamespaceIdentifier{Id: params.NamespaceID, ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: params.ClusterID}},
		Filter: &storageControllerVastApi.ListFilesystemsRequest_Filter{
			Names: params.Names,
		},
	}
	resp, err := c.VastFilesystemSvcClient.ListFilesystems(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list filesystems: %w", err)
	}
	return resp.Filesystems, nil
}

func (c *StorageControllerClient) GetVastFilesystem(ctx context.Context, params *GetFilesystemParams) (*storageControllerVastApi.Filesystem, error) {
	req := &storageControllerVastApi.GetFilesystemRequest{
		FilesystemId: &storageControllerVastApi.FilesystemIdentifier{
			NamespaceId: &storageControllerApi.NamespaceIdentifier{Id: params.NamespaceID, ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: params.ClusterID}},
			Id:          params.FilesystemID,
		},
	}
	resp, err := c.VastFilesystemSvcClient.GetFilesystem(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get filesystem: %w", err)
	}

	return resp.Filesystem, nil
}

func (c *StorageControllerClient) CreateVastFilesystem(ctx context.Context, params *CreateFilesystemParams) (*storageControllerVastApi.Filesystem, error) {
	req := &storageControllerVastApi.CreateFilesystemRequest{
		NamespaceId: &storageControllerApi.NamespaceIdentifier{Id: params.NamespaceID, ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: params.ClusterID}},
		Name:        params.Name,
		Path:        params.Path,
		TotalBytes:  params.TotalBytes,
		Protocols:   params.Protocols,
	}
	resp, err := c.VastFilesystemSvcClient.CreateFilesystem(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem: %w", err)
	}
	return resp.Filesystem, nil
}

func (c *StorageControllerClient) UpdateVastFilesystem(ctx context.Context, params *UpdateFilesystemParams) (*storageControllerVastApi.Filesystem, error) {

	req := &storageControllerVastApi.UpdateFilesystemRequest{
		FilesystemId: &storageControllerVastApi.FilesystemIdentifier{
			NamespaceId: &storageControllerApi.NamespaceIdentifier{
				ClusterId: &storageControllerApi.ClusterIdentifier{
					Uuid: params.ClusterID,
				},
				Id: params.NamespaceID,
			},
			Id: params.FilesystemID,
		},
		NewTotalBytes: &params.NewTotalBytes,
	}
	resp, err := c.VastFilesystemSvcClient.UpdateFilesystem(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update filesystem: %w", err)
	}
	return resp.Filesystem, nil
}

func (c *StorageControllerClient) DeleteVastFilesystem(ctx context.Context, params *DeleteFilesystemParams) error {
	req := &storageControllerVastApi.DeleteFilesystemRequest{
		FilesystemId: &storageControllerVastApi.FilesystemIdentifier{
			NamespaceId: &storageControllerApi.NamespaceIdentifier{Id: params.NamespaceID, ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: params.ClusterID}},
			Id:          params.FilesystemID,
		},
	}
	_, err := c.VastFilesystemSvcClient.DeleteFilesystem(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete filesystem: %w", err)
	}
	return nil
}
