// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package storagecontroller

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
)

type UserMetadata struct {
	Role              string
	NamespaceUser     string
	NamespacePassword string
	NamespaceName     string
	UUID              string
	NamespaceId       string
}

type UserProperties struct {
	NewUser         string
	NewUserPassword string
}

type User struct {
	Metadata   UserMetadata
	Properties UserProperties
}

type DeletUserData struct {
	NamespaceUser     string
	NamespacePassword string
	NamespaceName     string
	UsertoBeDeleted   string
	UUID              string
}

type DeleteUserData struct {
	ClusterUUID string
	NamespaceID string
	UserID      string
}

type UpdateUserpassword struct {
	NamespaceUser     string
	NamespacePassword string
	NamespaceName     string
	UsertoBeUpdated   string
	NewPassword       string
	UUID              string
}

// Checks if the User with given name exists. There is no `head` method, so we will try to get
// the User and check the failed message. User lookup might failed because of:
// - grpc connection error
// - invalid credentials
// error is populated with right message in those cases
func (client *StorageControllerClient) IsUserExists(ctx context.Context, queryParams User) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.IsUserExists").Start()
	defer span.End()
	logger.Info("checking if user exists")

	clusterUuid := queryParams.Metadata.UUID
	authCtx := newBasicAuthContext(queryParams.Metadata.NamespaceUser, queryParams.Metadata.NamespacePassword)

	// START to be removed after ID is saved
	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, queryParams.Metadata.NamespaceName)
	if !exists {
		logger.Info("could not find user by name, namespace does not exists")
		return exists, nil
	}
	if err != nil {
		logger.Error(err, "error in finding user by name, namespace does not exists")
		return exists, err
	}
	_, exists, err = client.getUserByName(ctx, ns.Id, queryParams.Properties.NewUser, authCtx)
	// END to be removed after ID is saved

	if err != nil {
		logger.Error(err, "error in finding user by name")
	}
	return exists, nil
}

// Creates a User with given auth credentials and auth role for access control
func (client *StorageControllerClient) CreateUser(ctx context.Context, user User) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.CreateUser").Start()
	defer span.End()
	logger.Info("starting user creation")

	clusterUuid := user.Metadata.UUID
	authCtx := newBasicAuthContext(user.Metadata.NamespaceUser, user.Metadata.NamespacePassword)

	// START to be removed after ID is saved
	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, user.Metadata.NamespaceName)
	if !exists {
		logger.Info("could not find namespace by name")
	}
	if err != nil {
		logger.Error(err, "error in finding namespace by name")
	}
	// END to be removed after ID is saved

	// Make a gRPC call to add a new user.
	_, err = client.UserSvcClient.CreateUser(ctx, &stcnt_api.CreateUserRequest{
		NamespaceId:  ns.Id,
		UserName:     user.Properties.NewUser,
		UserPassword: user.Properties.NewUserPassword,
		Role:         stcnt_api.User_ROLE_REGULAR,
		AuthCtx:      authCtx,
	})
	if err != nil {
		logger.Error(err, "error in creating a user in the controller.")
		return err
	}
	return nil
}
func (client *StorageControllerClient) CreateVastUser(ctx context.Context, user User) (*stcnt_api.CreateUserResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.CreateVastUser").Start()
	defer span.End()
	logger.Info("starting vast fs user creation")

	userResponse := &stcnt_api.CreateUserResponse{}
	userResponse, err := client.UserSvcClient.CreateUser(ctx, &stcnt_api.CreateUserRequest{
		NamespaceId: &stcnt_api.NamespaceIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: user.Metadata.UUID,
			},
			Id: user.Metadata.NamespaceId,
		},
		UserName:     user.Properties.NewUser,
		UserPassword: user.Properties.NewUserPassword,
		Role:         3,
	})
	if err != nil {
		logger.Error(err, "error in creating a vast user in the controller.")
		return userResponse, err
	}
	return userResponse, nil
}
func (client *StorageControllerClient) DeleteUser(ctx context.Context, queryParams DeletUserData) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.DeleteUser").Start()
	defer span.End()
	logger.Info("delete user object")

	clusterUuid := queryParams.UUID
	authCtx := newBasicAuthContext(queryParams.NamespaceUser, queryParams.NamespacePassword)

	// START to be removed after ID is saved
	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, queryParams.NamespaceName)
	if !exists {
		logger.Info("could not find user by name, namespace does not exists", logkeys.Namespace, ns.Name)
		return nil
	}
	if err != nil {
		logger.Error(err, "error in finding namespace by name")
		return err
	}
	user, exists, err := client.getUserByName(ctx, ns.Id, queryParams.UsertoBeDeleted, authCtx)
	if !exists {
		logger.Info("could not find user by name")
		return nil
	}
	if err != nil {
		logger.Error(err, "error in finding user by name")
		return err
	}
	// END to be removed after ID is saved

	_, err = client.UserSvcClient.DeleteUser(ctx, &stcnt_api.DeleteUserRequest{
		UserId:  user.Id,
		AuthCtx: authCtx,
	})
	if err != nil {
		logger.Error(err, "error in Deleting User data in the controller ")
		return err
	}
	return nil
}

func (client *StorageControllerClient) DeleteVastUser(ctx context.Context, queryParams DeleteUserData) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.DeleteVastUser").Start()
	defer span.End()
	logger.Info("delete vast user object")

	// Construct the UserIdentifier
	userIdentifier := &stcnt_api.UserIdentifier{
		NamespaceId: &stcnt_api.NamespaceIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: queryParams.ClusterUUID,
			},
			Id: queryParams.NamespaceID,
		},
		Id: queryParams.UserID,
	}

	// Construct the DeleteUserRequest
	deleteUserRequest := &stcnt_api.DeleteUserRequest{
		UserId: userIdentifier,
	}

	// Call the DeleteUser method
	_, err := client.UserSvcClient.DeleteUser(ctx, deleteUserRequest)
	if err != nil {
		logger.Error(err, "error in deleting user data in the controller")
		return err
	}
	return nil
}

func (client *StorageControllerClient) UpdateUserPassword(ctx context.Context, queryParams UpdateUserpassword) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.UpdateUserPassword").Start()
	defer span.End()
	logger.Info("update user object with new password")

	clusterUuid := queryParams.UUID
	authCtx := newBasicAuthContext(queryParams.NamespaceUser, queryParams.NamespacePassword)

	// START to be removed after ID is saved
	ns, exists, err := client.getNamespaceByName(ctx, clusterUuid, queryParams.NamespaceName)
	if !exists {
		logger.Info("could not find namespace by name, namespace does not exists")
		return nil
	}
	if err != nil {
		logger.Error(err, "error in finding namespace by name")
		return err
	}
	user, exists, err := client.getUserByName(ctx, ns.Id, queryParams.UsertoBeUpdated, authCtx)
	if !exists {
		logger.Info("could not find user by name")
		return nil
	}
	if err != nil {
		logger.Error(err, "error in finding user by name")
		return err
	}
	// END to be removed after ID is saved

	_, err = client.UserSvcClient.UpdateUserPassword(ctx, &stcnt_api.UpdateUserPasswordRequest{
		UserId:      user.Id,
		NewPassword: queryParams.NewPassword,
		AuthCtx:     authCtx,
	})
	if err != nil {
		logger.Error(err, "error in updating User data in the controller")
		return err
	}
	return nil
}
