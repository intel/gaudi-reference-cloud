// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthzClient struct {
	AuthzServiceClient pb.AuthzServiceClient
}

func NewAuthzClient(ctx context.Context, resolver grpcutil.Resolver) (*AuthzClient, error) {
	logger := log.FromContext(ctx).WithName("AuthzClient.InitAuthzClient")
	addr, err := resolver.Resolve(ctx, "authz")
	if err != nil {
		logger.Error(err, "grpc resolver not able to connect", "addr", addr)
		return nil, err
	}
	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		return nil, err
	}
	authzServiceClient := pb.NewAuthzServiceClient(conn)
	return &AuthzClient{
		AuthzServiceClient: pb.AuthzServiceClient(authzServiceClient),
	}, nil
}

func (authzClient *AuthzClient) AssignSystemRole(ctx context.Context, req *pb.RoleRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("AuthzClient.AssignSystemRole")
	logger.Info("assign authz systemrole")
	resp, err := authzClient.AuthzServiceClient.AssignSystemRole(ctx, req)
	if err != nil {
		logger.Error(err, "unable to assign system role", "cloudAccountId", req.CloudAccountId, "systemRole", req.SystemRole)
		return nil, err
	}
	return resp, nil
}

func (authzClient *AuthzClient) UnassignSystemRole(ctx context.Context, roleRequest *pb.RoleRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("AuthzClient.UnassignSystemRole")
	logger.Info("unassign authz systemrole")
	resp, err := authzClient.AuthzServiceClient.UnassignSystemRole(ctx, roleRequest)
	if err != nil {
		logger.Error(err, "unable to unassign system role", "cloudAccountId", roleRequest.CloudAccountId, "systemRole", roleRequest.SystemRole)
		return nil, err
	}
	return resp, nil
}

func (authzClient *AuthzClient) SystemRoleExists(ctx context.Context, roleRequest *pb.RoleRequest) (bool, error) {
	logger := log.FromContext(ctx).WithName("AuthzClient.SystemRoleExists")
	logger.V(9).Info("check authz systemrole exist")
	resp, err := authzClient.AuthzServiceClient.SystemRoleExists(ctx, roleRequest)
	if err != nil {
		logger.Error(err, "unable to check if system role exists", "cloudAccountId", roleRequest.CloudAccountId, "systemrole", roleRequest.SystemRole)
		return false, err
	}
	return resp.Exist, nil
}

func (authzClient *AuthzClient) RemoveUserFromCloudAccountRole(ctx context.Context, roleRequest *pb.CloudAccountRoleUserRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("AuthzClient.RemoveUserFromCloudAccountRole").WithValues("cloudAccountId", roleRequest.CloudAccountId, "cloudAccountRoleId", roleRequest.Id)
	logger.V(9).Info("remove user from cloud account role authorization")
	resp, err := authzClient.AuthzServiceClient.RemoveUserFromCloudAccountRole(ctx, roleRequest)
	if err != nil {
		logger.Error(err, "failed to remove user from cloud account role")
		return nil, err
	}
	return resp, nil
}

func (authzClient *AuthzClient) AddUserToCloudAccountRole(ctx context.Context, req *pb.CloudAccountRoleUserRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("AuthzClient.AddUserToCloudAccountRole").WithValues("cloudAccountId", req.CloudAccountId)
	logger.Info("adding user to cloud account role", "cloudAccountRoleId", req.Id)
	_, err := authzClient.AuthzServiceClient.AddUserToCloudAccountRole(ctx, req)
	if err != nil {
		logger.Error(err, "unable to add user to cloud account role.", "cloudAccountRoleId", req.Id)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (authzClient *AuthzClient) AssignDefaultCloudAccountRole(ctx context.Context, req *pb.AssignDefaultCloudAccountRoleRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("AuthzClient.AssignDefaultCloudAccountRole").WithValues("cloudAccountId", req.CloudAccountId)
	_, err := authzClient.AuthzServiceClient.AssignDefaultCloudAccountRole(ctx, req)
	if err != nil {
		logger.Error(err, "unable to assign the default role")
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (authzClient *AuthzClient) DefaultCloudAccountRoleAssigned(ctx context.Context, req *pb.DefaultCloudAccountRoleAssignedRequest) (bool, error) {
	logger := log.FromContext(ctx).WithName("AuthzClient.DefaultCloudAccountRoleAssigned").WithValues("cloudAccountId", req.CloudAccountId)
	resp, err := authzClient.AuthzServiceClient.DefaultCloudAccountRoleAssigned(ctx, req)
	if err != nil {
		logger.Error(err, "unable to check if default role is assigned")
		return false, err
	}
	return resp.Assigned, nil
}
