// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package storagecontroller

import (
	"context"
	"fmt"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	storageControllerWekaApi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"
)

type FilesystemMetadata struct {
	FileSystemName string
	Encrypted      bool
	AuthRequired   bool
	User           string
	Password       string
	NamespaceName  string
	UUID           string
	Backend        string
}

type FilesystemProperties struct {
	FileSystemCapacity string
}

type Filesystem struct {
	Metadata   FilesystemMetadata
	Properties FilesystemProperties
}

// Checks if the Filesystem with given name exists. There is no `head` method, so we will try to get
// the Filesystem and check the failed message. Filesystem lookup might failed because of:
// - grpc connection error
// - invalid credentials
// error is populated with right message in those cases
func (client *StorageControllerClient) IsFilesystemExists(ctx context.Context, queryParams FilesystemMetadata, deleteFlag bool) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.IsFilesystemExists").Start()
	defer span.End()
	logger.Info("checking if filesystem exists")

	clusterUuid := queryParams.UUID
	authCtx := newBasicAuthContext(queryParams.User, queryParams.Password)

	// Need to be removed after ID is saved
	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, queryParams.NamespaceName)
	if err != nil {
		logger.Error(err, "error in finding filesystem by name, namespace does not exists ")
		return exists, err
	}
	if !exists {
		logger.Info("namespace does not exists")
		return exists, nil
	}
	_, exists, err = client.getFilesystemByName(ctx, ns.Id, queryParams.FileSystemName, authCtx)
	if err != nil {
		logger.Error(err, "error in finding filesystem by name ")
	}
	if !exists && deleteFlag {
		logger.Info("could not find filesystem by name ")
	}

	// END to be removed after ID is saved

	return exists, nil
}

func intoFileSystem(fsResponse *storageControllerWekaApi.Filesystem) Filesystem {
	quota := strconv.FormatUint(fsResponse.GetCapacity().GetTotalBytes(), 10)
	filesystem := Filesystem{
		Metadata: FilesystemMetadata{
			FileSystemName: fsResponse.Name,
			Encrypted:      fsResponse.IsEncrypted,
			Backend:        fsResponse.Backend,
		},
		Properties: FilesystemProperties{
			FileSystemCapacity: quota,
		},
	}
	return filesystem
}

func intoFileSystemArray(fsResponses []*storageControllerWekaApi.Filesystem) []Filesystem {
	var filesystems []Filesystem
	for _, fsResponse := range fsResponses {
		quota := strconv.FormatUint(fsResponse.GetCapacity().GetTotalBytes(), 10)
		filesystem := Filesystem{
			Metadata: FilesystemMetadata{
				FileSystemName: fsResponse.Name,
				Encrypted:      fsResponse.IsEncrypted,
				Backend:        fsResponse.Backend,
			},
			Properties: FilesystemProperties{
				FileSystemCapacity: quota,
			},
		}
		filesystems = append(filesystems, filesystem)
	}
	return filesystems
}

// Creates a filesystem with given quota and basic auth for access control and returns the cluster Addr which it's run on
func (client *StorageControllerClient) CreateFilesystem(ctx context.Context, filesystem Filesystem) (Filesystem, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.CreateFilesystem").Start()
	defer span.End()

	logger.Info("starting filesystem creation")

	clusterUuid := filesystem.Metadata.UUID
	authCtx := newBasicAuthContext(filesystem.Metadata.User, filesystem.Metadata.Password)
	filesystemObject := Filesystem{}

	// START Need to be removed after ID is saved
	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, filesystem.Metadata.NamespaceName)
	if err != nil {
		logger.Error(err, "error in finding namespace by name ")
		return Filesystem{}, err
	}
	if !exists {
		logger.Info("could not find namespace by name ")
		return Filesystem{}, fmt.Errorf("Namespace does not exist")
	}
	// END to be removed after ID is saved
	capacity, err := strconv.ParseInt(filesystem.Properties.FileSystemCapacity, 10, 64)
	if err != nil {
		logger.Error(err, "cannot parse capacity string")
	}
	logger.Info("createFilesystem sds input params", logkeys.Namespace, ns.Id, logkeys.Name, filesystem.Metadata.FileSystemName, logkeys.FilesystemCapacity, filesystem.Properties.FileSystemCapacity)
	resp, err := client.WekaFilesystemSvcClient.CreateFilesystem(ctx, &storageControllerWekaApi.CreateFilesystemRequest{
		NamespaceId:  ns.Id,
		Name:         filesystem.Metadata.FileSystemName,
		TotalBytes:   uint64(capacity),
		Encrypted:    filesystem.Metadata.Encrypted,
		AuthRequired: filesystem.Metadata.AuthRequired,
		AuthCtx:      authCtx,
	})
	if err != nil {
		logger.Error(err, "Error in creating a FileSystem in the controller ")
		return Filesystem{}, err
	}

	if resp != nil {
		logger.Info("fs details response ", logkeys.Filesystem, resp.Filesystem)
		filesystemObject = intoFileSystem(resp.Filesystem)
		return filesystemObject, nil
	} else {
		return Filesystem{}, err
	}

}

// Gets filesystem object from the storage controller
func (client *StorageControllerClient) GetFilesystem(ctx context.Context, queryParams FilesystemMetadata) (Filesystem, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.GetFilesystem")
	logger.Info("get filesystem object")
	res := Filesystem{}
	return res, nil
}

