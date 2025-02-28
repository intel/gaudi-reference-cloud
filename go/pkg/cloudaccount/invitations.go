// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	authz "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InvitationService struct {
	pb.UnimplementedCloudAccountInvitationServiceServer
	session               *sql.DB
	notificationClient    *NotificationClient
	cfg                   *config.Config
	authzClient           *authz.AuthzClient
	userCredentialsClient *UserCredentialsClient
}

type Invite struct {
	adminAccountId      string
	memberEmail         string
	invitationCode      string
	expiry              *time.Time
	notes               string
	cloudAccountRoleIds []string
	invitationState     pb.InvitationState
}

func (svc *InvitationService) tryCreate(ctx context.Context, dbsession *sql.DB, req *pb.InviteRequestList) error {
	logger := log.FromContext(ctx).WithName("InvitationService.tryCreate")
	logger.Info("create invite query: enter")
	defer logger.Info("create invite query: return")

	inviteList := []*Invite{}
	logger.Info("Request", "invites", req.Invites)

	for _, inv := range req.Invites {
		logger.Info("Processing invite", "invite", inv)

		err := ValidateRequest(ctx, dbsession, inv, req.CloudAccountId)
		if err != nil {
			return err
		}
		invCode, err := GetInvitationCode()
		if err != nil {
			logger.Error(err, "unable to get invite code")
			return err
		}

		inviteList = append(inviteList, &Invite{
			adminAccountId:      req.CloudAccountId,
			memberEmail:         inv.MemberEmail,
			invitationCode:      invCode,
			invitationState:     pb.InvitationState_INVITE_STATE_PENDING_ACCEPT,
			expiry:              convertTimestampToTime(inv.Expiry),
			notes:               inv.Note,
			cloudAccountRoleIds: inv.CloudAccountRoleIds,
		})
	}
	args, argsStr := protodb.AddArrayArgValues(
		[]any{},
		inviteList,
		func(inv *Invite) []any {
			return []any{inv.adminAccountId, inv.memberEmail, inv.invitationCode, inv.invitationState, inv.expiry, inv.notes, inv.cloudAccountRoleIds}
		},
	)
	query := "INSERT INTO members (admin_account_id,member_email,invitation_code,invitation_state,expiry,notes,cloud_account_role_ids) VALUES" + argsStr
	_, err := dbsession.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Error(err, FailedToInsertDBData)
		return errors.New(FailedToInsertDBData)
	}
	for _, invite := range inviteList {
		if err = svc.SendInvitationEmail(ctx, invite.invitationCode, invite.memberEmail, invite.adminAccountId, invite.notes); err != nil {
			logger.Error(err, "couldn't send email", "invite", invite.adminAccountId)
		}
	}
	if err != nil {
		return status.Errorf(codes.Internal, "unable to send all invite")
	}
	return nil
}

func (svc *InvitationService) CreateInvite(ctx context.Context, req *pb.InviteRequestList) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InvitationService.CreateInvite").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	if req.Invites == nil || len(req.Invites) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, invites")
	}
	if req.CloudAccountId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, cloudAccountId")
	}
	err := checkMembersAddLimit(ctx, req.CloudAccountId)
	if err != nil {
		logger.Error(err, err.Error())
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	for ii := 0; ii < maxRetries; ii++ {
		err := svc.tryCreate(ctx, svc.session, req)
		if err == nil {
			return &emptypb.Empty{}, nil
		}
		pgErr := &pgconn.PgError{}
		if !errors.As(err, &pgErr) ||
			pgErr.Code != kErrUniqueViolation {
			return nil, err
		}
		logger.Info("retrying invitation code generation after id collision", "iter", ii)
	}

	return &emptypb.Empty{}, nil
}

