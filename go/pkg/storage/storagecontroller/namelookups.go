// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package storagecontroller

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"

	storageControllerApi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	storageWekaApi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"
)

// This file is temporary solution for backward compatibility with the name/uuid lookups.
// Operator should save id returned from resource creation, store it and provide where needed instead of name
// TODO: Remove as soon as possible and replace with ID, will only work with Weka backend
func (client *StorageControllerClient) getNamespaceByName(
	ctx context.Context,
	clusterUuid string,
	namespaceName string,
) (*storageControllerApi.Namespace, bool, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.getNamespaceByName")
	logger.Info("ListNamespace sds input params", logkeys.ClusterId, clusterUuid, logkeys.Namespace, namespaceName)
	resp, err := client.NamespaceSvcClient.ListNamespaces(ctx, &storageControllerApi.ListNamespacesRequest{
		ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: clusterUuid},
		Filter: &storageControllerApi.ListNamespacesRequest_Filter{
			Names: []string{namespaceName},
		},
	})
	if err != nil {
		logger.Error(err, "error calling ListNamespaces method")
		return nil, false, err
	}

	if len(resp.Namespaces) > 1 {
		err = fmt.Errorf("%d!=1", len(resp.Namespaces))
		logger.Error(err, "unexpected number of namespaces returned by name")
		return nil, false, err
	} else if len(resp.Namespaces) == 0 {
		return nil, false, nil
	}

	return resp.Namespaces[0], true, nil
}

func (client *StorageControllerClient) getUserByName(
	ctx context.Context,
	namespaceId *storageControllerApi.NamespaceIdentifier,
	username string,
	auth *storageControllerApi.AuthenticationContext,
) (*storageControllerApi.User, bool, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.getUserByName")
	logger.Info("ListUsers sds input params", logkeys.Namespace, namespaceId, logkeys.UserName, username)
	resp, err := client.UserSvcClient.ListUsers(ctx, &storageControllerApi.ListUsersRequest{
		NamespaceId: namespaceId,
		AuthCtx:     auth,
		Filter: &storageControllerApi.ListUsersRequest_Filter{
			Names: []string{username},
		},
	})
	if err != nil {
		logger.Error(err, "error calling ListUsers method")
		return nil, false, err
	}

	if len(resp.Users) > 1 {
		err = fmt.Errorf("%d!=1", len(resp.Users))
		logger.Error(err, "unexpected number of users returned by name")
		return nil, false, err
	} else if len(resp.Users) == 0 {
		return nil, false, nil
	}

	return resp.Users[0], true, nil
}

func (client *StorageControllerClient) getFilesystemByName(
	ctx context.Context,
	namespaceId *storageControllerApi.NamespaceIdentifier,
	fsName string,
	auth *storageControllerApi.AuthenticationContext,
) (*storageWekaApi.Filesystem, bool, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.getFilesystemByName")
	logger.Info("ListFilesystem sds input params", logkeys.Namespace, namespaceId, logkeys.FilesystemName, fsName)
	resp, err := client.WekaFilesystemSvcClient.ListFilesystems(ctx, &storageWekaApi.ListFilesystemsRequest{
		NamespaceId: namespaceId,
		AuthCtx:     auth,
		Filter: &storageWekaApi.ListFilesystemsRequest_Filter{
			Names: []string{fsName},
		},
	})
	if err != nil {
		logger.Error(err, "error calling ListFilesystems method")
		return nil, false, err
	}

	if len(resp.Filesystems) > 1 {
		err = fmt.Errorf("%d!=1", len(resp.Filesystems))
		logger.Error(err, "unexpected number of filesystems returned by name")
		return nil, false, err
	} else if len(resp.Filesystems) == 0 {
		return nil, false, nil
	}

	return resp.Filesystems[0], true, nil
}

func (client *StorageControllerClient) readAllFilesystems(
	ctx context.Context,
	namespaceId *storageControllerApi.NamespaceIdentifier,
	auth *storageControllerApi.AuthenticationContext,
) ([]*storageWekaApi.Filesystem, bool, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.readAllFilesystems")
	logger.Info("ListFilesystems sds input params", logkeys.Namespace, namespaceId)
	resp, err := client.WekaFilesystemSvcClient.ListFilesystems(ctx, &storageWekaApi.ListFilesystemsRequest{
		NamespaceId: namespaceId,
		AuthCtx:     auth,
	})
	if err != nil {
		logger.Error(err, "error calling list filesystems method")
		return nil, false, err
	}

	if len(resp.Filesystems) >= 1 {
		logger.Info("filesystems returned from the get all filesystems")
		return resp.Filesystems, true, nil
	} else {
		return []*storageWekaApi.Filesystem{}, false, nil
	}
}

func (client *StorageControllerClient) getNamespaces(
	ctx context.Context,
	clusterUuid string,
) ([]*storageControllerApi.Namespace, bool, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.getNamespaces")
	resp, err := client.NamespaceSvcClient.ListNamespaces(ctx, &storageControllerApi.ListNamespacesRequest{
		ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: clusterUuid},
	})
	if err != nil {
		logger.Error(err, "error calling list namespaces method")
		return nil, false, err
	}
	if len(resp.Namespaces) == 0 {
		return nil, false, nil
	}
	return resp.Namespaces, true, nil
}
