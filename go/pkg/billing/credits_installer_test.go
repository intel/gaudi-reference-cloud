// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreditsInstaller(t *testing.T) {
	clientCredit := pb.NewBillingCreditServiceClient(clientConn)
	ctx, cancel := context.WithCancel(context.Background())
	creditsInstallSchedulerInterval := uint16(1)
	creditsExpiryMinimumInterval := uint16(31)
	client := pb.NewBillingCouponServiceClient(clientConn)
	creator := "idc-admin"
	currentTime := timestamppb.Now()
	intelUser := "std_" + uuid.NewString() + "example.com"

	intelAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  intelUser,
		Owner: intelUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_INTEL,
	})
	intelCloudAccountId := intelAcct.GetId()

	// Create Coupon
	coupon, err := client.Create(context.Background(),
		&pb.BillingCouponCreate{
			Expires: timestamppb.New(currentTime.AsTime().AddDate(0, 0, 90)),
			Amount:  500,
			NumUses: 4,
			Creator: creator,
		})
	if err != nil {
		t.Fatalf("create coupon failed: %v", err)
	}

	err = CheckCouponCode(coupon.Code)
	if err != nil {
		t.Errorf("%v", err)
	}

	// Redeem coupon and disable the intantaneous installation of coupons
	TestScheduler = true
	_, err = client.Redeem(context.Background(),
		&pb.BillingCouponRedeem{
			Code:           coupon.Code,
			CloudAccountId: intelCloudAccountId,
		})
	if err != nil {
		t.Errorf("redeem coupon: %v", err)
	}

	db, err := Test.ManagedDb.Open(ctx)
	if err != nil {
		t.Fatalf("failed to open the database connection: %v", err)
	}
	creditsInstallSched, err := NewCreditsInstallScheduler(db, creditsInstallSchedulerInterval, creditsExpiryMinimumInterval)
	if err != nil {
		t.Fatalf("failed to create credits install scheduler: %v", err)
	}

	stopChan := make(chan struct{})
	done := make(chan struct{})

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(1)
		SyncWait.Store(&wg)
		defer SyncWait.Store(nil)
		defer wg.Done()
		defer close(done)

		creditsInstallSched.StartCreditsInstallScheduler(ctx)

		// Wait for the stop signal or context cancellation
		select {
		case <-stopChan:
			creditsInstallSched.StopCreditsInstallScheduler()
		case <-ctx.Done():
			// Context was canceled, stop the scheduler
			creditsInstallSched.StopCreditsInstallScheduler()
		}
	}()

	// Set a timer to stop the scheduler after a specific duration
	timer := time.NewTimer(3 * time.Second)

	// Wait for the timer or scheduler to complete
	select {
	case <-timer.C:
		close(stopChan)
	case <-done:
		timer.Stop()
	}

	// Read unapplied credit balance for the cloudaccountId
	resp, err := clientCredit.ReadUnappliedCreditBalance(ctx, &pb.BillingAccount{CloudAccountId: intelCloudAccountId})
	if err != nil {
		t.Fatalf("Read unapplied credit balance: %v", err)
	}
	assert.Equal(t, resp.UnappliedAmount, float64(500))

	// Cancel the context and wait for the scheduler to finish
	cancel()
	<-done
}