func (svc *InvitationService) ReadInvite(ctx context.Context, inviteFilter *pb.InviteFilter) (*pb.InviteList, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InvitationService.ReadInvite").WithValues("cloudAccountId", inviteFilter.AdminAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	invListObj := pb.InviteList{}

	readQueryBase := "SELECT member_email,invitation_state,expiry,notes FROM members"
	params := protodb.NewProtoToSql(inviteFilter)
	filter := params.GetFilter()
	if filter != "" {
		readQueryBase += " " + filter
	}

	rows, err := svc.session.QueryContext(ctx, readQueryBase, params.GetValues()...)
	if err != nil {
		logger.Error(err, FailedToFetchData)
		return nil, errors.New(FailedToFetchData)
	}
	defer rows.Close()
	for rows.Next() {
		invObj := pb.Invite{}
		var expTime *time.Time
		if err := rows.Scan(&invObj.MemberEmail, &invObj.InvitationState, &expTime, &invObj.Note); err != nil {
			logger.Error(err, "error reading result row.")
			return nil, err
		}
		if expTime != nil {
			invObj.Expiry = timestamppb.New(*expTime)
		}
		invListObj.Invites = append(invListObj.Invites, &invObj)
	}
	memberLimitData, err := getMembersAddLimitData(ctx, inviteFilter.AdminAccountId)
	if err != nil {
		return nil, err
	}
	invListObj.CanAddMoreMembers = memberLimitData.can_add_more_members
	invListObj.MembersLimit = memberLimitData.members_limit
	invListObj.MemberCount = memberLimitData.member_count
	return &invListObj, nil
}

func (svc *InvitationService) fetchInvitationState(ctx context.Context, dbsession *sql.DB, req *pb.InviteResendRequest) (*pb.InvitationState, error) {
	logger := log.FromContext(ctx).WithName("InvitationService.fetchInvitationState")
	logger.Info("BEGIN")
	defer logger.Info("END")
	query := "SELECT invitation_state FROM members WHERE admin_account_id=$1 AND member_email=$2"
	rows, err := dbsession.QueryContext(ctx,
		query,
		req.AdminAccountId, req.MemberEmail)
	if err != nil {
		logger.Error(err, FailedToDBQuery)
		return nil, errors.New(FailedToDBQuery)
	}
	defer rows.Close()

	var invitationState pb.InvitationState
	for rows.Next() {
		if err = rows.Scan(&invitationState); err != nil {
			logger.Error(err, "error while fetching invitationState of member", "cloudAccountId", req.AdminAccountId)
			return nil, status.Errorf(codes.Unknown, "failed to fetch invitation state")
		}
	}

	return &invitationState, err
}

