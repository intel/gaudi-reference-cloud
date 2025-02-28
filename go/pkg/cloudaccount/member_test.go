// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"sort"
	"testing"

	"github.com/google/uuid"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestMembers(t *testing.T) {
	user := "muser@example.com"
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())

	_, err := caClient.Create(context.Background(),
		&pb.CloudAccountCreate{
			Name:  user,
			Owner: user,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
		})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	acct, err := caClient.GetByName(context.Background(), &pb.CloudAccountName{Name: user})
	if err != nil {
		t.Fatalf("read account: %v", err)
	}

	membClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	members, err := membClient.ReadMembers(context.Background(),
		&pb.CloudAccountId{Id: acct.GetId()})
	if err != nil {
		t.Fatalf("read members: %v", err)
	}
	if len(members.Members) != 1 || members.Members[0] != user {
		t.Errorf("members for newly created cloud account are wrong")
	}

	t.Run("addMembers", func(t *testing.T) { testAddMembers(t, membClient, acct) })
	t.Run("removeMembers", func(t *testing.T) { testRemoveMembers(t, membClient, acct) })
}

var testMembers []string = []string{"user1@example.com", "user2@example.com", "user3.example.com"}

func testAddMembers(t *testing.T, mbClient pb.CloudAccountMemberServiceClient,
	acct *pb.CloudAccount) {

	_, err := mbClient.AddMembers(context.Background(), &pb.CloudAccountMembers{
		CloudAccountId: acct.GetId(),
		Members:        testMembers,
	})
	if err != nil {
		t.Fatalf("add members: %v", err)
	}

	members, err := mbClient.ReadMembers(context.Background(),
		&pb.CloudAccountId{Id: acct.GetId()})
	if err != nil {
		t.Fatalf("read members: %v", err)
	}

	if len(members.Members) != len(testMembers)+1 {
		t.Errorf("member list wrong after adding members")
		return
	}

	checkMembers := make([]string, len(testMembers))
	copy(checkMembers, testMembers)
	checkMembers = append(checkMembers, acct.GetName())

	sort.Strings(checkMembers)
	sort.Strings(members.Members)

	for ii := range checkMembers {
		if checkMembers[ii] != members.Members[ii] {
			t.Errorf("member list wrong after adding members")
			break
		}
	}
}

func testRemoveMembers(t *testing.T, mbClient pb.CloudAccountMemberServiceClient,
	acct *pb.CloudAccount) {
	_, err := mbClient.RemoveMembers(context.Background(), &pb.CloudAccountMembers{
		CloudAccountId: acct.GetId(),
		Members:        testMembers,
	})
	if err != nil {
		t.Fatalf("remove members: %v", err)
	}

	members, err := mbClient.ReadMembers(context.Background(),
		&pb.CloudAccountId{Id: acct.GetId()})
	if err != nil {
		t.Fatalf("read members: %v", err)
	}

	if len(members.Members) != 1 || members.Members[0] != acct.GetName() {
		t.Errorf("members for cloud account are wrong after removal")
	}
}

////

// func TestReadUserCloudAccounts(t *testing.T) {

// 	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())
// 	mbClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())

// 	const kNumUsers = 100
// 	const kNumMultiAcctUsers = 5
// 	const kNumAcctsPerMultiUser = 4

// 	start := time.Now()
// 	acctMap := make(map[int]string)
// 	for uu := 0; uu < kNumUsers; uu++ {
// 		user := fmt.Sprintf("u%v@example.com", uu)
// 		_, err := caClient.Create(context.Background(),
// 			&pb.CloudAccountCreate{
// 				Name:  user,
// 				Owner: user,
// 				Tid:   uuid.NewString(),
// 				Oid:   uuid.NewString(),
// 				Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
// 			})
// 		if err != nil {
// 			t.Fatalf("create account %v: %v", user, err)
// 		}
// 		acct, err := caClient.GetByName(context.Background(), &pb.CloudAccountName{Name: user})
// 		if err != nil {
// 			t.Fatalf("read account %v: %v", user, err)
// 		}
// 		acctMap[uu] = acct.GetId()
// 	}
// 	now := time.Now()
// 	t.Logf("%v: created %v users", now.Sub(start), kNumUsers)
// 	start = now

