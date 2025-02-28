// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateInvitation(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreateInvitation")
	logger.Info("BEGIN")
	defer logger.Info("End")
	user := "premuser" + uuid.NewString() + "@example.com"
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())

	cloudAccountId, err := caClient.Create(context.Background(),
		&pb.CloudAccountCreate{
			Name:  user,
			Owner: user,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	invitationServiceClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())

	expiry := timestamppb.New(time.Now().AddDate(0, 1, 0))
	memberEmail := "someemail@gmail.com"
	invite := &pb.InviteRequest{
		MemberEmail: memberEmail,
		Expiry:      expiry,
		Note:        "Inviting you to join our IDC cloud account so we can collaborate better 10 allowed ui characters . , ; : @ _ - ",
	}
	invites := []*pb.InviteRequest{}
	invites = append(invites, invite)

	invitationsRequestList := &pb.InviteRequestList{
		CloudAccountId: cloudAccountId.Id,
		Invites:        invites,
	}

	query := `insert into admin_otp (cloud_account_id,member_email,otp_code,otp_state,expiry,created_at,updated_at)
	values ($1,$2,$3,1,NOW()::timestamp,NOW()::timestamp,NOW()::timestamp)`
	otpCode := "123456"
	_, err = db.ExecContext(ctx, query, cloudAccountId.Id, memberEmail, otpCode)
	if err != nil {
		t.Fatalf("error encountered while inserting in otp table: %v", err)
	}
	_, err = invitationServiceClient.CreateInvite(ctx, invitationsRequestList)

	if err != nil {
		t.Fatalf("failed to create invitation: %v", err)
	}
}

func createInvitation(ctx context.Context, t *testing.T, user string, acctType pb.AccountType, memberEmail string) *pb.CloudAccountId {
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())

	cloudAccountId, err := caClient.Create(context.Background(),
		&pb.CloudAccountCreate{
			Name:  user,
			Owner: user,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  acctType,
		})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	invitationServiceClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())

	expiry := timestamppb.New(time.Now().AddDate(0, 1, 0))
	invite := &pb.InviteRequest{
		MemberEmail: memberEmail,
		Expiry:      expiry,
		Note:        "test_note",
	}
	invites := []*pb.InviteRequest{}
	invites = append(invites, invite)

	invitationsRequestList := &pb.InviteRequestList{
		CloudAccountId: cloudAccountId.Id,
		Invites:        invites,
	}

	query := `insert into admin_otp (cloud_account_id,member_email,otp_code,otp_state,expiry,created_at,updated_at)
	values ($1,$2,$3,1,NOW()::timestamp,NOW()::timestamp,NOW()::timestamp)`
	otpCode := "123456"
	_, err = db.ExecContext(ctx, query, cloudAccountId.Id, memberEmail, otpCode)
	if err != nil {
		t.Fatalf("error encountered while inserting in otp table: %v", err)
	}
	_, err = invitationServiceClient.CreateInvite(ctx, invitationsRequestList)

	if err != nil {
		t.Fatalf("failed to create invitation: %v", err)
	}
	return cloudAccountId
}

func TestGetCloudAcctsForInvitees(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestGetCloudAcctsForInvitees")
	logger.Info("BEGIN")
	defer logger.Info("End")

	user := "premuser" + uuid.NewString() + "@example.com"
	memberInviteEmail := "member" + uuid.NewString() + "@intel.com"
	createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberInviteEmail)
	membClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	userCloudAccounts, err := membClient.ReadUserCloudAccounts(ctx, &pb.CloudAccountUser{UserName: memberInviteEmail})

	if err != nil {
		t.Fatalf("failed to read user cloud accounts: %v", err)
	}

	logger.Info("the length of member accounts is", "len", len(userCloudAccounts.MemberAccount))
}