func (svc *InvitationService) ResendInvite(ctx context.Context, req *pb.InviteResendRequest) (*pb.InviteResendResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudAccountInvitationService.ResendInvite").WithValues("cloudAccountId", req.AdminAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	otpRetryLimitIntervalDuration := int32(config.Cfg.OtpRetryLimitIntervalDuration)

	resp := pb.InviteResendResponse{
		Message:   "",
		Blocked:   false,
		RetryLeft: otpRetryLimitIntervalDuration,
	}

	resend_attempt_reached, retry_left, err := failedOtpLogs.CheckThresholdReached(ctx, db, pb.OtpType_INVITATION_RESEND, req.GetAdminAccountId())
	resp.RetryLeft = int32(retry_left)
	if err != nil {
		logger.Error(err, "error in checking attempt number")
		return &resp, errors.New(FailedToFetchAttemptNumber)
	}
	logger.Info("ThresholdReachedData", "is_resend_attempt_reached", resend_attempt_reached)
	error_message := fmt.Sprintf("max resend limit reached, retry after %d minute", otpRetryLimitIntervalDuration)

	if resend_attempt_reached {
		err := failedOtpLogs.WriteAttempt(ctx, db, pb.OtpType_INVITATION_RESEND, req.GetAdminAccountId())
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

	invitation_state, err := svc.fetchInvitationState(ctx, svc.session, req)
	if err != nil {
		logger.Error(err, FailedToFetchData)
		return nil, status.Errorf(codes.Unknown, "failed to fetch invitation state")
	}

	if *invitation_state.Enum() == *pb.InvitationState_INVITE_STATE_ACCEPTED.Enum() {
		logger.Error(err, "resend invite failed, invite already exists and accepted ")
		return nil, status.Errorf(codes.AlreadyExists, "resend invite failed, invite already exists and accepted")
	}

	if req.AdminAccountId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, adminAccountId")
	}
	valid, err := validateAccountType(ctx, req.AdminAccountId)
	if !valid {
		return nil, err
	}
	valid, err = validateEmail(ctx, req.MemberEmail)
	if !valid {
		return nil, err
	}

	query := "SELECT invitation_code, notes FROM members"
	params := protodb.NewProtoToSql(req)
	filter := params.GetFilter()
	if filter != "" {
		query += " " + filter
	}

	var code string
	var notes string
	if err := svc.session.QueryRowContext(ctx, query, params.GetValues()...).Scan(&code, &notes); err != nil {
		if err == sql.ErrNoRows {
			logger.Error(err, "no matching invite found")
			return nil, status.Errorf(codes.NotFound, "no matching invite found")
		}
		logger.Error(err, "unable to query")
		return nil, status.Errorf(codes.Internal, FailedToDBQuery)
	}

	if err := svc.SendInvitationEmail(ctx, code, req.MemberEmail, req.AdminAccountId, notes); err != nil {
		logger.Error(err, "couldn't send email")
		return nil, status.Errorf(codes.Internal, "unable to resend invite")
	}

	err = failedOtpLogs.WriteAttempt(ctx, db, pb.OtpType_INVITATION_RESEND, req.GetAdminAccountId())
	if err != nil {
		logger.Error(err, "Failed to write attempt on otp logs")
	}
	if retry_left > 0 {
		resp.RetryLeft = int32(retry_left) - 1
	}
	if resp.RetryLeft-1 < 0 {
		resp.Message = error_message
		resp.Blocked = true
		return &resp, nil
	}
	updateQuery := "UPDATE members SET invitation_state= $1 WHERE member_email = $2"
	result, err := db.ExecContext(ctx, updateQuery, pb.InvitationState_INVITE_STATE_PENDING_ACCEPT, req.MemberEmail)
	if err != nil {
		logger.Error(err, "error encountered while updating db")
		return nil, status.Errorf(codes.Internal, FailedToUpdateDB)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if rowsAffected < 1 {
		return nil, status.Errorf(codes.NotFound, "resend invite from (%v) to (%v) failed", req.AdminAccountId, req.MemberEmail)
	}

	return &resp, nil
}

func (svc *InvitationService) RevokeInvite(ctx context.Context, req *pb.InviteRevokeRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InvitationService.RevokeInvite").WithValues("cloudAccountId", req.AdminAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	if req.AdminAccountId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, missing AdminAccountId")
	}

	if req.MemberEmail == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, missing MemberEmail")
	}

	if req.InvitationState == pb.InvitationState_INVITE_STATE_UNSPECIFIED {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, unkown InvitationState")
	}

	if req.InvitationState != pb.InvitationState_INVITE_STATE_PENDING_ACCEPT {
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "revoke not supported for invitation_state (%v)", req.InvitationState)
	}

	query := "DELETE FROM members"
	params := protodb.NewProtoToSql(req)
	filter := params.GetFilter()
	if filter != "" {
		query += " " + filter
	}

	result, err := svc.session.ExecContext(ctx, query, params.GetValues()...)
	if err != nil {
		logger.Error(err, FailedToDeleteData)
		return nil, errors.New(FailedToDeleteData)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if rowsAffected < 1 {
		return nil, status.Errorf(codes.NotFound, "the invite from (%v) to (%v) does not exist", req.AdminAccountId, req.MemberEmail)
	}

	err = invalidateAdminOtpState(ctx, svc.session, req.AdminAccountId, req.MemberEmail)
	if err != nil {
		logger.Error(err, "error encountered while updating db")
		return nil, status.Errorf(codes.Internal, FailedToUpdateDB)
	}

	return &emptypb.Empty{}, nil
}