// 	for uu := 0; uu < kNumMultiAcctUsers; uu++ {
// 		user := fmt.Sprintf("u%v@example.com", uu)
// 		for aa := 0; aa < kNumAcctsPerMultiUser; aa++ {
// 			ii := (uu + 1 + aa*kNumUsers/kNumMultiAcctUsers) % kNumUsers
// 			members := pb.CloudAccountMembers{CloudAccountId: acctMap[ii]}
// 			members.Members = append(members.Members, user)
// 			_, err := mbClient.AddMembers(context.Background(), &members)
// 			if err != nil {
// 				t.Fatalf("add members %v: %v", user, err)
// 			}
// 		}
// 	}

// 	now = time.Now()
// 	t.Logf("%v: added %v members", now.Sub(start), kNumMultiAcctUsers*kNumAcctsPerMultiUser)

// 	for uu := 0; uu < kNumMultiAcctUsers; uu++ {
// 		user := fmt.Sprintf("u%v@example.com", uu)
// 		stream, err := mbClient.ReadUserCloudAccounts(context.Background(),
// 			&pb.CloudAccountUser{UserName: user})
// 		if err != nil {
// 			t.Errorf("read cloud accounts: %v: %v", user, err)
// 			break
// 		}
// 		actualAccts := []string{}
// 		for {
// 			acct, err := stream.Recv()
// 			if err != nil {
// 				if errors.Is(err, io.EOF) {
// 					break
// 				}
// 				t.Errorf("read cloud account: %v", err)
// 				break
// 			}
// 			actualAccts = append(actualAccts, acct.GetId())
// 		}
// 		sort.Strings(actualAccts)

// 		desiredAccts := append([]string(nil), acctMap[uu])
// 		for aa := 0; aa < kNumAcctsPerMultiUser; aa++ {
// 			ii := (uu + 1 + aa*kNumUsers/kNumMultiAcctUsers) % kNumUsers
// 			desiredAccts = append(desiredAccts, acctMap[ii])
// 		}
// 		sort.Strings(desiredAccts)

// 		if len(actualAccts) != len(desiredAccts) {
// 			t.Errorf("%v: # of accounts is wrong", user)
// 			continue
// 		}
// 		for ii, actual := range actualAccts {
// 			if actual != desiredAccts[ii] {
// 				t.Errorf("%v: accounts are wrong", user)
// 				break
// 			}
// 		}
// 	}

// 	now = time.Now()
// 	t.Logf("%v: read cloud accounts for %v users", now.Sub(start), kNumMultiAcctUsers)
// 	start = now

// 	for _, id := range acctMap {
// 		_, err := caClient.Delete(context.Background(),
// 			&pb.CloudAccountId{
// 				Id: id,
// 			})
// 		if err != nil {
// 			t.Fatalf("delete account %v: %v", id, err)
// 		}
// 	}

// 	t.Logf("%v: deleted %v cloud accounts", time.Since(start), len(acctMap))
// }

func TestReadActiveMembers(t *testing.T) {
	ctx := context.Background()

	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudAccountMemberService.ReadUserCloudAccounts").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("End")

	mbClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())

	insertQuery := "INSERT INTO cloud_accounts (id, name, owner, type, tid, oid) VALUES ($1, $2, $3, $4, $5, $6)"
	id := "900000000110"

	admin := fmt.Sprintf("adminaaa%v@intel.com", 0)
	tid := fmt.Sprintf("4015bb99-0522-4387-b47e-c821595dd70%v", 0)
	oid := fmt.Sprintf("61befbee-0607-47c5-b140-c4509dgek80%v", 0)
	_, err := db.ExecContext(ctx, insertQuery, id, admin, admin, "5", tid, oid)
	if err != nil {
		t.Fatalf("failed to insert test record into the database: %v", err)
	}

	var InvitationStates []int = []int{1, 2, 5, 6}
	for uu := 0; uu < 4; uu++ {
		invitation_code := fmt.Sprintf("31%v", uu)
		insertQuery := "INSERT INTO members (admin_account_id, member_email, invitation_state, invitation_code) VALUES ($1, $2, $3, $4)"
		_, m_err := db.ExecContext(ctx, insertQuery, id, fmt.Sprintf("aaa1%v@intel.com", uu), InvitationStates[uu], invitation_code)
		if m_err != nil {
			t.Fatalf("failed to insert test record into the database: %v", m_err)
		}

	}

	res, err := mbClient.ReadActiveMembers(context.Background(), &pb.CloudAccountId{Id: id})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(res.Members) != 1 {
		t.Fatal("Account data couldn't be read successfully")
	}

}