// Gets filesystem object from the storage controller
func (client *StorageControllerClient) DeleteFilesystem(ctx context.Context, queryParams FilesystemMetadata) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.DeleteFilesystem").Start()
	defer span.End()

	logger.Info("delete filesystem object")

	clusterUuid := queryParams.UUID
	authCtx := newBasicAuthContext(queryParams.User, queryParams.Password)

	// Need to be removed after ID is saved
	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, queryParams.NamespaceName)
	if err != nil {
		logger.Error(err, "error in finding filesystem by name, namespace does not exists ")
		return err
	}
	if !exists {
		logger.Info("could not find filesystem by name, namespace does not exists ")
		return nil
	}
	fs, exists, err := client.getFilesystemByName(ctx, ns.Id, queryParams.FileSystemName, authCtx)
	if err != nil {
		logger.Error(err, "error in finding filesystem by name ")
		return err
	}
	if !exists {
		logger.Info("could not find filesystem by name ")
		return nil
	}
	// END to be removed after ID is saved
	logger.Info("deleteFilesystem sds input params", logkeys.Filesystem, fs.Id)
	_, err = client.WekaFilesystemSvcClient.DeleteFilesystem(ctx, &storageControllerWekaApi.DeleteFilesystemRequest{
		FilesystemId: fs.Id,
		AuthCtx:      authCtx,
	})
	if err != nil {
		logger.Error(err, "Error in Deleting a FileSystem in the controller ")
		return err
	}
	return nil
}

// Get All Filesystems in a namespace. There is no `head` method, so we will try to get
// the Filesystem and check the failed message. Filesystem lookup might failed because of:
// - grpc connection error
// - invalid credentials
// error is populated with right message in those cases
func (client *StorageControllerClient) GetAllFileSystems(ctx context.Context, queryParams FilesystemMetadata) ([]Filesystem, bool, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.GetAllFileSystems")
	logger.Info("Getting all Filesystem from a Namespace")

	clusterUuid := queryParams.UUID
	authCtx := newBasicAuthContext(queryParams.User, queryParams.Password)

	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, queryParams.NamespaceName)
	if err != nil {
		logger.Error(err, "error in finding filesystem by name, namespace does not exists ")
		return nil, exists, err
	}
	if !exists {
		logger.Info("given namespace does not exists", logkeys.Namespace, ns.Name)
		return nil, exists, nil
	}

	fsList, exists, err := client.readAllFilesystems(ctx, ns.Id, authCtx)
	if err != nil {
		logger.Error(err, "error in finding filesystems ")
		return nil, false, err
	}
	if !exists {
		logger.Info("could not find any filesystem for the ns ", logkeys.Namespace, ns.Name)
		return nil, exists, nil
	}

	fileSystemList := intoFileSystemArray(fsList)

	logger.Info("Filesystem Array", logkeys.FilesystemList, fileSystemList)

	return fileSystemList, exists, nil
}

// Updates a filesystem with given quota
func (client *StorageControllerClient) UpdateFilesystem(ctx context.Context, filesystem Filesystem) (Filesystem, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.UpdateFilesystem").Start()
	defer span.End()

	logger.Info("starting update filesystem ")

	clusterUuid := filesystem.Metadata.UUID
	authCtx := newBasicAuthContext(filesystem.Metadata.User, filesystem.Metadata.Password)

	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, filesystem.Metadata.NamespaceName)
	if err != nil {
		logger.Error(err, "error in finding namespace by name ")
		return Filesystem{}, err
	}
	if !exists {
		logger.Info("could not find namespace by name ", logkeys.Namespace, ns.Name)
		return Filesystem{}, nil
	}

	filesystemObject, exists, err := client.getFilesystemByName(ctx, ns.Id, filesystem.Metadata.FileSystemName, authCtx)
	if err != nil {
		logger.Error(err, "error in finding filesystem by name ", logkeys.FilesystemName, filesystem.Metadata.FileSystemName)
		return Filesystem{}, err
	}
	if !exists {
		logger.Info("could not find filesystem by name", logkeys.FilesystemName, filesystem.Metadata.FileSystemName)
		return Filesystem{}, nil
	}

	logger.Info("filesystem update specs", logkeys.FilesystemMetadata, filesystem.Metadata)
	capacity, err := strconv.ParseInt(filesystem.Properties.FileSystemCapacity, 10, 64)
	if err != nil {
		logger.Error(err, "cannot parse capacity string")
		return Filesystem{}, err
	}
	newFilesystem := Filesystem{}
	newName := filesystem.Metadata.FileSystemName
	newTotalBytes := uint64(capacity)
	newAuthRequired := filesystem.Metadata.AuthRequired
	logger.Info("UpdateFilesystem sds input params", logkeys.Filesystem, filesystemObject.Id, logkeys.FilesystemName, newName, logkeys.FilesystemCapacity, filesystem.Properties.FileSystemCapacity)
	resp, err := client.WekaFilesystemSvcClient.UpdateFilesystem(ctx, &storageControllerWekaApi.UpdateFilesystemRequest{
		FilesystemId:    filesystemObject.Id,
		NewName:         &newName,
		NewTotalBytes:   &newTotalBytes,
		NewAuthRequired: &newAuthRequired,
		AuthCtx:         authCtx,
	})
	if err != nil {
		logger.Error(err, "error in updating a filesystem in the contorller ")
		return Filesystem{}, err
	}

	if resp != nil {
		logger.Info("fs details response: ", logkeys.Filesystem, resp.Filesystem)
		newFilesystem = intoFileSystem(resp.Filesystem)
		return newFilesystem, nil
	} else {
		return Filesystem{}, err
	}
}