func TestGetCloudAcctsForInviteesAfterRevoke(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestGetCloudAcctsForInviteesAfterRevoke")
	logger.Info("BEGIN")
	defer logger.Info("End")

	user := "premuser" + uuid.NewString() + "@example.com"
	memberToRevokeInviteEmail := "member-revoke" + uuid.NewString() + "@intel.com"
	createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberToRevokeInviteEmail)

	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())
	acct, err := caClient.GetByName(context.Background(), &pb.CloudAccountName{Name: user})
	if err != nil {
		t.Fatalf("failed to read account: %v", err)
	}

	invitationServiceClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())
	invitationServiceClient.RevokeInvite(ctx, &pb.InviteRevokeRequest{
		AdminAccountId:  acct.Id,
		MemberEmail:     memberToRevokeInviteEmail,
		InvitationState: pb.InvitationState_INVITE_STATE_PENDING_ACCEPT,
	})
	if err != nil {
		t.Fatalf("failed to revoke member account: %v", err)
	}

	membClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	userCloudAccounts, err := membClient.ReadUserCloudAccounts(ctx, &pb.CloudAccountUser{UserName: memberToRevokeInviteEmail})

	if err != nil {
		t.Fatalf("failed to read user cloud accounts: %v", err)
	}

	logger.Info("the length of member account is", "len", len(userCloudAccounts.MemberAccount))
}

func TestGetCloudAcctsForInviteesAfterRevokeOnlyActive(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestGetCloudAcctsForInviteesAfterRevokeOnlyActive")
	logger.Info("BEGIN")
	defer logger.Info("End")

	user := "premiumuser" + uuid.NewString() + "@example.com"
	memberToRevokeInviteEmail := "member-revoke" + uuid.NewString() + "@intel.com"
	createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberToRevokeInviteEmail)

	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())
	acct, err := caClient.GetByName(context.Background(), &pb.CloudAccountName{Name: user})
	if err != nil {
		t.Fatalf("failed to read account: %v", err)
	}

	invitationServiceClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())
	invitationServiceClient.RevokeInvite(ctx, &pb.InviteRevokeRequest{
		AdminAccountId:  acct.Id,
		MemberEmail:     memberToRevokeInviteEmail,
		InvitationState: pb.InvitationState_INVITE_STATE_PENDING_ACCEPT,
	})
	if err != nil {
		t.Fatalf("failed to revoke member account: %v", err)
	}

	membClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	onlyActive := true
	userCloudAccounts, err := membClient.ReadUserCloudAccounts(ctx, &pb.CloudAccountUser{
		UserName:   memberToRevokeInviteEmail,
		OnlyActive: &onlyActive,
	})

	if err != nil {
		t.Fatalf("failed to read user cloud accounts: %v", err)
	}

	logger.Info("the length of member account is", "len", len(userCloudAccounts.MemberAccount))
}

func TestGetCloudAcctsForInviteesWithCloudAcct(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestGetCloudAcctsForInviteesWithCloudAcct")
	logger.Info("BEGIN")
	defer logger.Info("End")

	user := "premuser" + uuid.NewString() + "@example.com"
	memberInviteEmail := "member" + uuid.NewString() + "@intel.com"
	createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberInviteEmail)
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())

	_, err := caClient.Create(context.Background(),
		&pb.CloudAccountCreate{
			Name:  memberInviteEmail,
			Owner: memberInviteEmail,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
		})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	membClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	userCloudAccounts, err := membClient.ReadUserCloudAccounts(ctx, &pb.CloudAccountUser{UserName: memberInviteEmail})

	if err != nil {
		t.Fatalf("failed to read user cloud accounts: %v", err)
	}

	logger.Info("the length of member accounts is", "len", len(userCloudAccounts.MemberAccount))
}

func TestReadInvites(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestReadInvites")
	logger.Info("BEGIN")
	defer logger.Info("End")
	assert := assert.New(t)
	expectedInvites := 1
	user := "premuser" + uuid.NewString() + "@example.com"
	memberInviteEmail := "member" + uuid.NewString() + "@proton.me"

	caId := createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberInviteEmail)

	inviteClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())
	invites, err := inviteClient.ReadInvite(ctx, &pb.InviteFilter{AdminAccountId: caId.GetId()})
	if err != nil {
		t.Logf("failed to read invites: %v", err)
	}
	logger.Info("the number of invites", "len", len(invites.GetInvites()))
	assert.Equal(len(invites.GetInvites()), expectedInvites)
}

