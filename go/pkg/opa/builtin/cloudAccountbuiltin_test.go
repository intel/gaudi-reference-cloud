// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package builtin

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/open-policy-agent/opa/rego"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
)

// Setup CloudAccount service
func TestMain(m *testing.M) {
	ctx := context.Background()
	authz.EmbedService(ctx)
	cloudaccount.EmbedService(ctx)
	grpcutil.StartTestServices(ctx)
	defer grpcutil.StopTestServices()
	m.Run()
}

func TestCloudAccount(t *testing.T) {
	// Create a cloud account on Cloudaccount service
	acc1 := createCloudAccount(t)
	defer deleteCloudAccount(t, acc1.Id)
	// Fetch the cloud account id and url of cloud account service.
	targetURL := cloudaccount.ClientConn().Target()
	cloudAccountBuiltIn := getCloudAccountBuiltIn(t, targetURL)
	// Create a custom Rego function
	customRegoFn := rego.Function1(
		getCloudAccountByIdDecl,
		cloudAccountBuiltIn.getCloudAccountByIdImpl,
	)

	//Invoke the custom rego function.
	var builder strings.Builder
	builder.WriteString(`cloudaccount.getById("`)
	builder.WriteString(acc1.Id)
	builder.WriteString("\")")
	fnString := builder.String()
	r := rego.New(rego.Query(fnString), customRegoFn)
	rs, err := r.Eval(context.Background())

	// Verification
	assert.NoError(t, err, "Error observed when evaluating cloudaccount.getById")
	verifyRegoResult(t, &rs, acc1)

	//Toggle the flags in the cloud account service
	updateCloudAccount(t, acc1.Id)

	// Verify if the rego result reflects this change.
	rs, err = r.Eval(context.Background())
	assert.NoError(t, err, "Error observed when evaluating cloudaccount.getById")
	resMap := rs[0].Expressions[0].Value.(map[string]interface{})
	assert.Equal(t, true, resMap["enrolled"].(bool))
	assert.Equal(t, true, resMap["terminatePaidServices"].(bool))
}

func TestGetCloudAccountByName(t *testing.T) {
	// Create a cloud account on Cloudaccount service
	acc1 := createCloudAccount(t)
	defer deleteCloudAccount(t, acc1.Id)

	targetURL := cloudaccount.ClientConn().Target()
	cloudAccountBuiltIn := getCloudAccountBuiltIn(t, targetURL)
	// Create a custom Rego function
	customRegoFn := rego.Function1(
		getCloudAccountByNameDecl,
		cloudAccountBuiltIn.getCloudAccountByNameImpl,
	)

	//Invoke the custom rego function.
	var builder strings.Builder
	builder.WriteString(`cloudaccount.getByName("`)
	builder.WriteString("user@example.com")
	builder.WriteString("\")")
	fnString := builder.String()
	r := rego.New(rego.Query(fnString), customRegoFn)
	rs, err := r.Eval(context.Background())

	// Verification.
	assert.NoError(t, err, "Error observed when evaluating cloudaccount.getByName")
	verifyRegoResult(t, &rs, acc1)

	//Toggle the flags in the cloud account service
	updateCloudAccount(t, acc1.Id)

	// Verify if the rego result reflects this change.
	rs, err = r.Eval(context.Background())
	assert.NoError(t, err, "Error observed when evaluating cloudaccount.getByName")
	resMap := rs[0].Expressions[0].Value.(map[string]interface{})
	assert.Equal(t, true, resMap["enrolled"].(bool))
	assert.Equal(t, true, resMap["terminatePaidServices"].(bool))
}

func TestInvalidCloudAccountId(t *testing.T) {

	targetURL := cloudaccount.ClientConn().Target()
	cloudAccountBuiltIn := getCloudAccountBuiltIn(t, targetURL)

	// Create a custom Rego function to invoke getCloudAccountById
	customRegoFn := rego.Function1(
		getCloudAccountByIdDecl,
		cloudAccountBuiltIn.getCloudAccountByIdImpl,
	)
	//Invoke the custom rego function.
	var builder1 strings.Builder
	builder1.WriteString(`cloudaccount.getById("`)
	builder1.WriteString("12345678")
	builder1.WriteString("\")")
	fnString := builder1.String()
	r := rego.New(rego.Query(fnString), customRegoFn)
	rs, err := r.Eval(context.Background())

	// Verification
	assert.NoError(t, err, "Error observed when evaluating cloudaccount.getById")
	if len(rs) != 0 {
		t.Error("Expected a nil for an user that does not exist")
	}
}

