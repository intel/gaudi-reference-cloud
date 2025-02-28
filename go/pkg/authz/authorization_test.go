// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	psqlwatcher "github.com/IguteChung/casbin-psql-watcher"
	"github.com/casbin/casbin/v2"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type CheckTest struct {
	subject        string
	cloudAccountId string
	path           string
	payload        any
	verb           string
	expected       bool
}

func TestCheck(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "468310882074", Subject: "alex.jordan@examplecorp.com", SystemRole: "cloud_account_admin"},
		{CloudAccountId: "468310882074", Subject: "jose.lopez@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "468310882074", Subject: "jose@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "468310882075", Subject: "jose.lopez@examplecorp.com", SystemRole: "cloud_account_admin"},
	}

	// creating permission to allow
	cloudAccountRoles := []*pb.CloudAccountRole{
		{Alias: "Alias1", CloudAccountId: "468310882074", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"jose.lopez@examplecorp.com"}}}

	checks := []CheckTest{
		{
			cloudAccountId: "468310882074",
			subject:        "jose.lopez@examplecorp.com",
			path:           "/cloud_account/468310882074/instance?t=1723588117721",
			verb:           "POST",
			payload: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"instancename": structpb.NewStringValue("othervm"),
				},
			},
			expected: false,
		},
		{
			cloudAccountId: "468310882074",
			subject:        "jose.lopez@examplecorp.com",
			path:           "/v1/authorization/cloudaccounts/468310882074/users/jose.lopez@examplecorp.com?t=1723588117721",
			verb:           "GET",
			payload:        &structpb.Struct{},
			expected:       true,
		},
		{
			cloudAccountId: "468310882074",
			subject:        "jose@examplecorp.com",
			path:           "/v1/authorization/cloudaccounts/468310882074/users/jose.lopez@examplecorp.com?t=1723588117721",
			verb:           "GET",
			payload:        &structpb.Struct{},
			expected:       false,
		},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// when
	for _, check := range checks {
		allowed, err := client.CheckInternal(context.Background(), &pb.AuthorizationRequestInternal{
			CloudAccountId: check.cloudAccountId,
			Path:           check.path,
			Verb:           check.verb,
			Payload:        check.payload.(*structpb.Struct),
			User:           &pb.UserIdentification{Email: check.subject, EnterpriseId: "eid"},
		})

		// then
		if err != nil {
			t.Fatalf("check authz call failed (%v)", err)
		}
		if allowed.Allowed != check.expected {
			t.Fatalf("check call from '%v' to '%v' with verb '%v' failed, got (%v) but expected was (%v)",
				check.subject, check.path, check.verb, allowed.Allowed, check.expected)
		}
	}
}

func TestValidationAssignSystemRolesAllows(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "*", Subject: "john.doe@intel.com", SystemRole: "intel_admin"},
		{CloudAccountId: "*", Subject: "jane.smith@intel.com", SystemRole: "intel_admin"},
		{CloudAccountId: "168310882074", Subject: "alex.jordan@examplecorp.com", SystemRole: "cloud_account_admin"},
		{CloudAccountId: "168310882074", Subject: "jose.lopez@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	// creating permission to allow
	cloudAccountRoles := []*pb.CloudAccountRole{
		{Alias: "Alias1", CloudAccountId: "168310882074", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"jose.lopez@examplecorp.com"}}}

	checks := []CheckTest{
		{
			cloudAccountId: "B",
			subject:        "john.doe@intel.com",
			path:           "/cloud_account/B/instance?t=1723588117721",
			verb:           "POST",
			payload: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"instancename": structpb.NewStringValue("othervm"),
				},
			},
			expected: true,
		},
		{
			cloudAccountId: "C",
			subject:        "jane.smith@intel.com",
			path:           "/cloud_account/C/instance?metadata.filterType=ComputeGeneral&t=1723584418687",
			verb:           "POST",
			payload: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"instancename": structpb.NewStringValue("othervm"),
				},
			},
			expected: true,
		},
		{
			cloudAccountId: "168310882074",
			subject:        "alex.jordan@examplecorp.com",
			path:           "/cloud_account/168310882074/instance?metadata.filterType=ComputeGeneral",
			verb:           "POST",
			payload: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"instancename": structpb.NewStringValue("othervm"),
				},
			},
			expected: true,
		},
		{
			cloudAccountId: "168310882074",
			subject:        "jose.lopez@examplecorp.com",
			path:           "/cloud_account/168310882074/instance",
			verb:           "GET",
			payload: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"instancename": structpb.NewStringValue("othervm"),
				},
			},
			expected: true,
		},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// when
	for _, check := range checks {
		allowed, err := client.CheckInternal(context.Background(), &pb.AuthorizationRequestInternal{
			CloudAccountId: check.cloudAccountId,
			Path:           check.path,
			Verb:           check.verb,
			Payload:        check.payload.(*structpb.Struct),
			User:           &pb.UserIdentification{Email: check.subject, EnterpriseId: "eid"},
		})

		// then
		if err != nil {
			t.Fatalf("check authz call failed (%v)", err)
		}
		if allowed.Allowed != check.expected {
			t.Fatalf("check call from '%v' to '%v' with verb '%v' failed, got (%v) but expected was (%v)",
				check.subject, check.path, check.verb, allowed.Allowed, check.expected)
		}
	}
}

func TestAssignSystemRoleWithWildcardForNonAdmin(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	//Given
	roleRequest := &pb.RoleRequest{
		CloudAccountId: "*",
		Subject:        "non.admin@example.com",
		SystemRole:     "cloud_account_member",
	}

	//When assigning system role
	_, err := client.AssignSystemRole(context.Background(), roleRequest)

	if err == nil {
		t.Fatalf("expected error when assigning non-admin role with wildcard cloud account id, but got none")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("expected 'InvalidArgument' error, but got: %v", err)
	}

	//Then expect error to happen
	expectedErrMsg := "cannot assign non-admin role with wildcard cloud account id"
	if !strings.Contains(st.Message(), expectedErrMsg) {
		t.Fatalf("expected error message to contain '%s', but got: %s", expectedErrMsg, st.Message())
	}
}

func TestAssignSystemRoleWithWildcardForCloudAccountAdmin(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	//Given
	roleRequest := &pb.RoleRequest{
		CloudAccountId: "*",
		Subject:        "cloud.admin@example.com",
		SystemRole:     "cloud_account_admin",
	}

	//When assigning system role
	_, err := client.AssignSystemRole(context.Background(), roleRequest)

	if err == nil {
		t.Fatalf("expected error when assigning non-admin role with wildcard cloud account id, but got none")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("expected 'InvalidArgument' error, but got: %v", err)
	}

	//Then expect error to happen
	expectedErrMsg := "cannot assign non-admin role with wildcard cloud account id"
	if !strings.Contains(st.Message(), expectedErrMsg) {
		t.Fatalf("expected error message to contain '%s', but got: %s", expectedErrMsg, st.Message())
	}
}

func TestAssignSystemRoleForCloudAccountWithExistingUser(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	//Given
	roleRequest := &pb.RoleRequest{
		CloudAccountId: "251291230984",
		Subject:        "cloud.admin@example.com",
		SystemRole:     "cloud_account_admin",
	}

	//When assigning system role
	_, err := client.AssignSystemRole(context.Background(), roleRequest)
	if err != nil {
		t.Fatalf("failed error while assign role")
	}
	roleRequest.SystemRole = "cloud_account_member"
	//When assigning system role
	_, err = client.AssignSystemRole(context.Background(), roleRequest)
	if err == nil {
		t.Fatalf("expected error while assign role to user already assign to the cloud account")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.AlreadyExists {
		t.Fatalf("expected 'AlreadyExists' error, but got: %v", err)
	}

	//Then expect error to happen
	expectedErrMsg := "provided user already has a role assigned in the given cloud account"
	if !strings.Contains(st.Message(), expectedErrMsg) {
		t.Fatalf("expected error message to contain '%s', but got: %s", expectedErrMsg, st.Message())
	}

	roleRequest.SystemRole = "cloud_account_admin"
	//When assigning system role
	_, err = client.AssignSystemRole(context.Background(), roleRequest)
	if err == nil {
		t.Fatalf("expected error while assign role to user already assign to the cloud account")
	}

	st, ok = status.FromError(err)
	if !ok || st.Code() != codes.AlreadyExists {
		t.Fatalf("expected 'AlreadyExists' error, but got: %v", err)
	}

	//Then expect error to happen
	if !strings.Contains(st.Message(), expectedErrMsg) {
		t.Fatalf("expected error message to contain '%s', but got: %s", expectedErrMsg, st.Message())
	}

}

func TestAssignInvalidSystemRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	//Given
	roleRequest := &pb.RoleRequest{
		CloudAccountId: "123",
		Subject:        "cloud.admin@example.com",
		SystemRole:     "cloud_account_invalid",
	}

	//When assigning system role
	_, err := client.AssignSystemRole(context.Background(), roleRequest)

	if err == nil {
		t.Fatalf("expected error when assigning invalid systemrole")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("expected 'InvalidArgument' error, but got: %v", err)
	}

	//Then expect error to happen
	expectedErrMsg := "provided systemrole is not valid"
	if !strings.Contains(st.Message(), expectedErrMsg) {
		t.Fatalf("expected error message to contain '%s', but got: %s", expectedErrMsg, st.Message())
	}
}

func TestValidationAssignSystemRolesDenies(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "251296150984", Subject: "james.thompson@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	// Deny CloudAccountRole will overide Allow CloudAccountRole
	cloudAccountRoles := []*pb.CloudAccountRole{
		{Alias: "Alias3", CloudAccountId: "251296150984", Effect: pb.CloudAccountRole_deny, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"james.thompson@examplecorp.com"}},
		{Alias: "Alias2", CloudAccountId: "251296150984", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"james.thompson@examplecorp.com"}},
	}

	checks := []CheckTest{
		{
			cloudAccountId: "251296150984",
			subject:        "james.thompson@examplecorp.com",
			path:           "/cloud_account/251296150984/instance",
			verb:           "GET",
			payload: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"instancename": structpb.NewStringValue("othervm"),
				},
			},
			expected: false,
		},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	for _, check := range checks {
		// when
		allowed, err := client.CheckInternal(context.Background(), &pb.AuthorizationRequestInternal{
			CloudAccountId: check.cloudAccountId,
			Path:           check.path,
			Verb:           check.verb,
			Payload:        check.payload.(*structpb.Struct),
			User:           &pb.UserIdentification{Email: check.subject, EnterpriseId: "eid"},
		})

		// then
		if err != nil {
			t.Fatalf("check authz call failed (%v)", err)
		}
		if allowed.Allowed != check.expected {
			t.Fatalf("check call from '%v' to '%v' with verb '%v' failed, got (%v) but expected was (%v)",
				check.subject, check.path, check.verb, allowed.Allowed, check.expected)
		}
	}
}

func TestAuthorizationService_Check_ValidationError(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given an invalid AuthorizationRequest with empty fields that should fail validation
	invalidAuthRequest := &pb.AuthorizationRequestInternal{
		CloudAccountId: "",
		Path:           "",
		Verb:           "",
		Payload:        nil,
		User:           &pb.UserIdentification{Email: "", EnterpriseId: ""},
	}

	ctx := context.Background()

	// When calling the Check function with an invalid request
	_, err := client.CheckInternal(ctx, invalidAuthRequest)

	// Then expect a validation error
	if err == nil {
		t.Fatalf("expected validation error when calling Check with invalid request parameters, but got none")
	}

	// Check for the specific error code and message
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("expected 'InvalidArgument' error code, but got: %v", err)
	}
	if !strings.Contains(st.Message(), "invalid request parameters") {
		t.Fatalf("expected error message to contain 'invalid request parameters', but got: %s", st.Message())
	}
}

func TestNoPermissionShouldDeny(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "038243323293", Subject: "chris.wilson@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	// CloudAccountRoles (Permissions not created)

	checks := []CheckTest{
		{
			cloudAccountId: "038243323293",
			subject:        "chris.wilson@examplecorp.com",
			path:           "/cloud_account/038243323293/instance",
			verb:           "GET",
			payload: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"instancename": structpb.NewStringValue("othervm"),
				},
			},
			expected: false, // there isn't a permission allow it
		},
	}

	err := assignSystemRolesBulk(client, systemRoles)

	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// When - check
	for _, check := range checks {
		allowed, err := client.CheckInternal(context.Background(), &pb.AuthorizationRequestInternal{
			CloudAccountId: check.cloudAccountId,
			Path:           check.path,
			Verb:           check.verb,
			Payload:        check.payload.(*structpb.Struct),
			User:           &pb.UserIdentification{Email: check.subject, EnterpriseId: "eid"},
		})

		// then
		if err != nil {
			t.Fatalf("check authz call failed (%v)", err)
		}
		if allowed.Allowed != check.expected {
			t.Fatalf("check call from '%v' to '%v' with verb '%v' failed, got (%v) but expected was (%v)",
				check.subject, check.path, check.verb, allowed.Allowed, check.expected)
		}
	}
}

func TestCreateCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "720728960768", Subject: "emily.thompson@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given specified cloudAccountRole, When creating the cloudAccountRole
	cloudAccountRole, err := client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "Alias0", CloudAccountId: "720728960768", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"emily.thompson@examplecorp.com"}})
	// Then, expect to not return error
	if err != nil {
		t.Fatalf("error creating cloudAccountRole  (%v)", err)
	}
	// Then expect to have a new id set
	if cloudAccountRole.Id == "" {
		t.Fatalf("error id should have been set to cloudAccountRole when created (%v)", err)
	}

	// Given specified cloudAccountRole with invalid resourceType, When creating the cloudAccountRole
	_, err = client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "Alias01", CloudAccountId: "720728760768", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instanceInvalid", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"tom.thompson@examplecorp.com"}})
	// Then, expect to return error for invalid resourceType
	if err == nil {
		t.Fatalf("should throw error provided resource type is invalid")
	}

	// Given specified cloudAccountRole with invalid action, When creating the cloudAccountRole
	_, err = client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "Alias011", CloudAccountId: "720828760768", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"InvalidAction"}}}, Users: []string{"tom.thompson@examplecorp.com"}})
	// Then, expect to return error for invalid action
	if err == nil {
		t.Fatalf("should throw error provided action is invalid for the resource")
	}

	// Given specified cloudAccountRole, When creating the cloudAccountRole
	_, err = client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "Alias$!&~()", CloudAccountId: "720728960768", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"emily.thompson@examplecorp.com"}})
	// Then, expect to return error since alias do not match regex
	if !strings.Contains(err.Error(), "InvalidArgument") {
		t.Fatalf("error should be returned since alias doesn't meet input requirements")
	}

	// Given specified cloudAccountRole, When creating the cloudAccountRole
	_, err = client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "Alias23455", CloudAccountId: "720728960768", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "<asf?$", Actions: []string{"read"}}}, Users: []string{"emily.thompson@examplecorp.com"}})
	// Then, expect to return error since resourceId do not match regex
	if !strings.Contains(err.Error(), "InvalidArgument") {
		t.Fatalf("error should be returned since resourceId doesn't meet input requirements")
	}
}

func TestCreateCloudAccountRoleLimits(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "990728960768", Subject: "emily.thompson@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// 100 defined in the testing.go config setup
	for i := 0; i < 100; i++ {
		_, err := client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: fmt.Sprintf("Alias%d", i), CloudAccountId: "990728960768", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"emily.thompson@examplecorp.com"}})
		if err != nil {
			t.Fatalf("error creating cloudAccountRole  (%v)", err)
		}
	}

	_, err = client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "Alias99999", CloudAccountId: "990728960768", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"emily.thompson@examplecorp.com"}})

	if err == nil {
		t.Fatalf("error should be limit reached  (%v)", err)
	}

}

func TestCreateCloudAccountRoleUserActionEmpty(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "720728960722", Subject: "user1@test.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "720728960722", Subject: "user2@test.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given specified cloudAccountRole, When creating the cloudAccountRole with empty users and actions
	cloudAccountRole, err := client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "AliasUserEmptyCloudAccountRole", CloudAccountId: "720728960722", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{}})
	// Then, expect to not return error
	if err != nil {
		t.Fatalf("error creating cloudAccountRole  (%v)", err)
	}
	// Then expect to have a new id set
	if cloudAccountRole.Id == "" {
		t.Fatalf("error id should have been set to cloudAccountRole when created (%v)", err)
	}

	// add a new user1 to the cloudAccountRole
	_, err = client.AddUserToCloudAccountRole(context.Background(), &pb.CloudAccountRoleUserRequest{CloudAccountId: "720728960722", Id: cloudAccountRole.Id, UserId: "user1@test.com"})
	if err != nil {
		t.Fatalf("error adding user1 to cloudAccountRole  (%v)", err)
	}

	// when getting the cloudAccountRole user1 should exist
	cloudAccountRoleUser1, _ := client.GetCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{CloudAccountId: "720728960722", Id: cloudAccountRole.Id})
	if len(cloudAccountRoleUser1.Users) == 0 && cloudAccountRoleUser1.Users[0] != "user1@test.com" {
		t.Fatalf("added user1 not in cloudAccountRole")
	}

	// add a new user to the cloudAccountRole user2
	_, err = client.AddUserToCloudAccountRole(context.Background(), &pb.CloudAccountRoleUserRequest{CloudAccountId: "720728960722", Id: cloudAccountRole.Id, UserId: "user2@test.com"})
	if err != nil {
		t.Fatalf("error adding user2 to cloudAccountRole  (%v)", err)
	}

	// when getting the cloudAccountRole user should exist
	cloudAccountRoleUser2, _ := client.GetCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{CloudAccountId: "720728960722", Id: cloudAccountRole.Id})
	if len(cloudAccountRoleUser2.Users) != 2 && cloudAccountRoleUser2.Users[1] != "user2@test.com" {
		t.Fatalf("added user2 not in cloudAccountRole")
	}
}

func TestGetCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given, creation of specific cloudAccountRole
	cloudAccountRoleCreated, err := client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "Alias01", CloudAccountId: "720728960768", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{}})
	if err != nil {
		t.Fatalf("error creating cloudAccountRole  (%v)", err)
	}

	// When, GetCloudAccountRole
	cloudAccountRole, err := client.GetCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{CloudAccountId: "720728960768", Id: cloudAccountRoleCreated.Id})
	if err != nil {
		t.Fatalf("error getting cloudAccountRole  (%v)", err)
	}

	// Then, the retuned cloudAccountRole should match with the specified
	if cloudAccountRole.Id != cloudAccountRoleCreated.Id {
		t.Fatalf("error id for created and consulted cloudAccount role should match  (%v)", err)
	}

}

func TestUpdateCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "720728960767", Subject: "emily.thompson@examplecorp2.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "720728960767", Subject: "chris.thompson@examplecorp2.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)

	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	cloudAccountRoleCreated, err := client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{
		Alias:          "Alias001",
		CloudAccountId: "720728960767",
		Effect:         pb.CloudAccountRole_allow,
		Permissions: []*pb.CloudAccountRole_Permission{
			{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}},
			{ResourceType: "instance", ResourceId: "will-be-deleted", Actions: []string{"delete"}},
			{ResourceType: "instance", ResourceId: "will-be-deleted-2", Actions: []string{"delete"}},
		},
		Users: []string{"emily.thompson@examplecorp2.com"},
	})

	if len(cloudAccountRoleCreated.Permissions) != 3 {
		t.Fatalf("error cloudAccountRoleCreated permission now should be 3")
	}

	if err != nil {
		t.Fatalf("error creating cloudAccountRole  (%v)", err)
	}

	_, updateErr := client.UpdateCloudAccountRole(context.Background(), &pb.CloudAccountRoleUpdate{
		Id:             cloudAccountRoleCreated.Id,
		Alias:          "Alias02",
		CloudAccountId: "720728960767",
		Effect:         pb.CloudAccountRoleUpdate_deny,
		Users:          []string{"emily.thompson@examplecorp2.com", "chris.thompson@examplecorp2.com"},
		Permissions: []*pb.CloudAccountRoleUpdate_Permission{
			{
				Id:           &cloudAccountRoleCreated.Permissions[0].Id,
				ResourceType: "instance",
				ResourceId:   "*",
				Actions:      []string{"read"},
			},
			{
				ResourceType: "instance", // should create this permission
				ResourceId:   "2",
				Actions:      []string{"delete"},
			},
		},
	})

	if updateErr != nil {
		t.Fatalf("error updating cloudAccountRole  (%v)", updateErr)
	}

	cloudAccountRoleUpdated, getErr := client.GetCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{CloudAccountId: "720728960767", Id: cloudAccountRoleCreated.Id})
	if getErr != nil {
		t.Fatalf("error getting cloudAccountRole  (%v)", getErr)
	}

	if cloudAccountRoleUpdated.CloudAccountId != cloudAccountRoleCreated.CloudAccountId {
		t.Fatalf("error cloud account id for update and consulted cloudAccount role should match")
	}

	if cloudAccountRoleUpdated.Alias != "Alias02" {
		t.Fatalf("error alias was not updated")
	}

	if len(cloudAccountRoleUpdated.Users) != 2 {
		t.Fatalf("error users now should be 2")
	}

	if len(cloudAccountRoleUpdated.Permissions) != 2 {
		t.Fatalf("error permissions now should be 2")
	}

	if cloudAccountRoleUpdated.Effect.String() != pb.CloudAccountRole_deny.String() {
		t.Fatalf("error effect now should be deny")
	}

	_, updateErrBadAlias := client.UpdateCloudAccountRole(context.Background(), &pb.CloudAccountRoleUpdate{
		Id:             cloudAccountRoleCreated.Id,
		Alias:          "Alias$!&~()",
		CloudAccountId: "720728960767",
		Effect:         pb.CloudAccountRoleUpdate_deny,
		Users:          []string{"emily.thompson@examplecorp2.com", "chris.thompson@examplecorp2.com"},
		Permissions: []*pb.CloudAccountRoleUpdate_Permission{
			{
				Id:           &cloudAccountRoleCreated.Permissions[0].Id,
				ResourceType: "instance",
				ResourceId:   "*",
				Actions:      []string{"read"},
			},
			{
				ResourceType: "instance", // should create this permission
				ResourceId:   "2",
				Actions:      []string{"delete"},
			},
		},
	})

	// Then, expect to return error since alias do not match regex
	if !strings.Contains(updateErrBadAlias.Error(), "InvalidArgument ") {
		t.Fatalf("error should be returned since alias doesn't meet input requirements")
	}

	_, updateErrBadResourceId := client.UpdateCloudAccountRole(context.Background(), &pb.CloudAccountRoleUpdate{
		Id:             cloudAccountRoleCreated.Id,
		Alias:          "Alias00343",
		CloudAccountId: "720728960767",
		Effect:         pb.CloudAccountRoleUpdate_deny,
		Users:          []string{"emily.thompson@examplecorp2.com", "chris.thompson@examplecorp2.com"},
		Permissions: []*pb.CloudAccountRoleUpdate_Permission{
			{
				Id:           &cloudAccountRoleCreated.Permissions[0].Id,
				ResourceType: "instance",
				ResourceId:   "*",
				Actions:      []string{"read"},
			},
			{
				ResourceType: "instance", // should create this permission
				ResourceId:   "2<>",
				Actions:      []string{"delete"},
			},
		},
	})

	// Then, expect to return error since resourceId do not match regex
	if !strings.Contains(updateErrBadResourceId.Error(), "InvalidArgument ") {
		t.Fatalf("error should be returned since resourceId doesn't meet input requirements")
	}

}