func TestResendInvites(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestResendInvites")
	logger.Info("BEGIN")
	defer logger.Info("End")
	user := "premuser" + uuid.NewString() + "@example.com"
	memberInviteEmail := "member" + uuid.NewString() + "@proton.me"

	caId := createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberInviteEmail)

	inviteClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())
	_, err := inviteClient.ResendInvite(ctx, &pb.InviteResendRequest{AdminAccountId: caId.GetId(), MemberEmail: memberInviteEmail})
	if err != nil {
		t.Logf("failed to resend invite: %v", err)
	}
}

func TestResendNonExistingInvites(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestResendNonExistingInvites")
	logger.Info("BEGIN")
	defer logger.Info("End")
	expErrMsg := "no matching invite found"
	nonexistingInvite := "nonexistingEmail@example.com"
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())

	caId, err := caClient.Create(context.Background(),
		&pb.CloudAccountCreate{
			Name:  nonexistingInvite,
			Owner: nonexistingInvite,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	inviteClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())
	_, err = inviteClient.ResendInvite(ctx, &pb.InviteResendRequest{AdminAccountId: caId.GetId(), MemberEmail: nonexistingInvite})
	assert.Containsf(t, err.Error(), expErrMsg, "expected error containing %q, got %s", expErrMsg, err)
}

func TestSendInviteCode(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestSendInviteCode")
	logger.Info("BEGIN")
	defer logger.Info("End")

	user := "premuser" + uuid.NewString() + "@example.com"
	memberInviteEmail := "member" + uuid.NewString() + "@proton.me"

	caId := createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberInviteEmail)

	inviteClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())
	_, err := inviteClient.SendInviteCode(ctx, &pb.SendInviteCodeRequest{AdminAccountId: caId.GetId(),
		MemberEmail: memberInviteEmail})
	if err != nil {
		t.Logf("failed to send invitecode: %v", err)
	}
}

func TestValidateInviteCode(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreateInvitation")
	logger.Info("BEGIN")
	defer logger.Info("End")
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())
	invitationServiceClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())

	user := "premuser" + uuid.NewString() + "@example.com"
	memberInviteEmail := "member" + uuid.NewString() + "@intel.com"
	cloudAccountId := createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberInviteEmail)

	var inv_code string
	countQuery := "SELECT invitation_code FROM members WHERE admin_account_id = $1"
	err := db.QueryRowContext(ctx, countQuery, cloudAccountId.Id).Scan(&inv_code)
	if err != nil {
		t.Fatalf("failed to query the database: %v", err)
	}

	req := &pb.ValidateInviteCodeRequest{
		MemberEmail:         memberInviteEmail,
		AdminCloudAccountId: cloudAccountId.Id,
		InviteCode:          inv_code,
	}

	res, err := invitationServiceClient.ValidateInviteCode(ctx, req)

	if err != nil {
		t.Fatalf("An error occured: Invalid Invitation Code: %v", err)
	}
	if !res.Valid {
		t.Fatal("Invitation Code should be Valid")
	}
	if res.InvitationState != pb.InvitationState_INVITE_STATE_ACCEPTED {
		t.Fatal("Invitation State should be accepted")
	}
	mbClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	var mems []string = []string{req.MemberEmail}
	_, err = mbClient.RemoveMembers(context.Background(), &pb.CloudAccountMembers{
		CloudAccountId: user,
		Members:        mems,
	})
	if err != nil {
		t.Fatalf("couldn't remove members: %v", err)
	}
	_, d_err := caClient.Delete(context.Background(),
		&pb.CloudAccountId{
			Id: user,
		})
	if err != nil {
		t.Fatalf("couldn't delete account %v: %v", user, d_err)
	}
}

