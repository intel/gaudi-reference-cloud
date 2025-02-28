// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package user_credentials

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cognitoutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserCredentialsService struct {
	pb.UnimplementedUserCredentialsServiceServer
	cognitoUtil        *cognitoutil.COGNITOUtil
	customScope        string
	cloudAccountClient pb.CloudAccountServiceClient
	dbClient           *sql.DB
}

var (
	cloudAccountLocks *CloudAccountLocks
)

func NewUserCredentialsService(cognitoUtil *cognitoutil.COGNITOUtil, db *sql.DB, tokenScope string, cloudAccountClient pb.CloudAccountServiceClient) (*UserCredentialsService, error) {
	if cloudAccountClient == nil {
		return nil, fmt.Errorf("cloudaccount client is required")
	}
	return &UserCredentialsService{cognitoUtil: cognitoUtil, dbClient: db, customScope: tokenScope, cloudAccountClient: cloudAccountClient}, nil
}

func (srv *UserCredentialsService) CreateUserCredentials(ctx context.Context, request *pb.CreateUserCredentialsRequest) (*pb.ClientCredentials, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UserCredentialsService.CreateUser").WithValues("cloudAccountId", request.CloudaccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	msg, err := ValidateRequestField(request.AppClientName, "App Client Name")
	if err != nil {
		logger.V(9).Info("invalid argument", "msg", msg)
		return nil, err
	}
	userEmail, err := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.EmailClaim)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	cloudAcct, err := srv.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: request.CloudaccountId})
	logger.Info("create user credentials api invoked", "appClientName", request.AppClientName)
	if err != nil {
		logger.V(9).Error(err, "Error invoking cloudAccount to get ID")
		return nil, status.Errorf(codes.Internal, "Cannot fetch details of cloudAccount ID")
	}
	cloudAccountLocks.Lock(ctx, request.CloudaccountId)
	defer cloudAccountLocks.Unlock(ctx, request.CloudaccountId)
	//Check if the user has more than 2 credentials and unique app clientname
	query := "SELECT appclient_name FROM cloudaccount_user_credentials WHERE cloudaccount_id=$1 AND user_email=$2"
	rows, err := srv.dbClient.QueryContext(ctx, query, request.CloudaccountId, userEmail)
	if err != nil {
		logger.Error(err, "failed to read cloud account user credentials ", "cloudaccount_id", request.CloudaccountId, "context", "QueryContext")
		return nil, status.Errorf(codes.Internal, "application client cannot be created")
	}

	defer rows.Close()
	count := 0
	resp := pb.UpdateUserCredentialsRequest{}
	for rows.Next() {
		count++
		if err := rows.Scan(&resp.AppClientName); err != nil {
			return nil, err
		}
		// To keep appclientName Unique
		if resp.AppClientName == request.AppClientName {
			logger.Error(err, "AppClient Name already Exists")
			return nil, status.Errorf(codes.AlreadyExists, "AppClient Name already Exists, Please use different Name %v", request.AppClientName)
		}
	}
	if count >= 2 {
		logger.V(9).Error(err, "User Email cannot have more than 2 credentials")
		return nil, status.Errorf(codes.OutOfRange, "user email cannot have more than 2 credentials %v", userEmail)
	}

	// Create new app client in Cognito
	// Save new user credentials clientID in database
	clientId, clientSecret, err := CreateAppClient(ctx, request.AppClientName, userEmail, srv.customScope, cloudAcct, srv.cognitoUtil, srv.dbClient)
	if err != nil {
		logger.V(9).Error(err, "error creating app client")
		return nil, status.Errorf(codes.Internal, "error creating app client")
	}

	return &pb.ClientCredentials{
		CloudaccountId: request.CloudaccountId,
		ClientId:       clientId,
		ClientSecret:   clientSecret,
	}, nil
}

