// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	config "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz/config"
)

type TestLoggingParams struct {
	id                  string
	cloudAccountId      string
	cloudAccountRoleIds []string
	eventType           string
	additionalInfo      map[string]interface{}
	eventDate           time.Time
}

const (
	baseSelectLogQuery = `SELECT COUNT(*) FROM audit WHERE event_date >= $1;`
)

func assertRemainingAuditLogState(t *testing.T, expectedCount int, cutoffTime time.Time) {
	const baseSelectLogQueryRemaining = `SELECT COUNT(*) FROM audit WHERE event_date >= $1;`

	var count int
	err := db.QueryRow(baseSelectLogQueryRemaining, cutoffTime).Scan(&count)
	if err != nil {
		t.Fatalf("failed to fetch remaining audit logs: %v", err)
	}

	if count != expectedCount {
		t.Fatalf("expected %d remaining audit logs, but got %d", expectedCount, count)
	}
}

func buildAuditLogScheduler(retentionPeriodInDays int) *AuditLoggingCleanupScheduler {
	cfg := config.NewDefaultConfig()
	cfg.Features.AuditLogging.SchedulerTime = 0 // run scheduler immediately for testing purposes

	scheduler := NewAuditLoggingCleanupScheduler(cfg, db)
	scheduler.retentionPeriodInDays = retentionPeriodInDays

	return scheduler
}

func TestAuditLogCleanup(t *testing.T) {
	_ = test.ClientConn()
	ctx := context.Background()
	test.auditLogging.enabled = true
	retentionPeriodInDays := 1

	now := time.Now()
	additionalInfo := map[string]interface{}{"allowed": true}

	logs := []TestLoggingParams{
		{id: uuid.NewString(), cloudAccountId: "acc1", eventType: "delete", additionalInfo: additionalInfo, eventDate: now.Add(-25 * time.Hour)}, // Should be deleted
		{id: uuid.NewString(), cloudAccountId: "acc2", eventType: "check", additionalInfo: additionalInfo, eventDate: now.Add(-23 * time.Hour)},  // Should be kept
		{id: uuid.NewString(), cloudAccountId: "acc3", eventType: "update", additionalInfo: additionalInfo, eventDate: now.Add(-24 * time.Hour)}, // Should be deleted
		{id: uuid.NewString(), cloudAccountId: "acc4", eventType: "create", additionalInfo: additionalInfo, eventDate: now},                      // Current log, should be kept
	}

	// Insert logs into audit table
	for _, log := range logs {
		query := `
			INSERT INTO audit (id, cloud_account_id, event_type, additional_info, event_date) 
			VALUES ($1, $2, $3, $4, $5);
		`
		_, err := db.ExecContext(ctx, query, log.id, log.cloudAccountId, log.eventType, log.additionalInfo, log.eventDate)
		if err != nil {
			t.Fatalf("failed to insert audit log: %v", err)
		}
	}

	// Run the scheduler cleanup
	scheduler := buildAuditLogScheduler(retentionPeriodInDays)
	err := scheduler.CleanUpAuditLogs(context.Background())
	if err != nil {
		t.Fatalf("scheduler failed: %v", err)
	}

	// Verify the state of the logs
	retentionPeriod := time.Hour * 24 * time.Duration(retentionPeriodInDays)
	cutoffTime := now.Add(-retentionPeriod)

	// Logs older than 24 hours should be deleted
	// Two logs should remain: the 23-hour and current ones
	assertRemainingAuditLogState(t, 2, cutoffTime)
}