func TestUpdateCloudAccountRoleWithNoPermissions(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "920728960769", Subject: "emily.thompson@examplecorp2.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "920728960769", Subject: "chris.thompson@examplecorp2.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)

	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	cloudAccountRoleCreated, err := client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{
		Alias:          "Alias001",
		CloudAccountId: "920728960769",
		Effect:         pb.CloudAccountRole_allow,
		Permissions: []*pb.CloudAccountRole_Permission{
			{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}},
			{ResourceType: "instance", ResourceId: "will-stay", Actions: []string{"delete"}},
			{ResourceType: "instance", ResourceId: "will-stay-2", Actions: []string{"delete"}},
			{ResourceType: "instance", ResourceId: "will-stay-3", Actions: []string{"delete"}},
			{ResourceType: "instance", ResourceId: "will-stay-4", Actions: []string{"delete"}},
		},
		Users: []string{"emily.thompson@examplecorp2.com"},
	})

	if len(cloudAccountRoleCreated.Permissions) != 5 {
		t.Fatalf("error cloudAccountRoleCreated permission now should be 5")
	}

	if err != nil {
		t.Fatalf("error creating cloudAccountRole  (%v)", err)
	}

	_, updateErr := client.UpdateCloudAccountRole(context.Background(), &pb.CloudAccountRoleUpdate{
		Id:             cloudAccountRoleCreated.Id,
		Alias:          "Alias02",
		CloudAccountId: "920728960769",
		Effect:         pb.CloudAccountRoleUpdate_deny,
		Users:          []string{"emily.thompson@examplecorp2.com", "chris.thompson@examplecorp2.com"},
	})

	if updateErr != nil {
		t.Fatalf("error updating cloudAccountRole  (%v)", updateErr)
	}

	cloudAccountRoleUpdated, getErr := client.GetCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{CloudAccountId: "920728960769", Id: cloudAccountRoleCreated.Id})
	if getErr != nil {
		t.Fatalf("error getting cloudAccountRole  (%v)", getErr)
	}

	if cloudAccountRoleUpdated.CloudAccountId != cloudAccountRoleCreated.CloudAccountId {
		t.Fatalf("error cloud account id for update and consulted cloudAccount role should match")
	}

	if cloudAccountRoleUpdated.Alias != "Alias02" {
		t.Fatalf("error alias was not updated")
	}

	if len(cloudAccountRoleUpdated.Users) != 2 {
		t.Fatalf("error users now should be 2")
	}

	if len(cloudAccountRoleUpdated.Permissions) != 5 {
		t.Fatalf("error permissions now should be 5")
	}

	if cloudAccountRoleUpdated.Effect.String() != pb.CloudAccountRole_deny.String() {
		t.Fatalf("error effect now should be deny")
	}
}

func TestUpdateCloudAccountRoleInvalidPermissions(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())
	cloudAccountRoleCreated, err := client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "Alias001", CloudAccountId: "720728960768", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"emily.thompson@examplecorp.com"}})
	if err != nil {
		t.Fatalf("error creating cloudAccountRole  (%v)", err)
	}

	notFoundId := "notfound"

	_, updateErr := client.UpdateCloudAccountRole(context.Background(), &pb.CloudAccountRoleUpdate{Id: cloudAccountRoleCreated.Id, Alias: "Alias02", CloudAccountId: "720728960768", Effect: pb.CloudAccountRoleUpdate_deny, Users: []string{"emily.thompson@examplecorp2.com", "chris.thompson@examplecorp2.com"}, Permissions: []*pb.CloudAccountRoleUpdate_Permission{{Id: &notFoundId, ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}})

	if updateErr == nil {
		t.Fatalf("updating cloudAccountRole should fail")
		return
	}

	cloudAccountRoleUpdated, getErr := client.GetCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{CloudAccountId: "720728960768", Id: cloudAccountRoleCreated.Id})
	if getErr != nil {
		t.Fatalf("error getting cloudAccountRole  (%v)", getErr)
	}

	if cloudAccountRoleUpdated.CloudAccountId != cloudAccountRoleCreated.CloudAccountId {
		t.Fatalf("error cloud account id for update and consulted cloudAccount role should match")
	}

	if cloudAccountRoleUpdated.Alias != "Alias001" {
		t.Fatalf("error alias was updated")
	}

}

func TestCreateCloudAccountRoleUniqueKeyValidationDuplicate(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "430158012468", Subject: "emily.thompson@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given, creation of specific cloudAccountRole
	_, err = client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "AliasDup1", CloudAccountId: "430158012468", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"emily.thompson@examplecorp.com"}})
	if err != nil {
		t.Fatalf("error creating cloudAccountRole  (%v)", err)
	}

	// When, trying to create duplicate alias and CloudAccountId
	_, err = client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "AliasDup1", CloudAccountId: "430158012468", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"emily.thompson@examplecorp.com"}})
	if err == nil {
		t.Fatalf("should return duplicate key value violates unique constraint error ")
	}
	// Then, expect to return error alias already exist
	if !strings.Contains(err.Error(), "AlreadyExists") {
		t.Fatalf("should return duplicate key value violates unique constraint")
	}
}

func TestQueryCloudAccountRoles(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())
	resourceType := "instance"
	cloudAcountId := "345678901234"

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: cloudAcountId, Subject: "unique.admin01@raremail.com", SystemRole: "cloud_account_admin"},
		{CloudAccountId: cloudAcountId, Subject: "unique.user01@raremail.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: cloudAcountId, Subject: "distinctive.alias03@infrequent.co.uk", SystemRole: "cloud_account_member"},
		{CloudAccountId: cloudAcountId, Subject: "random.person02@uncommondomain.org", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given
	cloudAccountRoles := []*pb.CloudAccountRole{
		{
			Alias:          "AliasQ1",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "*", Actions: []string{"read"}}},
			Users:          []string{"unique.user01@raremail.com"},
		},
		{
			Alias:          "AliasQ2",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "*", Actions: []string{"update"}}},
			Users:          []string{"distinctive.alias03@infrequent.co.uk"},
		},
		{
			Alias:          "AliasQ3",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "0", Actions: []string{"delete"}}},
			Users:          []string{"random.person02@uncommondomain.org"},
		},
		{
			Alias:          "AliasQ4",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "1", Actions: []string{"delete"}}},
			Users:          []string{"random.person02@uncommondomain.org"},
		},
		{
			Alias:          "AliasQ5",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "2", Actions: []string{"delete"}}},
			Users:          []string{"random.person02@uncommondomain.org"},
		},
		{
			Alias:          "AliasQ6",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "3", Actions: []string{"delete"}}},
			Users:          []string{"random.person02@uncommondomain.org"},
		},
		{
			Alias:          "AliasQ7",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "4", Actions: []string{"delete"}}},
			Users:          []string{"random.person02@uncommondomain.org"},
		},
		{
			Alias:          "AliasQ8",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "5", Actions: []string{"delete"}}},
			Users:          []string{"random.person02@uncommondomain.org"},
		},
		{
			Alias:          "AliasQ9",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "6", Actions: []string{"delete"}}},
			Users:          []string{"random.person02@uncommondomain.org"},
		},
		{
			Alias:          "AliasQ10",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "7", Actions: []string{"delete"}}},
			Users:          []string{"random.person02@uncommondomain.org"},
		},
		{
			Alias:          "AliasQ11",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "8", Actions: []string{"delete"}}},
			Users:          []string{"random.person02@uncommondomain.org"},
		},
		{
			Alias:          "AliasQ12",
			CloudAccountId: cloudAcountId,
			Effect:         pb.CloudAccountRole_allow,
			Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: resourceType, ResourceId: "9", Actions: []string{"delete"}}},
			Users:          []string{"random.person02@uncommondomain.org"},
		},
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	ctx := CreateContextWithToken("unique.admin01@raremail.com")

	// When - query existing queryCloudAccountsRoles for specific CloudAccountId and resourceType
	cloudAccountRolesResp, err := client.QueryCloudAccountRoles(ctx, &pb.CloudAccountRoleQuery{CloudAccountId: "345678901234", ResourceType: &resourceType})
	if err != nil {
		t.Fatalf("failed to query cloudAccountRoles (%v)", err)
	}

	// Then, expect to get the specified CloudAccountsRoles
	if len(cloudAccountRolesResp.CloudAccountRoles) != 12 {
		t.Fatalf("should have returned only items with cloudAccount %v ", cloudAcountId)
	}

	// When queryCloudAccountsRoles with size 1 page 1
	var size uint32 = 1
	cloudAccountRolesResp, err = client.QueryCloudAccountRoles(ctx, &pb.CloudAccountRoleQuery{CloudAccountId: "345678901234", ResourceType: &resourceType, Size: &size})
	if err != nil {
		t.Fatalf("failed to query cloudAccountRoles (%v)", err)
	}

	// Then, expect to return just 1 queryCloudAccountsRole
	if len(cloudAccountRolesResp.CloudAccountRoles) != 1 {
		t.Fatalf("should have returned only 1 item with cloudAccount %v since size is %v", cloudAcountId, size)
	}

	// When queryCloudAccountsRoles with a non existing CloudAccountId or resourceType
	resourceType = "nothing"
	cloudAccountRolesResp, err = client.QueryCloudAccountRoles(ctx, &pb.CloudAccountRoleQuery{CloudAccountId: "test", ResourceType: &resourceType, Size: &size})
	if err != nil {
		t.Fatalf("failed to query cloudAccountRoles (%v)", err)
	}

	// Then, expect to return 0 CloudAccountsRoles
	if len(cloudAccountRolesResp.CloudAccountRoles) != 0 {
		t.Fatalf("should have returned 0 cloudAccountRoles ")
	}
}

func TestAddRemoveUserToCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "770955320586", Subject: "unique.user87@raremail.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "770955320586", Subject: "unique.user87@evenrareremail.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given
	cloudAccountRole := pb.CloudAccountRole{
		Alias:          "AnotherAlias",
		CloudAccountId: "770955320586",
		Effect:         pb.CloudAccountRole_allow,
		Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "testvm", Actions: []string{"read"}}},
		Users:          []string{"unique.user87@raremail.com"},
	}
	createdRole, err := client.CreateCloudAccountRole(context.Background(), &cloudAccountRole)
	if err != nil {
		t.Fatalf("failed creating cloud account roles for the test (%v)", err)
	}

	// When adding another user
	newUser := "unique.user87@evenrareremail.com"
	_, err = client.AddUserToCloudAccountRole(context.Background(), &pb.CloudAccountRoleUserRequest{
		Id:             createdRole.Id,
		CloudAccountId: createdRole.CloudAccountId,
		UserId:         newUser,
	})

	// Expect no errors
	if err != nil {
		t.Fatalf("failed adding user to the cloud account role (%v)", err)
	}

	// Expect the role to contain the new user
	updatedRole, err := client.GetCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{
		Id:             createdRole.Id,
		CloudAccountId: createdRole.CloudAccountId})
	if err != nil {
		t.Fatalf("failed get cloud account role (%v)", err)
	}

	if !slices.Contains(updatedRole.Users, newUser) {
		t.Fatal("the role doesn't contain the new user")
	}

	// Given the recently added user
	// When calling the remove function
	_, err = client.RemoveUserFromCloudAccountRole(context.Background(), &pb.CloudAccountRoleUserRequest{
		Id:             createdRole.Id,
		CloudAccountId: createdRole.CloudAccountId,
		UserId:         newUser,
	})

	// Expect no errors
	if err != nil {
		t.Fatalf("failed removing the user from the cloud account role (%v)", err)
	}

	// Expect the role to not contain the new user
	updatedRole, err = client.GetCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{
		Id:             createdRole.Id,
		CloudAccountId: createdRole.CloudAccountId})

	if slices.Contains(updatedRole.Users, newUser) {
		t.Fatal("the role contains the deleted user")
	}

	// Given user request data with empty cloud account
	userRequest := pb.CloudAccountRoleUserRequest{
		Id:             createdRole.Id,
		CloudAccountId: "",
		UserId:         newUser,
	}

	// When calling the add user function
	_, err = client.AddUserToCloudAccountRole(context.Background(), &userRequest)

	// Expect a validation error
	if err == nil {
		t.Fatal("empty cloud account should return a validation error")
	}

	// Given the same user request data with empty cloud account
	// When calling the remove user function
	_, err = client.RemoveUserFromCloudAccountRole(context.Background(), &userRequest)

	// Expect a validation error
	if err == nil {
		t.Fatal("empty cloud account should return a validation error")
	}
}

func TestRemoveResourceFromCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "770955320584", Subject: "unique.user87@raremail.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given
	cloudAccountRole := pb.CloudAccountRole{
		Alias:          "AnotherAlias",
		CloudAccountId: "770955320584",
		Effect:         pb.CloudAccountRole_allow,
		Permissions: []*pb.CloudAccountRole_Permission{
			{ResourceType: "instance", ResourceId: "testvm", Actions: []string{"read"}},
			{ResourceType: "instance", ResourceId: "testvm1", Actions: []string{"read"}},
		},
		Users: []string{"unique.user87@raremail.com"},
	}
	createdCloudAccountRole, err := client.CreateCloudAccountRole(context.Background(), &cloudAccountRole)
	if err != nil {
		t.Fatalf("failed creating cloud account roles for the test (%v)", err)
	}

	_, err = client.RemoveResourceFromCloudAccountRole(context.Background(), &pb.CloudAccountRoleResourceRequest{
		CloudAccountId: "770955320584",
		ResourceId:     "testvm",
		ResourceType:   "instance",
	})
	if err != nil {
		t.Fatalf("failed to remove resource from cloud account role (%v)", err)
	}

	cloudAccountRoleUpdated, err := client.GetCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{CloudAccountId: "770955320584", Id: createdCloudAccountRole.Id})
	if err != nil {
		t.Fatalf("failed to get cloud account role (%v)", err)
	}

	if len(cloudAccountRoleUpdated.Permissions) != 1 {
		t.Fatalf("permissions len should be 1 since one resource was removed (%v)", err)
	}
}

