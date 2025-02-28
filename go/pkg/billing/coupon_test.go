// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"testing"

	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/assert"
)

func TestCoupon(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCoupon")
	logger.Info("BEGIN")
	defer logger.Info("End")

	client := pb.NewBillingCouponServiceClient(clientConn)
	creator := "idc-admin"
	currentTime := timestamppb.Now()

	// Test Create Coupon
	coupon, err := client.Create(ctx,
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

	intelUser := "std_" + uuid.NewString() + "example.com"
	// Test Redeem Coupon
	intelAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  intelUser,
		Owner: intelUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_INTEL,
	})
	intelCloudAccountId := intelAcct.GetId()
	_, err = client.Redeem(ctx,
		&pb.BillingCouponRedeem{
			Code:           coupon.Code,
			CloudAccountId: intelCloudAccountId,
		})
	if err != nil {
		t.Errorf("redeem coupon: %v", err)
	}

	// Test Redeem Same coupon
	_, err = client.Redeem(context.Background(),
		&pb.BillingCouponRedeem{
			Code:           coupon.Code,
			CloudAccountId: intelCloudAccountId,
		})
	if err == nil {
		t.Errorf("should throw error since the copon (%v) is already redeemed by the cloudAccountId (%v)", coupon.Code, intelCloudAccountId)
	}

	// Test Read Coupons
	res, err := client.Read(context.Background(),
		&pb.BillingCouponFilter{
			Code: &coupon.Code,
		})
	if err != nil {
		t.Fatalf("coupon read failed: %v", err)
	}
	assert.Equal(t, res.Coupons[0].Redemptions[0].Code, coupon.Code)
	assert.Equal(t, res.Coupons[0].Redemptions[0].CloudAccountId, intelCloudAccountId)

	// Test Disable Coupon
	_, err = client.Disable(context.Background(),
		&pb.BillingCouponDisable{
			Code: coupon.Code,
		})
	if err != nil {
		t.Errorf("Disable Coupon: %v", err)
	}

	// Test Redeem a Disabled coupon
	_, err = client.Redeem(context.Background(),
		&pb.BillingCouponRedeem{
			Code:           coupon.Code,
			CloudAccountId: intelCloudAccountId,
		})
	if err == nil {
		t.Errorf("should throw error since coupon is already disabled")
	}
}

