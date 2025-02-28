// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type CloudAccountSvcClientInterface interface {
	GetCloudAccount(ctx context.Context, accountId *pb.CloudAccountId) (*pb.CloudAccount, error)
	GetCloudAccountType(ctx context.Context, accountId *pb.CloudAccountId) (pb.AccountType, error)
	GetAllCloudAccount(ctx context.Context) ([]*pb.CloudAccount, error)
}