func (svc *InvitationService) RemoveInvite(ctx context.Context, req *pb.InviteUpdateRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InvitationService.RemoveInvite").WithValues("cloudAccountId", req.AdminAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("removing invite req", "req", req)
	if req.AdminAccountId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, missing AdminAccountId")
	}
	if req.MemberEmail == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, missing MemberEmail")
	}
	if req.InvitationState != pb.InvitationState_INVITE_STATE_ACCEPTED {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, invite state")
	}

	// UnassignSystemRole for member on authz
	if svc.cfg.Authz.Enabled {
		if _, err := svc.authzClient.UnassignSystemRole(ctx, &pb.RoleRequest{CloudAccountId: req.AdminAccountId, Subject: req.MemberEmail, SystemRole: pb.SystemRole_cloud_account_member.String()}); err != nil {
			logger.Error(err, "couldn't unassign system role")
			return nil, status.Errorf(codes.Internal, "unable to remove invite, could not unassign system role")
		}
		// verify systemrole was removed from authz
		exist, err := svc.authzClient.SystemRoleExists(ctx, &pb.RoleRequest{CloudAccountId: req.AdminAccountId, Subject: req.MemberEmail, SystemRole: pb.SystemRole_cloud_account_member.String()})
		if err != nil {
			logger.Error(err, "couldn't verify if system role exist")
			return nil, status.Errorf(codes.Internal, "couldn't verify if system role exist")
		}
		if exist {
			logger.Error(errors.New("system role not removed"), "couldn't remove systemrole still exist")
			return nil, status.Errorf(codes.Internal, "unable to remove invite, could not unassign system role")
		}
		// remove user from all related cloudAccountRoles
		if _, err := svc.authzClient.RemoveUserFromCloudAccountRole(ctx, &pb.CloudAccountRoleUserRequest{CloudAccountId: req.AdminAccountId, UserId: req.MemberEmail}); err != nil {
			logger.Error(err, "couldn't remove user cloud account roles")
			return nil, status.Errorf(codes.Internal, "unable to remove invite")
		}
	}
	logger.Info("removing credential", "admin cloud account id", req.AdminAccountId)
	removeUserCredentialsRequest := &pb.RemoveMemberUserCredentialsRequest{CloudaccountId: req.AdminAccountId, Revoked: "true", MemberEmail: req.MemberEmail}
	logger.Info("removing user credential", "removeUserCredentialsRequest", removeUserCredentialsRequest)
	if _, err := svc.userCredentialsClient.RemoveMemberUserCredentials(ctx, removeUserCredentialsRequest); err != nil {
		logger.Error(err, "couldn't remove user credential")
		return nil, status.Errorf(codes.Internal, "unable to revoke user credential")
	}

	logger.Info("removing invite", "cloudAccountId", req.AdminAccountId)
	query := "DELETE FROM members " +
		"WHERE admin_account_id=$1 AND member_email=$2 AND invitation_state=$3"
	_, err := svc.session.ExecContext(ctx, query, req.AdminAccountId, req.MemberEmail, pb.InvitationState_INVITE_STATE_ACCEPTED)
	if err != nil {
		logger.Error(err, FailedToDeleteData)
		return nil, errors.New(FailedToDeleteData)
	}
	err = invalidateAdminOtpState(ctx, svc.session, req.AdminAccountId, req.MemberEmail)
	if err != nil {
		logger.Error(err, "error encountered while updating db")
		return nil, status.Errorf(codes.Internal, FailedToUpdateDB)
	}
	return &emptypb.Empty{}, nil
}

func (svc *InvitationService) SendInvitationEmail(ctx context.Context, code string, memberEmail string, adminAccountId string, invitationNotes string) error {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvitationService.SendInvitationEmail").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info(" send invitation email invoked for ", "code", code)

	cloudAccount, err := GetCloudAccount(ctx, adminAccountId)
	if err != nil {
		logger.Error(err, "unable to get cloud account", "adminAccountId", adminAccountId)
		return err
	}

	emailRequest := GetInvitationEmailRequest(code, memberEmail, cloudAccount.GetOwner(), invitationNotes)
	logger.Info("sending email")
	if _, err := svc.notificationClient.SendEmailNotification(ctx, emailRequest); err != nil {
		logger.Error(err, "couldn't send email")
		return status.Errorf(codes.Internal, "unable to send invite")
	}
	return nil
}