func TestInvalidCloudAccountName(t *testing.T) {
	targetURL := cloudaccount.ClientConn().Target()
	cloudAccountBuiltIn := getCloudAccountBuiltIn(t, targetURL)
	// Create a custom Rego function
	customRegoFn := rego.Function1(
		getCloudAccountByNameDecl,
		cloudAccountBuiltIn.getCloudAccountByNameImpl,
	)

	//Invoke the custom rego function.
	var builder strings.Builder
	builder.WriteString(`cloudaccount.getByName("`)
	builder.WriteString("userInvalid@example.com")
	builder.WriteString("\")")
	fnString := builder.String()
	r := rego.New(rego.Query(fnString), customRegoFn)
	rs, err := r.Eval(context.Background())

	// Verification.
	assert.NoError(t, err, "Error observed when evaluating cloudaccount.getByName")
	if len(rs) != 0 {
		t.Error("Expected a nil for an user that does not exist")
	}
}

func getString(val any) string {
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

func getBool(val any) bool {
	if bb, ok := val.(bool); ok {
		return bb
	}
	return false
}

func getEnumNumber(t *testing.T, val any) protoreflect.EnumNumber {
	if ii, ok := val.(json.Number); ok {
		i64, err := ii.Int64()
		if err != nil {
			t.Fatal(err)
		}
		return protoreflect.EnumNumber(i64)
	}
	return 0
}

// Helper function to verify the resultSet
func verifyRegoResult(t *testing.T, rs *rego.ResultSet, acc *pb.CloudAccount) {
	resMap := (*rs)[0].Expressions[0].Value.(map[string]interface{})
	t.Logf("resMap=%v", resMap)
	assert.Equal(t, acc.Id, getString(resMap["id"]))
	assert.Equal(t, acc.ParentId, getString(resMap["parentId"]))
	assert.Equal(t, acc.Name, getString(resMap["name"]))
	assert.Equal(t, acc.Owner, getString(resMap["owner"]))
	assert.Equal(t, acc.Type.Number(), getEnumNumber(t, resMap["type"]))
	assert.Equal(t, acc.BillingAccountCreated, getBool(resMap["billingAccountCreated"]))
	assert.Equal(t, acc.Enrolled, getBool(resMap["enrolled"]))
	assert.Equal(t, acc.LowCredits, getBool(resMap["lowCredits"]))
	assert.Equal(t, acc.TerminatePaidServices, getBool(resMap["terminatePaidServices"]))
	assert.Equal(t, acc.TerminateMessageQueued, getBool(resMap["terminateMessageQueued"]))
}

// Helper function to toggle the flags in the cloud account.
func updateCloudAccount(t *testing.T, id string) {
	client := pb.NewCloudAccountServiceClient(cloudaccount.ClientConn())
	res, err := client.GetById(context.Background(), &pb.CloudAccountId{Id: id})
	if err != nil {
		assert.Fail(t, "Failed to fetch cloud account by id")
	}
	updatedTerminateFlag := !res.TerminatePaidServices
	updatedEnrolledFlag := !res.Enrolled

	_, err = client.Update(context.Background(), &pb.CloudAccountUpdate{Id: id,
		TerminatePaidServices: &updatedTerminateFlag,
		Enrolled:              &updatedEnrolledFlag})
	if err != nil {
		assert.Fail(t, "Failed to toggle flags. ")
	}
}

// Helper function to create a cloud account.
func createCloudAccount(t *testing.T) *pb.CloudAccount {
	acctCreate := pb.CloudAccountCreate{
		Name:  "user@example.com",
		Owner: "user@example.com",
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}

	client := pb.NewCloudAccountServiceClient(cloudaccount.ClientConn())
	_, err := client.Create(context.Background(), &acctCreate)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	account, err := client.GetByName(context.Background(), &pb.CloudAccountName{Name: acctCreate.Name})
	if err != nil {
		t.Fatalf("read account: %v", err)
	}
	return account
}

func deleteCloudAccount(t *testing.T, id string) {
	client := pb.NewCloudAccountServiceClient(cloudaccount.ClientConn())
	_, _ = client.Delete(context.Background(), &pb.CloudAccountId{
		Id: id,
	})
}

func getCloudAccountBuiltIn(t *testing.T, cloudAccountUri string) *CloudAccountBuiltIn {
	conn, err := grpc.Dial(cloudAccountUri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	println("Invoke GRPC Dial to ", cloudAccountUri)
	assert.NoError(t, err, "Failed to create GRPC client")
	return NewCloudAccountBuiltIn(conn)
}
