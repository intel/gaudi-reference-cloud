// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package builtin

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/open-policy-agent/opa/rego"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestAuthzCheck(t *testing.T) {
	err := os.Setenv("AUTHZ_ENABLED", "true")
	if err != nil {
		t.Fatalf("Error setting environment variable: %v", err)
	}
	// Fetch url of authz service.
	targetURL := authz.ClientConn().Target()
	authzBuiltIn := getAuthzBuiltIn(t, targetURL)
	// Create a custom Rego function
	customRegoFn := rego.FunctionDyn(
		checkAuthz,
		authzBuiltIn.authCheck,
	)

	// Invoke the custom rego function.
	var builder strings.Builder
	builder.WriteString(`authzService.check("admin@intel.com", "123456","*","/v1/authorization/resources","GET",{})`)
	fnString := builder.String()
	r := rego.New(rego.Query(fnString), customRegoFn)
	rs, err := r.Eval(context.Background())

	// Verification
	assert.NoError(t, err, "Error observed when evaluating authzService.check")
	verifyCheckRegoResult(t, &rs, true)
}

func TestAuthzCheckDeny(t *testing.T) {
	err := os.Setenv("AUTHZ_ENABLED", "true")
	if err != nil {
		t.Fatalf("Error setting environment variable: %v", err)
	}
	// Fetch url of authz service.
	targetURL := authz.ClientConn().Target()
	authzBuiltIn := getAuthzBuiltIn(t, targetURL)
	// Create a custom Rego function
	customRegoFn := rego.FunctionDyn(
		checkAuthz,
		authzBuiltIn.authCheck,
	)

	// Invoke the custom rego function.
	var builder strings.Builder
	builder.WriteString(`authzService.check("adminInvalid@intel.com", "123456","*","/v1/authorization/resources","GET",{})`)
	fnString := builder.String()
	r := rego.New(rego.Query(fnString), customRegoFn)
	rs, err := r.Eval(context.Background())

	// Verification
	assert.NoError(t, err, "Error observed when evaluating authzService.check")
	verifyCheckRegoResult(t, &rs, false)
}

func TestAuthzDisabled(t *testing.T) {
	err := os.Setenv("AUTHZ_ENABLED", "false")
	if err != nil {
		t.Fatalf("Error setting environment variable: %v", err)
	}
	// Fetch url of authz service.
	targetURL := authz.ClientConn().Target()
	authzBuiltIn := getAuthzBuiltIn(t, targetURL)
	// Create a custom Rego function
	customRegoFn := rego.FunctionDyn(
		checkAuthz,
		authzBuiltIn.authCheck,
	)

	// Invoke the custom rego function.
	var builder strings.Builder
	builder.WriteString(`authzService.check("adminInvalid@intel.com","1234567","*","/v1/authorization/resources","GET",{})`)
	fnString := builder.String()
	r := rego.New(rego.Query(fnString), customRegoFn)
	rs, err := r.Eval(context.Background())

	// Verification
	assert.NoError(t, err, "Error observed when evaluating authzService.check")
	verifyCheckRegoResult(t, &rs, true)
}

func TestAuthzCheckCollectionResource(t *testing.T) {
	err := os.Setenv("AUTHZ_ENABLED", "true")
	if err != nil {
		t.Fatalf("Error setting environment variable: %v", err)
	}

	// Fetch url of authz service.
	targetURL := authz.ClientConn().Target()
	authzBuiltIn := getAuthzBuiltIn(t, targetURL)

	client := pb.NewAuthzServiceClient(authz.ClientConn())

	_, err = client.AssignSystemRole(context.Background(), &pb.RoleRequest{
		CloudAccountId: "168310882044",
		Subject:        "cloud_account_member@ABC.com",
		SystemRole:     pb.SystemRole_cloud_account_member.String()})
	if err != nil {
		t.Fatalf(" error assign role: %v", err)
	}

	CloudAccountRoleCreate := pb.CloudAccountRole{
		Alias:          "AliasBuiltIn1",
		CloudAccountId: "168310882044",
		Effect:         pb.CloudAccountRole_allow,
		Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"list", "create"}}},
		Users:          []string{"cloud_account_member@ABC.com"},
	}

	_, err = client.CreateCloudAccountRole(context.Background(), &CloudAccountRoleCreate)
	if err != nil {
		t.Fatalf("create cloud account role account: %v", err)
	}

	// Create a custom Rego function
	customRegoFn := rego.FunctionDyn(
		checkAuthz,
		authzBuiltIn.authCheck,
	)

	// create token

	// Invoke the custom rego function.
	var builder strings.Builder
	builder.WriteString(`authzService.check("cloud_account_member@ABC.com", "99999999","168310882044","/v1/authorization/cloud_accounts/168310882044/roles","GET",{"name":"b"})`)
	fnString := builder.String()
	r := rego.New(rego.Query(fnString), customRegoFn)
	rs, err := r.Eval(context.Background())

	// Verification
	assert.NoError(t, err, "Error observed when evaluating authzService.check")
	verifyCheckRegoResult(t, &rs, true)
}

func TestAuthzCheckResourceIdWithCollectionAction(t *testing.T) {
	err := os.Setenv("AUTHZ_ENABLED", "true")
	if err != nil {
		t.Fatalf("Error setting environment variable: %v", err)
	}

	// Fetch url of authz service.
	targetURL := authz.ClientConn().Target()
	authzBuiltIn := getAuthzBuiltIn(t, targetURL)

	client := pb.NewAuthzServiceClient(authz.ClientConn())

	_, err = client.AssignSystemRole(context.Background(), &pb.RoleRequest{
		CloudAccountId: "168310883044",
		Subject:        "cloud_account_member@ABC.com",
		SystemRole:     pb.SystemRole_cloud_account_member.String()})
	if err != nil {
		t.Fatalf(" error assign role: %v", err)
	}

	CloudAccountRoleCreate := pb.CloudAccountRole{
		Alias:          "AliasBuiltIn3",
		CloudAccountId: "168310883044",
		Effect:         pb.CloudAccountRole_allow,
		Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"create"}}},
		Users:          []string{"cloud_account_member@ABC.com"},
	}

	_, err = client.CreateCloudAccountRole(context.Background(), &CloudAccountRoleCreate)
	if err != nil {
		t.Fatalf("create cloud account role account: %v", err)
	}

	// Create a custom Rego function
	customRegoFn := rego.FunctionDyn(
		checkAuthz,
		authzBuiltIn.authCheck,
	)

	// Invoke the custom rego function.
	var builder strings.Builder
	builder.WriteString(`authzService.check("cloud_account_member@ABC.com","99999999","168310883044","/cloud_account/168310883044/instance","POST",{"name":"b"})`)
	fnString := builder.String()
	r := rego.New(rego.Query(fnString), customRegoFn)
	rs, err := r.Eval(context.Background())

	// Verification
	assert.NoError(t, err, "Error observed when evaluating authzService.check")
	verifyCheckRegoResult(t, &rs, true)
}

func getAuthzBuiltIn(t *testing.T, authzUri string) *AuthzBuiltIn {
	conn, err := grpc.Dial(authzUri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	println("Invoke GRPC Dial to ", authzUri)
	assert.NoError(t, err, "Failed to create GRPC client")
	return NewAuthzBuiltIn(conn)
}

func verifyCheckRegoResult(t *testing.T, rs *rego.ResultSet, expected bool) {
	resMap := (*rs)[0].Expressions[0].Value
	assert.Equal(t, expected, resMap)
}
