// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidationCloudAccountID(t *testing.T) {
	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	_, err := client.GetById(context.Background(), &pb.CloudAccountId{Id: ""})
	if err == nil {
		t.Fatalf("Error expected while getting Invalid Id")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("Expected Invalid Argument")
	}

	_, err = client.Delete(context.Background(), &pb.CloudAccountId{Id: ""})
	if err == nil {
		t.Fatalf("Error expected while getting Invalid Id")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("Expected Invalid Argument")
	}
}
func TestCreateOneAccount(t *testing.T) {
	oid := uuid.NewString()
	acctCreate := pb.CloudAccountCreate{
		Name:        "user@example.com",
		Owner:       "user@example.com",
		Tid:         uuid.NewString(),
		Oid:         oid,
		Type:        pb.AccountType_ACCOUNT_TYPE_STANDARD,
		PersonId:    "9849912",
		CountryCode: "US",
	}

	start := time.Now()

	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	id, err := client.Create(context.Background(), &acctCreate)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	end := time.Now()
	getByNameAndVerify(t, &acctCreate, id, start, end)
	getByOidAndVerify(t, &acctCreate, id, start, end)
	getByPersonIdAndVerify(t, &acctCreate, id, start, end)

	_, err = client.GetByOid(context.Background(), &pb.CloudAccountOid{Oid: oid})
	if err != nil {
		t.Errorf("Error while fetching by Oid")
	}
	_, err = client.Delete(context.Background(), &pb.CloudAccountId{
		Id: id.GetId(),
	})
	if err != nil {
		t.Errorf("error deleting account: %v", err)
	}

	_, err = client.GetByName(context.Background(), &pb.CloudAccountName{Name: acctCreate.GetName()})
	if err == nil {
		t.Errorf("account still exists after deletion")
	} else if status.Code(err) != codes.NotFound {
		t.Errorf("error reading account: %v", err)
	}

}

func TestCompare(t *testing.T) {
	flag := true
	country := "IN"
	ctx := context.Background()
	timeStamp := timestamppb.Now()

	acct := pb.CloudAccount{
		Id:                     "id",
		Name:                   "user@example.com",
		Owner:                  "user@example.com",
		Tid:                    "uid-01",
		Oid:                    "uid-02",
		Type:                   pb.AccountType_ACCOUNT_TYPE_STANDARD,
		ParentId:               "",
		BillingAccountCreated:  flag,
		Enrolled:               flag,
		LowCredits:             flag,
		CreditsDepleted:        timeStamp,
		TerminatePaidServices:  flag,
		TerminateMessageQueued: flag,
		PaidServicesAllowed:    flag,
		CountryCode:            country,
	}

	// No changes from the existing account data.
	acctCreateReq := pb.CloudAccountCreate{
		Name:                   "user@example.com",
		Owner:                  "user@example.com",
		Tid:                    "uid-01",
		Oid:                    "uid-02",
		Type:                   pb.AccountType_ACCOUNT_TYPE_STANDARD,
		ParentId:               "",
		BillingAccountCreated:  &flag,
		Enrolled:               &flag,
		LowCredits:             &flag,
		CreditsDepleted:        timeStamp,
		TerminatePaidServices:  &flag,
		TerminateMessageQueued: &flag,
		PaidServicesAllowed:    &flag,
		CountryCode:            country,
	}

	update := compare(ctx, &acct, &acctCreateReq)
	if update != nil {
		t.Errorf("Expected a nil value given that there are no changes")
	}

	// With updates
	flag = false
	country = "US"
	acctCreateReq = pb.CloudAccountCreate{
		Name:                   "user@example1.com",
		Owner:                  "user@example1.com",
		Tid:                    "uid-03",
		Oid:                    "uid-02",
		Type:                   pb.AccountType_ACCOUNT_TYPE_ENTERPRISE,
		ParentId:               "t1",
		BillingAccountCreated:  &flag,
		Enrolled:               &flag,
		LowCredits:             &flag,
		CreditsDepleted:        timeStamp,
		TerminatePaidServices:  &flag,
		TerminateMessageQueued: &flag,
		PaidServicesAllowed:    &flag,
		CountryCode:            country,
	}

	update = compare(ctx, &acct, &acctCreateReq)
	if update == nil {
		t.Error("Expected a non nil value given that there are no changes")
	}

	// With partial update
	flag = true
	acctCreateReq = pb.CloudAccountCreate{
		Name:  "user@example1.com",
		Owner: "user@example1.com",
		Oid:   "uid-02",
		Type:  pb.AccountType_ACCOUNT_TYPE_ENTERPRISE,
	}

	update = compare(ctx, &acct, &acctCreateReq)
	if update == nil {
		t.Error("Expected a non nil value given that there are no changes")
	}

	acctCreateReq = pb.CloudAccountCreate{
		Name: "user@example1.com",
		Oid:  "uid-02",
	}
	update = compare(ctx, &acct, &acctCreateReq)
	if update == nil {
		t.Error("Expected a non nil value given that there are no changes")
	}

}