func TestReadUserCloudAccounts(t *testing.T) {
	ctx := context.Background()

	onlyActive := true
	cloudAccountUser1 := &pb.CloudAccountUser{
		UserName:   "m1@intel.com",
		OnlyActive: nil,
	}
	cloudAccountUser2 := &pb.CloudAccountUser{
		UserName:   "m1@intel.com",
		OnlyActive: &onlyActive,
	}
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudAccountMemberService.ReadUserCloudAccounts").WithValues("UserName", cloudAccountUser1.UserName).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("End")

	mbClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())

	adminIds := make([]string, 0)
	No_of_Admin_accts := 4
	// 4 new cloud account records are being inserted into the database inside the loop below
	for ind := 0; ind < No_of_Admin_accts; ind++ {
		insertQuery := "INSERT INTO cloud_accounts (id, name, owner, type, tid, oid) VALUES ($1, $2, $3, $4, $5, $6)"
		id := fmt.Sprintf("00000000011%v", ind)

		adminIds = append(adminIds, id)
		admin := fmt.Sprintf("a%v@intel.com", ind)
		tid := fmt.Sprintf("4015bb99-0522-4387-b47e-c821596dd70%v", ind)
		oid := fmt.Sprintf("61befbee-0607-47c5-b140-c4509dfek80%v", ind)
		_, err := db.ExecContext(ctx, insertQuery, id, admin, admin, "5", tid, oid)
		if err != nil {
			t.Fatalf("failed to insert test record into the database: %v", err)
		}
	}

	var InvitationStates []int = []int{1, 2, 5, 6}
	// 4 new member account records are being inserted into the database inside the loop below
	for uu := 0; uu < 4; uu++ {
		user := adminIds[uu]
		invitation_code := fmt.Sprintf("00%v", uu)
		insertQuery := "INSERT INTO members (admin_account_id, member_email, invitation_state, invitation_code) VALUES ($1, $2, $3, $4)"
		_, m_err := db.ExecContext(ctx, insertQuery, user, "m1@intel.com", InvitationStates[uu], invitation_code)
		if m_err != nil {
			t.Fatalf("failed to insert test record into the database: %v", m_err)
		}

	}

	res, err := mbClient.ReadUserCloudAccounts(context.Background(), cloudAccountUser1)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	actualAccts := make([]string, 0)

	for _, acct := range res.MemberAccount {
		actualAccts = append(actualAccts, acct.GetId())
	}
	if len(actualAccts) != No_of_Admin_accts {
		t.Fatal("Account data couldn't be read successfully")
	}

	// for  active invitations
	No_Active_Accts := 2
	nres, err := mbClient.ReadUserCloudAccounts(context.Background(), cloudAccountUser2)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	ActiveAccts := make([]string, 0)
	for _, acct := range nres.MemberAccount {
		ActiveAccts = append(ActiveAccts, acct.GetId())
		if (acct.InvitationState != pb.InvitationState_INVITE_STATE_PENDING_ACCEPT) && (acct.InvitationState != pb.InvitationState_INVITE_STATE_ACCEPTED) {
			t.Fatal("Invitation State is inactive")
		}
	}
	if len(ActiveAccts) != No_Active_Accts {
		t.Fatal("Active Account's data couldn't be read successfully")
	}
	var mems []string = []string{"m1@intel.com"}
	for i, usr := range actualAccts {
		_, err = caClient.Delete(context.Background(),
			&pb.CloudAccountId{
				Id: usr,
			})
		if err != nil {
			t.Fatalf("delete account %v: %v", usr, err)
		}

		_, err := mbClient.RemoveMembers(context.Background(), &pb.CloudAccountMembers{
			CloudAccountId: adminIds[i],
			Members:        mems,
		})
		if err != nil {
			t.Fatalf("remove members: %v", err)
		}
	}
}

