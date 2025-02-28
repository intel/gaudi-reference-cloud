// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestNullDatabase(t *testing.T) {
	// Given a nil database

	//When initializing the audit logger
	_, err := NewAuditLogging(nil, true)

	//Expect an error
	if err == nil {
		t.Fatalf("the initialization should fail")
	}
}

func TestWrongAdditionalInfo(t *testing.T) {
	_ = test.ClientConn()
	ctx := context.Background()

	// Given an structure that refuses to serialize as additionalInfo
	additionalInfo := map[string]interface{}{"allowed": make(chan int)}
	decisionId := uuid.NewString()
	logginParams := LoggingParams{id: decisionId, cloudAccountId: "856295372056", eventType: "check", additionalInfo: additionalInfo}

	// When trying to log it
	test.auditLogging.enabled = true
	test.auditLogging.Logging(ctx, logginParams)

	// Expect a marshaling error, but log should exist even without additional info
	// Expect the logexists function to return true
	result := LogExists(test.auditLogging, decisionId)
	if result != true {
		t.Fatalf("the log should be created even with nil additional info")
	}
}

func TestAuditDisabledLogging(t *testing.T) {
	_ = test.ClientConn()
	ctx := context.Background()

	// Given the following log information
	decisionId := uuid.NewString()
	additionalInfo := map[string]interface{}{"allowed": true}
	logginParams := LoggingParams{id: decisionId, cloudAccountId: "856295372056", eventType: "check", additionalInfo: additionalInfo}

	// When the audit log is disabled, and a log is created
	test.auditLogging.enabled = false
	test.auditLogging.Logging(ctx, logginParams)

	// Then expect the logexists function to return false
	result := LogExists(test.auditLogging, decisionId)
	if result == true {
		t.Fatalf("the function should have returned false")
	}
}

func TestAuditLoggingUpdate(t *testing.T) {
	_ = test.ClientConn()
	ctx := context.Background()
	test.auditLogging.enabled = true

	// Given the following log information
	decisionId := uuid.NewString()
	additionalInfo := map[string]interface{}{"allowed": true}
	logginParams := LoggingParams{id: decisionId, cloudAccountId: "856295372056", eventType: "check", additionalInfo: additionalInfo}

	// When the log is created 2 times with the same ID
	test.auditLogging.Logging(ctx, logginParams)
	cloudAccountRoleIds := []string{uuid.NewString()}
	logginParams = LoggingParams{id: decisionId, cloudAccountRoleIds: cloudAccountRoleIds}
	test.auditLogging.Logging(ctx, logginParams)

	// Then expect the logexists function to return true
	result := LogExists(test.auditLogging, decisionId)
	if result == false {
		t.Fatalf("the function should have returned true")
	}
}

func TestAuditLogDoesNotExists(t *testing.T) {
	_ = test.ClientConn()
	ctx := context.Background()
	test.auditLogging.enabled = true

	// Given that the log table should be empty

	// When asking if a log exists (new uuid)
	decisionId := uuid.NewString()
	result := LogExists(test.auditLogging, decisionId)

	// Then expect the function to return false
	if result == true {
		t.Fatalf("the function should have returned false")
	}

	// Given that a log is created
	additionalInfo := map[string]interface{}{"allowed": true}
	logginParams := LoggingParams{id: decisionId, cloudAccountId: "856295372056", eventType: "check", additionalInfo: additionalInfo}
	test.auditLogging.Logging(ctx, logginParams)

	// When the function is called with another or non existent uuid
	decisionId = uuid.NewString()
	result = LogExists(test.auditLogging, decisionId)

	// Then expect the function to return false
	if result == true {
		t.Fatalf("the function should have returned false")
	}
}

func TestAuditLogExists(t *testing.T) {
	_ = test.ClientConn()
	ctx := context.Background()
	test.auditLogging.enabled = true

	// Given the following log information
	decisionId := uuid.NewString()
	additionalInfo := map[string]interface{}{"allowed": true}
	logginParams := LoggingParams{id: decisionId, cloudAccountId: "856295372056", eventType: "check", additionalInfo: additionalInfo}

	// When the log is created
	test.auditLogging.Logging(ctx, logginParams)

	// Then expect the function to return true
	result := LogExists(test.auditLogging, decisionId)
	if result == false {
		t.Fatalf("the function should have returned true")
	}
}

func LogExists(s *AuditLogging, decisionId string) bool {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM audit WHERE id = $1)"
	err := s.db.QueryRow(query, decisionId).Scan(&exists)
	if err != nil {
		logger.Error(err, "error checking if the log exists", "decisionId", decisionId)
	}
	return exists
}