func TestEnsureAPI(t *testing.T) {
	trueBool := true
	acctCreate := pb.CloudAccountCreate{
		Name:                  "user@example.com",
		Owner:                 "user@example.com",
		Tid:                   uuid.NewString(),
		Oid:                   uuid.NewString(),
		Type:                  pb.AccountType_ACCOUNT_TYPE_STANDARD,
		ParentId:              "",
		BillingAccountCreated: &trueBool,
		Enrolled:              &trueBool,
		LowCredits:            &trueBool,
	}
	start := time.Now()
	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	// Invoke Ensure API
	acct, err := client.Ensure(context.Background(), &acctCreate)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	end := time.Now()

	//Verify
	getByNameAndVerify(t, &acctCreate, &pb.CloudAccountId{Id: acct.Id}, start, end)
}

func getByNameAndVerify(t *testing.T, acctCreate *pb.CloudAccountCreate, id *pb.CloudAccountId,
	start time.Time, end time.Time) {
	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	acctOut, err := client.GetByName(context.Background(), &pb.CloudAccountName{Name: acctCreate.Name})
	if err != nil {
		t.Fatalf("read account: %v", err)
	}
	verifyCloudAccount(t, acctCreate, id, acctOut, start, end)
}

func getByOidAndVerify(t *testing.T, acctCreate *pb.CloudAccountCreate, id *pb.CloudAccountId,
	start time.Time, end time.Time) {
	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	acctOut, err := client.GetByOid(context.Background(), &pb.CloudAccountOid{Oid: acctCreate.Oid})
	if err != nil {
		t.Fatalf("read account: %v", err)
	}
	verifyCloudAccount(t, acctCreate, id, acctOut, start, end)
}

func getByPersonIdAndVerify(t *testing.T, acctCreate *pb.CloudAccountCreate, id *pb.CloudAccountId,
	start time.Time, end time.Time) {
	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	acctOut, err := client.GetByPersonId(context.Background(), &pb.CloudAccountPersonId{Personid: acctCreate.PersonId})
	if err != nil {
		t.Fatalf("read account: %v", err)
	}
	verifyCloudAccount(t, acctCreate, id, acctOut, start, end)
}

func verifyCloudAccount(t *testing.T, acctCreate *pb.CloudAccountCreate, id *pb.CloudAccountId, acctOut *pb.CloudAccount,
	start time.Time, end time.Time) {
	if acctOut.GetId() != id.GetId() {
		t.Errorf("output id %v doesn't match expected id %v", acctOut.GetId(),
			id.GetId())
	}
	if acctOut.GetName() != acctCreate.GetName() {
		t.Errorf("output name %v doesn't match input name %v", acctOut.GetName(),
			acctCreate.GetName())
	}
	if acctOut.GetOwner() != acctCreate.GetOwner() {
		t.Errorf("output owner %v doesn't match input owner %v", acctOut.GetOwner(),
			acctCreate.GetOwner())
	}
	created := acctOut.GetCreated().AsTime()

	if created.Before(start) || created.After(end) {
		t.Errorf("created timestamp %v out of range", created)
	}
	if acctOut.GetCountryCode() != acctCreate.GetCountryCode() {
		t.Errorf("output name %v doesn't match input name %v", acctOut.GetCountryCode(),
			acctCreate.GetCountryCode())
	}
	if acctCreate.PaidServicesAllowed == nil {
		if acctOut.PaidServicesAllowed == true {
			t.Error("Default value of PaidServicesAllowed should be false")
		}
	} else {
		if *acctCreate.PaidServicesAllowed != acctOut.PaidServicesAllowed {
			t.Error("Input value for PaidServicesAllowed is different from read value")
		}
	}
}

