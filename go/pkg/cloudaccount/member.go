// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CloudAccountMemberService struct {
	pb.UnimplementedCloudAccountMemberServiceServer
}

const (
	FailedToDeleteData = "failed to delete data"
	FailedToUpdateData = "failed to update data"
)

func (ms *CloudAccountMemberService) ReadActiveMembers(ctx context.Context, cloudAccountId *pb.CloudAccountId) (*pb.CloudAccountMembers, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountMemberService.ReadActiveMembers").WithValues("cloudAccountId", cloudAccountId.GetId()).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	rows, err := db.QueryContext(ctx,
		"SELECT member_email from members WHERE admin_account_id = $1 AND invitation_state = $2 AND revoked = false",
		cloudAccountId.GetId(), pb.InvitationState_INVITE_STATE_ACCEPTED)

	if err != nil {
		log.Error(err, FailedToFetchData)
		return nil, errors.New(FailedToFetchData)
	}
	defer rows.Close()

	members := pb.CloudAccountMembers{CloudAccountId: cloudAccountId.GetId()}
	for rows.Next() {
		member := ""
		if err = rows.Scan(&member); err != nil {
			log.Error(err, "Error while fetching members", "cloudAccountId", cloudAccountId.Id)
			return nil, err
		}
		members.Members = append(members.Members, member)
	}
	return &members, nil
}

func (ms *CloudAccountMemberService) ReadMembers(ctx context.Context,
	cloudAccountId *pb.CloudAccountId) (*pb.CloudAccountMembers, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountMemberService.ReadMembers").WithValues("cloudAccountId", cloudAccountId.GetId()).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	log.V(9).Info("ReadMembers invoked", "cloudAccountId", cloudAccountId.Id)
	rows, err := db.QueryContext(ctx,
		"SELECT member from cloud_account_members WHERE account_id = $1",
		cloudAccountId.GetId())

	if err != nil {
		log.Error(err, FailedToFetchData)
		return nil, errors.New(FailedToFetchData)
	}
	defer rows.Close()

	members := pb.CloudAccountMembers{}
	for rows.Next() {
		member := ""
		if err = rows.Scan(&member); err != nil {
			log.Error(err, "Error while fetching members", "cloudAccountId", cloudAccountId.Id)
			return nil, err
		}
		members.Members = append(members.Members, member)
	}
	return &members, nil
}

func (ms *CloudAccountMemberService) AddMembers(ctx context.Context,
	members *pb.CloudAccountMembers) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountMemberService.AddMembers").WithValues("cloudAccountId", members.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	log.V(9).Info("AddMembers invoked", "cloudAccountId", members.CloudAccountId)

	builder := strings.Builder{}
	args := []any{}
	args = append(args, members.CloudAccountId)

	for ii, member := range members.Members {
		if ii > 0 {
			_, err := builder.WriteRune(',')
			if err != nil {
				return nil, err
			}
		}
		_, err := builder.WriteString("($1, $")
		if err != nil {
			return nil, err
		}

		_, err = builder.WriteString(strconv.Itoa(ii + 2))
		if err != nil {
			return nil, err
		}

		_, err = builder.WriteRune(')')
		if err != nil {
			return nil, err
		}

		args = append(args, member)
	}
	_, err := db.ExecContext(ctx,
		"INSERT INTO cloud_account_members (account_id, member) VALUES "+
			builder.String(), args...)
	if err != nil {
		log.Error(err, FailedToInsertDBData)
		return &emptypb.Empty{}, errors.New(FailedToInsertDBData)
	}
	return &emptypb.Empty{}, err
}

func (ms *CloudAccountMemberService) RemoveMembers(ctx context.Context,
	members *pb.CloudAccountMembers) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountMemberService.RemoveMembers").WithValues("cloudAccountId", members.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	log.V(9).Info("RemoveMembers invoked", "cloudAccountId", members.CloudAccountId)

	builder := strings.Builder{}
	args := []any{}
	args = append(args, members.CloudAccountId)

	_, err := builder.WriteRune('(')
	if err != nil {
		return nil, err
	}

	for ii, member := range members.Members {
		if ii > 0 {
			_, err := builder.WriteRune(',')
			if err != nil {
				return nil, err
			}
		}
		_, err := builder.WriteRune('$')
		if err != nil {
			return nil, err
		}

		_, err = builder.WriteString(strconv.Itoa(ii + 2))
		if err != nil {
			return nil, err
		}

		args = append(args, member)
	}

	_, err = builder.WriteRune(')')
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx,
		"DELETE FROM cloud_account_members WHERE account_id = $1 and member IN "+
			builder.String(), args...)
	if err != nil {
		log.Error(err, FailedToDeleteData)
	}
	return &emptypb.Empty{}, err
}

