// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type InviteStatesTest struct {
	name       string
	expiration string
	initial    int
	expected   int // 0 = removed
}

const (
	baseEmail           = "idcs-invite-test-%s@intel.com"
	baseInsertSeedQuery = `INSERT INTO members (admin_account_id,member_email,invitation_code,invitation_state,expiry,notes) 
		VALUES ((SELECT id FROM cloud_accounts LIMIT 1),'%s', SUBSTR(MD5(RANDOM()::text), 0, 10), %d,%s,'');`
)

func createInviteSeed(state InviteStatesTest, t *testing.T) {
	email := fmt.Sprintf(baseEmail, state.name)

	_, err := db.Exec(fmt.Sprintf(baseInsertSeedQuery, email, state.initial, state.expiration))

	if err != nil {
		t.Fatal(err, "failed to seed invitations")
	}
}

func fetchInviteStateSeed(state InviteStatesTest, selectQuery string, t *testing.T) (string, int, string) {
	email := fmt.Sprintf(baseEmail, state.name)

	var current int
	var expiry string

	rows, err := db.Query(fmt.Sprintf(selectQuery, email))

	if err != nil {
		t.Fatal(err, "failed to run query that asserts the expiration result")
	}

	defer rows.Close()
	rows.Next()
	rows.Scan(&current, &expiry)
	return email, current, expiry
}

func buildScheduler() *InvitationsExpiryScheduler {
	cfg := config.NewDefaultConfig()

	cfg.InvitationsExpirySchedulerBatchSize = 100
	cfg.InvitationsExpirySchedulerTime = 0

	invitationsExpiryScheduler := NewInvitationsExpiryScheduler(cfg, db, nil)
	return invitationsExpiryScheduler
}

func assertStateSeeds(states *[]InviteStatesTest, t *testing.T) {
	// for every seeded invite, check the expected result after the scheduler execution

	selectQuery := "SELECT invitation_state, expiry FROM members WHERE member_email = '%s'"

	for _, state := range *states {
		email, current, expiry := fetchInviteStateSeed(state, selectQuery, t)

		if state.expected != current {
			t.Logf("failed invite %s with the state %d but expected %d. now is marked as %s, expiry %s", email, current, state.expected, time.Now(), expiry)
			t.Fail()
		}
	}
}

func TestMarkAndNotifyExpiryPossibleStatesInvitations(t *testing.T) {
	// given

	// base states
	states := []InviteStatesTest{
		{name: "pending", expiration: "NOW() - interval '1 day'", initial: int(pb.InvitationState_INVITE_STATE_PENDING_ACCEPT), expected: 0},
		{name: "pending-keep", expiration: "NOW() + interval '1 day'", initial: int(pb.InvitationState_INVITE_STATE_PENDING_ACCEPT), expected: int(pb.InvitationState_INVITE_STATE_PENDING_ACCEPT)},
		{name: "accepted", expiration: "NOW() - interval '1 day'", initial: int(pb.InvitationState_INVITE_STATE_ACCEPTED), expected: int(pb.InvitationState_INVITE_STATE_ACCEPTED)},
		{name: "revoked", expiration: "NOW() - interval '1 day'", initial: int(pb.InvitationState_INVITE_STATE_REVOKED), expected: int(pb.InvitationState_INVITE_STATE_REVOKED)},
		{name: "expired", expiration: "NOW() - interval '1 day'", initial: int(pb.InvitationState_INVITE_STATE_EXPIRED), expected: 0},
		{name: "rejected", expiration: "NOW() - interval '1 day'", initial: int(pb.InvitationState_INVITE_STATE_REJECTED), expected: 0},
		{name: "removed", expiration: "NOW() - interval '1 day'", initial: int(pb.InvitationState_INVITE_STATE_REMOVED), expected: 0},
	}

	for _, state := range states {
		createInviteSeed(state, t)
	}

	t.Logf("%d invites were created", len(states))

	invitationsExpiryScheduler := buildScheduler()

	// when
	// runs the scheduler
	err := invitationsExpiryScheduler.RemoveAndNotifyAll(context.Background())

	if err != nil {
		t.Fatal(err, "failed to expire invitations")
	}

	// then
	assertStateSeeds(&states, t)
}

func TestMarkAndNotifyExpiryBatchInvitations(t *testing.T) {
	// given

	states := []InviteStatesTest{}
	invitationsExpiryScheduler := buildScheduler()

	// generate invites
	for i := 0; i < invitationsExpiryScheduler.batchSize*3; i++ {
		piece := InviteStatesTest{name: fmt.Sprintf("pending-%d-", i), expiration: "NOW() - interval '1 day'", initial: int(pb.InvitationState_INVITE_STATE_PENDING_ACCEPT), expected: 0}
		states = append(states, piece)
	}

	for _, state := range states {
		createInviteSeed(state, t)
	}

	t.Logf("%d invites were created", len(states))

	// when
	// runs the scheduler
	err := invitationsExpiryScheduler.RemoveAndNotifyAll(context.Background())

	if err != nil {
		t.Fatal(err, "failed to expire invitations")
	}

	// then
	assertStateSeeds(&states, t)
}