func TestCreateOneAccountUnspecifiedType(t *testing.T) {
	acctCreate := pb.CloudAccountCreate{
		Name:  "user@example.com",
		Owner: "user@example.com",
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_UNSPECIFIED,
	}

	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	_, err := client.Create(context.Background(), &acctCreate)
	if err == nil {
		// Invalid error expected.
		t.Fatalf("Error expected during cloud account create: %v", err)
	}
	returnStatus, ok := status.FromError(err)
	if !ok || returnStatus.Code() != codes.InvalidArgument {
		t.Fatalf("Invalid error returned, Invalid Argument expected")
	}
}

func TestIdCollisionDetect(t *testing.T) {
	saveMaxId := maxId
	maxId = 1
	defer func() { maxId = saveMaxId }()

	acctCreate := pb.CloudAccountCreate{
		Name:  "1@example.com",
		Owner: "1@example.com",
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}

	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	id, err := client.Create(context.Background(), &acctCreate)
	if err != nil {
		t.Fatal(err)
	}

	if id.GetId() != "000000000000" {
		t.Errorf("unexpected account id %v", id)
	}

	acctCreate.Name = "2@example.com"
	acctCreate.Owner = "2@example.com"
	acctCreate.Oid = uuid.NewString()

	_, err = client.Create(context.Background(), &acctCreate)
	if err == nil {
		t.Fatalf("client.Create should have failed due to id collision")
	}

	if !strings.Contains(err.Error(), kErrUniqueViolation) {
		t.Fatalf("unexpected error: %v", err)
	}

	maxId = saveMaxId
	if _, err = client.Create(context.Background(), &acctCreate); err != nil {
		t.Fatal(err)
	}
}

func TestIdCollision(t *testing.T) {
	if os.Getenv("CLOUDACCOUNT_TEST_ID_COLLISION") == "" {
		t.SkipNow()
	}
	ctx := context.Background()
	svc := CloudAccountService{}
	acctCreate := pb.CloudAccountCreate{
		Type: pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}
	var ii int
	for ii = 0; ; ii++ {
		user := fmt.Sprintf("user%v@example.com", ii)
		acctCreate.Name = user
		acctCreate.Owner = user
		acctCreate.Tid = uuid.NewString()
		acctCreate.Oid = uuid.NewString()
		if _, err := svc.tryCreate(ctx, &acctCreate); err != nil {
			pgErr := &pgconn.PgError{}
			if !errors.As(err, &pgErr) || pgErr.Code != kErrUniqueViolation {
				t.Fatal(err)
			}
			break
		}
	}
	t.Logf("failed after %v iterations", ii)
}

func TestIsValidId(t *testing.T) {
	id := MustNewId()
	if !IsValidId(id) {
		t.Errorf("id %v should be valid", id)
	}

	id = "123456789012"
	if !IsValidId(id) {
		t.Errorf("id %v should be valid", id)
	}

	id = "12345678901"
	if IsValidId(id) {
		t.Errorf("id %v is too short and should be invalid", id)
	}

	id = "1234567890123"
	if IsValidId(id) {
		t.Errorf("id %v is too long and should be invalid", id)
	}

	id = "12345a789012"
	if IsValidId(id) {
		t.Errorf("id %v contains non-digits and should be invalid", id)
	}
}

func TestAccountUniqueName(t *testing.T) {
	acctCreate := pb.CloudAccountCreate{
		Name:  "unique@example.com",
		Owner: "unique@example.com",
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
	}
	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	if _, err := client.Create(context.Background(), &acctCreate); err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err := client.Create(context.Background(), &acctCreate)
	if err == nil {
		t.Errorf("duplicate account should not be allowed by database constraints")
	}
	t.Logf("err from duplicate create: %v", err)
}

