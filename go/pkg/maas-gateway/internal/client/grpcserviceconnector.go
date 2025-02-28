// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"crypto/tls"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	// AWS Cognito client for auth token generation, and required by global grpc-proxy
	cognitoEnabled, _ = strconv.ParseBool(os.Getenv("IDC_COGNITO_ENABLED"))
	cognitoURL, _     = url.Parse(os.Getenv("IDC_COGNITO_ENDPOINT"))
)

type GrpcServiceConnector struct {
	logger logr.Logger
}

func NewGrpcServiceConnector(log logr.Logger) *GrpcServiceConnector {
	logger := log.WithName("GrpcServiceConnector")

	return &GrpcServiceConnector{
		logger: logger,
	}
}

func (c *GrpcServiceConnector) GetIdcConnection(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error

	if cognitoEnabled {
		// create the cognitoClient to access AWS Cognito
		cognitoClient, err := authutil.NewCognitoClient(&authutil.CognitoConfig{
			URL:     cognitoURL,
			Timeout: 1 * time.Minute,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to read AWS Cognito credentials, %s:%s", logkeys.Address, addr)
		}

		conn, err = grpcutil.NewClient(ctx, addr,
			grpc.WithPerRPCCredentials(authutil.NewCognitoAuth(ctx, cognitoClient)))
		if err != nil {
			return nil, errors.Wrapf(err, "Not able to connect to gRPC service using grpcutil.NewClient, %s:%s", logkeys.Address, addr)
		}
	} else {
		conn, err = grpcutil.NewClient(ctx, addr)
		if err != nil {
			return nil, errors.Wrapf(err, "Not able to connect to gRPC service using grpcutil.NewClient, %s:%s", logkeys.Address, addr)
		}
	}

	return conn, nil
}

func (c *GrpcServiceConnector) GetIksConnection(_ context.Context, addr string) (*grpc.ClientConn, error) {
	backoffConfig := backoff.DefaultConfig
	backoffConfig.BaseDelay = 5 * time.Second

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoffConfig,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to the service dispatcher")
	}
	return conn, nil
}
