// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/lib/pq"
)

func timeUntilNext(now time.Time, hour int) time.Duration {
	nextOccurrence := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())

	if now.After(nextOccurrence) {
		// if the current time is later than the next occurrence, schedule the next occurrence for the same hour tomorrow.
		nextOccurrence = nextOccurrence.AddDate(0, 0, 1)
	}

	return time.Until(nextOccurrence)
}

type InviteDetails struct {
	Id                uint64
	AdminAccountId    string
	AdminAccountEmail string
	MemberEmail       string
}

type InvitationsExpiryScheduler struct {
	cfg                 *config.Config
	db                  *sql.DB
	notificationClient  *NotificationClient
	ticker              *time.Ticker
	batchSize           int
	scheduledHour       int
	interruptionChannel chan bool
}

const (
	InvitationsExpiryGetAllInvitationsError             string = "invitations expiry: failed to get all invitations"
	InvitationsExpiryRowReadError                       string = "invitations expiry: failed to read invitation"
	InvitationsExpiryDeleteError                        string = "invitations expiry: failed to deleted expired invitations"
	InvitationsExpiryCloudAccountOwnerNotificationError string = "invitations expiry: failed to notify cloud account owner: %s"
)

func startInvitationsExpiryScheduler(s *InvitationsExpiryScheduler, ctx context.Context) {
	go s.Loop(ctx)
}

func NewInvitationsExpiryScheduler(cfg *config.Config, db *sql.DB, notificationClient *NotificationClient) *InvitationsExpiryScheduler {
	scheduledHour := int(cfg.InvitationsExpirySchedulerTime)
	now := time.Now()
	duration := timeUntilNext(now, scheduledHour)

	return &InvitationsExpiryScheduler{
		cfg:                 cfg,
		db:                  db,
		notificationClient:  notificationClient,
		ticker:              time.NewTicker(duration),
		scheduledHour:       scheduledHour,
		interruptionChannel: make(chan bool),
		batchSize:           cfg.InvitationsExpirySchedulerBatchSize,
	}
}

func (s *InvitationsExpiryScheduler) Loop(ctx context.Context) {
	_, logger, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvitationExpiryScheduler.Loop").Start()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	for {

		err := s.RemoveAndNotifyAll(ctx)

		if err != nil {
			logger.Error(err, InvitationsExpiryDeleteError)
		}

		select {
		case <-s.interruptionChannel:
			logger.Info("invitation expiry scheduler interrupted")
			return
		case tm := <-s.ticker.C:
			if tm.IsZero() {
				logger.Info("invitation expiry scheduler interrupted")
				return
			}
			// reset the ticker to tick next day at the defined hour
			s.ticker.Reset(timeUntilNext(time.Now(), s.scheduledHour))
		}
	}
}

func (s *InvitationsExpiryScheduler) RemoveAndNotifyAll(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvitationsExpiryScheduler.RemoveAndNotifyAll").Start()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	defer span.End()

	var totalAffected int64 = 0

	expiryTime := time.Now()

	selectExpiryInvitesQuery := fmt.Sprintf(`
		SELECT m.id, m.admin_account_id , acc.owner, m.member_email
			FROM members as m
			INNER JOIN cloud_accounts as acc ON acc.id = m.admin_account_id
			WHERE m.expiry < $1 AND m.invitation_state not in (%d, %d) ORDER BY m.id LIMIT $2`,
		pb.InvitationState_INVITE_STATE_ACCEPTED, pb.InvitationState_INVITE_STATE_REVOKED)

	deleteExpiryInvitesQuery := "DELETE FROM members WHERE id = ANY($1)"

	for {
		// TODO: golang 1.21 introduces clear() to optimize memory consumption
		inviteIds := make([]uint64, 0, s.batchSize)

		rows, err := s.db.Query(selectExpiryInvitesQuery, expiryTime, s.batchSize)

		if err != nil {
			logger.Error(err, InvitationsExpiryGetAllInvitationsError)
			return err
		}

		defer rows.Close()

		for rows.Next() {
			invite := &InviteDetails{}

			if err := rows.Scan(&invite.Id, &invite.AdminAccountId, &invite.AdminAccountEmail, &invite.MemberEmail); err != nil {
				logger.Error(err, InvitationsExpiryRowReadError)
				return err
			}

			if s.notificationClient != nil && s.cfg.GetInviteExpiryEmail() {
				// notify cloud account owner about the invitation expiry
				if err := s.notificationClient.SendInvitationExpiredEmail(ctx, "Invitation Expired", invite.AdminAccountEmail, invite.MemberEmail, s.cfg.GetInviteExpiredTemplate()); err != nil {
					logger.Error(err, fmt.Sprintf(InvitationsExpiryCloudAccountOwnerNotificationError, invite.AdminAccountId))
				}
			}

			// marks the invite id to be removed later in batches
			inviteIds = append(inviteIds, invite.Id)
		}

		expirationResult, err := s.db.ExecContext(ctx, deleteExpiryInvitesQuery, pq.Array(inviteIds))

		if err != nil {
			logger.Error(err, InvitationsExpiryDeleteError)
			return err
		}

		affected, _ := expirationResult.RowsAffected()

		totalAffected += affected

		if len(inviteIds) < s.batchSize {
			break
		}

	}

	logger.Info("invitations were removed", "count", fmt.Sprint(totalAffected), "expiryTime", expiryTime)

	return nil
}
