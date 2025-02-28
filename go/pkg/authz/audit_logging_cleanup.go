// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"
	"database/sql"

	"time"

	config "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

func timeUntilNext(now time.Time, hour int) time.Duration {
	nextOccurrence := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())

	if now.After(nextOccurrence) {
		// if the current time is later than the next occurrence, schedule the next occurrence for the same hour tomorrow.
		nextOccurrence = nextOccurrence.AddDate(0, 0, 1)
	}

	return time.Until(nextOccurrence)
}

type AuditLoggingCleanupScheduler struct {
	cfg                   *config.Config
	db                    *sql.DB
	ticker                *time.Ticker
	batchSize             int
	scheduledHour         int
	interruptionChannel   chan bool
	retentionPeriodInDays int
}

func startAuditLoggingCleanupScheduler(s *AuditLoggingCleanupScheduler, ctx context.Context) {
	go s.Loop(ctx)
}

func NewAuditLoggingCleanupScheduler(cfg *config.Config, db *sql.DB) *AuditLoggingCleanupScheduler {
	scheduledHour := int(cfg.Features.AuditLogging.SchedulerTime)
	retentionPeriodInDays := int(cfg.Features.AuditLogging.RetentionPeriodInDays)
	now := time.Now()
	duration := timeUntilNext(now, scheduledHour)

	return &AuditLoggingCleanupScheduler{
		cfg:                   cfg,
		db:                    db,
		ticker:                time.NewTicker(duration),
		scheduledHour:         scheduledHour,
		retentionPeriodInDays: retentionPeriodInDays,
		interruptionChannel:   make(chan bool),
	}
}

func (s *AuditLoggingCleanupScheduler) Loop(ctx context.Context) {
	_, logger, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuditLoggingCleanupScheduler.Loop").Start()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	for {

		err := s.CleanUpAuditLogs(ctx)

		if err != nil {
			logger.Error(err, "Failed to clean up audit logs")
		}

		select {
		case <-s.interruptionChannel:
			logger.Info("Audit logging cleanup scheduler interrupted")
			return
		case tm := <-s.ticker.C:
			if tm.IsZero() {
				logger.Info("Audit logging cleanup scheduler interrupted")
				return
			}
			// reset the ticker to tick next day at the defined hour
			s.ticker.Reset(timeUntilNext(time.Now(), s.scheduledHour))
		}
	}
}

func (s *AuditLoggingCleanupScheduler) CleanUpAuditLogs(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AuditLoggingCleanupScheduler.CleanupAuditLogs").Start()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	defer span.End()

	retentionPeriod := time.Hour * 24 * time.Duration(s.retentionPeriodInDays)
	cutoffTime := time.Now().Add(-retentionPeriod)
	logger.V(9).Info("Calculated cutoff time for log cleanup", "cutoffTime", cutoffTime)

	query := `DELETE FROM audit WHERE event_date < $1`
	result, err := s.db.ExecContext(ctx, query, cutoffTime)

	if err != nil {
		logger.Error(err, "Failed to clean up audit logs")
	}

	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		logger.Error(err, "Failed to get the number of deleted rows")
		return err
	}

	logger.Info("Cleaned up old audit logs", "rowsDeleted", rowsDeleted, "cutoffTime", cutoffTime, "retentionPeriodInDays", s.retentionPeriodInDays)
	return nil
}