func TestAddCloudAccountRolesToUser(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "170955320583", Subject: "jose.lopez@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "170955320583", Subject: "user1.thompson@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "370955320583", Subject: "user3.thompson@examplecorp.com", SystemRole: "cloud_account_admin"},
		{CloudAccountId: "470955320583", Subject: "user4.thompson@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	ctx := CreateContextWithToken("user3.thompson@examplecorp.com")

	// create cloud account role for user 1
	createdRole1, err := client.CreateCloudAccountRole(ctx, &pb.CloudAccountRole{Alias: "AddCloudAccountRolesToUseraAlias1", CloudAccountId: "170955320583", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "1", Actions: []string{"read"}}}, Users: []string{"jose.lopez@examplecorp.com"}})
	if err != nil {
		t.Fatalf("failed creating the cloud account role for the test (%v)", err)
	}
	createdRole2, err := client.CreateCloudAccountRole(ctx, &pb.CloudAccountRole{Alias: "AddCloudAccountRolesToUseraAlias2", CloudAccountId: "170955320583", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "2", Actions: []string{"read"}}}, Users: []string{}})
	if err != nil {
		t.Fatalf("failed creating the cloud account role for the test (%v)", err)
	}
	// When adding roles to user 1
	expectedRolesUser1 := []string{createdRole1.Id, createdRole2.Id}
	_, err = client.AddCloudAccountRolesToUser(ctx, &pb.CloudAccountRolesUserRequest{CloudAccountId: "170955320583", CloudAccountRoleIds: expectedRolesUser1, UserId: "user1.thompson@examplecorp.com"})
	if err != nil {
		t.Fatalf("failed adding cloud account roles to user (%v)", err)
	}
	user1, err := client.GetUser(ctx, &pb.GetUserRequest{CloudAccountId: "170955320583", UserId: "user1.thompson@examplecorp.com"})
	if err != nil {
		t.Fatalf("failed to get user (%v)", err)
	}
	gotRolesUser1 := []string{}
	for _, cloudAccountRole := range user1.CloudAccountRoles {
		gotRolesUser1 = append(gotRolesUser1, cloudAccountRole.CloudAccountRoleId)
	}
	// Then expect user 1 to have only createdRole1  and createdRole2
	if !containsOnly(gotRolesUser1, expectedRolesUser1) {
		t.Errorf("unexpected cloudAccountRoles: got %v, want %v", gotRolesUser1, expectedRolesUser1)
	}

	// When user 2 testing validation for add role for user as cloud_account_admin
	createdRole3, err := client.CreateCloudAccountRole(ctx, &pb.CloudAccountRole{Alias: "AddCloudAccountRolesToUseraAlias3", CloudAccountId: "370955320583", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "1", Actions: []string{"read"}}}})
	if err != nil {
		t.Fatalf("failed creating the system roles for the test (%v)", err)
	}
	_, err = client.AddCloudAccountRolesToUser(ctx, &pb.CloudAccountRolesUserRequest{CloudAccountId: "370955320583", CloudAccountRoleIds: []string{createdRole3.Id}, UserId: "user3.thompson@examplecorp.com"})
	// Then, expect to return error provided user is not a member of the cloud account
	if !strings.Contains(err.Error(), "provided user is not a member of the cloud account") {
		t.Fatalf("should return provided user is not a member of the cloud account")
	}

	// When adding roles to user 4 role 2 does not exist on specified cloud account
	_, err = client.AddCloudAccountRolesToUser(ctx, &pb.CloudAccountRolesUserRequest{CloudAccountId: "470955320583", CloudAccountRoleIds: []string{createdRole2.Id}, UserId: "user4.thompson@examplecorp.com"})
	// Then, expect to return error because provided role doesn't exist on specified cloud account
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("not found")
	}
}

func TestRemoveCloudAccountRolesToUser(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "110955320583", Subject: "jose.lopez@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "110955320583", Subject: "user1.thompson@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "330955320583", Subject: "user3.thompson@examplecorp.com", SystemRole: "cloud_account_admin"},
		{CloudAccountId: "440955320583", Subject: "user4.thompson@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	ctx := CreateContextWithToken("user3.thompson@examplecorp.com")

	// create cloud account role for user 1
	createdRole1, err := client.CreateCloudAccountRole(ctx, &pb.CloudAccountRole{Alias: "RemoveCloudAccountRolesToUseraAlias1", CloudAccountId: "110955320583", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "1", Actions: []string{"read"}}}, Users: []string{"jose.lopez@examplecorp.com"}})
	if err != nil {
		t.Fatalf("failed creating the cloud account role for the test (%v)", err)
	}
	createdRole2, err := client.CreateCloudAccountRole(ctx, &pb.CloudAccountRole{Alias: "RemoveCloudAccountRolesToUseraAlias2", CloudAccountId: "110955320583", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "2", Actions: []string{"read"}}}, Users: []string{}})
	if err != nil {
		t.Fatalf("failed creating the cloud account role for the test (%v)", err)
	}

	createdRole3, err := client.CreateCloudAccountRole(ctx, &pb.CloudAccountRole{Alias: "RemoveCloudAccountRolesToUseraAlias3", CloudAccountId: "110955320583", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "3", Actions: []string{"read"}}}, Users: []string{}})
	if err != nil {
		t.Fatalf("failed creating the cloud account role for the test (%v)", err)
	}
	// When adding roles to user 1
	expectedRolesUser1 := []string{createdRole1.Id, createdRole3.Id}
	_, err = client.AddCloudAccountRolesToUser(ctx, &pb.CloudAccountRolesUserRequest{CloudAccountId: "110955320583", CloudAccountRoleIds: []string{createdRole1.Id, createdRole2.Id, createdRole3.Id}, UserId: "user1.thompson@examplecorp.com"})
	if err != nil {
		t.Fatalf("failed adding cloud account roles to user (%v)", err)
	}
	// When removing roles to user 1
	_, err = client.RemoveCloudAccountRolesFromUser(ctx, &pb.CloudAccountRolesUserRequest{CloudAccountId: "110955320583", CloudAccountRoleIds: []string{createdRole2.Id}, UserId: "user1.thompson@examplecorp.com"})
	if err != nil {
		t.Fatalf("failed adding cloud account roles to user (%v)", err)
	}
	user1, err := client.GetUser(ctx, &pb.GetUserRequest{CloudAccountId: "110955320583", UserId: "user1.thompson@examplecorp.com"})
	if err != nil {
		t.Fatalf("failed to get user (%v)", err)
	}
	gotRolesUser1 := []string{}
	for _, cloudAccountRole := range user1.CloudAccountRoles {
		gotRolesUser1 = append(gotRolesUser1, cloudAccountRole.CloudAccountRoleId)
	}
	// Then expect user 1 to have only createdRole1  and createdRole2
	if !containsOnly(gotRolesUser1, expectedRolesUser1) {
		t.Errorf("unexpected cloudAccountRoles: got %v, want %v", gotRolesUser1, expectedRolesUser1)
	}

	// When user 2 testing validation for remove role for user as cloud_account_admin
	createdRole4, err := client.CreateCloudAccountRole(ctx, &pb.CloudAccountRole{Alias: "AddCloudAccountRolesToUseraAlias4", CloudAccountId: "330955320583", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "1", Actions: []string{"read"}}}})
	if err != nil {
		t.Fatalf("failed creating the system roles for the test (%v)", err)
	}
	_, err = client.RemoveCloudAccountRolesFromUser(ctx, &pb.CloudAccountRolesUserRequest{CloudAccountId: "330955320583", CloudAccountRoleIds: []string{createdRole4.Id}, UserId: "user3.thompson@examplecorp.com"})
	// Then, expect to return error provided user is not a member of the cloud account
	if !strings.Contains(err.Error(), "provided user is not a member of the cloud account") {
		t.Fatalf("should return provided user is not a member of the cloud account")
	}

	// When removing roles to user 4 role 2 does not exist on specified cloud account
	_, err = client.RemoveCloudAccountRolesFromUser(ctx, &pb.CloudAccountRolesUserRequest{CloudAccountId: "440955320583", CloudAccountRoleIds: []string{createdRole2.Id}, UserId: "user4.thompson@examplecorp.com"})
	// Then, expect to return error because provided role doesn't exist on specified cloud account
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("not found")
	}
}

func TestAddPermissionToCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())
	ctx := context.Background()

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "770955326586", Subject: "unique.user97@raremail.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given
	cloudAccountRole := pb.CloudAccountRole{
		Alias:          "permission1Alias",
		CloudAccountId: "770955326586",
		Effect:         pb.CloudAccountRole_allow,
		Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "testvm", Actions: []string{"read"}}},
		Users:          []string{"unique.user97@raremail.com"},
	}
	createdRole, err := client.CreateCloudAccountRole(ctx, &cloudAccountRole)
	if err != nil {
		t.Fatalf("failed creating the cloud account roles for the test (%v)", err)
	}

	// when adding a new permission
	_, err = client.AddPermissionToCloudAccountRole(ctx, &pb.CloudAccountRolePermissionRequest{
		CloudAccountId:     "770955326586",
		CloudAccountRoleId: createdRole.Id,
		Permission:         &pb.CloudAccountRole_Permission{ResourceType: "cloudaccount", ResourceId: "*", Actions: []string{"create"}},
	})
	if err != nil {
		t.Fatalf("failed adding permission to cloudAccountRole (%v)", err)
	}

	cloudAccountRoleUpdated, _ := client.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: "770955326586", Id: createdRole.Id})

	// expect to have permission added
	if len(cloudAccountRoleUpdated.Permissions) != 2 {
		t.Fatalf("should exist 2 permissions to this cloud account role")
	}

	// expect permission to match
	// if cloudAccountRoleUpdated.Permissions[1].ResourceType == "cloudaccount" &&
	// 	cloudAccountRoleUpdated.Permissions[1].ResourceId == "*" &&
	// 	cloudAccountRoleUpdated.Permissions[1].Actions[0] == "create" {
	// 	fmt.Printf("cloudAccountRoleUpdated.Permissions[1].ResourceType %s \n", cloudAccountRoleUpdated.Permissions[1].ResourceType)
	// 	fmt.Printf("cloudAccountRoleUpdated.Permissions[1].ResourceId %s \n", cloudAccountRoleUpdated.Permissions[1].ResourceId)
	// 	fmt.Printf("cloudAccountRoleUpdated.Permissions[1].Actions[0] %s \n", cloudAccountRoleUpdated.Permissions[1].Actions[0])
	// 	fmt.Printf("%+v\n", cloudAccountRoleUpdated)
	// 	t.Fatalf("added permission do not match")
	// }

	// when adding permission with unique key constraint
	_, err = client.AddPermissionToCloudAccountRole(ctx, &pb.CloudAccountRolePermissionRequest{
		CloudAccountId:     "770955326586",
		CloudAccountRoleId: createdRole.Id,
		Permission:         &pb.CloudAccountRole_Permission{ResourceType: "cloudaccount", ResourceId: "*", Actions: []string{"create"}},
	})

	// Then, expect to return error alias already exist
	if !strings.Contains(err.Error(), "AlreadyExists") {
		t.Fatalf("should return duplicate key value violates unique constraint type and cloudAccountId")
	}

	// when adding permission with invalid resourceId
	_, err = client.AddPermissionToCloudAccountRole(ctx, &pb.CloudAccountRolePermissionRequest{
		CloudAccountId:     "770955326586",
		CloudAccountRoleId: createdRole.Id,
		Permission:         &pb.CloudAccountRole_Permission{ResourceType: "cloudaccount", ResourceId: "*<>", Actions: []string{"create"}},
	})
	if !strings.Contains(err.Error(), "InvalidArgument") {
		t.Fatalf("should return field resourceId with value *<> doesn't meet input requirements")
	}

}

func TestAddPermissionLimitsToCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())
	ctx := context.Background()

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "778155326586", Subject: "unique.user97@raremail.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given
	cloudAccountRole := pb.CloudAccountRole{
		Alias:          "permission1Alias",
		CloudAccountId: "778155326586",
		Effect:         pb.CloudAccountRole_allow,
		Permissions:    []*pb.CloudAccountRole_Permission{},
		Users:          []string{"unique.user97@raremail.com"},
	}
	createdRole, err := client.CreateCloudAccountRole(ctx, &cloudAccountRole)
	if err != nil {
		t.Fatalf("failed creating the cloud account roles for the test (%v)", err)
	}

	// 100 defined in the testing.go config setup
	for i := 0; i < 100; i++ {
		// when adding a new permission
		_, err = client.AddPermissionToCloudAccountRole(ctx, &pb.CloudAccountRolePermissionRequest{
			CloudAccountId:     "778155326586",
			CloudAccountRoleId: createdRole.Id,
			Permission:         &pb.CloudAccountRole_Permission{ResourceType: "cloudaccount", ResourceId: fmt.Sprintf("%v", i), Actions: []string{"create"}},
		})
		if err != nil {
			t.Fatalf("failed adding permission to cloudAccountRole (%v)", err)
		}
	}

	_, err = client.AddPermissionToCloudAccountRole(ctx, &pb.CloudAccountRolePermissionRequest{
		CloudAccountId:     "778155326586",
		CloudAccountRoleId: createdRole.Id,
		Permission:         &pb.CloudAccountRole_Permission{ResourceType: "cloudaccount", ResourceId: "99999", Actions: []string{"create"}},
	})

	if err == nil {
		t.Fatalf("should reach permission limit (%v)", err)
	}
}

func TestRemovePermissionFromCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())
	ctx := context.Background()

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "770955322586", Subject: "unique.user98@raremail.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given
	cloudAccountRole := pb.CloudAccountRole{
		Alias:          "permission11Alias",
		CloudAccountId: "770955322586",
		Effect:         pb.CloudAccountRole_allow,
		Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "testvm", Actions: []string{"read"}}},
		Users:          []string{"unique.user98@raremail.com"},
	}
	createdRole, err := client.CreateCloudAccountRole(ctx, &cloudAccountRole)
	if err != nil {
		t.Fatalf("failed creating the cloud account roles for the test (%v)", err)
	}

	// when adding a new permission
	_, err = client.AddPermissionToCloudAccountRole(ctx, &pb.CloudAccountRolePermissionRequest{
		CloudAccountId:     "770955322586",
		CloudAccountRoleId: createdRole.Id,
		Permission:         &pb.CloudAccountRole_Permission{ResourceType: "cloudaccount", ResourceId: "*", Actions: []string{"create"}},
	})
	if err != nil {
		t.Fatalf("failed adding permission to cloudAccountRole (%v)", err)
	}

	cloudAccountRoleUpdated, _ := client.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: "770955322586", Id: createdRole.Id})

	_, err = client.RemovePermissionFromCloudAccountRole(ctx, &pb.CloudAccountRolePermissionId{
		CloudAccountId:     "770955322586",
		CloudAccountRoleId: createdRole.Id,
		Id:                 cloudAccountRoleUpdated.Permissions[0].Id,
	})
	if err != nil {
		t.Fatalf("failed delete permission from cloudAccountRole (%v)", err)
	}

	cloudAccountRoleUpdated, _ = client.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: "770955322586", Id: createdRole.Id})
	if len(cloudAccountRoleUpdated.Permissions) != 1 {
		t.Fatalf("permissions should have 1 permission")
	}

	_, err = client.RemovePermissionFromCloudAccountRole(ctx, &pb.CloudAccountRolePermissionId{
		CloudAccountId:     "770955322586",
		CloudAccountRoleId: createdRole.Id,
		Id:                 cloudAccountRoleUpdated.Permissions[0].Id,
	})
	if err != nil {
		t.Fatalf("failed delete permission from cloudAccountRole (%v)", err)
	}

	cloudAccountRoleUpdated, _ = client.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: "770955322586", Id: createdRole.Id})
	if len(cloudAccountRoleUpdated.Permissions) != 0 {
		t.Fatalf("permissions should have 0 permission")
	}

	// when adding a new permission
	_, err = client.AddPermissionToCloudAccountRole(ctx, &pb.CloudAccountRolePermissionRequest{
		CloudAccountId:     "770955322586",
		CloudAccountRoleId: createdRole.Id,
		Permission:         &pb.CloudAccountRole_Permission{ResourceType: "cloudaccount", ResourceId: "*", Actions: []string{"create"}},
	})
	if err != nil {
		t.Fatalf("failed adding permission to cloudAccountRole (%v)", err)
	}

	cloudAccountRoleUpdated, _ = client.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: "770955322586", Id: createdRole.Id})
	if len(cloudAccountRoleUpdated.Permissions) != 1 {
		t.Fatalf("permissions should have 1 permission")
	}

}

func TestUpdatePermissionFromCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())
	ctx := context.Background()

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "770855322586", Subject: "unique.user99@raremail.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given
	cloudAccountRole := pb.CloudAccountRole{
		Alias:          "permission111Alias",
		CloudAccountId: "770855322586",
		Effect:         pb.CloudAccountRole_allow,
		Permissions:    []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "testvm", Actions: []string{"read"}}},
		Users:          []string{"unique.user99@raremail.com"},
	}
	createdRole, err := client.CreateCloudAccountRole(ctx, &cloudAccountRole)
	if err != nil {
		t.Fatalf("failed creating the cloud account roles for the test (%v)", err)
	}

	getCloudAccountRole, _ := client.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: "770855322586", Id: createdRole.Id})

	_, err = client.UpdatePermissionCloudAccountRole(ctx, &pb.CloudAccountRolePermissionRequest{
		CloudAccountId: getCloudAccountRole.CloudAccountId, CloudAccountRoleId: getCloudAccountRole.Id, Permission: &pb.CloudAccountRole_Permission{
			Id:           getCloudAccountRole.Permissions[0].Id,
			ResourceType: "instance",
			ResourceId:   "abc",
			Actions:      []string{"read"}}})
	if err != nil {
		t.Fatalf("failed updating permission from cloudAccountRole (%v)", err)
	}

	getCloudAccountRoleUpdated, _ := client.GetCloudAccountRole(ctx, &pb.CloudAccountRoleId{CloudAccountId: "770855322586", Id: createdRole.Id})
	permission := getCloudAccountRoleUpdated.Permissions[0]
	if permission.ResourceType == "instance" && permission.Actions[0] == "read" && permission.ResourceId == "testvm" {
		t.Fatalf("failed permission updated do not match")
	}

	// when updating a permission with invalid resourceId
	_, err = client.UpdatePermissionCloudAccountRole(ctx, &pb.CloudAccountRolePermissionRequest{
		CloudAccountId: getCloudAccountRole.CloudAccountId, CloudAccountRoleId: getCloudAccountRole.Id, Permission: &pb.CloudAccountRole_Permission{
			Id:           getCloudAccountRole.Permissions[0].Id,
			ResourceType: "instance",
			ResourceId:   "ab<&$>c",
			Actions:      []string{"read"}}})

	if !strings.Contains(err.Error(), "InvalidArgument") {
		t.Fatalf("should return error field resourceId with value ab<&$>c doesn't meet input requirements")
	}
}

func TestUnassignSystemRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())
	// Given the assigned system role
	roleRequest := &pb.RoleRequest{
		CloudAccountId: "770955320589",
		Subject:        "unique.user87@raremail.com",
		SystemRole:     "cloud_account_member",
	}
	_, err := client.AssignSystemRole(context.Background(), roleRequest)

	if err != nil {
		t.Fatalf("not expecting an error while assigning a system role")
	}
	existResponse, err := client.SystemRoleExists(context.Background(), roleRequest)
	if err != nil {
		t.Fatal("not expecting errors after checking the system role existence")
	} else {
		if existResponse.Exist == false {
			t.Fatal("expecting the system role to exist after assigning")
		}
	}

	// When unassigning the same system role
	_, err = client.UnassignSystemRole(context.Background(), roleRequest)

	// Expect no errors
	if err != nil {
		t.Fatal("not expecting any errors unassigning a system role")
	}

	// Expect the system role to not exists
	existResponse, err = client.SystemRoleExists(context.Background(), roleRequest)
	if err != nil {
		t.Fatal("not expecting errors after checking the system role existence")
	} else {
		if existResponse.Exist == true {
			t.Fatal("expecting the system role to not exist after unassigning")
		}
	}

	// Given the same role request, with empty cloud account ID
	roleRequest = &pb.RoleRequest{
		CloudAccountId: "",
		Subject:        "unique.user87@raremail.com",
		SystemRole:     "cloud_account_member",
	}

	// When assign
	_, err = client.AssignSystemRole(context.Background(), roleRequest)

	// Expect validation error
	if err == nil {
		t.Fatal("expecting a validation error due to empty cloud account ID")
	}

	// When unassign
	_, err = client.UnassignSystemRole(context.Background(), roleRequest)

	// Expect validation error
	if err == nil {
		t.Fatal("expecting a validation error due to empty cloud account ID")
	}

	// When checking if role exists
	_, err = client.SystemRoleExists(context.Background(), roleRequest)

	// Expect validation error
	if err == nil {
		t.Fatal("expecting a validation error due to empty cloud account ID")
	}

}

func TestRemoveCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "720728960763", Subject: "emily.thompson@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given the following cloudAccountRole created
	cloudAccountRole, err := client.CreateCloudAccountRole(context.Background(), &pb.CloudAccountRole{Alias: "AliasRemoveCloudAccountRole", CloudAccountId: "720728960763", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"emily.thompson@examplecorp.com"}})
	if err != nil {
		t.Fatalf("error creating cloudAccountRole  (%v)", err)
	}

	// When removing the specified CloudAccountRole
	_, err = client.RemoveCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{CloudAccountId: cloudAccountRole.CloudAccountId, Id: cloudAccountRole.Id})
	if err != nil {
		t.Fatalf("error deleting cloudAccountRole  (%v)", err)
	}

	// Then, expect CloudAccountRole to not exist since was deleted
	cloudAccountRoleNotExist, err := client.GetCloudAccountRole(context.Background(), &pb.CloudAccountRoleId{CloudAccountId: cloudAccountRole.CloudAccountId, Id: cloudAccountRole.Id})
	if err == nil {
		t.Fatalf("cloudaccountrole not found should have been returned")
	}

	if cloudAccountRoleNotExist != nil {
		t.Fatalf("cloudAccountRole deleted should not exist")
	}

}

func TestCreatePolicy(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given a new policy
	policy := &pb.PolicyRequest{
		Subject:    "cloud_account_member",
		Object:     "/cloud_account/:cloud_account_id/vnet",
		Action:     "WRITE",
		Expression: "true",
	}

	// When create policy is called
	_, err := client.CreatePolicy(context.Background(), policy)

	// Expect no errors:
	if err != nil {
		t.Fatalf("there should be no errors creating a policy (%v)", err)
	}

	// When create policy is called with the same values
	_, err = client.CreatePolicy(context.Background(), policy)

	// Expect no errors:
	if err != nil {
		t.Fatalf("there should be no errors creating a policy that already exists (%v)", err)
	}
}

func TestCreatePolicyWrongData(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given empty data
	policy := &pb.PolicyRequest{
		Subject:    "",
		Object:     "",
		Action:     "",
		Expression: "",
	}

	// When create policy is called
	_, err := client.CreatePolicy(context.Background(), policy)

	// Expect validation errors:
	if err == nil {
		t.Fatalf("there should be a validation error when creating the policy with empty data  (%v)", err)
	}
}

func TestRemovePolicy(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given the current policy: cloud_account_member;/cloud_account/:cloud_account_id/instance;DELETE;"true"
	// Loaded from the policy.csv
	policy := &pb.PolicyRequest{
		Subject:    "cloud_account_member",
		Object:     "/cloud_account/:cloud_account_id/instance",
		Action:     "WRITE",
		Expression: "true",
	}

	// When remove policy is called
	_, err := client.RemovePolicy(context.Background(), policy)

	// Expect no errors:
	if err != nil {
		t.Fatalf("there should be no errors deleting a policy that exists(%v)", err)
	}

	// When remove policy is called with the same values
	_, err = client.RemovePolicy(context.Background(), policy)

	// Expect no errors:
	if err != nil {
		t.Fatalf("there should be no errors deleting a policy that doesn't exists (%v)", err)
	}

}