func TestValidateExpiredInviteCode(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestValidateExpiredInviteCode")
	logger.Info("BEGIN")
	defer logger.Info("End")

	expiredInviteEmail := "expired@intel.com"
	adminAccountId := "admin123"
	inviteCode := strconv.Itoa(rand.Intn(9999999-1000000) + 1000000)
	expirationTime := time.Now().AddDate(0, 0, -1)

	insertQuery := "INSERT INTO members (member_email, admin_account_id, invitation_state, invitation_code, expiry) VALUES ($1, $2, $3, $4, $5)"
	_, err := db.ExecContext(ctx, insertQuery, expiredInviteEmail, adminAccountId, 1, inviteCode, expirationTime)
	if err != nil {
		t.Fatalf("failed to insert expired invitation into the database: %v", err)
	}

	invitationService := &InvitationService{}
	req := &pb.ValidateInviteCodeRequest{
		MemberEmail:         expiredInviteEmail,
		AdminCloudAccountId: adminAccountId,
		InviteCode:          inviteCode,
	}

	resp, err := invitationService.ValidateInviteCode(ctx, req)

	if err == nil {
		t.Fatalf("Expected an error for expired invitation, but got no error.%v", err)
	}

	if resp.Valid {
		t.Fatal("Expected validation failure, but got a valid response.")
	}
	if resp.InvitationState != pb.InvitationState_INVITE_STATE_EXPIRED {
		t.Fatal("Expected Invitation State to be EXPIRED, but got:", resp.InvitationState)
	}
	mbClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	var mems []string = []string{expiredInviteEmail}
	_, err = mbClient.RemoveMembers(context.Background(), &pb.CloudAccountMembers{
		CloudAccountId: req.AdminCloudAccountId,
		Members:        mems,
	})
	if err != nil {
		t.Fatalf("remove members: %v", err)
	}
}

func TestRemoveInvite(t *testing.T) {

	ctx := context.Background()
	user := "premuser" + uuid.NewString() + "@example.com"
	memberEmail := "member" + uuid.NewString() + "@intel.com"

	cloudAccountId := createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberEmail)
	testRecord := &pb.InviteUpdateRequest{
		AdminAccountId:  cloudAccountId.Id,
		MemberEmail:     memberEmail,
		InvitationState: pb.InvitationState_INVITE_STATE_ACCEPTED,
	}

	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InvitationService.RemoveInvite").WithValues("CloudAccountId", testRecord.AdminAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("End")
	invitationService := &InvitationService{
		session:               db,
		cfg:                   test.cfg,
		authzClient:           &authz.AuthzClient{AuthzServiceClient: pb.NewAuthzServiceClient(test.clientConnAuthz)},
		userCredentialsClient: &UserCredentialsClient{userCredentialsServiceClient: pb.NewUserCredentialsServiceClient(test.clientConnCredentials)},
	}

	var count int
	countQuery := "SELECT COUNT(*) FROM members WHERE admin_account_id = $1 AND member_email = $2"
	err := db.QueryRowContext(ctx, countQuery, cloudAccountId.Id, testRecord.MemberEmail).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query the database: %v", err)
	}

	if count == 0 {
		t.Fatal("Record was not inserted into the database.")
	}

	_, err = invitationService.RemoveInvite(ctx, testRecord)
	if err != nil {
		t.Fatalf("RemoveInvite function failed: %v", err)
	}

	countQuery = "SELECT COUNT(*) FROM members WHERE admin_account_id = $1 AND member_email = $2"
	err = db.QueryRowContext(ctx, countQuery, user, testRecord.MemberEmail).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query the database: %v", err)
	}

	if count != 0 {
		t.Fatal("Record was not deleted from the database.")
	}
}

func TestRevokeInvite(t *testing.T) {

	ctx := context.Background()
	user := "premuser" + uuid.NewString() + "@example.com"
	memberEmail := "member" + uuid.NewString() + "@intel.com"

	cloudAccountId := createInvitation(ctx, t, user, pb.AccountType_ACCOUNT_TYPE_PREMIUM, memberEmail)
	testRecord := &pb.InviteRevokeRequest{
		AdminAccountId:  cloudAccountId.Id,
		MemberEmail:     memberEmail,
		InvitationState: pb.InvitationState_INVITE_STATE_PENDING_ACCEPT,
	}

	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InvitationService.RevokeInvite").WithValues("CloudAccountId", testRecord.AdminAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("End")

	invitationService := &InvitationService{
		session: db,
	}

	var count int
	countQuery := "SELECT COUNT(*) FROM members WHERE admin_account_id = $1 AND member_email = $2"
	err := db.QueryRowContext(ctx, countQuery, cloudAccountId.Id, testRecord.MemberEmail).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query the database: %v", err)
	}

	if count == 0 {
		t.Fatal("Record was not inserted into the database.")
	}

	_, err = invitationService.RevokeInvite(ctx, testRecord)
	if err != nil {
		t.Fatalf("RevokeInvite function failed: %v", err)
	}

	countQuery = "SELECT COUNT(*) FROM members WHERE admin_account_id = $1 AND member_email = $2"
	err = db.QueryRowContext(ctx, countQuery, user, testRecord.MemberEmail).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query the database: %v", err)
	}

	if count != 0 {
		t.Fatal("Record was not deleted from the database.")
	}

}

