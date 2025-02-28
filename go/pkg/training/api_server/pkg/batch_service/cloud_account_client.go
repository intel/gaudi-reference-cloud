// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"
	"io"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
)

type CloudAccountSvcClient struct {
	CloudAccountClient pb.CloudAccountServiceClient
}

func NewCloudAccountClient(ctx context.Context, resolver grpcutil.Resolver) (*CloudAccountSvcClient, error) {
	logger := log.FromContext(ctx).WithName("CloudAccountClient.NewCloudAccountClient")
	var cloudAccountConn *grpc.ClientConn

	cloudAccountAddr := os.Getenv("CLOUDACCOUNT_ADDR")
	if cloudAccountAddr == "" {
		cloudAccountAddr, err := resolver.Resolve(ctx, "cloudaccount")
		if err != nil {
			logger.Error(err, "grpc resolver not able to resolve", "addr", cloudAccountAddr)
			return nil, err
		}
	}

	cloudAccountConn, err := grpcConnect(ctx, cloudAccountAddr)
	if err != nil {
		return nil, err
	}
	ca := pb.NewCloudAccountServiceClient(cloudAccountConn)
	return &CloudAccountSvcClient{CloudAccountClient: ca}, nil
}

func (cloudAccount *CloudAccountSvcClient) GetCloudAccount(ctx context.Context, accountId *pb.CloudAccountId) (*pb.CloudAccount, error) {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.GetCloudAccount")
	account, err := cloudAccount.CloudAccountClient.GetById(ctx, accountId)
	if err != nil {
		logger.Error(err, "error in cloud account client response")
		return nil, err
	}
	logger.Info("cloudaccount response", "account", account)
	return account, nil
}

func (cloudAccount *CloudAccountSvcClient) GetCloudAccountType(ctx context.Context, accountId *pb.CloudAccountId) (pb.AccountType, error) {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.GetCloudAccountType")
	account, err := cloudAccount.CloudAccountClient.GetById(ctx, accountId)
	if err != nil {
		logger.Error(err, "error in cloud account client response")
		return pb.AccountType_ACCOUNT_TYPE_UNSPECIFIED, err
	}
	logger.Info("cloudaccount response", "account", account)
	return account.GetType(), nil
}

func (cloudAccount *CloudAccountSvcClient) GetAllCloudAccount(ctx context.Context) ([]*pb.CloudAccount, error) {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.GetAllCloudAccount")
	cloudAccounts := []*pb.CloudAccount{}
	accStream, err := cloudAccount.CloudAccountClient.Search(ctx, &pb.CloudAccountFilter{})
	if err != nil {
		logger.Error(err, "error in cloud account client response")
		return nil, err
	}
	recvError := make(chan error)
	go func() {
		for {
			currAcc, err := accStream.Recv()
			if err == io.EOF {
				recvError <- nil //close(done)
				return
			}
			if err != nil {
				logger.Error(err, "failed to read from stream")
				recvError <- err
				return
			}
			cloudAccounts = append(cloudAccounts, currAcc)
		}
	}()

	recvErr := <-recvError
	if recvErr != nil {
		return nil, recvErr
	}

	logger.Info("cloudaccount response", "# cloudaccount ", len(cloudAccounts))
	return cloudAccounts, nil
}