func testUpdateField(t *testing.T, id string, label string, client pb.CloudAccountServiceClient,
	newVal interface{}) {

	acct, err := client.GetById(context.Background(), &pb.CloudAccountId{Id: id})
	if err != nil {
		t.Errorf("%v getById: %v", label, err)
	}

	acctUpdate := pb.CloudAccountUpdate{Id: id}
	msgUpdate := acctUpdate.ProtoReflect()
	fdUpdate := msgUpdate.Descriptor().Fields().ByName(protoreflect.Name(label))
	msgUpdate.Set(fdUpdate, protoreflect.ValueOf(newVal))

	_, err = client.Update(context.Background(), &acctUpdate)
	if err != nil {
		t.Errorf("%v: update: %v", label, err)
		return
	}

	acctVerify, err := client.GetById(context.Background(), &pb.CloudAccountId{Id: id})
	if err != nil {
		t.Errorf("%v getById: %v", label, err)
	}

	msgVerify := acctVerify.ProtoReflect()
	msgVerifyFields := msgVerify.Descriptor().Fields()

	msg := acct.ProtoReflect()
	msgFields := msg.Descriptor().Fields()
	for ii := 0; ii < msgVerifyFields.Len(); ii++ {
		msgVerifyFd := msgVerifyFields.Get(ii)
		msgVerifyVal := msgVerify.Get(msgVerifyFd)
		verifyMsgIf := ifFromVal(msgVerifyFd, msgVerifyVal)

		if msgVerifyFd.Name() == protoreflect.Name(label) {
			upIf := newVal
			msg, ok := upIf.(protoreflect.Message)
			if ok {
				upIf = msg.Interface()
			}
			if !valsEqual(upIf, verifyMsgIf) {
				t.Errorf("%v: verify %v does not match update %v", label, msgVerifyVal, upIf)
			}
		} else {
			msgFd := msgFields.ByName(msgVerifyFd.Name())
			msgVal := msg.Get(msgFd)
			msgIf := ifFromVal(msgFd, msgVal)

			if !valsEqual(msgIf, verifyMsgIf) {
				t.Errorf("when setting %v: %v changed from %v to %v when it shouldn't have",
					label, msgVerifyFd.Name(), msgIf, verifyMsgIf)
			}
		}
	}
}

func TestAccountUpdate(t *testing.T) {
	acctCreate := pb.CloudAccountCreate{
		Name:  "update@example.com",
		Owner: "update@example.com",
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}
	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	_, err := client.Create(context.Background(), &acctCreate)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	acct, err := client.GetByName(context.Background(), &pb.CloudAccountName{Name: acctCreate.Name})
	if err != nil {
		t.Fatalf("read account: %v", err)
	}
	if acct.GetName() != acctCreate.GetName() ||
		acct.GetOwner() != acctCreate.GetOwner() ||
		acct.GetType() != acctCreate.GetType() {
		t.Errorf("account from API doesn't match")
	}

	testUpdateField(t, acct.GetId(), "parentId", client, MustNewId())
	testUpdateField(t, acct.GetId(), "parentId", client, "")
	testUpdateField(t, acct.GetId(), "owner", client, "newowner@example.com")
	testUpdateField(t, acct.GetId(), "type", client,
		protoreflect.EnumNumber(pb.AccountType_ACCOUNT_TYPE_PREMIUM))
	testUpdateField(t, acct.GetId(), "billingAccountCreated", client, true)
	testUpdateField(t, acct.GetId(), "billingAccountCreated", client, false)
	testUpdateField(t, acct.GetId(), "enrolled", client, true)
	testUpdateField(t, acct.GetId(), "enrolled", client, false)
	testUpdateField(t, acct.GetId(), "lowCredits", client, true)
	testUpdateField(t, acct.GetId(), "lowCredits", client, false)
	testUpdateField(t, acct.GetId(), "terminatePaidServices", client, true)
	testUpdateField(t, acct.GetId(), "terminatePaidServices", client, false)

	testUpdateField(t, acct.GetId(), "creditsDepleted", client, timestamppb.Now().ProtoReflect())

	testUpdateField(t, acct.GetId(), "terminateMessageQueued", client, true)
	testUpdateField(t, acct.GetId(), "terminateMessageQueued", client, false)
	testUpdateField(t, acct.GetId(), "paidServicesAllowed", client, true)
	testUpdateField(t, acct.GetId(), "paidServicesAllowed", client, false)
	testUpdateField(t, acct.GetId(), "countryCode", client, "IN")
}