func (ms *CloudAccountMemberService) ReadUserCloudAccounts(ctx context.Context,
	cloudAccountUser *pb.CloudAccountUser) (*pb.MemberAccount, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudAccountMemberService.ReadUserCloudAccounts").WithValues("userName", cloudAccountUser.UserName).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	resp := pb.MemberAccount{}
	adminAccountQuery := "SELECT id, name, owner, type FROM cloud_accounts WHERE name = $1"

	adminAccountrows, err := db.QueryContext(ctx, adminAccountQuery, cloudAccountUser.UserName)
	if err != nil {
		logger.Error(err, FailedToFetchData)
		return nil, errors.New(FailedToFetchData)
	}

	defer adminAccountrows.Close()
	for adminAccountrows.Next() {
		memberAccount := pb.MemberCloudAccount{}

		var id sql.NullString
		var name sql.NullString
		var owner sql.NullString

		if err := adminAccountrows.Scan(&id, &name, &owner, &memberAccount.Type); err != nil {
			logger.Error(err, "error reading result row.")
			return nil, err
		}

		if id.Valid {
			memberAccount.Id = id.String
		} else {
			memberAccount.Id = ""
		}

		if name.Valid {
			memberAccount.Name = name.String
		} else {
			memberAccount.Name = ""
		}

		if owner.Valid {
			memberAccount.Owner = owner.String
		} else {
			memberAccount.Owner = ""
		}

		resp.MemberAccount = append(resp.MemberAccount, &memberAccount)
	}

	// Member Accounts Query
	memberAccountsQuery := "SELECT id, name, owner, type FROM cloud_accounts WHERE id IN (SELECT admin_account_id from members WHERE member_email = $1)"
	memberAccountsRows, err := db.QueryContext(ctx, memberAccountsQuery, cloudAccountUser.UserName)
	if err != nil {
		logger.Error(err, FailedToFetchData)
		return nil, errors.New(FailedToFetchData)
	}
	defer memberAccountsRows.Close()

	for memberAccountsRows.Next() {
		memberAccount := pb.MemberCloudAccount{}

		var id sql.NullString
		var name sql.NullString
		var owner sql.NullString

		if err := memberAccountsRows.Scan(&id, &name, &owner, &memberAccount.Type); err != nil {
			logger.Error(err, "error reading result row.")
			return nil, err
		}

		if id.Valid {
			memberAccount.Id = id.String
		} else {
			memberAccount.Id = ""
		}

		if name.Valid {
			memberAccount.Name = name.String
		} else {
			memberAccount.Name = ""
		}

		if owner.Valid {
			memberAccount.Owner = owner.String
		} else {
			memberAccount.Owner = ""
		}

		inviteStatusQuery := "SELECT invitation_state FROM members WHERE admin_account_id = $1 AND member_email = $2"
		inviteStatusRow, err := db.QueryContext(ctx, inviteStatusQuery, memberAccount.Id, cloudAccountUser.UserName)
		if err != nil {
			logger.Error(err, FailedToFetchData)
			return nil, errors.New(FailedToFetchData)
		}
		defer inviteStatusRow.Close()

		if !inviteStatusRow.Next() {
			return nil, status.Errorf(codes.NotFound, "Invite status not found for %v", memberAccount.Id)
		}

		if err := inviteStatusRow.Scan(&memberAccount.InvitationState); err != nil {
			logger.Error(err, "error reading result row.")
			return nil, err
		}

		if cloudAccountUser.OnlyActive == nil {
			resp.MemberAccount = append(resp.MemberAccount, &memberAccount)
		}

		if (cloudAccountUser.OnlyActive != nil) &&
			(*cloudAccountUser.OnlyActive) &&
			((memberAccount.InvitationState == pb.InvitationState_INVITE_STATE_PENDING_ACCEPT) ||
				(memberAccount.InvitationState == pb.InvitationState_INVITE_STATE_ACCEPTED)) {
			resp.MemberAccount = append(resp.MemberAccount, &memberAccount)
		}

	}

	return &resp, nil
}

