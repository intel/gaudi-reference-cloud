// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserCredentialsClient struct {
	userCredentialsServiceClient pb.UserCredentialsServiceClient
	cfg                          *config.Config
}

func NewUserCredentialsClient(ctx context.Context, resolver grpcutil.Resolver, cfg *config.Config) (*UserCredentialsClient, error) {
	logger := log.FromContext(ctx).WithName("UserCredentialsClient.NewUserCredentialsClient")
	addr, err := resolver.Resolve(ctx, "user-credentials")
	if err != nil {
		logger.Error(err, "grpc resolver not able to connect", "addr", addr)
		return nil, err
	}
	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		return nil, err
	}
	userCredentialsServiceClient := pb.NewUserCredentialsServiceClient(conn)
	return &UserCredentialsClient{
		userCredentialsServiceClient: pb.UserCredentialsServiceClient(userCredentialsServiceClient),
		cfg:                          cfg,
	}, nil
}

func (userCredentialsClient *UserCredentialsClient) RemoveMemberUserCredentials(ctx context.Context, req *pb.RemoveMemberUserCredentialsRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("UserCredentialsClient.RemoveMemberUserCredentials").WithValues("cloudAccountId", req.CloudaccountId)
	logger.Info("remove member user credentials")
	resp, err := userCredentialsClient.userCredentialsServiceClient.RemoveMemberUserCredentials(ctx, req)
	if err != nil {
		logger.Error(err, "unable to remove user credential", "resp", resp)
		return nil, err
	}
	return resp, nil
}
