// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"database/sql"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type FailedOtpLogs struct {
}

var failedOtpLogs FailedOtpLogs

func (fol *FailedOtpLogs) CountAttempts(ctx context.Context, dbsession *sql.DB, otp_type pb.OtpType, cloud_account_id string, minutes int) (int, error) {
	// CountAttempts retrieves the number of failed OTP attempts for a specific type and cloud account within a given time interval.
	//
	// Parameters:
	// ctx (context.Context): The context for the operation.
	// dbsession (*sql.DB): The database session to use for querying the OTP logs.
	// otp_type (int): The type of OTP (e.g., INVITATION_VALIDATE).
	// cloud_account_id (string): The ID of the cloud account.
	// minutes (int): The time interval in minutes to consider for failed attempts.
	//
	// Returns:
	// int: The number of failed OTP attempts within the specified time interval.
	// error: An error if any occurred during the operation, nil otherwise.

	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("FailedOtpLogs.CountAttempts").WithValues("cloudAccountId", cloud_account_id).Start()
	defer span.End()
	logger.Info("Begin", "otp_type", otp_type)
	defer logger.Info("END")
	count := 0
	var interval_latest_timestamp sql.NullString
	query := `
	SELECT MAX(requested_at) AS latest_timestamp FROM otp_logs
	WHERE requested_at >= NOW() - make_interval(mins := $1) and
	log_type = $2 and
	cloud_account_id = $3
	`
	err := dbsession.QueryRowContext(ctx, query, minutes, otp_type, cloud_account_id).Scan(&interval_latest_timestamp)
	if err != nil {
		logger.Error(err, "failed to fetch lastone minute timestamp")
		return 0, err
	}

	if interval_latest_timestamp.Valid {
		query = `SELECT COUNT(*) FROM otp_logs where requested_at > $1::TIMESTAMPTZ - make_interval(mins := $2) and
		log_type = $3 and
		cloud_account_id = $4
		`
		logger.Info(query)
		err = dbsession.QueryRowContext(ctx, query, interval_latest_timestamp.String, minutes, otp_type, cloud_account_id).Scan(&count)
		if err != nil {
			logger.Error(err, "failed to fetch count")
			return 0, err
		}
	}

	return count, nil
}

func (fol *FailedOtpLogs) CheckThresholdReached(ctx context.Context, dbsession *sql.DB, otp_type pb.OtpType, cloud_account_id string) (bool, int, error) {
	// CheckThresholdReached checks if the number of failed OTP attempts for a specific type and cloud account
	// has reached the configured threshold within the specified interval.
	//
	// Parameters:
	// ctx (context.Context): The context for the operation.
	// dbsession (*sql.DB): The database session to use for querying the OTP logs.
	// otp_type (int): The type of OTP (e.g., INVITATION_VALIDATE).
	// cloud_account_id (string): The ID of the cloud account.
	//
	// Returns:
	// bool: True if the threshold has been reached, false otherwise.
	// int: The number of failed OTP attempts within the specified interval left.
	// error: An error if any occurred during the operation, nil otherwise.

	threshold := int(config.Cfg.OtpRetryLimit)
	minutes := int(config.Cfg.OtpRetryLimitIntervalDuration)
	count, err := fol.CountAttempts(ctx, dbsession, otp_type, cloud_account_id, minutes)
	var retry_left int = threshold - count
	if retry_left < 0 {
		retry_left = 0
	}
	if err != nil {
		return false, retry_left, err
	}
	if count >= threshold {
		return true, retry_left, nil
	}
	return false, retry_left, nil
}

func (fol *FailedOtpLogs) WriteAttempt(ctx context.Context, dbsession *sql.DB, otp_type pb.OtpType, cloud_account_id string) error {
	// WriteAttempt writes a new OTP log entry to the database.
	// It first deletes old log entries (older than 10 minutes) to maintain a clean database.
	// Then, it inserts a new row into the otp_logs table with the provided OTP type and cloud account ID.
	//
	// Parameters:
	// ctx (context.Context): The context for the operation.
	// dbsession (*sql.DB): The database session to use for querying the OTP logs.
	// otp_type (int): The type of OTP (e.g., INVITATION_VALIDATE).
	// cloud_account_id (string): The ID of the cloud account.
	//
	// Returns:
	// error: An error if any occurred during the operation, nil otherwise.

	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("FailedOtpLogs.WriteAttempt").WithValues("cloudAccountId", cloud_account_id).Start()
	defer span.End()
	logger.Info("Begin", "otp_type", otp_type)
	defer logger.Info("END")
	fol.DeleteOldAttempts(ctx, dbsession)
	query := "INSERT INTO otp_logs (log_type,cloud_account_id,requested_at) VALUES ($1, $2, NOW())"
	_, err := dbsession.ExecContext(ctx, query, otp_type, cloud_account_id)
	if err != nil {
		logger.Error(err, "failed to insert row")
		return err
	}
	return nil
}

func (fol *FailedOtpLogs) DeleteOldAttempts(ctx context.Context, dbsession *sql.DB) {
	// DeleteOldAttempts deletes old OTP log entries from the database.
	// It removes entries that are older than 10 minutes to maintain a clean database.
	//
	// Parameters:
	// ctx (context.Context): The context for the operation.
	// dbsession (*sql.DB): The database session to use for querying the OTP logs.
	//
	// Returns:
	// This function does not return any value.

	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("FailedOtpLogs.DeleteOldAttempts").Start()
	defer span.End()
	logger.Info("Begin")
	defer logger.Info("END")
	query := "delete FROM otp_logs where requested_at < NOW() - INTERVAL '10 minute'"
	_, err := dbsession.ExecContext(ctx, query)
	if err != nil {
		logger.Error(err, "failed to delete old log")
	}
}
