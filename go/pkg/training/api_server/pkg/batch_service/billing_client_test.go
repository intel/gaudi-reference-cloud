// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"
	"errors"
	"os"
	"testing"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestBillingClient(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution with env var", func(t *testing.T) {
		mockResolver := NewMockResolver(func(ctx context.Context, s string) (string, error) {
			// not actually connecting to anything, but needs to be "localhost" to not use TLS
			return "localhost:port", nil
		})

		// set environment variable
		os.Setenv("PRODUCTCATALOG_ADDR", "localhost:port")

		// run the function under test
		productClient, err := NewProductClient(ctx, mockResolver)
		assert.NoError(t, err)
		assert.NotNil(t, productClient)
	})

	t.Run("resolver error", func(t *testing.T) {
		mockResolver := NewMockResolver(func(ctx context.Context, s string) (string, error) {
			return "", errors.New("resolver error")
		})

		// set environment variable
		os.Setenv("PRODUCTCATALOG_ADDR", "")

		// run the function under test
		productClient, err := NewProductClient(ctx, mockResolver)
		assert.Error(t, err)
		assert.Nil(t, productClient)
	})
}

func TestGetCloudAccountCredits(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution", func(t *testing.T) {

		// Create timestamp instances for lastUpdated and expirationDate timestamps
		lastUpdatedTimestamp := &timestamppb.Timestamp{Seconds: 1711060625, Nanos: 884908000}
		expirationTimestamp := &timestamppb.Timestamp{Seconds: 1711060699}

		// Create mock BillingCredit instances
		createdTimestamp := &timestamppb.Timestamp{Seconds: 1711060625, Nanos: 882938000}
		expirationTimestampBillingCredit := &timestamppb.Timestamp{Seconds: 1711060699}

		mockBillingCredit := &v1.BillingCredit{
			Created:         createdTimestamp,
			Expiration:      expirationTimestampBillingCredit,
			CloudAccountId:  "test-account-id",
			Reason:          v1.BillingCreditReason_CREDIT_INITIAL,
			OriginalAmount:  1.0,
			RemainingAmount: 1.0,
			CouponCode:      "test-coupon-code",
		}

		// Create a slice of BillingCredit containing the mock instance
		mockBillingCredits := []*v1.BillingCredit{mockBillingCredit}

		mockGetCloudAccountCredits := func(ctx context.Context, in *v1.BillingCreditFilter, opts ...grpc.CallOption) (*v1.BillingCreditResponse, error) {
			return &v1.BillingCreditResponse{
				TotalRemainingAmount: 1.0,
				TotalUsedAmount:      0.0,
				TotalUnAppliedAmount: 0.0,
				LastUpdated:          lastUpdatedTimestamp,
				ExpirationDate:       expirationTimestamp,
				Credits:              mockBillingCredits,
			}, nil
		}

		mockBillingCreditServiceClient := NewMockBillingCreditServiceClient(mockGetCloudAccountCredits)
		billingClient := &BillingClient{
			BillingCreditServiceClient: mockBillingCreditServiceClient,
			BillingCouponServiceClient: nil,
		}

		// run the function under test
		creditsRes, err := billingClient.GetCloudAccountCredits(ctx, mockBillingCredit.CloudAccountId, true)
		assert.NoError(t, err)
		assert.Equal(t, mockBillingCredit.CouponCode, creditsRes[0].CouponCode)
	})

	t.Run("error in fetching billing credit client response", func(t *testing.T) {
		mockGetCloudAccountCredits := func(ctx context.Context, in *v1.BillingCreditFilter, opts ...grpc.CallOption) (*v1.BillingCreditResponse, error) {
			return nil, errors.New("error in fetching billing credit client response")
		}

		mockBillingCreditServiceClient := NewMockBillingCreditServiceClient(mockGetCloudAccountCredits)
		billingClient := &BillingClient{
			BillingCreditServiceClient: mockBillingCreditServiceClient,
			BillingCouponServiceClient: nil,
		}

		// run the function under test
		creditsRes, err := billingClient.GetCloudAccountCredits(ctx, "test-account-id", true)
		assert.Error(t, err)
		assert.Nil(t, creditsRes)

	})
}

