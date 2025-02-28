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
	"net/mail"
	"regexp"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	maxRetries               = 10
	invCodeLen               = 1_000_000_00
	noteInputRegexValidation = `^[[:alpha:]\d\s.,;:@_-]+$` // UI XSS validation
)

type CloudAdminGrantAccessData struct {
	can_add_more_members bool
	members_limit        int32
	member_count         int32
}

type AppClientCloudAccounts struct {
	clientId       string
	cloudAccountId string
	userEmail      string
	countryCode    string
}

func validateAccountType(ctx context.Context, id string) (bool, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ValidateAccountType").WithValues("cloudAccountId", id).Start()
	defer span.End()
	logger.V(9).Info("validate account type invoked for ", "cloudAccountId", id)
	var accountType int
	query := "select type from cloud_accounts where id = $1"
	switch err := db.QueryRowContext(ctx, query, id).Scan(&accountType); err {
	case sql.ErrNoRows:
		return false, status.Errorf(codes.NotFound, "Cloud account %v not found", id)
	case nil:
		if pb.AccountType(accountType) != pb.AccountType_ACCOUNT_TYPE_INTEL && pb.AccountType(accountType) != pb.AccountType_ACCOUNT_TYPE_PREMIUM && pb.AccountType(accountType) != pb.AccountType_ACCOUNT_TYPE_ENTERPRISE {
			return false, status.Errorf(codes.PermissionDenied, "Operation is allowed only for Cloudaccount type Intel, Premium or Enterprise")
		} else {
			return true, nil
		}
	default:
		logger.Error(err, "error searching record in db")
		return false, status.Errorf(codes.Internal, "%v", InvalidRequestData)
	}
}

func validateEmail(ctx context.Context, email string) (bool, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ValidateEmailAddress").Start()
	defer span.End()
	logger.V(9).Info("validate email address invoked for given input email")
	_, err := mail.ParseAddress(email)
	if err != nil {
		logger.Error(err, "Not a valid email address")
		return false, status.Errorf(codes.InvalidArgument, err.Error())
	}
	return true, nil
}

func validateInputRegex(ctx context.Context, input string, regex string, inputName string) (bool, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ValidateInputwithRegex").WithValues("input", input, "regex", regex, "inputName", inputName).Start()
	defer span.End()
	match := regexp.MustCompile(regex).MatchString(input)
	if !match {
		logger.Error(errors.New("input not valid"), "input doesn't meet validation requirements")
		return false, status.Errorf(codes.InvalidArgument, "input %v validation doesn't meet the requirements", inputName)
	}
	return true, nil
}

func checkAdminOtpStateVerified(ctx context.Context, dbsession *sql.DB, cloudAccountId string, memberEmail string) (bool, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("checkAdminOtpStateVerified").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("Start")
	defer logger.Info("End")
	var adminOtpVerified int
	query := "SELECT count(*) FROM admin_otp WHERE otp_state = $1 AND cloud_account_id = $2 AND member_email = $3;"
	if err := dbsession.QueryRowContext(ctx, query, pb.OtpState_OTP_STATE_ACCEPTED, cloudAccountId, memberEmail).Scan(&adminOtpVerified); err != nil {
		return false, status.Errorf(codes.Internal, FailedToDBQuery)
	}
	if adminOtpVerified == 0 {
		return false, status.Errorf(codes.InvalidArgument, "Admin otp not verified")
	}
	return true, nil
}

func invalidateAdminOtpState(ctx context.Context, dbsession *sql.DB, cloudAccountId string, memberEmail string) error {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("invalidateAdminOtpState").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("Start")
	defer logger.Info("End")
	query := "UPDATE admin_otp SET otp_state = $1 WHERE cloud_account_id = $2 AND member_email = $3 AND otp_state = $4;"
	if _, err := dbsession.ExecContext(ctx, query, pb.OtpState_OTP_STATE_INVALID, cloudAccountId, memberEmail, pb.OtpState_OTP_STATE_ACCEPTED); err != nil {
		return status.Errorf(codes.Internal, FailedToUpdateDB)
	}
	return nil
}