func doSearch(client pb.CloudAccountServiceClient, filter *pb.CloudAccountFilter) ([]*pb.CloudAccount, error) {

	res, err := client.Search(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	results := []*pb.CloudAccount{}
	for {
		acct, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		results = append(results, acct)
	}
	return results, nil
}

func filterDiff(filter protoreflect.Message, obj protoreflect.Message,
	cb func(fieldName string, val any) bool) {
	filterFields := filter.Descriptor().Fields()
	objFields := obj.Descriptor().Fields()
	for ii := 0; ii < filterFields.Len(); ii++ {
		filterFd := filterFields.Get(ii)
		if filter.Has(filterFd) {
			filterVal := filter.Get(filterFd)
			objFd := objFields.ByName(filterFd.Name())
			objVal := obj.Get(objFd)
			objIf := ifFromVal(objFd, objVal)
			if !valsEqual(ifFromVal(filterFd, filterVal), objIf) {
				if !cb(string(filterFd.Name()), objIf) {
					return
				}
			}
		}
	}
}

func searchAndCheck(t *testing.T, client pb.CloudAccountServiceClient, filter *pb.CloudAccountFilter,
	accts []pb.CloudAccountCreate) {
	results, err := doSearch(client, filter)
	if err != nil {
		t.Fatalf("search error: %v", err)
	}

	msgFilter := filter.ProtoReflect()
	ids := map[string]bool{}
	for _, acct := range results {
		if ids[acct.GetId()] {
			t.Errorf("duplicate result for account %v", acct.GetId())
		}
		ids[acct.GetId()] = true
		filterDiff(msgFilter, acct.ProtoReflect(),
			func(fieldName string, val any) bool {
				t.Errorf("wrong %v %v for account %v", fieldName, val, acct.GetName())
				return true
			})
	}
	expectedMatches := len(accts)
	for ii := range accts {
		acct := &accts[ii]
		filterDiff(msgFilter, acct.ProtoReflect(),
			func(fieldName string, val any) bool {
				expectedMatches--
				return false
			})
	}
	if expectedMatches != len(results) {
		t.Errorf("expected %v matches, got %v matches", expectedMatches, len(results))
	}
}

func TestAccountSearch(t *testing.T) {
	tt := true
	accts := []pb.CloudAccountCreate{
		{Name: "std1@example.com", Owner: "std1@example.com", Tid: uuid.NewString(), Oid: uuid.NewString(), Type: pb.AccountType_ACCOUNT_TYPE_STANDARD},
		{Name: "std2@example.com", Owner: "std2@example.com", Tid: uuid.NewString(), Oid: uuid.NewString(), Type: pb.AccountType_ACCOUNT_TYPE_STANDARD},
		{Name: "prem1@example.com", Owner: "prem1@example.com", Tid: uuid.NewString(), Oid: uuid.NewString(), Type: pb.AccountType_ACCOUNT_TYPE_PREMIUM},
		{Name: "prem2@example.com", Owner: "prem2@example.com", Tid: uuid.NewString(), Oid: uuid.NewString(), Type: pb.AccountType_ACCOUNT_TYPE_PREMIUM},
		{Name: "enrolled1@example.com", Owner: "enrolled1@example.com", Tid: uuid.NewString(), Oid: uuid.NewString(), Type: pb.AccountType_ACCOUNT_TYPE_PREMIUM, Enrolled: &tt},
		{Name: "enrolled2@example.com", Owner: "enrolled2@example.com", Tid: uuid.NewString(), Oid: uuid.NewString(), Type: pb.AccountType_ACCOUNT_TYPE_PREMIUM, Enrolled: &tt},
	}

	if _, err := db.Exec("DELETE FROM cloud_accounts"); err != nil {
		t.Fatalf("error removing cloud accounts: %v", err)
	}

	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	for ii := range accts {
		_, err := client.Create(context.Background(), &accts[ii])
		if err != nil {
			t.Fatalf("create account failed: %v", err)
		}
	}

	acctType := pb.AccountType_ACCOUNT_TYPE_STANDARD
	searchAndCheck(t, client, &pb.CloudAccountFilter{Type: &acctType}, accts)

	acctType = pb.AccountType_ACCOUNT_TYPE_PREMIUM
	searchAndCheck(t, client, &pb.CloudAccountFilter{Type: &acctType}, accts)

	enrolled := true
	searchAndCheck(t, client, &pb.CloudAccountFilter{Enrolled: &enrolled}, accts)
	// check multiple filters
	searchAndCheck(t, client, &pb.CloudAccountFilter{Type: &acctType, Enrolled: &enrolled}, accts)
}

func checkParent(t *testing.T, client pb.CloudAccountServiceClient,
	id string, name string, parentId string) {

	acctById, err := client.GetById(context.Background(),
		&pb.CloudAccountId{Id: id})
	if err != nil {
		t.Fatalf("GetById: %v", err)
	}
	if acctById.ParentId != parentId {
		t.Errorf("GetById: expecting parent %v, got %v", parentId, acctById.ParentId)
	}

	acctByName, err := client.GetByName(context.Background(), &pb.CloudAccountName{Name: name})
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if acctByName.ParentId != parentId {
		t.Errorf("GetByName: expecting parent %v, got %v", parentId, acctByName.ParentId)
	}

	res, err := client.Search(context.Background(), &pb.CloudAccountFilter{Id: &id})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	acctFound, err := res.Recv()
	if err != nil {
		t.Fatalf("Recv: %v", err)
	}
	if acctFound.ParentId != parentId {
		t.Errorf("Search: expecting parent %v, got %v", parentId, acctFound.ParentId)
	}
}

func TestParentId(t *testing.T) {
	acctChild := pb.CloudAccountCreate{
		Name:  "child@example.com",
		Owner: "child@example.com",
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}

	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	id, err := client.Create(context.Background(), &acctChild)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	checkParent(t, client, id.Id, acctChild.Name, "")

	acctParent := pb.CloudAccountCreate{
		Name:  "parent@example.com",
		Owner: "parent@example.com",
		Tid:   acctChild.Tid,
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}
	parentId, err := client.Create(context.Background(), &acctParent)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	_, err = client.Update(context.Background(),
		&pb.CloudAccountUpdate{Id: id.Id, ParentId: &parentId.Id})
	if err != nil {
		t.Fatalf("update child: %v", err)
	}
	checkParent(t, client, id.Id, acctChild.Name, parentId.Id)

	noParent := ""
	_, err = client.Update(context.Background(),
		&pb.CloudAccountUpdate{Id: id.Id, ParentId: &noParent})
	if err != nil {
		t.Fatalf("update child: %v", err)
	}
	checkParent(t, client, id.Id, acctChild.Name, "")
}

func TestAuthzSystemRoleAdminAssigned(t *testing.T) {
	// Given CloudAccountCreate
	oid := uuid.NewString()
	acctCreate := pb.CloudAccountCreate{
		Name:        "authzuser@example.com",
		Owner:       "authzuser@example.com",
		Tid:         uuid.NewString(),
		Oid:         oid,
		Type:        pb.AccountType_ACCOUNT_TYPE_STANDARD,
		PersonId:    "9849912",
		CountryCode: "US",
	}

	// When creating cloud account  authz systemrole should be assign for cloud_account_admin
	client := pb.NewCloudAccountServiceClient(test.ClientConn())
	id, err := client.Create(context.Background(), &acctCreate)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	authzlient := pb.NewAuthzServiceClient(test.clientConnAuthz)
	existsResponse, _ := authzlient.SystemRoleExists(context.Background(), &pb.RoleRequest{CloudAccountId: id.Id, Subject: "authzuser@example.com", SystemRole: "cloud_account_admin"})

	if test.cfg.Authz.Enabled {
		// Then systemRole should exist
		if !existsResponse.Exist {
			t.Fatalf("error system role should exist")
		}
	} else {
		// Then systemRole should not exist
		if existsResponse.Exist {
			t.Fatalf("error system role should not exist")
		}
	}

}
