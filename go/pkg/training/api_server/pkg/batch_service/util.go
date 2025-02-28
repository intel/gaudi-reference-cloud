// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// used for AWS Cognito client to generate auth-token
	cognitoEnabled, _ = strconv.ParseBool(os.Getenv("IDC_COGNITO_ENABLED"))
	cognitoURL, _     = url.Parse(os.Getenv("IDC_COGNITO_ENDPOINT"))
)

func grpcConnect(ctx context.Context, addr string) (*grpc.ClientConn, error) {

	var conn *grpc.ClientConn
	var err error

	logger := log.FromContext(ctx)

	if cognitoEnabled {
		// create the cognitoClient to access AWS Cognito
		cognitoClient, err := authutil.NewCognitoClient(&authutil.CognitoConfig{
			URL:     cognitoURL,
			Timeout: 1 * time.Minute,
		})
		if err != nil {
			logger.Error(err, "unable to read AWS Cognito credentials", "addr", addr)
			os.Exit(1)
		}

		// prefetch the access token to access global: cloudaccount svc
		token, err := cognitoClient.GetGlobalAuthToken(ctx)
		if err != nil {
			logger.Error(err, "unable to get AWS Cognito token", "addr", addr)
			os.Exit(1)
		}
		logger.V(9).Info("Prefetched Cognito Token: ", "cognitoToken", token)

		conn, err = grpcutil.NewClient(ctx, addr,
			grpc.WithPerRPCCredentials(authutil.NewCognitoAuth(ctx, cognitoClient)))
		if err != nil {
			logger.Error(err, "Not able to connect to gRPC service using grpcutil.NewClient", "addr", addr)
			os.Exit(1)
		}
	} else {
		conn, err = grpcutil.NewClient(ctx, addr)
		if err != nil {
			logger.Error(err, "Not able to connect to gRPC service using grpcutil.NewClient", "addr", addr)
			os.Exit(1)
		}
	}

	return conn, nil
}

func ToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func ToTime(ts *timestamppb.Timestamp) time.Time {
	return ts.AsTime()
}

// FNV32a hashes using fnv32a algorithm
// func FNV32a(text string) uint32 {
// 	algorithm := fnv.New32a()
// 	algorithm.Write([]byte(text))
// 	return algorithm.Sum32()
// }

func FNV32a(text string) (uint32, error) {
	if text == "" {
		return 0, fmt.Errorf("input string is empty")
	}

	algorithm := fnv.New32a()
	_, err := algorithm.Write([]byte(text))
	if err != nil {
		return 0, err
	}
	return algorithm.Sum32(), nil
}