func ValidateRequest(ctx context.Context, dbsession *sql.DB, inv *pb.InviteRequest, cloudAccountId string) error {
	logger := log.FromContext(ctx).WithName("ValidateRequest")
	logger.Info("validate invite request: enter")
	defer logger.Info("validate invite request: return")

	var memberExists string
	valid, err := validateEmail(ctx, inv.MemberEmail)
	if !valid {
		return err
	}

	valid, err = validateInputRegex(ctx, inv.Note, noteInputRegexValidation, "note")
	if !valid {
		return err
	}

	valid, err = checkAdminOtpStateVerified(ctx, dbsession, cloudAccountId, inv.MemberEmail)
	if !valid {
		return err
	}

	query := "select invitation_code from members where member_email = $1 AND admin_account_id = $2;"
	if err := dbsession.QueryRowContext(ctx, query, inv.MemberEmail, cloudAccountId).Scan(&memberExists); err != nil {
		if err != sql.ErrNoRows {
			logger.Error(err, FailedToDBQuery)
			return status.Errorf(codes.Internal, FailedToDBQuery)
		}
	}
	if memberExists != "" {
		return status.Errorf(codes.InvalidArgument, "memberEmail already exists")
	}

	if inv.Expiry != nil {
		if inv.Expiry.AsTime().Before(timestamppb.Now().AsTime()) {
			return status.Errorf(codes.InvalidArgument, "invalid input argument, request is expired")
		}
	}

	logger.Info("Valid invite request for", "memberEmail", inv.MemberEmail)
	return nil
}

func GetInvitationCode() (string, error) {
	id, err := newCode()
	if err != nil {
		return "", fmt.Errorf("unable to get invite code: %w", err)
	}
	return id, nil
}

func newCode() (string, error) {
	intId, err := rand.Int(rand.Reader, big.NewInt(invCodeLen))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08d", intId), nil
}

func GetCloudAccount(ctx context.Context, cloudAccountId string) (*pb.CloudAccount, error) {
	logger := log.FromContext(ctx).WithName("GetCloudAccount")
	logger.Info("Get CloudAccount invoked for ", "cloudAccountId", cloudAccountId)
	obj := pb.CloudAccount{}
	readParams := protodb.NewSqlToProto(&obj, fieldOpts...)

	query := fmt.Sprintf("SELECT %v from cloud_accounts WHERE id = $1",
		readParams.GetNamesString())

	rows, err := db.QueryContext(ctx, query, cloudAccountId)
	if err != nil {
		logger.Error(err, FailedToDBQuery)
		return nil, status.Errorf(codes.Internal, FailedToDBQuery)
	}

	defer rows.Close()
	if !rows.Next() {
		return nil, status.Errorf(codes.NotFound, "cloud account %v not found", cloudAccountId)
	}
	if err = readParams.Scan(rows); err != nil {
		logger.Error(err, "Error observed while fetching Cloud Account", "id", cloudAccountId)
		return nil, status.Errorf(codes.Internal, InvalidRequestData)
	}
	return &obj, nil
}

func convertTimestampToTime(timestamp *timestamppb.Timestamp) *time.Time {
	if timestamp != nil {
		timeValue := timestamp.AsTime()
		return &timeValue
	}
	return nil
}

func GetInvitationEmailRequest(code string, memberEmail string, accountOwner string, invitationNotes string) *pb.EmailRequest {
	templateData := map[string]string{
		"invitationLink":  config.Cfg.GetInviteLink(),
		"invitationCode":  code,
		"accountOwner":    accountOwner,
		"invitationNotes": invitationNotes,
	}
	emailRequest := &pb.EmailRequest{
		MessageType:  "InviteNotification",
		ServiceName:  "CloudAccountInvitationService",
		Recipient:    memberEmail,
		Sender:       config.Cfg.GetSenderEmail(),
		TemplateName: config.Cfg.GetInviteTemplate(),
		TemplateData: templateData,
	}
	return emailRequest
}

func GetInvitationAcceptEmailRequest(code string, memberEmail string, accountOwner string) *pb.EmailRequest {
	templateData := map[string]string{
		"invitationCode": code,
		"accountOwner":   accountOwner,
	}
	emailRequest := &pb.EmailRequest{
		MessageType:  "InviteAcceptNotification",
		ServiceName:  "CloudAccountInvitationService",
		Recipient:    memberEmail,
		Sender:       config.Cfg.GetSenderEmail(),
		TemplateName: config.Cfg.GetInviteAcceptTemplate(),
		TemplateData: templateData,
	}
	return emailRequest
}

