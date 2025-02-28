// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/lib/pq"
)

type AuditLogging struct {
	db      *sql.DB
	enabled bool
}

type LoggingParams struct {
	id                  string
	cloudAccountId      string
	cloudAccountRoleIds []string
	eventType           string
	additionalInfo      map[string]interface{}
}

func NewAuditLogging(db *sql.DB, enabled bool) (*AuditLogging, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &AuditLogging{
		db:      db,
		enabled: enabled,
	}, nil
}

func (s *AuditLogging) Logging(ctx context.Context, params LoggingParams) {
	if s.enabled {
		ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AuditLogging.Logging").Start()
		defer span.End()
		logger.Info("BEGIN")
		defer logger.Info("END")

		jsonStr, jsonErr := json.Marshal(params.additionalInfo)
		if jsonErr != nil {
			logger.Error(jsonErr, "error marshaling additional info")
		}
		if params.id == "" {
			params.id = uuid.NewString()
		}

		var id string

		if params.eventType == "" {
			var existingEventType sql.NullString
			checkQuery := "SELECT event_type FROM audit WHERE id = $1"
			err := s.db.QueryRowContext(ctx, checkQuery, params.id).Scan(&existingEventType)
			if err != nil && err != sql.ErrNoRows {
				logger.Error(err, "error checking existing log")
				return

			}
			if existingEventType.Valid {
				params.eventType = existingEventType.String
			} else {
				logger.Error(err, "error inserting the log, no event type found", "executionId", params.id)
			}
		}

		if len(params.cloudAccountRoleIds) == 0 {
			var existingRoleIds pq.StringArray
			checkQuery := "SELECT cloud_account_role_ids FROM audit WHERE id = $1"
			err := s.db.QueryRowContext(ctx, checkQuery, params.id).Scan(&existingRoleIds)
			if err != nil && err != sql.ErrNoRows {
				logger.Error(err, "error checking existing log for cloud_account_role_ids")
				return
			}
			params.cloudAccountRoleIds = existingRoleIds
		}

		query := "INSERT INTO audit (id, cloud_account_id, cloud_account_role_ids, event_type, additional_info) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO UPDATE SET additional_info = jsonb_set(audit.additional_info, '{more_info}', $5, true) RETURNING id"

		err := s.db.QueryRowContext(ctx, query, params.id, params.cloudAccountId, pq.Array(params.cloudAccountRoleIds), params.eventType, jsonStr).Scan(&id)

		if err != nil {
			logger.Error(err, "error inserting the log", "executionId", params.id)
		} else {
			logger.Info("log successfully added")
		}
	}
}