func TestRemovePolicyWrongData(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given empty data
	policy := &pb.PolicyRequest{
		Subject:    "",
		Object:     "",
		Action:     "",
		Expression: "",
	}

	// When remove policy is called
	_, err := client.RemovePolicy(context.Background(), policy)

	// Expect validation errors:
	if err == nil {
		t.Fatalf("there should be a validation error when deleting the policy with empty data  (%v)", err)
	}

	// Given wrong data
	policy = &pb.PolicyRequest{
		Subject:    "x",
		Object:     "x",
		Action:     "x",
		Expression: "x",
	}

	// When remove policy is called
	_, err = client.RemovePolicy(context.Background(), policy)

	// Expect no errors:
	if err != nil {
		t.Fatalf("there should be no errors deleting a policy that doesn't exists (%v)", err)
	}
}

func TestLookupCloudAccountMember(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "168310882075", Subject: "jose.lopez@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "168310882075", Subject: "maria.lopez@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	// creating permission to allow
	cloudAccountRoles := []*pb.CloudAccountRole{
		{Alias: "Alias0001", CloudAccountId: "168310882075", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "instance001", Actions: []string{"read"}}, {ResourceType: "instance", ResourceId: "instance003", Actions: []string{"read"}}}, Users: []string{"jose.lopez@examplecorp.com"}},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	ctx := CreateContextWithToken("jose.lopez@examplecorp.com")

	// when lookup for specific resource
	lookupResp, err := client.Lookup(ctx, &pb.LookupRequest{CloudAccountId: "168310882075", ResourceType: "instance", Action: "read", ResourceIds: []string{"instance001", "instance002", "instance003"}})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}

	// expect to be only 2 resources
	if len(lookupResp.ResourceIds) == 2 && lookupResp.ResourceIds[0] != "instance001" && lookupResp.ResourceIds[1] != "instance003" {
		t.Fatalf("expect to have only two instances")
	}

	ctx = CreateContextWithToken("maria.lopez@examplecorp.com")

	// when lookup for specific resource
	lookupResp, err = client.Lookup(ctx, &pb.LookupRequest{CloudAccountId: "168310882075", ResourceType: "instance", Action: "read", ResourceIds: []string{"instance001", "instance002", "instance003"}})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}

	// expect to be 0 resources since maria doesn't have acces to anything on specific cloudAccount
	if len(lookupResp.ResourceIds) != 0 {
		t.Fatalf("expect to have 0 instances")
	}

}

/*
func TestLookupCloudAccountMemberWildCard(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "168310882079", Subject: "josefina.lopez@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "168310882079", Subject: "perez.lopez@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	// creating permission to allow
	cloudAccountRoles := []*pb.CloudAccountRole{
		{Alias: "Alias0001", CloudAccountId: "168310882079", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"*"}},                                                                                                                             // 1 all users can 'read' all instances
		{Alias: "Alias0002", CloudAccountId: "168310882079", Effect: pb.CloudAccountRole_deny, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "instance001", Actions: []string{"delete"}}, {ResourceType: "instance", ResourceId: "instance003", Actions: []string{"delete"}}}, Users: []string{"josefina.lopez@examplecorp.com"}}, // 2 deny Josefina to delete instance001 and instance003
		{Alias: "Alias0003", CloudAccountId: "168310882079", Effect: pb.CloudAccountRole_deny, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"update"}}}, Users: []string{"*"}},                                                                                                                            // 3 deny all users to update all instances
		{Alias: "Alias0004", CloudAccountId: "168310882079", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"delete"}}}, Users: []string{"*"}},                                                                                                                           // trying to allow delete all instances to all users
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	ctx := CreateContextWithToken("josefina.lopez@examplecorp.com")

	// 1 when lookup for read instance000", "instance001", "instance002", "instance003", "instance004
	lookupResp, err := client.Lookup(ctx, &pb.LookupRequest{CloudAccountId: "168310882079", ResourceType: "instance", Action: "read", ResourceIds: []string{"instance000", "instance001", "instance002", "instance003", "instance004"}})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}

	// 1 expect to  all users can 'read' all instances
	if len(lookupResp.ResourceIds) != 5 && lookupResp.ResourceIds[0] != "instance000" && lookupResp.ResourceIds[1] != "instance001" && lookupResp.ResourceIds[2] != "instance002" && lookupResp.ResourceIds[3] != "instance003" && lookupResp.ResourceIds[4] != "instance004" {
		t.Fatalf("expect to have only 5 instances")
	}

	// 2 when lookup for Josefina to delete  "instance000", "instance001", "instance002", "instance003", "instance004"
	lookupResp, err = client.Lookup(ctx, &pb.LookupRequest{CloudAccountId: "168310882079", ResourceType: "instance", Action: "delete", ResourceIds: []string{"instance000", "instance001", "instance002", "instance003", "instance004"}})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}

	// 2  Josefina can delete All instances because of Alias0004 but cannot delete instance001 instance003 because of Alias0002
	if len(lookupResp.ResourceIds) != 3 && lookupResp.ResourceIds[0] != "instance000" && lookupResp.ResourceIds[1] != "instance002" && lookupResp.ResourceIds[2] != "instance004" {
		t.Fatalf("expect to have only 3 instances")
	}

	// 3 when lookup for Josefina to update "instance000", "instance001", "instance002", "instance003", "*"
	lookupResp, err = client.Lookup(ctx, &pb.LookupRequest{CloudAccountId: "168310882079", ResourceType: "instance", Action: "update", ResourceIds: []string{"instance000", "instance001", "instance002", "instance003", "*"}})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}

	// 3  Josefina cannot update all instances because of alias Alias0003
	if len(lookupResp.ResourceIds) != 0 {
		t.Fatalf("expect to have 0 instances")
	}

	ctx = CreateContextWithToken("perez.lopez@examplecorp.com")

	// 1 when lookup for read instance000", "instance001", "instance002", "instance003", "instance004
	lookupResp, err = client.Lookup(ctx, &pb.LookupRequest{CloudAccountId: "168310882079", ResourceType: "instance", Action: "read", ResourceIds: []string{"instance000", "instance001", "instance002", "instance003", "instance004"}})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}
	// 1 expect to  all users can 'read' all instances
	if len(lookupResp.ResourceIds) != 5 && lookupResp.ResourceIds[0] != "instance000" && lookupResp.ResourceIds[1] != "instance001" && lookupResp.ResourceIds[2] != "instance002" && lookupResp.ResourceIds[3] != "instance003" && lookupResp.ResourceIds[4] != "instance004" {
		t.Fatalf("expect to have only 5 instances")
	}

	// 2 when lookup for Perez to delete  "instance000", "instance001", "instance002", "instance003", "instance004"
	lookupResp, err = client.Lookup(ctx, &pb.LookupRequest{CloudAccountId: "168310882079", ResourceType: "instance", Action: "delete", ResourceIds: []string{"instance000", "instance001", "instance002", "instance003", "instance004"}})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}

	// 2  Perez can delete All instances because of Alias0004
	if len(lookupResp.ResourceIds) != 5 && lookupResp.ResourceIds[0] != "instance000" && lookupResp.ResourceIds[1] != "instance001" && lookupResp.ResourceIds[2] != "instance002" && lookupResp.ResourceIds[3] != "instance003" && lookupResp.ResourceIds[4] != "instance004" {
		t.Fatalf("expect to have only 5 instances")
	}

	// denying everything on specific cloudAccount
	err = createCloudAccountRolesRoleBulk(client, []*pb.CloudAccountRole{
		{Alias: "Alias0005", CloudAccountId: "168310882079", Effect: pb.CloudAccountRole_deny, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read", "update", "delete"}}}, Users: []string{"*"}},
	})
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// after denying when lookup for read instance000", "instance001", "instance002", "instance003", "instance004
	lookupResp, err = client.Lookup(ctx, &pb.LookupRequest{CloudAccountId: "168310882079", ResourceType: "instance", Action: "read", ResourceIds: []string{"instance000", "instance001", "instance002", "instance003", "instance004"}})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}

	// after denying perez cannot read anything
	if len(lookupResp.ResourceIds) != 0 {
		t.Fatalf("expect to have only 0 instances")
	}
}*/

func TestActions(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "*", Subject: "user0@intel.com", SystemRole: pb.SystemRole_intel_admin.String()},
		{CloudAccountId: "168310882070", Subject: "user1@examplecorp.com", SystemRole: pb.SystemRole_cloud_account_member.String()},
		{CloudAccountId: "168310882070", Subject: "user2@examplecorp.com", SystemRole: pb.SystemRole_cloud_account_member.String()},
		{CloudAccountId: "168310882070", Subject: "user3@examplecorp.com", SystemRole: pb.SystemRole_cloud_account_admin.String()},
		{CloudAccountId: "*", Subject: "user4@examplecorp.com", SystemRole: pb.SystemRole_intel_admin.String()},
	}

	// creating permission to allow
	cloudAccountRoles := []*pb.CloudAccountRole{
		{Alias: "Alias0001", CloudAccountId: "168310882070", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "instance001", Actions: []string{"read", "update"}}, {ResourceType: "instance", ResourceId: "instance003", Actions: []string{"read", "update"}}}, Users: []string{"user1@examplecorp.com"}},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// user1 (cloud_account_member) allowed actions should be (read,update) because there is a cloudAccountRole assign to user1
	ctx := CreateContextWithToken("user1@examplecorp.com")
	allowedActions, err := client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310882070", ResourceType: "instance", ResourceId: "instance003"})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}
	if len(allowedActions.Actions) != 2 && allowedActions.Actions[0] != "read" && allowedActions.Actions[1] != "update" {
		t.Fatalf("allowedActions not expected for user1")
	}

	// user1 with invalid ResourceType
	_, err = client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310882070", ResourceType: "invalidResourceType", ResourceId: "instance003"})
	if err == nil {
		t.Fatalf("should return error for invalid resourceType,Resource invalidResourceType not found")
	}

	// user1  deny the access (cloud_account_member) allowed actions should be empty since cloudAccountRole was created to deny the access
	cloudAccountRoles = []*pb.CloudAccountRole{
		{Alias: "DenyUser1", CloudAccountId: "168310882070", Effect: pb.CloudAccountRole_deny, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "instance001", Actions: []string{"read", "update"}}}, Users: []string{"user1@examplecorp.com"}},
	}
	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}
	allowedActions, err = client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310882070", ResourceType: "instance", ResourceId: "instance001"})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}
	if len(allowedActions.Actions) != 0 {
		t.Fatalf("allowedActions should be empty since deny cloudAccountRole rule was created for user1")
	}

	// user2 (cloud_account_member) allowed actions should be empty because there's no cloudAccountRole for user2
	ctx = CreateContextWithToken("user2@examplecorp.com")
	allowedActions, err = client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310882070", ResourceType: "instance", ResourceId: "instance001"})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}
	if len(allowedActions.Actions) != 0 {
		t.Fatalf("allowedActions not expected for user2, expect 0")
	}

	// user3 (cloud_account_admin) allowed actions should be all allowed actions since user3 is admin
	ctx = CreateContextWithToken("user3@examplecorp.com")
	allowedActions, err = client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310882070", ResourceType: "instance", ResourceId: "instance003"})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}
	if len(allowedActions.Actions) != 5 {
		t.Fatalf("allowedActions not expected for user3, expect to have all allowed actions for the resource")
	}

	// user0 (intel_admin) allowed actions should be all allowed actions since user0 is admin
	ctx = CreateContextWithToken("user0@intel.com")
	allowedActions, err = client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "*", ResourceType: "instance", ResourceId: "instance001"})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}
	if len(allowedActions.Actions) == 0 {
		t.Fatalf("allowedActions not expected for user0, expect to have all allowed actions for the resource")
	}

}

