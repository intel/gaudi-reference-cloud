// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"database/sql"
	"errors"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type InvitationMemberService struct {
	pb.UnimplementedCloudAccountInvitationMemberServiceServer
	session *sql.DB
}

func (svc *InvitationMemberService) RejectInvite(ctx context.Context, req *pb.InviteUpdateRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InvitationMemberService.RevokeInvite").WithValues("cloudAccountId", req.AdminAccountId).Start()
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
		return &emptypb.Empty{}, status.Errorf(codes.NotFound, "reject not supported for invitation_state (%v)", req.InvitationState)
	}

	logger.Info("rejecting member invite", "cloudAccountId", req.AdminAccountId)
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
		logger.Error(err, "error encountered while invalidating otp")
		return nil, status.Errorf(codes.Internal, "failed to invalidate Otp")
	}

	return &emptypb.Empty{}, nil
}