func TestInvalidCoupon(t *testing.T) {
	client := pb.NewBillingCouponServiceClient(clientConn)
	currentTime := timestamppb.Now()

	// Test Create Coupon where amount is zero
	_, err := client.Create(context.Background(),
		&pb.BillingCouponCreate{
			Expires: timestamppb.New(currentTime.AsTime().AddDate(0, 0, 90)),
			NumUses: 4,
		})
	if err == nil {
		t.Fatalf("should throw error because the coupon amount needs to be greater than 0")
	}

	// Test Create coupon where expiration time is smaller than start time
	_, err = client.Create(context.Background(),
		&pb.BillingCouponCreate{
			Expires: timestamppb.New(currentTime.AsTime().AddDate(0, 0, 30)),
			Amount:  500,
			NumUses: 4,
			Start:   timestamppb.New(currentTime.AsTime().AddDate(0, 0, 31)),
		})
	if err == nil {
		t.Fatalf("should throw error because expiration time cannot be smaller than start time")
	}

	// Test Create coupon where expiration time is smaller than creation time
	_, err = client.Create(context.Background(),
		&pb.BillingCouponCreate{
			Expires: timestamppb.New(currentTime.AsTime().AddDate(0, 0, -35)),
			Amount:  500,
			NumUses: 4,
			Start:   timestamppb.New(currentTime.AsTime().AddDate(0, 0, 30)),
		})
	if err == nil {
		t.Fatalf("should throw error because expiration time cannot be smaller than current time")
	}

	// Test Create coupon with max uses limit
	_, err = client.Create(context.Background(),
		&pb.BillingCouponCreate{
			Expires: timestamppb.New(currentTime.AsTime().AddDate(0, 0, 35)),
			Amount:  500,
			NumUses: 100,
			Start:   timestamppb.New(currentTime.AsTime().AddDate(0, 0, 5)),
		})
	if err == nil {
		t.Fatalf("should throw error because number of uses limit reached")
	}

	//  Create a Valid Coupon
	coupon, err := client.Create(context.Background(),
		&pb.BillingCouponCreate{
			Expires: timestamppb.New(currentTime.AsTime().AddDate(0, 0, 90)),
			Start:   timestamppb.New(currentTime.AsTime().AddDate(0, 0, 2)),
			Amount:  500,
			NumUses: 4,
		})
	if err != nil {
		t.Fatalf("create coupon failed: %v", err)
	}

	err = CheckCouponCode(coupon.Code)
	if err != nil {
		t.Errorf("%v", err)
	}

	// Check an Invalid Coupon
	invalidCoupon := "Z#S9-&77A-TRE#"
	err = CheckCouponCode(invalidCoupon)
	if err == nil {
		t.Errorf("should throw error because coupon code can be of characters which are capital letters and digits except letters I and O")
	}

	intelUser := "std_" + uuid.NewString() + "example.com"
	intelAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  intelUser,
		Owner: intelUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_INTEL,
	})
	intelCloudAccountId := intelAcct.GetId()

	// Test redempetion of a nil coupon code
	_, err = client.Redeem(context.Background(),
		&pb.BillingCouponRedeem{
			CloudAccountId: intelCloudAccountId,
		})
	if err == nil {
		t.Errorf("should throw error because coupon code is missing")
	}

	// Test redemption of a non existing coupon
	nonExistingCouponCode := "ZZZZ-ZZZZ-AAAA"
	_, err = client.Redeem(context.Background(),
		&pb.BillingCouponRedeem{
			Code:           nonExistingCouponCode,
			CloudAccountId: intelCloudAccountId,
		})
	if err == nil {
		t.Errorf("redeeming a non-existing coupon should throw error")
	}

	// Test redemption of a coupon by an non existing cloudAccountID
	nonExistingCloudAccountId := "123412341234"
	_, err = client.Redeem(context.Background(),
		&pb.BillingCouponRedeem{
			Code:           coupon.Code,
			CloudAccountId: nonExistingCloudAccountId,
		})
	if err == nil {
		t.Errorf("should throw error because the coupon cannnot be redeemed by an invalid cloudAccountId")
	}

	// Test redempetion of a coupon before start time
	_, err = client.Redeem(context.Background(),
		&pb.BillingCouponRedeem{
			Code:           coupon.Code,
			CloudAccountId: intelCloudAccountId,
		})
	if err == nil {
		t.Errorf("should throw error because the coupon cannnot be redeemed before the start time of coupon")
	}

	// Test Read non existing coupon code
	_, err = client.Read(context.Background(),
		&pb.BillingCouponFilter{
			Code: &nonExistingCouponCode,
		})
	if err == nil {
		t.Errorf("should throw error because coupon code (%v) does not exists", nonExistingCouponCode)
	}

	// Test Disable without the coupon code
	_, err = client.Disable(context.Background(),
		&pb.BillingCouponDisable{
			Disabled: currentTime,
		})
	if err == nil {
		t.Errorf("should throw error because coupon code cannot be empty")
	}

	// Test Disable non existing coupon code
	_, err = client.Disable(context.Background(),
		&pb.BillingCouponDisable{
			Code:     nonExistingCouponCode,
			Disabled: currentTime,
		})
	if err == nil {
		t.Errorf("should throw error because coupon code does not exists")
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
	service := &BillingCouponService{}

	// Define the redeemedDate for the test
	redeemedDate := time.Now()
	creator := "test.admin@intel.com"
	// Set up the expected rows that the query will return
	rows := sqlmock.NewRows([]string{"code", "cloud_account_id", "redeemed", "installed", "creator"}).
		AddRow("test-code", "test-cloud-account-id", redeemedDate, true, "test.admin@intel.com")

	// Expect a query to execute with the correct SQL and arguments
	mock.ExpectQuery("SELECT code, cloud_account_id, redeemed, installed FROM redemptions WHERE redeemed >= \\$1").
		WithArgs(redeemedDate.Format(time.RFC3339)).
		WillReturnRows(rows)

	// Call the function under test
	redemptions, err := service.getRedemptions(context.Background(), db, redeemedDate, creator)

	// Assert that there was no error returned
	assert.NilError(t, err)

	// Assert that the correct number of redemptions were returned
	assert.Equal(t, len(redemptions), 1)

	// Assert that the redemption data is correct
	assert.Equal(t, redemptions[0].Code, "test-code")
	assert.Equal(t, redemptions[0].CloudAccountId, "test-cloud-account-id")
	assert.DeepEqual(t, redemptions[0].Redeemed, timestamppb.New(redeemedDate))
	assert.Equal(t, redemptions[0].Installed, true)
	assert.Equal(t, redemptions[0].Creator, "test.admin@intel.com")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