func TestActionsWildCard(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "*", Subject: "user00@intel.com", SystemRole: pb.SystemRole_intel_admin.String()},
		{CloudAccountId: "168310882075", Subject: "user1@examplecorp.com", SystemRole: pb.SystemRole_cloud_account_member.String()},
		{CloudAccountId: "168310882075", Subject: "user2@examplecorp.com", SystemRole: pb.SystemRole_cloud_account_member.String()},
		{CloudAccountId: "168310882075", Subject: "user3@examplecorp.com", SystemRole: pb.SystemRole_cloud_account_admin.String()},
		{CloudAccountId: "*", Subject: "user40@examplecorp.com", SystemRole: pb.SystemRole_intel_admin.String()},
	}

	// creating permission to allow
	cloudAccountRoles := []*pb.CloudAccountRole{
		{Alias: "Alias00001", CloudAccountId: "168310882075", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "instance001", Actions: []string{"update"}}}, Users: []string{"user1@examplecorp.com"}},
		{Alias: "Alias00003", CloudAccountId: "168310882075", Effect: pb.CloudAccountRole_deny, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "instance001", Actions: []string{"delete"}}, {ResourceType: "instance", ResourceId: "instance003", Actions: []string{"delete"}}}, Users: []string{"user1@examplecorp.com"}},
		{Alias: "Alias00004", CloudAccountId: "168310882075", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"user1@examplecorp.com"}},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// user1 (cloud_account_member) allowed actions should be (update,read) because there is a cloudAccountRole assign to user1
	ctx := CreateContextWithToken("user1@examplecorp.com")
	allowedActions, err := client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310882075", ResourceType: "instance", ResourceId: "instance001"})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}

	if len(allowedActions.Actions) != 2 && allowedActions.Actions[0] != "update" && allowedActions.Actions[1] != "read" {
		t.Fatalf("allowedActions not expected for user1")
	}

	// user1 with invalid ResourceType
	_, err = client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310882075", ResourceType: "invalidResourceType", ResourceId: "instance003"})
	if err == nil {
		t.Fatalf("should return error for invalid resourceType,Resource invalidResourceType not found")
	}

	// user1  deny the access (cloud_account_member) allowed actions should be empty since cloudAccountRole was created to deny the access
	cloudAccountRoles = []*pb.CloudAccountRole{
		{Alias: "DenyUser1", CloudAccountId: "168310882075", Effect: pb.CloudAccountRole_deny, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "instance001", Actions: []string{"read", "update"}}, {ResourceType: "instance", ResourceId: "instance003", Actions: []string{"read", "update"}}}, Users: []string{"user1@examplecorp.com"}},
	}
	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}
	allowedActions, err = client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310882075", ResourceType: "instance", ResourceId: "instance001"})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}
	if len(allowedActions.Actions) != 0 {
		t.Fatalf("allowedActions should be empty since deny cloudAccountRole rule was created for user1")
	}

	// user3 (cloud_account_admin) allowed actions should be all allowed actions since user3 is admin
	ctx = CreateContextWithToken("user3@examplecorp.com")
	allowedActions, err = client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310882075", ResourceType: "instance", ResourceId: "instance003"})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}
	if len(allowedActions.Actions) != 5 {
		t.Fatalf("allowedActions not expected for user3, expect to have all allowed actions for the resource")
	}

	// user0 (intel_admin) allowed actions should be all allowed actions since user0 is admin
	ctx = CreateContextWithToken("user00@intel.com")
	allowedActions, err = client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "*", ResourceType: "instance", ResourceId: "instance001"})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}
	if len(allowedActions.Actions) == 0 {
		t.Fatalf("allowedActions not expected for user0, expect to have all allowed actions for the resource")
	}

}

func TestActionsWithActionTypeCollection(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "168310582076", Subject: "user1@examplecorp.com", SystemRole: pb.SystemRole_cloud_account_member.String()},
	}

	// creating permission to allow
	cloudAccountRoles := []*pb.CloudAccountRole{
		{Alias: "Alias000001", CloudAccountId: "168310582076", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"create", "list"}}}, Users: []string{"user1@examplecorp.com"}},
		{Alias: "Alias000002", CloudAccountId: "168310582076", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"user1@examplecorp.com"}},
		{Alias: "Alias000003", CloudAccountId: "168310582076", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "abc", Actions: []string{"delete"}}}, Users: []string{"user1@examplecorp.com"}},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// user1 (cloud_account_member) allowed actions should be (read,create,list) because those are the allowed actions in general for this user for all the resources in this type
	ctx := CreateContextWithToken("user1@examplecorp.com")
	allowedActions, err := client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310582076", ResourceType: "instance"})
	if err != nil {
		t.Fatalf("error executing actions (%v)", err)
	}

	if len(allowedActions.Actions) != 3 && allowedActions.Actions[0] != "read" && allowedActions.Actions[1] != "update" && allowedActions.Actions[2] != "list" {
		t.Fatalf("allowedActions not expected for user1")
	}

	// user1 (cloud_account_member) allowed actions should be (read,create,list,delete) because those are the allowed actions in general for this user for all the resources in this type and also add the action for the specific resource abc that is delete
	ctx = CreateContextWithToken("user1@examplecorp.com")
	allowedActions, err = client.Actions(ctx, &pb.ActionsRequest{CloudAccountId: "168310582076", ResourceType: "instance", ResourceId: "abc"})
	if err != nil {
		t.Fatalf("error executing actions (%v)", err)
	}

	if len(allowedActions.Actions) != 4 && allowedActions.Actions[0] != "read" && allowedActions.Actions[1] != "update" && allowedActions.Actions[2] != "list" && allowedActions.Actions[3] != "delete" {
		t.Fatalf("allowedActions not expected for user1")
	}

}

func TestLookupCloudAccountAdmin(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "168310882072", Subject: "maria.lopez@examplecorp.com", SystemRole: "cloud_account_admin"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	ctx := CreateContextWithToken("maria.lopez@examplecorp.com")

	// when lookup for specific resource
	lookupResp, err := client.Lookup(ctx, &pb.LookupRequest{CloudAccountId: "168310882072", ResourceType: "instance", Action: "read", ResourceIds: []string{"instance001", "instance002", "instance003"}})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}

	// expect to return all 3 instances since admin has access to everything
	if lookupResp.ResourceIds[0] != "instance001" && lookupResp.ResourceIds[1] != "instance002" && lookupResp.ResourceIds[2] != "instance003" {
		t.Fatalf("expect to have only three instances")
	}
}

func TestLookupIntelAdmin(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "*", Subject: "maria.perez@examplecorp.com", SystemRole: "intel_admin"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	ctx := CreateContextWithToken("maria.perez@examplecorp.com")

	// when lookup for specific resource
	lookupResp, err := client.Lookup(ctx, &pb.LookupRequest{CloudAccountId: "168310882023", ResourceType: "instance", Action: "read", ResourceIds: []string{"instance001", "instance003"}})
	if err != nil {
		t.Fatalf("error executing lookup (%v)", err)
	}

	// expect to return all 2 instances since intel admin has access to everything
	if lookupResp.ResourceIds[0] != "instance001" && lookupResp.ResourceIds[1] != "instance003" {
		t.Fatalf("expect to have only three instances")
	}
}

func TestListResourceDefinition(t *testing.T) {
	// Given
	client := pb.NewAuthzServiceClient(test.ClientConn())
	// When - query ListResourceDefinition
	resourceDefinitionsResp, err := client.ListResourceDefinition(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("failed to query ListResourceDefinition (%v)", err)
	}
	// Then should list all resources configured on resources.yaml
	if len(resourceDefinitionsResp.Resources) != 3 {
		t.Fatalf("resource definitions do not match")
	}
}

func TestListUsersByCloudAccount(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "*", Subject: "elsadmin@intel.com", SystemRole: "intel_admin"},
		{CloudAccountId: "168310882001", Subject: "alice.wonder@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "168310882001", Subject: "bob.builder@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "168310882001", Subject: "jess.builder@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "168310882002", Subject: "charlie.choco@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "168310882002", Subject: "diana.dreams@examplecorp.com", SystemRole: "cloud_account_member"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// creating cloud account roles
	cloudAccountRoles := []*pb.CloudAccountRole{
		{Alias: "Alias0000001", CloudAccountId: "168310882001", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"create", "list"}}}, Users: []string{"alice.wonder@examplecorp.com", "bob.builder@examplecorp.com"}},
		{Alias: "Alias0000002", CloudAccountId: "168310882001", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"create", "list"}}}, Users: []string{"bob.builder@examplecorp.com"}},
		{Alias: "Alias0000004", CloudAccountId: "168310882002", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "*", Actions: []string{"read"}}}, Users: []string{"charlie.choco@examplecorp.com"}},
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// Given
	cloudAccountId := "168310882001"
	expectedUsers := []string{"alice.wonder@examplecorp.com", "bob.builder@examplecorp.com", "jess.builder@examplecorp.com"}
	cloudAccountRolesMap := map[string][]string{
		"alice.wonder@examplecorp.com": {"Alias0000001"},
		"bob.builder@examplecorp.com":  {"Alias0000001", "Alias0000002"},
		"jess.builder@examplecorp.com": {},
	}
	unexpectedUsers := []string{
		"charlie.choco@examplecorp.com",
		"diana.dreams@examplecorp.com",
	}
	ctx := CreateContextWithToken("elsadmin@intel.com")

	request := &pb.ListUsersByCloudAccountRequest{
		CloudAccountId: cloudAccountId,
	}

	// When call the ListUsersByCloudAccount function
	response, err := client.ListUsersByCloudAccount(ctx, request)

	if err != nil {
		t.Fatalf("listUsersByCloudAccount returned an error: %v", err)
	}

	userIds := []string{}
	for _, user := range response.Users {
		// verify assigned cloudAccountRoles
		expectedcloudAccountRolesAlias := cloudAccountRolesMap[user.Id]
		userCloudAccountRoleAlias := []string{}
		for _, userCloudAccount := range user.CloudAccountRoles {
			userCloudAccountRoleAlias = append(userCloudAccountRoleAlias, userCloudAccount.Alias)
		}
		// Then expect that the response contains only the expected users.
		if !containsOnly(userCloudAccountRoleAlias, expectedcloudAccountRolesAlias) {
			t.Errorf("unexpected cloudAccountRoleAlias: got %v, want %v", expectedcloudAccountRolesAlias, expectedcloudAccountRolesAlias)
		}
		userIds = append(userIds, user.Id)
	}

	// Then expect that the response contains only the expected users.
	if !containsOnly(userIds, expectedUsers) {
		t.Errorf("ListUsersByCloudAccount returned unexpected users: got %v, want %v", userIds, expectedUsers)
	}
	// Then expect that the response does not contain any unexpected users.
	if containsAny(userIds, unexpectedUsers) {
		t.Errorf("ListUsersByCloudAccount returned users from other cloud accounts: got %v, should not contain %v", userIds, unexpectedUsers)
	}
}

func TestListUsersByCloudAccountWildcard(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "*", Subject: "max.power@examplecorp.com", SystemRole: "intel_admin"},
		{CloudAccountId: "*", Subject: "lucas.king@examplecorp.com", SystemRole: "intel_admin"},
		{CloudAccountId: "*", Subject: "olivia.queen@examplecorp.com", SystemRole: "intel_admin"},
		{CloudAccountId: "168310882023", Subject: "annie@examplecorp.com", SystemRole: "cloud_account_admin"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}
	// Given
	cloudAccountId := "*"
	expectedUsers := []string{"max.power@examplecorp.com", "lucas.king@examplecorp.com", "olivia.queen@examplecorp.com"}
	unexpectedUsers := []string{"annie@examplecorp.com"}

	// only an intel_admin can execute cloudAccountId as *
	ctx := CreateContextWithToken("max.power@examplecorp.com")

	request := &pb.ListUsersByCloudAccountRequest{
		CloudAccountId: cloudAccountId,
	}

	// When call the ListUsersByCloudAccount function
	response, err := client.ListUsersByCloudAccount(ctx, request)

	if err != nil {
		t.Fatalf("listUsersByCloudAccount returned an error: %v", err)
	}

	userIds := []string{}
	for _, user := range response.Users {
		userIds = append(userIds, user.Id)
	}

	// Then expect that the expected users are all present in the response.
	if !allExpectedUsersPresent(userIds, expectedUsers) {
		t.Errorf("listUsersByCloudAccount did not return all expected users: got %v, want at least %v", userIds, expectedUsers)
	}
	// Then expect that none of the unexpected users are not present in the response.
	if containsAny(userIds, unexpectedUsers) {
		t.Errorf("ListUsersByCloudAccount returned users from other cloud accounts: got %v, should not contain %v", userIds, unexpectedUsers)
	}
}

// containsOnly checks if all elements in got are present in want and no extras.
func containsOnly(got, want []string) bool {
	gotSet := make(map[string]struct{})
	for _, user := range got {
		gotSet[user] = struct{}{}
	}

	for _, user := range want {
		if _, exists := gotSet[user]; !exists {
			return false
		}
		delete(gotSet, user)
	}

	return len(gotSet) == 0
}

