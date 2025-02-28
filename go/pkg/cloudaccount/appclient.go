// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"errors"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	FailedToGetCloudAccount          = "failed to get cloud account"
	FailedToGetAppClientCloudAccount = "failed to get app client cloud account"
)

type CloudAccountAppClientService struct {
	pb.UnimplementedCloudAccountAppClientServiceServer
}

func (ms *CloudAccountAppClientService) GetAppClientCloudAccount(ctx context.Context,
	accountClient *pb.AccountClient) (*pb.CloudAccount, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountAppClientService.GetAppClientCloudAccount").WithValues("clientId", accountClient.ClientId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	appClientAcct, err := getAppClientCloudAccount(ctx, accountClient.ClientId)
	if err != nil {
		logger.Error(err, FailedToGetAppClientCloudAccount)
		return nil, errors.New(FailedToGetAppClientCloudAccount)
	}

	cloudAccount, err := GetCloudAccount(ctx, appClientAcct.cloudAccountId)
	if err != nil {
		logger.Error(err, FailedToGetAppClientCloudAccount)
		return nil, errors.New(FailedToGetAppClientCloudAccount)
	}

	// if not the owner, overwrite user_email, country_code in response object
	if appClientAcct.userEmail != cloudAccount.Name {
		cloudAccount.Name = appClientAcct.userEmail
		cloudAccount.CountryCode = appClientAcct.countryCode
	}
	return cloudAccount, nil
}