func (svc *InvitationService) SendAcceptInvitationEmail(ctx context.Context, code string, memberEmail string, adminAccountId string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvitationService.SendAcceptInvitationEmail").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("send invitation email invoked for ", "code", code)

	cloudAccount, err := GetCloudAccount(ctx, adminAccountId)
	if err != nil {
		logger.Error(err, "unable to get cloud account", "adminAccountId", adminAccountId)
		return err
	}

	emailRequest := GetInvitationAcceptEmailRequest(code, memberEmail, cloudAccount.GetOwner())
	logger.Info("sending email")
	if _, err := svc.notificationClient.SendEmailNotification(ctx, emailRequest); err != nil {
		logger.Error(err, "couldn't send email")
		return status.Errorf(codes.Internal, "unable to send invite")
	}
	return nil
}

func (svc *InvitationService) ValidateInviteCode(ctx context.Context, req *pb.ValidateInviteCodeRequest) (*pb.ValidateInviteCodeResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvitationService.ValidateInviteCode").WithValues("invitationCode", req.InviteCode).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("validate Invite Code")

	resp := &pb.ValidateInviteCodeResponse{
		Valid:           false,
		InvitationState: pb.InvitationState_INVITE_STATE_PENDING_ACCEPT,
	}

	inviteValidationQuery := "SELECT invitation_state, invitation_code, expiry, cloud_account_role_ids FROM members WHERE member_email = $1 AND admin_account_id = $2"
	inviteValidationRow, err := db.QueryContext(ctx, inviteValidationQuery, req.MemberEmail, req.AdminCloudAccountId)
	if err != nil {
		logger.Error(err, FailedToFetchData)
		return nil, errors.New(FailedToFetchData)
	}

	defer inviteValidationRow.Close()

	for inviteValidationRow.Next() {
		var inviteState int
		var cloudAccountRoleIds pq.StringArray
		var inviteCode sql.NullString
		var expirationTime time.Time

		if err := inviteValidationRow.Scan(&inviteState, &inviteCode, &expirationTime, &cloudAccountRoleIds); err != nil {
			logger.Error(err, "error reading result row.")
			return nil, err
		}

		resp.InvitationState = pb.InvitationState(inviteState)

		if resp.InvitationState != pb.InvitationState_INVITE_STATE_PENDING_ACCEPT {
			err = status.Errorf(codes.Internal, "invalid invitation state : "+resp.InvitationState.String())
			logger.Error(err, "invalid invitation state.")
			return resp, err
		}

		if inviteCode.Valid && req.InviteCode == inviteCode.String {
			// Expiration checking only on dates but not time, hence seting time to 00:00:00
			currentTime := time.Now().Truncate(24 * time.Hour)
			inviteCodeExpiryTime := timestamppb.New(expirationTime.Truncate(24 * time.Hour)).AsTime()
			if currentTime.Before(inviteCodeExpiryTime) || currentTime.Equal(inviteCodeExpiryTime) {
				logger.Info("valid invite code")
				resp.Valid = true
				resp.InvitationState = pb.InvitationState_INVITE_STATE_ACCEPTED
			} else {
				logger.Info("expired invite code")
				resp.InvitationState = pb.InvitationState_INVITE_STATE_EXPIRED
			}
		} else {
			logger.Info("invalid invite code")
			return resp, status.Errorf(codes.Internal, "Invalid invitation code")
		}

		updateQuery := "UPDATE members SET invitation_state = $1, updated_at = NOW()::timestamp WHERE member_email = $2 AND admin_account_id = $3"
		result, err := db.ExecContext(ctx, updateQuery, resp.InvitationState, req.MemberEmail, req.AdminCloudAccountId)
		if err != nil {
			logger.Error(err, FailedToUpdateData)
			resp.Valid = false
			resp.InvitationState = pb.InvitationState_INVITE_STATE_PENDING_ACCEPT
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logger.Error(err, FailedToUpdateData)
			resp.Valid = false
			resp.InvitationState = pb.InvitationState_INVITE_STATE_PENDING_ACCEPT
		}
		if rowsAffected < 1 {
			logger.Error(err, "entry to db does not exist")
			resp.Valid = false
			resp.InvitationState = pb.InvitationState_INVITE_STATE_PENDING_ACCEPT
		}

		if resp.InvitationState == pb.InvitationState_INVITE_STATE_EXPIRED {
			return resp, status.Errorf(codes.Internal, "Invitation code expired")
		}

		// set authz system role for cloud_account_member
		if svc.cfg.Authz.Enabled {
			if resp.InvitationState == pb.InvitationState_INVITE_STATE_ACCEPTED {
				if _, err := svc.authzClient.AssignSystemRole(ctx, &pb.RoleRequest{CloudAccountId: req.AdminCloudAccountId, Subject: req.MemberEmail, SystemRole: pb.SystemRole_cloud_account_member.String()}); err != nil {
					logger.Error(err, "couldn't assign systemrole")
					return nil, status.Errorf(codes.Internal, "unable to create systemrole")
				}

				// assign user to cloudAccountRoles
				for _, cloudAccountRoleId := range cloudAccountRoleIds {
					if _, err := svc.authzClient.AddUserToCloudAccountRole(ctx, &pb.CloudAccountRoleUserRequest{
						CloudAccountId: req.AdminCloudAccountId, Id: cloudAccountRoleId, UserId: req.MemberEmail,
					}); err != nil {
						logger.Error(err, "failed to add user to cloud account role", "cloudAccountRoleId", cloudAccountRoleId)
					}
				}

			}
		}

	}

	return resp, err
}