func getMembersAddLimitData(ctx context.Context, cloudAccountId string) (*CloudAdminGrantAccessData, error) {
	logger := log.FromContext(ctx).WithName("GetMembersAddLimitData")
	var limit int32
	var count int
	var accountType int

	accountTypeQuery := "select type from cloud_accounts where id = $1"
	err := db.QueryRowContext(ctx, accountTypeQuery, cloudAccountId).Scan(&accountType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "Cloud account %v not found", cloudAccountId)
		}
		logger.Error(err, "failed to fetch type of cloud account", "id", cloudAccountId)
		return nil, status.Errorf(codes.Unknown, FailedToFetchData)
	}

	switch pb.AccountType(accountType) {
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE:
		limit = config.Cfg.EnterpriseMemberInvitationLimit
	case pb.AccountType_ACCOUNT_TYPE_PREMIUM:
		limit = config.Cfg.PremiumMemberInvitationLimit
	case pb.AccountType_ACCOUNT_TYPE_INTEL:
		limit = config.Cfg.IntelMemberInvitationLimit
	default:
		return nil, errors.New("only intel, premium and enterprise accounts are supported")
	}

	states := fmt.Sprintf("%v, %v", pb.InvitationState_INVITE_STATE_ACCEPTED.Number(),
		pb.InvitationState_INVITE_STATE_PENDING_ACCEPT.Number())
	memberCountQuery := fmt.Sprintf("select count(id) from members where admin_account_id='%s' and invitation_state in (%s)",
		cloudAccountId, states)

	err = db.QueryRow(memberCountQuery).Scan(&count)
	if err != nil {
		logger.Error(err, "failed to fetch invitation count from cloud account", "id", cloudAccountId)
		return nil, status.Errorf(codes.Unknown, FailedToFetchData)
	}
	can_add_more_members := count < int(limit)
	return &CloudAdminGrantAccessData{can_add_more_members: can_add_more_members,
		members_limit: limit, member_count: int32(count)}, nil
}

func checkMembersAddLimit(ctx context.Context, cloudAccountId string) error {
	memberLimitData, err := getMembersAddLimitData(ctx, cloudAccountId)
	if err != nil {
		return err
	}
	if memberLimitData.can_add_more_members {
		return nil
	} else {
		return errors.New("member max add limit reached")
	}
}

func getAppClientCloudAccount(ctx context.Context, clientId string) (*AppClientCloudAccounts, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("getAppClientCloudAccount").WithValues("clientId", clientId).Start()
	defer span.End()
	logger.Info("Start")
	defer logger.Info("End")
	var rows *sql.Rows
	var err error
	var revoked sql.NullString
	var enabled sql.NullString
	appClientAcct := &AppClientCloudAccounts{}
	query := "SELECT user_email, country_code, cloudaccount_id, revoked, enabled FROM cloudaccount_user_credentials WHERE client_id=$1"
	rows, err = db.QueryContext(ctx, query, clientId)
	if err != nil {
		logger.Error(err, "failed to read cloud account user credentials ", "clientId", clientId, "context", "QueryContext")
		return appClientAcct, fmt.Errorf("cloudaccount doesn't exists: %w", err)
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("client cloud account %v not found", clientId)
	}
	if err := rows.Scan(&appClientAcct.userEmail, &appClientAcct.countryCode, &appClientAcct.cloudAccountId, &revoked, &enabled); err != nil {
		return appClientAcct, err
	}
	revokedAppClient := ""
	if revoked.Valid {
		revokedAppClient = revoked.String
	}
	enabledAppClient := ""
	if revoked.Valid {
		enabledAppClient = enabled.String
	}
	logger.Info("user credentials", "clienId: ", clientId, "cloudAccountId: ", appClientAcct.cloudAccountId, "revoked: ", revoked, "enabled: ", enabled)
	if revokedAppClient == "false" && enabledAppClient == "true" {
		appClientAcct.clientId = clientId
	}
	return appClientAcct, err
}
