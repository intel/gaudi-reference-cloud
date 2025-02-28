// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/secrets"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// StorageKMSServiceServer is used to implement pb.UnimplementedFileStorageServiceServer
type StorageKMSServiceServer struct {
	pb.UnimplementedStorageKMSPrivateServiceServer
	vault secrets.SecretManager
}

func NewStorageKMSService(vault *secrets.Vault) (*StorageKMSServiceServer, error) {

	return &StorageKMSServiceServer{vault: vault}, nil
}

func NewMockStorageKMSService(vault secrets.SecretManager) (*StorageKMSServiceServer, error) {

	return &StorageKMSServiceServer{vault: vault}, nil
}

func (scheduler *StorageKMSServiceServer) Put(ctx context.Context, in *pb.StoreSecretRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageKMSServiceServer.Put").WithValues(logkeys.KeyPath, in.KeyPath).Start()
	defer span.End()
	logger.Info("entering km store")
	defer logger.Info("returning from km store")

	interfaceMap := make(map[string]interface{})
	for key, value := range in.Secrets {
		interfaceMap[key] = value
	}

	secret, err := scheduler.vault.PutStorageSecrets(ctx, in.KeyPath, interfaceMap, true)
	if err != nil {
		logger.Error(err, "PUT Error")
	}

	logger.Info("Returned data from write", logkeys.Secret, secret)
	return &emptypb.Empty{}, nil
}

func (scheduler *StorageKMSServiceServer) Delete(ctx context.Context, in *pb.DeleteSecretRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageKMSServiceServer.Delete").WithValues(logkeys.KeyPath, in.KeyPath).Start()
	defer span.End()

	logger.Info("entering kms delete")
	defer logger.Info("returning from kms delete")

	err := scheduler.vault.DeleteStorageCredentials(ctx, in.KeyPath, true)
	if err != nil {
		logger.Error(err, "error deleting secrets", logkeys.SecretsPath, in.KeyPath)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (scheduler *StorageKMSServiceServer) Get(ctx context.Context, in *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageKMSServiceServer.Get").WithValues(logkeys.KeyPath, in.KeyPath).Start()
	defer span.End()

	logger.Info("entering kms get")
	defer logger.Info("returning from filesystem get")

	user, pass, userId, nsId, err := scheduler.vault.GetStorageCredentials(ctx, in.KeyPath, true)
	if err != nil {
		logger.Error(err, "error reading secrets", logkeys.SecretsPath, in.KeyPath)
		return nil, err
	}

	// Create a map[string]string and convert the values
	resp := pb.GetSecretResponse{
		Secrets: make(map[string]string),
	}
	resp.Secrets["username"] = user
	resp.Secrets["password"] = pass
	resp.Secrets["userId"] = userId
	resp.Secrets["namespaceId"] = nsId

	return &resp, nil
}

func (scheduler *StorageKMSServiceServer) PingPrivate(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("StorageKMSServiceServer.Ping")

	logger.Info("entering storage kms private Ping")
	defer logger.Info("returning from storage kms private Ping")

	return &emptypb.Empty{}, nil
}