func (svc *InvitationService) SendInviteCode(ctx context.Context, req *pb.SendInviteCodeRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvitationService.SendInviteCode").WithValues("cloudAccountId", req.AdminAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	if req.AdminAccountId == "" {
		logger.Info("missing input argument")
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, missing adminAccountId")
	}
	if req.MemberEmail == "" {
		logger.Info("missing input argument")
		return nil, status.Errorf(codes.InvalidArgument, "invalid input argument, missing memberEmail")
	}

	query := "SELECT invitation_code, expiry, invitation_state FROM members"
	params := protodb.NewProtoToSql(req)
	filter := params.GetFilter()
	if filter != "" {
		query += " " + filter
	}

	var inviteCode string
	var inviteState int32
	var expirationTime time.Time
	if err := svc.session.QueryRowContext(ctx, query, params.GetValues()...).Scan(&inviteCode, &expirationTime, &inviteState); err != nil {
		if err == sql.ErrNoRows {
			logger.Info(fmt.Sprintf("no matching invite found for memberEmail %v", req.MemberEmail))
			return nil, status.Error(codes.NotFound, fmt.Sprintf("no matching invite found for memberEmail %v", req.MemberEmail))
		}
		logger.Error(err, FailedToFetchData)
		return nil, status.Errorf(codes.Internal, FailedToFetchData)
	}

	if inviteState != int32(pb.InvitationState_INVITE_STATE_PENDING_ACCEPT) {
		logger.Info("Unexpected invitationState", "expected", pb.InvitationState_INVITE_STATE_PENDING_ACCEPT, "found", pb.InvitationState_name[inviteState])
		return nil, status.Error(codes.Internal, fmt.Sprintf("expected invitationState: %s, found invitationState: %s", pb.InvitationState_INVITE_STATE_PENDING_ACCEPT, pb.InvitationState_name[inviteState]))
	}

	// this check for any corner case where expired state is pending to be updated
	currentDate := time.Now().Truncate(24 * time.Hour)
	expirationDate := expirationTime.Truncate(24 * time.Hour)
	if expirationDate.Before(currentDate) {
		return nil, status.Error(codes.Internal, fmt.Sprintf("expired invite for memberEmail %v, expirationDate: %v", req.MemberEmail, expirationDate))
	}

	// Sending Email
	err := svc.SendAcceptInvitationEmail(ctx, inviteCode, req.MemberEmail, req.AdminAccountId)
	if err != nil {
		logger.Error(err, "unable to send invite")
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &emptypb.Empty{}, nil
}
