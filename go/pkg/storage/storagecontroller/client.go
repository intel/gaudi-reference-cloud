// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package storagecontroller

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	stcnt_vast_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/vast"
	stcnt_weka_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type StorageControllerClient struct {
	NamespaceSvcClient      stcnt_api.NamespaceServiceClient
	UserSvcClient           stcnt_api.UserServiceClient
	WekaFilesystemSvcClient stcnt_weka_api.FilesystemServiceClient
	ClusterSvcClient        stcnt_api.ClusterServiceClient
	S3ServiceClient         stcnt_api.S3ServiceClient
	StatefulSvcClient       stcnt_weka_api.StatefulClientServiceClient
	VastFilesystemSvcClient stcnt_vast_api.FilesystemServiceClient
}

// Init
// TODO: Remove clusterAddr input argument
func (client *StorageControllerClient) Init(ctx context.Context, serverAddr string, useMtls bool) error {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.Init")
	if serverAddr == "" {
		return fmt.Errorf("storage controller server address missing")
	}

	// Declare options for gRPC dial.
	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}
	clientConn := &grpc.ClientConn{}
	var err error
	if useMtls {
		clientConn, err = grpcutil.NewClient(ctx, serverAddr, dialOptions...)
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
		clientConn, err = grpc.Dial(serverAddr, dialOptions...)
	}
	if err != nil {
		logger.Error(err, "unable to obtain connection for storage controller", logkeys.ServerAddr, serverAddr)
		return fmt.Errorf("storage controller server grpc dial failed")
	}
	client.NamespaceSvcClient = stcnt_api.NewNamespaceServiceClient(clientConn)
	client.UserSvcClient = stcnt_api.NewUserServiceClient(clientConn)
	client.WekaFilesystemSvcClient = stcnt_weka_api.NewFilesystemServiceClient(clientConn)
	client.ClusterSvcClient = stcnt_api.NewClusterServiceClient(clientConn)
	client.S3ServiceClient = stcnt_api.NewS3ServiceClient(clientConn)
	client.StatefulSvcClient = stcnt_weka_api.NewStatefulClientServiceClient(clientConn)
	client.VastFilesystemSvcClient = stcnt_vast_api.NewFilesystemServiceClient(clientConn)

	return nil
}

func newBasicAuthContext(username, password string) *stcnt_api.AuthenticationContext {
	return &stcnt_api.AuthenticationContext{
		Scheme: &stcnt_api.AuthenticationContext_Basic_{
			Basic: &stcnt_api.AuthenticationContext_Basic{
				Principal:   username,
				Credentials: password,
			},
		},
	}
}