// containsAny checks if any of the elements in got are present in notWant.
func containsAny(got, notWant []string) bool {
	gotSet := make(map[string]struct{})
	for _, user := range got {
		gotSet[user] = struct{}{}
	}

	for _, user := range notWant {
		if _, exists := gotSet[user]; exists {
			return true
		}
	}

	return false
}

// allExpectedUsersPresent checks if all the elements in got are present in want.
func allExpectedUsersPresent(responseUsers, expectedUsers []string) bool {
	responseSet := make(map[string]struct{})
	for _, user := range responseUsers {
		responseSet[user] = struct{}{}
	}

	for _, expectedUser := range expectedUsers {
		if _, exists := responseSet[expectedUser]; !exists {
			return false
		}
	}

	return true
}

func TestGetUserRoles(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	systemRoles := []*pb.RoleRequest{
		{CloudAccountId: "*", Subject: "maria.perez1@intel.com", SystemRole: "intel_admin"},
		{CloudAccountId: "168310882021", Subject: "maria.perez1@intel.com", SystemRole: "cloud_account_admin"},
		{CloudAccountId: "168310882021", Subject: "john.doe1@examplecorp.com", SystemRole: "cloud_account_member"},
		{CloudAccountId: "168310882024", Subject: "jane.smith1@examplecorp.com", SystemRole: "cloud_account_admin"},
	}

	err := assignSystemRolesBulk(client, systemRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}

	// creating permission to allow
	cloudAccountRoles := []*pb.CloudAccountRole{
		{Alias: "Alias7344", CloudAccountId: "168310882021", Effect: pb.CloudAccountRole_allow, Permissions: []*pb.CloudAccountRole_Permission{{ResourceType: "instance", ResourceId: "instance001", Actions: []string{"read", "update"}}, {ResourceType: "instance", ResourceId: "instance003", Actions: []string{"read", "update"}}}, Users: []string{"john.doe1@examplecorp.com"}},
	}

	err = createCloudAccountRolesRoleBulk(client, cloudAccountRoles)
	if err != nil {
		t.Fatalf("failed creating system roles for test (%v)", err)
	}
	ctx := CreateContextWithToken("maria.perez1@intel.com")
	expectedRoles := []string{"cloud_account_member"}

	// When calling GetUserRolesByCloudAccount for a specific user and cloud account
	userRolesResp, err := client.GetUser(ctx, &pb.GetUserRequest{
		CloudAccountId: "168310882021",
		UserId:         "john.doe1@examplecorp.com",
	})
	if err != nil {
		t.Fatalf("error executing GetUserRolesByCloudAccount (%v)", err)
	}
	expectedRolesSet := make(map[string]bool)
	for _, role := range expectedRoles {
		expectedRolesSet[role] = true
	}

	userRolesSet := make(map[string]bool)
	for _, role := range userRolesResp.SystemRoles {
		userRolesSet[role] = true
	}

	// Then Expect to return all roles assigned to the user
	if !reflect.DeepEqual(userRolesSet, expectedRolesSet) {
		t.Fatalf("expected roles: %v, got: %v", expectedRoles, userRolesResp.SystemRoles)
	}

	// Then expect to contain the Alias for the assign cloudAccountRole
	if len(userRolesResp.CloudAccountRoles) != 1 &&
		userRolesResp.CloudAccountRoles[0].Alias != cloudAccountRoles[0].Alias {
		t.Fatalf("expected alias to be part of the user response: %v, got: %v", userRolesResp.CloudAccountRoles, cloudAccountRoles[0])
	}
}

func TestDefaultCloudAccountRoleAssigned(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	// When
	result, err := client.DefaultCloudAccountRoleAssigned(context.Background(), &pb.DefaultCloudAccountRoleAssignedRequest{CloudAccountId: "168310882075"})
	// Then

	if err != nil {
		t.Fatalf("error executing DefaultCloudAccountRoleAssigned (%v)", err)
	}

	if result.Assigned {
		t.Fatalf("expected default cloud account role to not be assigned")
	}
}

func TestAssignDefaultCloudAccountRole(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	ctx := CreateContextWithToken("admin@intel.com")

	// Given
	cloudAccountId := "168310992075"
	// When
	_, err := client.AssignDefaultCloudAccountRole(ctx, &pb.AssignDefaultCloudAccountRoleRequest{
		CloudAccountId: cloudAccountId,
		Members:        []string{"member@intel.com"},
		Admins:         []string{"admin@intel.com"},
	})

	// Then
	if err != nil {
		t.Fatalf("expected error when assigning default cloud account role, but got none")
	}

	// assert users
	users, err := client.ListUsersByCloudAccount(ctx, &pb.ListUsersByCloudAccountRequest{CloudAccountId: cloudAccountId})

	if err != nil {
		t.Fatalf("error executing ListUsersByCloudAccount (%v)", err)
	}

	if len(users.Users) != 2 {
		t.Fatalf("expected 2 users to be assigned to the cloud account, but have: %v", len(users.Users))
	}

	// assert roles
	cloudAccountRole, err := client.QueryCloudAccountRoles(ctx, &pb.CloudAccountRoleQuery{CloudAccountId: cloudAccountId})

	if err != nil {
		t.Fatalf("error executing QueryCloudAccountRoles (%v)", err)
	}

	if len(cloudAccountRole.CloudAccountRoles) != 1 {
		t.Fatalf("expected 1 cloud account role to be assigned to the cloud account")
	}

	if cloudAccountRole.CloudAccountRoles[0].Effect != pb.CloudAccountRole_allow {
		t.Fatalf("expected cloud account role to be allow")
	}

	if len(cloudAccountRole.CloudAccountRoles[0].Users) != 1 {
		t.Fatalf("expected 1 user to be assigned to the cloud account role")
	}

	if cloudAccountRole.CloudAccountRoles[0].Users[0] != "member@intel.com" {
		t.Fatalf("expected user to be member@intel.com")
	}

	if len(cloudAccountRole.CloudAccountRoles[0].Permissions) < 1 {
		t.Fatalf("expected at least 1 permission to be assigned to the cloud account role")
	}
}

func TestAssignDefaultCloudAccountRoleIdempotent(t *testing.T) {
	client := pb.NewAuthzServiceClient(test.ClientConn())

	// Given
	// When
	_, err := client.AssignDefaultCloudAccountRole(context.Background(), &pb.AssignDefaultCloudAccountRoleRequest{
		CloudAccountId: "168310882075",
		Members:        []string{"member@intel.com"},
		Admins:         []string{"admin@intel.com"},
	})

	if err != nil {
		t.Fatalf("expected error when assigning default cloud account role, but got none")
	}

	result, err := client.AssignDefaultCloudAccountRole(context.Background(), &pb.AssignDefaultCloudAccountRoleRequest{
		CloudAccountId: "168310882075",
		Members:        []string{"member@intel.com"},
		Admins:         []string{"admin@intel.com"},
	})

	if err != nil {
		t.Fatalf("expected error when assigning default cloud account role, but got none")
	}

	if result == nil {
		t.Fatalf("expected result to not be nil")
	}
}

func assignSystemRolesBulk(client pb.AuthzServiceClient, roles []*pb.RoleRequest) error {
	for _, role := range roles {
		_, err := client.AssignSystemRole(context.Background(), role)
		if err != nil {
			return err
		}
	}
	return nil
}

func createCloudAccountRolesRoleBulk(client pb.AuthzServiceClient, cloudAccountroles []*pb.CloudAccountRole) error {
	for _, cloudAccountrole := range cloudAccountroles {
		_, err := client.CreateCloudAccountRole(context.Background(), cloudAccountrole)
		if err != nil {
			return err
		}
	}
	return nil
}
func slicesEqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aMap := make(map[string]int)
	for _, item := range a {
		aMap[item]++
	}
	for _, item := range b {
		if _, ok := aMap[item]; !ok {
			return false
		}
		aMap[item]--
		if aMap[item] == 0 {
			delete(aMap, item)
		}
	}
	return len(aMap) == 0
}

func TestCasbinWatcher(t *testing.T) {

	a, err := NewAdapter(test.mdb.DatabaseURL)
	if err != nil {
		logger.Error(err, "failed to create new adapter", "databaseURL", test.mdb.DatabaseURL.String())
		t.Fatalf("failed to create new adapter(%v)", err)
	}

	// watcher to keep all casbin instances syncronized
	w, err := psqlwatcher.NewWatcherWithConnString(context.Background(), test.mdb.DatabaseURL.String(),
		psqlwatcher.Option{NotifySelf: false, Verbose: true, LocalID: "4304dd87-1aa9-4cbc-b895-f46e1f3f4e67"})
	if err != nil {
		t.Fatalf("failed to set NewWatcherWithConnString (%v)", err)
	}

	// Initialize the Enforcer with the model and the Gorm adapter
	enforcer, err := casbin.NewEnforcer(test.cfg.ModelFilePath, a)
	if err != nil {
		t.Fatalf("failed to create new enforcer (%v)", err)
	}

	// set the watcher for enforcer.
	err = enforcer.SetWatcher(w)
	if err != nil {
		t.Fatalf("failed to set new watcher (%v)", err)
	}

	// set the default callback to handle policy changes.
	err = w.SetUpdateCallback(func(string) {
		// Attempt to reload policies
		err := enforcer.LoadPolicy()
		if err != nil {
			t.Fatalf("failed watcher callback (%v)", err)
		}
	})
	if err != nil {
		t.Fatalf("failed to SetUpdateCallback on watcher (%v)", err)
	}

	// local instance of casbin adds two group policy
	enforcer.AddGroupingPolicy("person1@examplecorp.com", "cloud_account_member", "168310882075")
	enforcer.AddGroupingPolicy("person2@examplecorp.com", "cloud_account_member", "168310882075")

	time.Sleep(1 * time.Second)

	// verify that roles were created on remote instance
	client := pb.NewAuthzServiceClient(test.ClientConn())
	roleExists1, _ := client.SystemRoleExists(context.Background(), &pb.RoleRequest{CloudAccountId: "168310882075", Subject: "person1@examplecorp.com", SystemRole: "cloud_account_member"})
	roleExists2, _ := client.SystemRoleExists(context.Background(), &pb.RoleRequest{CloudAccountId: "168310882075", Subject: "person2@examplecorp.com", SystemRole: "cloud_account_member"})
	if !roleExists1.Exist || !roleExists2.Exist {
		t.Fatalf("roles were not created on remote casbin instance")
	}

	// remote instance of casbin adds a new system role as a groupPolicy
	_, err = client.AssignSystemRole(context.Background(), &pb.RoleRequest{CloudAccountId: "168310882075", Subject: "b@hello.com", SystemRole: "cloud_account_admin"})
	if err != nil {
		t.Fatalf("error when assign systemrole")
	}

	_, err = client.UnassignSystemRole(context.Background(), &pb.RoleRequest{CloudAccountId: "168310882075", Subject: "person1@examplecorp.com", SystemRole: "cloud_account_member"})
	if err != nil {
		t.Fatalf("error when unassign systemrole")
	}

	// add grouping policy
	enforcer.AddGroupingPolicy("person3@examplecorp.com", "cloud_account_member", "168310882075")

	// remote instance of casbin adds a new system role as a groupPolicy
	_, err = client.AssignSystemRole(context.Background(), &pb.RoleRequest{CloudAccountId: "168310882075", Subject: "al@hello.com", SystemRole: "cloud_account_admin"})
	if err != nil {
		t.Fatalf("error when assign systemrole")
	}

	// verification against db
	db, err := sql.Open("postgres", test.mdb.DatabaseURL.String())
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	rows, err := db.Query("SELECT ptype,v0,v1,v2 from CASBIN_RULE")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var dbcolums [][]string
	for rows.Next() {
		var column0, column1, column2, column3 string
		err := rows.Scan(&column0, &column1, &column2, &column3)
		if err != nil {
			log.Fatal(err)
		}
		if column0 == "g" {
			dbcolums = append(dbcolums, []string{column1, column2, column3})
		}
	}

	// wait for watcher to update the instances
	time.Sleep(1 * time.Second)

	remotePolicies := enforcer.GetGroupingPolicy()
	localPolicies := test.casbinEngine.GetGroupingPolicy()

	// test match local policies against remote policies and db policies
	for i, dbpolicies := range dbcolums {
		for j, dbpolicy := range dbpolicies {
			if dbpolicy != remotePolicies[i][j] && dbpolicy != localPolicies[i][j] {
				t.Fatalf("all policies should match")
			}
		}
	}
}
