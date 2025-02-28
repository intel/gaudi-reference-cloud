// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/assert"
)

func TestPingCreditService(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestPingCreditService")
	logger.Info("BEGIN")
	defer logger.Info("End")

	_, err := creditsClient.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("failed to ping cloudcredit service: %v", err)
	}
}
func TestGetRedemptions(t *testing.T) {
	// Create a new mock SQL database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Define your service
	service := &MockCloudCreditsService{
		RedemptionsFunc: func(ctx context.Context, redeemedDate time.Time) ([]*pb.CouponRedemption, error) {
			// Return mock data or error as needed for your test
			return []*pb.CouponRedemption{
				{
					Code:           "mock-code",
					CloudAccountId: "mock-account-id",
					Redeemed:       timestamppb.New(redeemedDate),
					Installed:      true,
				},
			}, nil
		},
	}

	// Define the redeemedDate for the test
	redeemedDate := time.Now()

	// Set up the expected rows that the query will return
	rows := sqlmock.NewRows([]string{"code", "cloud_account_id", "redeemed", "installed"}).
		AddRow("test-code", "test-cloud-account-id", redeemedDate, true)

	// Expect a query to execute with the correct SQL and arguments
	mock.ExpectQuery("SELECT code, cloud_account_id, redeemed, installed FROM redemptions WHERE redeemed >= \\$1").
		WithArgs(redeemedDate.Format(time.RFC3339)).
		WillReturnRows(rows)

	// Call the function under test
	redemptions, err := service.GetRedemptions(context.Background(), redeemedDate)

	// Assert that there was no error returned
	assert.NilError(t, err)

	// Assert that the correct number of redemptions were returned
	assert.Equal(t, len(redemptions), 1)

	// Assert that the redemption data is correct
	assert.Equal(t, redemptions[0].Code, "test-code")
	assert.Equal(t, redemptions[0].CloudAccountId, "test-cloud-account-id")
	assert.DeepEqual(t, redemptions[0].Redeemed, timestamppb.New(redeemedDate))
	assert.Equal(t, redemptions[0].Installed, true)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

type CloudCreditsServiceInterface interface {
	GetRedemptions(ctx context.Context, redeemedDate time.Time) ([]*pb.CouponRedemption, error)
}

type MockCloudCreditsService struct {
	CloudCreditsServiceInterface
	RedemptionsFunc func(ctx context.Context, redeemedDate time.Time) ([]*pb.CouponRedemption, error)
}

// GetRedemptions calls the mock implementation function.
func (m *MockCloudCreditsService) GetRedemptions(ctx context.Context, redeemedDate time.Time) ([]*pb.CouponRedemption, error) {
	return m.RedemptionsFunc(ctx, redeemedDate)
}