func TestUpdatePersonId(t *testing.T) {
	ctx := context.Background()

	cloud_acct_id := strconv.Itoa(rand.Intn(99999999-10000000) + 10000000)
	Person_id := strconv.Itoa(rand.Intn(99999999-10000000) + 10000000)
	acct := &pb.MemberPersonId{
		PersonId:    Person_id,
		MemberEmail: "member@intel.com",
	}
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountMemberService.UpdatePersonId").WithValues("PersonId", acct.PersonId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("End")

	mbClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	oldPersonId := strconv.Itoa(rand.Intn(99999999-10000000) + 10000000)
	invite_code := strconv.Itoa(rand.Intn(9999999-1000000) + 1000000)
	insertQuery := "INSERT INTO members (person_id, admin_account_id, member_email, invitation_state, invitation_code, updated_at, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, m_err := db.ExecContext(ctx, insertQuery, oldPersonId, cloud_acct_id, acct.MemberEmail, 1, invite_code, time.Now(), time.Now())
	if m_err != nil {
		t.Fatalf("failed to insert test record into the database: %v", m_err)
	}

	_, err := mbClient.UpdatePersonId(ctx, acct)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	var UpdatedPersonId string
	ValidateQuery := "SELECT person_id from members WHERE admin_account_id = $1 and member_email = $2"
	v_err := db.QueryRow(ValidateQuery, cloud_acct_id, acct.MemberEmail).Scan(&UpdatedPersonId)

	if v_err != nil {
		t.Fatalf("Expected no error in database query, got: %v", err)
	}

	if UpdatedPersonId != acct.PersonId {
		t.Fatalf("Expected person_id to be %s, got: %s", acct.PersonId, UpdatedPersonId)
	}
	var mems []string = []string{"member@intel.com"}
	_, err = mbClient.RemoveMembers(context.Background(), &pb.CloudAccountMembers{
		CloudAccountId: cloud_acct_id,
		Members:        mems,
	})
	if err != nil {
		t.Fatalf("remove members: %v", err)
	}
}

func TestGetMemberPersonId(t *testing.T) {
	ctx := context.Background()
	cl_acct := strconv.Itoa(rand.Intn(99999999-10000000) + 10000000)
	accountUser := &pb.AccountUser{
		UserName:       "member2@intel.com",
		CloudAccountId: &cl_acct,
	}
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudAccountMemberService.GetMemberPersonId").WithValues("UserName", accountUser.UserName, "CloudAccountId", accountUser.CloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("End")

	mbClient := pb.NewCloudAccountMemberServiceClient(test.ClientConn())
	newPersonId := strconv.Itoa(rand.Intn(99999999-10000000) + 10000000)
	invite_code := strconv.Itoa(rand.Intn(9999999-1000000) + 1000000)
	insertQuery := "INSERT INTO members (person_id, admin_account_id, member_email, invitation_state, invitation_code, updated_at, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, m_err := db.ExecContext(ctx, insertQuery, newPersonId, accountUser.CloudAccountId, accountUser.UserName, 1, invite_code, time.Now(), time.Now())
	if m_err != nil {
		t.Fatalf("failed to insert test record into the database: %v", m_err)
	}
	res, err := mbClient.GetMemberPersonId(ctx, accountUser)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if res.PersonId != newPersonId {
		t.Fatalf("Expected person_id %s, not received successfully", newPersonId)
	}
	var mems []string = []string{"member2@intel.com"}
	_, d_err := mbClient.RemoveMembers(context.Background(), &pb.CloudAccountMembers{
		CloudAccountId: cl_acct,
		Members:        mems,
	})
	if d_err != nil {
		t.Fatalf("remove members: %v", err)
	}
}
