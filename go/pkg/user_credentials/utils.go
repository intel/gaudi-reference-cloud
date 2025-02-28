// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package user_credentials

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cognitoutil"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/user_credentials/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var REGEX_PATTERN = regexp.MustCompile(`[!@#~$%^&*(),.?":{}|<>_]`)

type CloudAccountLocks struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func NewCloudAccountLocks(ctx context.Context) *CloudAccountLocks {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("NewCloudAccountLocks").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	return &CloudAccountLocks{
		locks: make(map[string]*sync.Mutex),
	}
}
func CreateAppClient(ctx context.Context, appClientName string, userEmail string, customScope string, cloudAcct *pb.CloudAccount, cognitoUtil *cognitoutil.COGNITOUtil, dbClient *sql.DB) (string, string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CreateAppClient").WithValues("cloudAccountId", cloudAcct.Id).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	clientId, clientSecret, err := cognitoUtil.CreateCredentials(ctx, appClientName, config.Cfg.GetAWSUserPool(), userEmail, customScope, cloudAcct.Id)
	logger.V(9).Info("create user credentials api invoked", "clienID", clientId, "clientSecret", clientSecret)
	if err != nil {
		logger.Error(err, "Error creating user pool client")
		return "", "", status.Errorf(codes.Internal, "Credentials cannot be created with the Client")
	}

	queryInsert := "INSERT INTO cloudaccount_user_credentials (cloudaccount_id, user_email, country_code, client_id, appclient_name) VALUES($1, $2, $3, $4, $5)"
	if _, err := dbClient.ExecContext(ctx, queryInsert, cloudAcct.Id, userEmail, cloudAcct.CountryCode, clientId, appClientName); err != nil {
		logger.Error(err, "error inserting user credentials into db", "query: ", queryInsert)
		return "", "", status.Errorf(codes.Internal, "Credentials cannot be created")
	}
	return clientId, clientSecret, nil
}

func DeleteAppClient(ctx context.Context, clientId string, cloudAccountId string, userEmail string, cognitoUtil *cognitoutil.COGNITOUtil, dbClient *sql.DB) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("DeleteAppClient").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("deleting appclient", "clientId", clientId)
	ok, err := cognitoUtil.DeleteCredentials(ctx, clientId, config.Cfg.GetAWSUserPool())
	if err != nil {
		logger.Error(err, "failed to delete credentials", "cloudAccountId", cloudAccountId, "ClientID", clientId)
		target := &types.ResourceNotFoundException{}
		if errors.As(err, &target) {
			logger.Error(err, "Resource not found", "cloudAccountId", cloudAccountId, "ClientID", clientId)
			return nil, status.Errorf(codes.NotFound, "ClientID not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete credentials")
	}
	if ok.Value {
		err := DeleteUserCredentials(ctx, userEmail, clientId, dbClient)
		if err != nil {
			logger.Error(err, "failed to delete cloud account", "ClientID", clientId)
			return nil, status.Errorf(codes.Internal, "cliend_id couldnot be deleted for the userEmail")
		}
	} else {
		return nil, status.Errorf(codes.Aborted, "unable to Delete Credentials")
	}
	return &emptypb.Empty{}, nil
}

func ValidateRequestField(field, fieldName string) (string, error) {
	if len(field) == 0 {
		msg := fmt.Sprintf("%s required", fieldName)
		err := status.Errorf(codes.InvalidArgument, msg)
		return msg, err
	}
	if REGEX_PATTERN.MatchString(field) {
		msg := fmt.Sprintf("%s Cannot contain special characters", fieldName)
		err := status.Errorf(codes.InvalidArgument, msg)
		return msg, err
	}
	return "", nil
}

func RemoveAppClients(ctx context.Context, cloudAccountId string, userEmail string, cognitoUtil *cognitoutil.COGNITOUtil, dbClient *sql.DB) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RemoveAppClients").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("remove appclients")
	query := "SELECT client_id FROM cloudaccount_user_credentials WHERE cloudaccount_id=$1 AND user_email=$2"
	rows, err := dbClient.QueryContext(ctx, query, cloudAccountId, userEmail)
	if err != nil {
		logger.Error(err, "failed to read cloud account user credentials ", "cloudAccountId", cloudAccountId, "context", "QueryContext")
		return nil, status.Errorf(codes.Internal, "clientId cannot be fetched")
	}

	defer rows.Close()
	resp := pb.UpdateUserCredentialsRequest{}
	for rows.Next() {
		if err := rows.Scan(&resp.ClientId); err != nil {
			return nil, err
		}

		_, err := DeleteAppClient(ctx, resp.ClientId, cloudAccountId, userEmail, cognitoUtil, dbClient)
		if err != nil {
			return nil, err
		}
	}
	return &emptypb.Empty{}, nil
}

func (cloudAccountLocks *CloudAccountLocks) Lock(ctx context.Context, cloudAccountID string) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("Lock").WithValues("cloudAccountID", cloudAccountID).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	cloudAccountLocks.mu.Lock()

	if _, exists := cloudAccountLocks.locks[cloudAccountID]; !exists {
		cloudAccountLocks.locks[cloudAccountID] = &sync.Mutex{}
	}
	cloudAccountLocks.mu.Unlock()
	cloudAccountLocks.locks[cloudAccountID].Lock()
}

func (cloudAccountLocks *CloudAccountLocks) Unlock(ctx context.Context, cloudAccountID string) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("Unlock").WithValues("cloudAccountID", cloudAccountID).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	cloudAccountLocks.mu.Lock()
	defer cloudAccountLocks.mu.Unlock()

	if lock, exists := cloudAccountLocks.locks[cloudAccountID]; exists {
		lock.Unlock()
	}
}

func DeleteUserCredentials(ctx context.Context, userEmail string, clientId string, dbClient *sql.DB) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("DeleteUserCredentials").WithValues("clientId", clientId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	select {
	case <-ctx.Done():
		logger.Error(ctx.Err(), "context canceled")
		return ctx.Err()
	default:
		tx, err := dbClient.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		queryDelete := `
				delete from cloudaccount_user_credentials where user_email = $1 AND client_id = $2		`
		_, err = dbClient.ExecContext(ctx, queryDelete, userEmail, clientId)
		if err != nil {
			logger.Error(err, "failed to delete user credentials for client id", "ClientID", clientId)
			return err
		}
		return tx.Commit()
	}

}

func ValidateEmailRequestField(field, fieldName string) (string, error) {
	if len(field) == 0 {
		msg := fmt.Sprintf("%s required", fieldName)
		err := status.Errorf(codes.InvalidArgument, msg)
		return msg, err
	}
	if GetEmailExclusionRegex().MatchString(field) {
		msg := fmt.Sprintf("%s Cannot contain special characters", fieldName)
		err := status.Errorf(codes.InvalidArgument, msg)
		return msg, err
	}
	return "", nil
}

func GetEmailExclusionRegex() *regexp.Regexp {
	pattern := fmt.Sprintf("[%s]", regexp.QuoteMeta(config.Cfg.GetEmailExclusionPattern()))
	re := regexp.MustCompile(pattern)
	return re
}