func TestAuthzSystemRoleMemberAssigned(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestSystemRoleAssignMember")
	logger.Info("BEGIN")
	defer logger.Info("End")
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())
	invitationServiceClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())
	authzlient := pb.NewAuthzServiceClient(test.clientConnAuthz)

	user := "premuser" + uuid.NewString() + "@example.com"
	memberInviteEmail := "member" + uuid.NewString() + "@intel.com"

	// Create cloudAccount
	cloudAccount, err := caClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  user,
			Owner: user,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// Creation of cloudAccountRole to Assign to user when the user is invited
	cloudAccountRole, err := authzlient.CreateCloudAccountRole(ctx, &pb.CloudAccountRole{
		Alias: "AliasInviteTest", CloudAccountId: cloudAccount.Id, Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "instance001", Actions: []string{"read"}}}, Users: []string{},
	})
	if err != nil {
		t.Fatalf("failed to create cloudAccountRole: %v", err)
	}

	expiry := timestamppb.New(time.Now().AddDate(0, 1, 0))
	invite := &pb.InviteRequest{
		MemberEmail:         memberInviteEmail,
		Expiry:              expiry,
		Note:                "test_note",
		CloudAccountRoleIds: []string{cloudAccountRole.Id},
	}
	invites := []*pb.InviteRequest{}
	invites = append(invites, invite)

	invitationsRequestList := &pb.InviteRequestList{
		CloudAccountId: cloudAccount.Id,
		Invites:        invites,
	}

	query := `insert into admin_otp (cloud_account_id,member_email,otp_code,otp_state,expiry,created_at,updated_at)
	values ($1,$2,$3,1,NOW()::timestamp,NOW()::timestamp,NOW()::timestamp)`
	otpCode := "123456"
	_, err = db.ExecContext(ctx, query, cloudAccount.Id, memberInviteEmail, otpCode)
	if err != nil {
		t.Fatalf("error encountered while inserting in otp table: %v", err)
	}

	_, err = invitationServiceClient.CreateInvite(ctx, invitationsRequestList)
	if err != nil {
		t.Fatalf("failed to create invitation: %v", err)
	}

	// Given Creation of invitation for member

	var inv_code string
	countQuery := "SELECT invitation_code FROM members WHERE admin_account_id = $1"
	err = db.QueryRowContext(ctx, countQuery, cloudAccount.Id).Scan(&inv_code)
	if err != nil {
		t.Fatalf("failed to query the database: %v", err)
	}

	req := &pb.ValidateInviteCodeRequest{
		MemberEmail:         memberInviteEmail,
		AdminCloudAccountId: cloudAccount.Id,
		InviteCode:          inv_code,
	}

	// When member accepts the invitation
	res, err := invitationServiceClient.ValidateInviteCode(ctx, req)

	if err != nil {
		t.Fatalf("An error occured: Invalid Invitation Code: %v", err)
	}
	if !res.Valid {
		t.Fatal("Invitation Code should be Valid")
	}
	if res.InvitationState != pb.InvitationState_INVITE_STATE_ACCEPTED {
		t.Fatal("Invitation State should be accepted")
	}

	cloudAccountRole, err = authzlient.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{Id: cloudAccountRole.Id, CloudAccountId: cloudAccount.Id})
	if err != nil {
		t.Fatalf("failed to get cloudAccountRole: %v", err)
	}
	if len(cloudAccountRole.Users) != 1 {
		t.Fatal("should have 1 users in the cloudAccountRole")
	}

	existsResponse, err := authzlient.SystemRoleExists(context.Background(), &pb.RoleRequest{CloudAccountId: cloudAccount.Id, Subject: memberInviteEmail, SystemRole: "cloud_account_member"})
	if err != nil {
		t.Fatalf("failed check if systemrole exists: %v", err)
	}
	if test.cfg.Authz.Enabled {
		// Then systemRole should exist for cloudAccountmember
		if !existsResponse.Exist {
			t.Fatalf("error system role should exist")
		}
	} else {
		// Then systemRole should not exist for cloudAccountmember
		if existsResponse.Exist {
			t.Fatalf("error system role should not exist")
		}
	}

	mbClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	var mems []string = []string{req.MemberEmail}
	_, err = mbClient.RemoveMembers(context.Background(), &pb.CloudAccountMembers{
		CloudAccountId: user,
		Members:        mems,
	})
	if err != nil {
		t.Fatalf("couldn't remove members: %v", err)
	}
	_, d_err := caClient.Delete(context.Background(),
		&pb.CloudAccountId{
			Id: user,
		})
	if err != nil {
		t.Fatalf("couldn't delete account %v: %v", user, d_err)
	}
}
func TestCreateInvitationWithInvalidInputNote(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreateInvitationWithInvalidNoteXSS")
	logger.Info("BEGIN")
	defer logger.Info("End")
	user := "premuser" + uuid.NewString() + "@example.com"
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())

	cloudAccountId, err := caClient.Create(context.Background(),
		&pb.CloudAccountCreate{
			Name:  user,
			Owner: user,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	invitationServiceClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())

	expiry := timestamppb.New(time.Now().AddDate(0, 1, 0))

	invites := []*pb.InviteRequest{
		{
			MemberEmail: "someemail@gmail.com",
			Expiry:      expiry,
			Note: `<img src=https://theitgear.com/wp-content/uploads/2024/06/intel-star-banner.png><br><br><b>Dear User, 
			       <br><br>Your account has been inactive. Please use the <a href=#>link</a>
				   to login and activate your account.<br><br>Thanks,
				   <br>IDC Team<br><br><br><br><hr>© 2024 Intel Corporation
				   <br><br><br><br><br><br>`,
		},
	}

	invitationsRequestList := &pb.InviteRequestList{
		CloudAccountId: cloudAccountId.Id,
		Invites:        invites,
	}

	_, err = invitationServiceClient.CreateInvite(ctx, invitationsRequestList)

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("expected 'InvalidArgument' error, but got: %v", err)
	}

	expectedErrMsg := "input note validation doesn't meet the requirements"
	if !strings.Contains(st.Message(), expectedErrMsg) {
		t.Fatalf("expected error message to contain '%s', but got: %s", expectedErrMsg, st.Message())
	}
}