func (ms *CloudAccountMemberService) GetCloudAccountsForOpa(ctx context.Context,
	accountUser *pb.AccountUser) (*pb.RelatedAccounts, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountMemberService.GetCloudAccountsForOpa").WithValues("userName", accountUser.UserName).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	resp := pb.RelatedAccounts{}

	adminAccountQuery := "SELECT id, name, owner FROM cloud_accounts WHERE name = $1"
	adminAccountrows, err := db.QueryContext(ctx, adminAccountQuery, accountUser.UserName)
	if err != nil {
		logger.Error(err, FailedToFetchData)
		return nil, errors.New(FailedToFetchData)
	}

	defer adminAccountrows.Close()
	for adminAccountrows.Next() {
		relatedAccount := pb.RelatedAccount{}

		var id sql.NullString
		var name sql.NullString
		var owner sql.NullString

		if err := adminAccountrows.Scan(&id, &name, &owner); err != nil {
			logger.Error(err, "error reading result row.")
			return nil, err
		}

		if id.Valid {
			relatedAccount.Id = id.String
		} else {
			relatedAccount.Id = ""
		}

		if name.Valid {
			relatedAccount.Name = name.String
		} else {
			relatedAccount.Name = ""
		}

		if owner.Valid {
			relatedAccount.Owner = owner.String
		} else {
			relatedAccount.Owner = ""
		}

		resp.RelatedAccounts = append(resp.RelatedAccounts, &relatedAccount)
	}

	// Member Accounts Query
	memberAccountsQuery := "SELECT id, name, owner FROM cloud_accounts WHERE id IN (SELECT admin_account_id from members WHERE member_email = $1)"
	memberAccountsRows, err := db.QueryContext(ctx, memberAccountsQuery, accountUser.UserName)
	if err != nil {
		logger.Error(err, FailedToFetchData)
		return nil, errors.New(FailedToFetchData)
	}
	defer memberAccountsRows.Close()

	for memberAccountsRows.Next() {
		memberAccount := pb.RelatedAccount{}

		var id sql.NullString
		var name sql.NullString
		var owner sql.NullString

		if err := memberAccountsRows.Scan(&id, &name, &owner); err != nil {
			logger.Error(err, "error reading result row.")
			return nil, err
		}

		if id.Valid {
			memberAccount.Id = id.String
		} else {
			memberAccount.Id = ""
		}

		if name.Valid {
			memberAccount.Name = name.String
		} else {
			memberAccount.Name = ""
		}

		if owner.Valid {
			memberAccount.Owner = owner.String
		} else {
			memberAccount.Owner = ""
		}

		inviteStatusQuery := "SELECT invitation_state FROM members WHERE admin_account_id = $1 AND member_email = $2"
		inviteStatusRow, err := db.QueryContext(ctx, inviteStatusQuery, memberAccount.Id, accountUser.UserName)
		if err != nil {
			logger.Error(err, FailedToFetchData)
			return nil, errors.New(FailedToFetchData)
		}
		defer inviteStatusRow.Close()

		if !inviteStatusRow.Next() {
			return nil, status.Errorf(codes.NotFound, "Invite status not found for %v", memberAccount.Id)
		}

		var invitationState pb.InvitationState
		if err := inviteStatusRow.Scan(&invitationState); err != nil {
			logger.Error(err, "error reading result row.")
			return nil, err
		}

		if invitationState == pb.InvitationState_INVITE_STATE_UNSPECIFIED ||
			invitationState == pb.InvitationState_INVITE_STATE_ACCEPTED {
			resp.RelatedAccounts = append(resp.RelatedAccounts, &memberAccount)
		}
	}

	return &resp, nil
}

func (ms *CloudAccountMemberService) UpdatePersonId(ctx context.Context,
	memberPersonId *pb.MemberPersonId) (*emptypb.Empty, error) {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountMemberService.UpdatePersonId").WithValues("personId", memberPersonId.PersonId).Start()
	defer span.End()
	logger.Info("update person id")
	defer logger.Info("update person id")

	updateQuery := "UPDATE members SET person_id = $1, updated_at = NOW()::timestamp WHERE member_email = $2"
	result, err := db.ExecContext(ctx, updateQuery, memberPersonId.PersonId, memberPersonId.MemberEmail)
	if err != nil {
		logger.Error(err, FailedToUpdateData)
		return nil, errors.New(FailedToUpdateData)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if rowsAffected < 1 {
		return nil, status.Errorf(codes.NotFound, "no entry in the db of member_email: (%v).  failed to update", memberPersonId.MemberEmail)
	}

	return &emptypb.Empty{}, err
}

func (ms *CloudAccountMemberService) GetMemberPersonId(ctx context.Context,
	accountUser *pb.AccountUser) (*pb.AccountPerson, error) {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountMemberService.GetMemberPersonId").WithValues("userName", accountUser.UserName, "cloudAccountId", accountUser.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	resp := pb.AccountPerson{}

	personIdQuery := "SELECT person_id FROM members WHERE admin_account_id = $1 AND member_email = $2"
	personIdRow, err := db.QueryContext(ctx, personIdQuery, accountUser.CloudAccountId, accountUser.UserName)
	if err != nil {
		logger.Error(err, FailedToFetchData)
		return nil, errors.New(FailedToFetchData)
	}
	defer personIdRow.Close()

	if !personIdRow.Next() {
		return nil, status.Errorf(codes.NotFound, "PersonId not found for %v", accountUser.UserName)
	}

	if err := personIdRow.Scan(&resp.PersonId); err != nil {
		logger.Error(err, "error reading result row.")
		return nil, err
	}

	return &resp, nil
}
