// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OtpService struct {
	notificationClient               *NotificationClient
	pb.UnimplementedOtpServiceServer // Used for forward compatability
}

const (
	OtpExpired                      = "otp is expired"
	OtpNotFound                     = "otp not found"
	FailedToReadOtp                 = "failed to read otp"
	FailedToFetchOTPError           = "failed to fetch otp_code"
	FailedToStartDBTransactionError = "error starting db transaction"
	FailedToInsertDBData            = "error inserting db data"
	FailedToUpdateDB                = "error updating db data"
	FailedToCommitDBTransaction     = "error committing db transaction"
	FailedToDBQuery                 = "unable to query"
	InvalidRequestData              = "invalid request data"
	FailedToFetchData               = "failed to fetch data"
	FailedToFetchAttemptNumber      = "failed to fetch retry attempt number"
)

func (svc *OtpService) CreateOtp(ctx context.Context, req *pb.OtpRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OtpService.CreateOtp").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()
	logger.Info("create otp: enter")
	defer logger.Info("create otp: return")

	err := svc.validateExistingEmail(ctx, req)
	if err != nil {
		logger.Error(err, "error validating email request")
		return nil, err
	}

	err = checkMembersAddLimit(ctx, req.CloudAccountId)
	if err != nil {
		logger.Error(err, err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	valid, err := svc.validateOtpRequest(ctx, req)
	if !valid {
		logger.Error(err, err.Error())
		return nil, status.Errorf(codes.InvalidArgument, InvalidRequestData)
	}
	otpData, err := svc.otpCreateHandler(ctx, req)
	if err != nil {
		logger.Error(err, "error otp create handler")
		return nil, err
	}
	logger.Info("created otp", "cloudAccountId", otpData.CloudAccountId)
	err = svc.otpEmailHandler(ctx, otpData)
	if err != nil {
		logger.Error(err, "error in otp email handler")
	}
	return &emptypb.Empty{}, nil
}

func GenerateOTP() (string, error) {
	intId, err := rand.Int(rand.Reader, big.NewInt(999999))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", intId), nil
}

func (svc *OtpService) VerifyOtp(ctx context.Context, req *pb.VerifyOtpRequest) (*pb.VerifyOtpResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OtpService.VerifyOtp").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()
	valid, err := svc.validateOtpRequest(ctx, &pb.OtpRequest{CloudAccountId: req.CloudAccountId, MemberEmail: req.MemberEmail})
	otpRetryLimitIntervalDuration := int32(config.Cfg.OtpRetryLimitIntervalDuration)
	if !valid {
		return nil, err
	}
	resp := pb.VerifyOtpResponse{
		Validated: false,
		OtpState:  pb.OtpState_OTP_STATE_PENDING,
		Message:   "",
		Blocked:   false,
		RetryLeft: otpRetryLimitIntervalDuration,
	}
	logger.Info("verify otp: enter")
	defer logger.Info("verify otp: return")

	validate_attempt_reached, retry_left, err := failedOtpLogs.CheckThresholdReached(ctx, db, pb.OtpType_INVITATION_VALIDATE, req.CloudAccountId)
	resp.RetryLeft = int32(retry_left)
	if err != nil {
		logger.Error(err, "error in checking attempt number")
		return &resp, errors.New(FailedToFetchAttemptNumber)
	}
	logger.Info("ThresholdReachedData", "is_validate_attempt_reached", validate_attempt_reached)
	error_message := fmt.Sprintf("max verification limit reached, retry after %d minute", otpRetryLimitIntervalDuration)

	if validate_attempt_reached {
		err := failedOtpLogs.WriteAttempt(ctx, db, pb.OtpType_INVITATION_VALIDATE, req.CloudAccountId)
		if err != nil {
			logger.Error(err, "Failed to write invalid attempt on otp logs")
		}
		if retry_left > 0 {
			resp.RetryLeft = int32(retry_left) - 1
		}
		logger.Info(error_message)
		resp.Message = error_message
		resp.Blocked = true
		return &resp, nil
	}

	obj := pb.Otp{}
	readParams := protodb.NewSqlToProto(&obj)
	query := fmt.Sprintf("SELECT %v FROM admin_otp WHERE otp_state = 0 AND cloud_account_id = $1 AND member_email = $2", readParams.GetNamesString())
	rows, err := db.QueryContext(ctx, query, req.CloudAccountId, req.MemberEmail)
	if err != nil {
		logger.Error(err, FailedToFetchOTPError)
		return &resp, errors.New(FailedToFetchOTPError)
	}
	defer rows.Close()

	foundRows := false
	failedAttempt := false
	for rows.Next() {
		foundRows = true
		if err = readParams.Scan(rows); err != nil {
			logger.Error(err, "error encountered while reading entry")
		}

		if time.Now().After(obj.Expiry.AsTime()) {
			logger.Info("otp has expired")
			updateOtpState(ctx, int(pb.OtpState_OTP_STATE_EXPIRED), req.OtpCode, req.CloudAccountId, req.MemberEmail)
			resp.OtpState = pb.OtpState_OTP_STATE_EXPIRED
			continue
		}

		if obj.OtpCode == req.OtpCode {
			logger.Info("otp code matched")
			updateOtpState(ctx, int(pb.OtpState_OTP_STATE_ACCEPTED), req.OtpCode, req.CloudAccountId, req.MemberEmail)
			resp.Validated = true
			resp.OtpState = pb.OtpState_OTP_STATE_ACCEPTED
			return &resp, nil
		} else {
			failedAttempt = true
		}
	}
	if !foundRows || failedAttempt {
		err := failedOtpLogs.WriteAttempt(ctx, db, pb.OtpType_INVITATION_VALIDATE, req.CloudAccountId)
		if err != nil {
			logger.Error(err, "Failed to write invalid attempt on otp logs")
		}
		if retry_left > 0 {
			resp.RetryLeft = int32(retry_left) - 1
		}
		if resp.RetryLeft-1 < 0 {
			resp.Message = error_message
			resp.Blocked = true
			return &resp, nil
		}
	}

	return &resp, nil
}

func (svc *OtpService) ResendOtp(ctx context.Context, req *pb.OtpRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OtpService.ResendOtp").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()
	logger.Info("resend otp api invoked", "cloudAccountId", req.CloudAccountId)
	valid, err := svc.validateOtpRequest(ctx, req)
	if !valid {
		return nil, err
	}
	otpData, err := svc.otpGet(ctx, req)
	if err != nil {
		logger.Error(err, err.Error())
		if err.Error() == OtpExpired {
			logger.Info("OTP Expired, creating a new OTP")
			otpData, err = svc.otpCreateHandler(ctx, req)
			if err != nil {
				logger.Error(err, "error otp handler")
				return nil, status.Errorf(codes.Unavailable, "failed to create otp")
			}
			logger.Info("created otp", "cloudAccountId", otpData.CloudAccountId)
		} else {
			return nil, status.Errorf(codes.NotFound, "otp for cloud account %v not found", req.CloudAccountId)
		}
	}
	err = svc.otpEmailHandler(ctx, otpData)
	if err != nil {
		logger.Error(err, "error in otp email handler")
	}
	return &emptypb.Empty{}, nil
}

func (svc *OtpService) otpGet(ctx context.Context, req *pb.OtpRequest) (*pb.Otp, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OtpService.otpGet").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("get otp invoked for ", "cloudAccountId", req.CloudAccountId)
	logger.Info("getting otp")

	obj := pb.Otp{}
	readParams := protodb.NewSqlToProto(&obj)
	query := fmt.Sprintf("SELECT %v FROM admin_otp WHERE otp_state = 0 AND cloud_account_id = $1 AND member_email = $2", readParams.GetNamesString())

	rows, err := db.QueryContext(ctx, query, req.CloudAccountId, req.MemberEmail)
	if err != nil {
		logger.Error(err, FailedToFetchOTPError)
		return nil, errors.New(FailedToFetchOTPError)
	}
	defer rows.Close()

	for rows.Next() {
		if err = readParams.Scan(rows); err != nil {
			logger.Error(err, "error encountered while reading entry")
			return nil, errors.New(FailedToReadOtp)
		}
		logger.Info("otp for cloud account ", "obj", &obj)
		if time.Now().Before(obj.Expiry.AsTime()) {
			logger.Info("existing otp", "cloudAccountId", obj.CloudAccountId)
			return &obj, nil
		} else {
			updateOtpState(ctx, int(pb.OtpState_OTP_STATE_EXPIRED), obj.OtpCode, obj.CloudAccountId, obj.MemberEmail)
			return nil, errors.New(OtpExpired)
		}
	}
	return nil, errors.New(OtpNotFound)
}

func (svc *OtpService) otpCreateHandler(ctx context.Context, req *pb.
	OtpRequest) (*pb.Otp, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OtpService.otpCreateHandler").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()
	logger.Info("send email notification")
	otp, err := GenerateOTP()
	if err != nil {
		logger.Error(err, "error generating otp")
		return nil, err
	}
	exp := timestamppb.New(time.Now().Add(time.Minute * 10))

	otpData := &pb.Otp{
		CloudAccountId: req.CloudAccountId,
		MemberEmail:    req.MemberEmail,
		OtpCode:        otp,
		OtpState:       pb.OtpState_OTP_STATE_PENDING,
		Expiry:         exp,
	}
	params := protodb.NewProtoToSql(otpData)
	values := params.GetValues()
	values = append(values, pb.OtpState_OTP_STATE_PENDING)

	query := fmt.Sprintf("INSERT INTO admin_otp (%v,otp_state) VALUES(%v,$5)", params.GetNamesString(), params.GetParamsString())
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, FailedToDBQuery)
		return nil, errors.New(FailedToDBQuery)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, query, values...); err != nil {
		logger.Error(err, "error inserting record into admin_otp table", "query", query)
		return nil, errors.New(FailedToInsertDBData)
	}
	if err := tx.Commit(); err != nil {
		logger.Error(err, "error committing db transaction")
		return nil, errors.New(FailedToCommitDBTransaction)
	}

	return otpData, nil
}

