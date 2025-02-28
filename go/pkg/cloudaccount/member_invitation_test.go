// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"

	"testing"

	"github.com/google/uuid"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestRejectInvite(t *testing.T) {

	ctx := context.Background()
	user := "premuser" + uuid.NewString() + "@example.com"
	memberEmail := "member" + uuid.NewString() + "@intel.com"

	cloudAccountId := createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberEmail)
	testRecord := &pb.InviteUpdateRequest{
		AdminAccountId:  cloudAccountId.Id,
		MemberEmail:     memberEmail,
		InvitationState: pb.InvitationState_INVITE_STATE_PENDING_ACCEPT,
	}

	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InvitationMemberService.RevokeInvite").WithValues("CloudAccountId", testRecord.AdminAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("End")

	invitationMemberService := &InvitationMemberService{
		session: db,
	}

	var count int
	countQuery := "SELECT COUNT(*) FROM members WHERE admin_account_id = $1 AND member_email = $2"
	err := db.QueryRowContext(ctx, countQuery, testRecord.AdminAccountId, testRecord.MemberEmail).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query the database: %v", err)
	}

	if count == 0 {
		t.Fatal("Record was not inserted into the database.")
	}

	_, err = invitationMemberService.RejectInvite(ctx, testRecord)
	if err != nil {
		t.Fatalf("RejectInvite function failed: %v", err)
	}

	countQuery = "SELECT COUNT(*) FROM members WHERE admin_account_id = $1 AND member_email = $2"
	err = db.QueryRowContext(ctx, countQuery, testRecord.AdminAccountId, testRecord.MemberEmail).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query the database: %v", err)
	}

	if count != 0 {
		t.Fatal("Record was not deleted from the database.")
	}
}