func TestAuthzSystemRoleMemberUnAssigned(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestAuthzSystemRoleMemberUnAssigned")
	logger.Info("BEGIN")
	defer logger.Info("End")
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())
	invitationServiceClient := pb.NewCloudAccountInvitationServiceClient(test.ClientConn())
	authzlient := pb.NewAuthzServiceClient(test.clientConnAuthz)

	user := "premuser" + uuid.NewString() + "@example.com"
	memberInviteEmail := "member" + uuid.NewString() + "@intel.com"

	// Create cloudAccount
	cloudAccount, err := caClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  user,
			Owner: user,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// Creation of cloudAccountRole to Assign to user when the user is invited
	cloudAccountRole, err := authzlient.CreateCloudAccountRole(ctx, &pb.CloudAccountRole{
		Alias: "AliasInviteTest", CloudAccountId: cloudAccount.Id, Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "instance001", Actions: []string{"read"}}}, Users: []string{},
	})
	if err != nil {
		t.Fatalf("failed to create cloudAccountRole: %v", err)
	}

	expiry := timestamppb.New(time.Now().AddDate(0, 1, 0))
	invite := &pb.InviteRequest{
		MemberEmail:         memberInviteEmail,
		Expiry:              expiry,
		Note:                "test_note",
		CloudAccountRoleIds: []string{cloudAccountRole.Id},
	}
	invites := []*pb.InviteRequest{}
	invites = append(invites, invite)

	invitationsRequestList := &pb.InviteRequestList{
		CloudAccountId: cloudAccount.Id,
		Invites:        invites,
	}

	query := `insert into admin_otp (cloud_account_id,member_email,otp_code,otp_state,expiry,created_at,updated_at)
	values ($1,$2,$3,1,NOW()::timestamp,NOW()::timestamp,NOW()::timestamp)`
	otpCode := "123456"
	_, err = db.ExecContext(ctx, query, cloudAccount.Id, memberInviteEmail, otpCode)
	if err != nil {
		t.Fatalf("error encountered while inserting in otp table: %v", err)
	}

	_, err = invitationServiceClient.CreateInvite(ctx, invitationsRequestList)
	if err != nil {
		t.Fatalf("failed to create invitation: %v", err)
	}

	// Given Creation of invitation for member
	var inv_code string
	countQuery := "SELECT invitation_code FROM members WHERE admin_account_id = $1"
	err = db.QueryRowContext(ctx, countQuery, cloudAccount.Id).Scan(&inv_code)
	if err != nil {
		t.Fatalf("failed to query the database: %v", err)
	}

	req := &pb.ValidateInviteCodeRequest{
		MemberEmail:         memberInviteEmail,
		AdminCloudAccountId: cloudAccount.Id,
		InviteCode:          inv_code,
	}

	// When member accepts the invitation
	res, err := invitationServiceClient.ValidateInviteCode(ctx, req)

	if err != nil {
		t.Fatalf("An error occured: Invalid Invitation Code: %v", err)
	}
	if !res.Valid {
		t.Fatal("Invitation Code should be Valid")
	}
	if res.InvitationState != pb.InvitationState_INVITE_STATE_ACCEPTED {
		t.Fatal("Invitation State should be accepted")
	}

	cloudAccountRole, err = authzlient.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{Id: cloudAccountRole.Id, CloudAccountId: cloudAccount.Id})
	if err != nil {
		t.Fatalf("failed to get cloudAccountRole: %v", err)
	}
	if len(cloudAccountRole.Users) != 1 {
		t.Fatal("should have 1 users in the cloudAccountRole")
	}

	existsResponse, err := authzlient.SystemRoleExists(context.Background(), &pb.RoleRequest{CloudAccountId: cloudAccount.Id, Subject: memberInviteEmail, SystemRole: "cloud_account_member"})
	if err != nil {
		t.Fatalf("failed check if systemrole exists: %v", err)
	}
	if test.cfg.Authz.Enabled {
		// Then systemRole should exist for cloudAccountmember
		if !existsResponse.Exist {
			t.Fatalf("error system role should exist")
		}
	} else {
		// Then systemRole should not exist for cloudAccountmember
		if existsResponse.Exist {
			t.Fatalf("error system role should not exist")
		}
	}

	invitationServiceClient.RemoveInvite(ctx, &pb.InviteUpdateRequest{AdminAccountId: cloudAccount.Id, MemberEmail: memberInviteEmail, InvitationState: res.InvitationState})
	existsResponse, err = authzlient.SystemRoleExists(context.Background(), &pb.RoleRequest{CloudAccountId: cloudAccount.Id, Subject: memberInviteEmail, SystemRole: "cloud_account_member"})
	if err != nil {
		t.Fatalf("failed check if systemrole exists: %v", err)
	}
	if test.cfg.Authz.Enabled {
		// Then systemRole should not exist for cloudAccountmember since invite was removed
		if existsResponse.Exist {
			t.Fatalf("error system role should not exist")
		}
	} else {
		// Then systemRole should not exist for cloudAccountmember
		if existsResponse.Exist {
			t.Fatalf("error system role should not exist")
		}
	}

	cloudAccountRole, err = authzlient.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{Id: cloudAccountRole.Id, CloudAccountId: cloudAccount.Id})
	if err != nil {
		t.Fatalf("failed to get cloudAccountRole: %v", err)
	}
	if len(cloudAccountRole.Users) != 0 {
		t.Fatal("should have 0 users in the cloudAccountRole since user member was removed from all cloud account roles")
	}

	mbClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	var mems []string = []string{req.MemberEmail}
	_, err = mbClient.RemoveMembers(context.Background(), &pb.CloudAccountMembers{
		CloudAccountId: user,
		Members:        mems,
	})
	if err != nil {
		t.Fatalf("couldn't remove members: %v", err)
	}
	_, d_err := caClient.Delete(context.Background(),
		&pb.CloudAccountId{
			Id: user,
		})
	if err != nil {
		t.Fatalf("couldn't delete account %v: %v", user, d_err)
	}
}