func (svc *OtpService) otpEmailHandler(ctx context.Context, req *pb.Otp) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OtpService.otpEmailHandler").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()
	logger.Info("send email notification")
	emailRequest, err := svc.notificationClient.GetOTPEmailRequest(ctx, req)
	if err != nil {
		logger.Error(err, "error getting otp email request")
		return err
	}
	resp, err := svc.notificationClient.SendEmailNotification(ctx, emailRequest)
	if err != nil {
		logger.Error(err, "error sending email notification")
		return err
	}
	logger.Info("email notification sent", "resp", resp)
	return nil
}

func updateOtpState(ctx context.Context, otpState int, otpCode string, cloudAccountId string, memberEmail string) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OtpService.updateOtpState").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	query := "UPDATE admin_otp SET otp_state = $1, updated_at = NOW()::timestamp WHERE cloud_account_id = $2 AND member_email = $3 AND otp_code = $4"
	result, err := db.ExecContext(ctx, query, otpState, cloudAccountId, memberEmail, otpCode)
	if err != nil {
		logger.Error(err, "error encountered while updating otp table")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(err, "error encountered while updating otp table")
	}
	if rowsAffected < 1 {
		logger.Error(err, "entry to update otp table does not exist")
	}
}

func (svc *OtpService) validateOtpRequest(ctx context.Context, req *pb.OtpRequest) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OtpService.validateOtpRequest").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()
	logger.Info("validate otp request: enter")
	defer logger.Info("validate otp request: return")
	if req.CloudAccountId == "" {
		return false, status.Errorf(codes.InvalidArgument, "invalid input argument, cloudAccountId")
	}

	valid, err := validateEmail(ctx, req.MemberEmail)
	if !valid {
		logger.Error(err, "invalid email")
		return false, err
	}
	return true, nil
}

func (svc *OtpService) validateExistingEmail(ctx context.Context, req *pb.OtpRequest) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OtpService.validateExistingEmail").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()
	logger.Info("validate email request: enter")
	defer logger.Info("validate email request: return")

	var memberExists string
	query := "SELECT invitation_code FROM members WHERE LOWER(member_email) = LOWER($1) AND admin_account_id = $2;"
	if err := db.QueryRowContext(ctx, query, strings.ToLower(req.MemberEmail), req.CloudAccountId).Scan(&memberExists); err != nil {
		if err != sql.ErrNoRows {
			logger.Error(err, "Failed to validate email")
			return status.Errorf(codes.Internal, FailedToDBQuery)
		}
	}
	if strings.TrimSpace(memberExists) != "" {
		return status.Errorf(codes.InvalidArgument, "memberEmail already exists")
	}

	return nil
}