func (srv *UserCredentialsService) GetUserCredentials(ctx context.Context, request *pb.GetUserCredentialRequest) (*pb.GetUserCredentialResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UserCredentialsService.GetUserCredentials").WithValues("cloudAccountId", request.CloudaccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("get user credentials api invoked", "CloudaccountId", request.CloudaccountId)
	userEmail, err := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.EmailClaim)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	obj := pb.GetUserCredentialResponse{}
	err = func() error {
		var rows *sql.Rows
		var err error
		query := "SELECT client_id, appclient_name, revoked, enabled, created FROM cloudaccount_user_credentials WHERE cloudaccount_id=$1 AND user_email=$2"
		rows, err = srv.dbClient.QueryContext(ctx, query, request.CloudaccountId, userEmail)
		if err != nil {
			logger.Error(err, "failed to read cloud account user credentials ", "cloudaccount_id", request.CloudaccountId, "context", "QueryContext")
			return status.Errorf(codes.Internal, "CloudaccountId or client_id doesn't exists")
		}

		defer rows.Close()

		for rows.Next() {
			resp := pb.UpdateUserCredentialsRequest{}
			getCredential := pb.GetUserCredential{}
			var created time.Time
			if err := rows.Scan(&resp.ClientId, &resp.AppClientName, &resp.Revoked, &resp.Enabled, &created); err != nil {
				return err
			}
			logger.Info("get user credentials response", "clienId", resp.ClientId, "appClientName: ", resp.AppClientName, "revoked :", resp.Revoked, "enabled: ", resp.Enabled)

			getCredential.AppClientName = resp.AppClientName
			getCredential.ClientId = resp.ClientId
			getCredential.Revoked = resp.Revoked
			getCredential.Created = timestamppb.New(created)
			//Check if the client_id is not revoked or disabled
			if resp.Revoked == "false" && resp.Enabled == "true" {
				obj.AppClients = append(obj.AppClients, &getCredential)
			}

		}
		return err
	}()
	return &obj, err
}

func (srv *UserCredentialsService) RemoveUserCredentials(ctx context.Context, request *pb.DeleteUserCredentialsRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UserCredentialsService.RemoveUserCredentials").WithValues("cloudAccountId", request.CloudaccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("Delete request for :", "cloudaccountid", "clientID", request.CloudaccountId, request.ClientId)
	msg, err := ValidateRequestField(request.ClientId, "Client id")
	if err != nil {
		logger.V(9).Info("invalid argument", "msg", msg)
		return nil, err
	}
	userEmail, extractErr := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.EmailClaim)
	if extractErr != nil {
		return nil, status.Error(codes.InvalidArgument, extractErr.Error())
	}
	logger.Info("userEmail from Jwt", "userEmail", userEmail)

	query := "SELECT COUNT(*) FROM cloudaccount_user_credentials WHERE cloudaccount_id=$1 AND user_email=$2 and client_id=$3"
	var rowCount int

	err = srv.dbClient.QueryRow(query, request.CloudaccountId, userEmail, request.ClientId).Scan(&rowCount)

	if rowCount == 0 {
		return nil, status.Error(codes.NotFound, "ClientID not Found")
	}
	if err != nil {
		logger.Error(err, "failed to delete cloud account", "ClientID", request.ClientId)
		return nil, status.Errorf(codes.Internal, "cliend_id couldnot be deleted")
	}

	// Delete the app client Credentials in Cognito
	// if the credentials are deleted from cognito delete from database
	// if  credentials are not deleted from cognito throw error

	_, err = DeleteAppClient(ctx, request.ClientId, request.CloudaccountId, userEmail, srv.cognitoUtil, srv.dbClient)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, err
}

func (srv *UserCredentialsService) RevokeUserCredentials(ctx context.Context, request *pb.RevokeUserCredentialsRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UserCredentialsService.RevokeUserCredentials").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	userEmail, extractErr := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.EmailClaim)
	if extractErr != nil {
		return nil, status.Error(codes.InvalidArgument, extractErr.Error())
	}
	// Delete the app client Credentials in Cognito
	// if the credentials are deleted from cognito delete from database
	// if  credentials are not deleted from cognito throw error
	_, err := RemoveAppClients(ctx, request.CloudaccountId, userEmail, srv.cognitoUtil, srv.dbClient)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (srv *UserCredentialsService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	_, log, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UserCredentialsService.Ping").Start()
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (srv *UserCredentialsService) RemoveMemberUserCredentials(ctx context.Context, request *pb.RemoveMemberUserCredentialsRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UserCredentialsService.RemoveMemberUserCredentials").WithValues("cloudAccountId", request.CloudaccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("remove member user credential")
	msg, err := ValidateRequestField(request.CloudaccountId, "cloud account id")
	if err != nil {
		logger.V(9).Info("invalid argument", "msg", msg)
		return nil, err
	}
	msg, err = ValidateEmailRequestField(request.MemberEmail, "member email")
	if err != nil {
		logger.V(9).Info("invalid argument", "msg", msg)
		return nil, err
	}
	_, err = RemoveAppClients(ctx, request.CloudaccountId, request.MemberEmail, srv.cognitoUtil, srv.dbClient)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (srv *UserCredentialsService) GetCloudAccountById(ctx context.Context, cloudAccountId string) (*pb.CloudAccount, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UserCredentialsMemberService.GetMemberUserCredentials").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	cloudAcct, err := srv.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil {
		logger.Error(err, "error invoking cloudAccount api to get ID")
		return nil, status.Errorf(codes.Internal, "failed to get cloud account")
	}
	return cloudAcct, err
}