func TestGetCouponExpiry(t *testing.T) {
	ctx := context.Background()

	t.Run("valid coupon check successful execution", func(t *testing.T) {

		// Create mock timestamp instances for created, start, expires, and actual timestamps
		createdTimestamp := &timestamppb.Timestamp{Seconds: 1711060129}
		startTimestamp := &timestamppb.Timestamp{Seconds: 1710964973}
		expiresTimestamp := &timestamppb.Timestamp{Seconds: 1742500400}
		actualTimestamp := &timestamppb.Timestamp{Seconds: 1742500300, Nanos: 884908000}

		// Create mock BillingCoupon instance
		mockBillingCoupon := &v1.BillingCoupon{
			Code:        "test-code",
			Creator:     "test-id",
			Created:     createdTimestamp,
			Start:       startTimestamp,
			Expires:     expiresTimestamp,
			Amount:      1.0,
			NumUses:     100,
			NumRedeemed: 1,
			IsStandard:  nil,
		}

		// Create a slice of BillingCoupon containing the mock instance
		mockBillingCoupons := []*v1.BillingCoupon{mockBillingCoupon}

		mockGetCouponExpiry := func(ctx context.Context, in *v1.BillingCouponFilter, opts ...grpc.CallOption) (*v1.BillingCouponResponse, error) {
			return &v1.BillingCouponResponse{
				Coupons: mockBillingCoupons,
			}, nil
		}

		mockBillingCouponServiceClient := NewMockBillingCouponServiceClient(mockGetCouponExpiry)
		billingClient := &BillingClient{
			BillingCreditServiceClient: nil,
			BillingCouponServiceClient: mockBillingCouponServiceClient,
		}

		// run the function under test
		couponRes, err := billingClient.GetCouponExpiry(ctx, mockBillingCoupon.Code)
		assert.NoError(t, err)
		assert.True(t, couponRes[0].GetExpires().Seconds > actualTimestamp.Seconds)
	})

	t.Run("invalid coupon check successful execution", func(t *testing.T) {

		// Create mock timestamp instances for created, start, expires, and actual timestamps
		createdTimestamp := &timestamppb.Timestamp{Seconds: 1711060129}
		startTimestamp := &timestamppb.Timestamp{Seconds: 1710964973}
		expiresTimestamp := &timestamppb.Timestamp{Seconds: 1742500300}
		actualTimestamp := &timestamppb.Timestamp{Seconds: 1842500300, Nanos: 884908000}

		// Create mock BillingCoupon instance
		mockBillingCoupon := &v1.BillingCoupon{
			Code:        "test-code",
			Creator:     "test-id",
			Created:     createdTimestamp,
			Start:       startTimestamp,
			Expires:     expiresTimestamp,
			Amount:      1.0,
			NumUses:     100,
			NumRedeemed: 1,
			IsStandard:  nil,
		}

		// Create a slice of BillingCoupon containing the mock instance
		mockBillingCoupons := []*v1.BillingCoupon{mockBillingCoupon}

		mockGetCouponExpiry := func(ctx context.Context, in *v1.BillingCouponFilter, opts ...grpc.CallOption) (*v1.BillingCouponResponse, error) {
			return &v1.BillingCouponResponse{
				Coupons: mockBillingCoupons,
			}, nil
		}

		mockBillingCouponServiceClient := NewMockBillingCouponServiceClient(mockGetCouponExpiry)
		billingClient := &BillingClient{
			BillingCreditServiceClient: nil,
			BillingCouponServiceClient: mockBillingCouponServiceClient,
		}

		// run the function under test
		couponRes, err := billingClient.GetCouponExpiry(ctx, mockBillingCoupon.Code)
		assert.NoError(t, err)
		assert.False(t, couponRes[0].GetExpires().Seconds > actualTimestamp.Seconds)
	})

	t.Run("error in fetching billing coupon client response", func(t *testing.T) {

		mockGetCouponExpiry := func(ctx context.Context, in *v1.BillingCouponFilter, opts ...grpc.CallOption) (*v1.BillingCouponResponse, error) {
			return nil, errors.New("error in fetching billing coupon client response")
		}

		mockBillingCouponServiceClient := NewMockBillingCouponServiceClient(mockGetCouponExpiry)
		billingClient := &BillingClient{
			BillingCreditServiceClient: nil,
			BillingCouponServiceClient: mockBillingCouponServiceClient,
		}

		// run the function under test
		couponRes, err := billingClient.GetCouponExpiry(ctx, "test-code")
		assert.Error(t, err)
		assert.Nil(t, couponRes)

	})
}
