// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStandardCredits(t *testing.T) {
	acct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  "standardcredit@example.com",
		Owner: "standardcredit@example.com",
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	})

	client := pb.NewBillingCreditServiceClient(clientConn)
	couponCient := pb.NewBillingCouponServiceClient(clientConn)

	creator := "idc-admin"
	currentTime := timestamppb.Now()

	isStandard := true
	coupon, err := couponCient.Create(context.Background(),
		&pb.BillingCouponCreate{
			Expires:    timestamppb.New(currentTime.AsTime().AddDate(0, 0, 90)),
			Amount:     500,
			NumUses:    4,
			Creator:    creator,
			IsStandard: &isStandard,
		})
	if err != nil {
		t.Fatalf("create coupon failed: %v", err)
	}

	_, err = couponCient.Redeem(context.Background(),
		&pb.BillingCouponRedeem{
			Code:           coupon.Code,
			CloudAccountId: acct.Id,
		})
	if err != nil {
		t.Errorf("redeem coupon: %v", err)
	}

	// Read Credits
	history := true
	_, err = client.Read(context.Background(), &pb.BillingCreditFilter{CloudAccountId: acct.Id, History: &history})
	if err != nil {
		t.Fatalf("Standard billing driver could not read credit balance %v", err)
	}

	// Read Internal Credits
	_, err = client.ReadInternal(context.Background(), &pb.BillingAccount{CloudAccountId: acct.Id})
	if err != nil {
		t.Fatalf("Standard billing driver cound not read credit balance %v", err)
	}

	//ReadUnappliedBalanceCredit
	resp, err := client.ReadUnappliedCreditBalance(context.Background(), &pb.BillingAccount{CloudAccountId: acct.Id})
	if err != nil {
		t.Fatalf("Standard billing driver cound not read unapllied credit balance %v", err)
	}
	assert.Equal(t, resp.UnappliedAmount, float64(500))
}

func TestAriaCredits(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestAriaCredits")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "example.com"
	acct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  premiumUser,
		Owner: premiumUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
	})

	client := pb.NewBillingCreditServiceClient(clientConn)

	currTime := time.Now()
	newDate := currTime.AddDate(0, 0, 35)
	expirationDate := timestamppb.New(newDate)

	//Create
	_, err := client.Create(ctx, &pb.BillingCredit{CloudAccountId: acct.Id, OriginalAmount: 100, Reason: 1, Expiration: expirationDate})
	if err != nil {
		t.Fatalf("Create credits failed: %v", err)
	}

	//TODO: Add Read to check if values for amount, reason code and expiration date are same

	//ReadUnappliedBalanceCredit
	resp, err := client.ReadUnappliedCreditBalance(ctx, &pb.BillingAccount{CloudAccountId: acct.Id})
	if err != nil {
		t.Fatalf("Read unapplied credit balance: %v", err)
	}
	assert.Equal(t, resp.UnappliedAmount, float64(2600))
}

func TestIntelCredits(t *testing.T) {
	acct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  "intelcredit@example.com",
		Owner: "intelcredit@example.com",
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_INTEL,
	})

	client := pb.NewBillingCreditServiceClient(clientConn)
	couponCient := pb.NewBillingCouponServiceClient(clientConn)

	creator := "idc-admin"
	currentTime := timestamppb.Now()

	coupon, err := couponCient.Create(context.Background(),
		&pb.BillingCouponCreate{
			Expires: timestamppb.New(currentTime.AsTime().AddDate(0, 0, 90)),
			Amount:  500,
			NumUses: 4,
			Creator: creator,
		})
	if err != nil {
		t.Fatalf("create coupon failed: %v", err)
	}

	_, err = couponCient.Redeem(context.Background(),
		&pb.BillingCouponRedeem{
			Code:           coupon.Code,
			CloudAccountId: acct.Id,
		})
	if err != nil {
		t.Errorf("redeem coupon: %v", err)
	}

	// Read Credits
	history := true
	_, err = client.Read(context.Background(), &pb.BillingCreditFilter{CloudAccountId: acct.Id, History: &history})
	if err != nil {
		t.Fatalf("Intel billing driver could not read credit balance %v", err)
	}

	// Read Internal Credits
	_, err = client.ReadInternal(context.Background(), &pb.BillingAccount{CloudAccountId: acct.Id})
	if err != nil {
		t.Fatalf("Intel billing driver cound not read credit balance %v", err)
	}

	//ReadUnappliedBalanceCredit
	resp, err := client.ReadUnappliedCreditBalance(context.Background(), &pb.BillingAccount{CloudAccountId: acct.Id})
	if err != nil {
		t.Fatalf("Intel billing driver cound not read unapllied credit balance %v", err)
	}
	assert.Equal(t, resp.UnappliedAmount, float64(500))

	invoiceClient := pb.NewBillingInvoiceServiceClient(clientConn)
	_, err = invoiceClient.Read(context.Background(), &pb.BillingInvoiceFilter{CloudAccountId: acct.Id})
	if err != nil {
		t.Fatalf("Intel billing driver cound not read invoice %v", err)
	}

	_, err = invoiceClient.ReadStatement(context.Background(), &pb.InvoiceId{CloudAccountId: acct.Id, InvoiceId: 000})
	if err == nil {
		t.Log("invoice readstatement unimplemented")
	}

}

func TestReadInternalAccountCredits(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestReadInternalAccountCredits")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "example.com"

	acct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  premiumUser,
		Owner: premiumUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
	})

	client := pb.NewBillingCreditServiceClient(clientConn)
	currTime := time.Now()
	newDate := currTime.AddDate(0, 0, 35)
	expirationDate := timestamppb.New(newDate)

	//Create
	_, err := client.Create(ctx, &pb.BillingCredit{CloudAccountId: acct.Id, OriginalAmount: 100, Reason: 1, Expiration: expirationDate})
	if err != nil {
		t.Fatalf("Create credits failed: %v", err)
	}

	res, err := client.ReadInternal(ctx,
		&pb.BillingAccount{CloudAccountId: acct.Id})
	if err != nil {
		t.Fatalf("error reading client credits: %v", err)
	}
	for {
		credit, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("error receving account credit: %v", err)
		}
		t.Logf("account credit: %v", credit)
	}
}

func TestReadAccountCredits(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestReadAccountCredits")
	logger.Info("BEGIN")
	defer logger.Info("End")

	premiumUser := "premium_" + uuid.NewString() + "example.com"

	acct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  premiumUser,
		Owner: premiumUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
	})

	client := pb.NewBillingCreditServiceClient(clientConn)
	currTime := time.Now()

	//Create 1st Service credit
	newDate := currTime.AddDate(0, 0, 35)
	expirationDate := timestamppb.New(newDate)

	_, err := client.Create(ctx, &pb.BillingCredit{CloudAccountId: acct.Id, OriginalAmount: 100, Reason: 1, Expiration: expirationDate, CouponCode: "code1"})
	if err != nil {
		t.Fatalf("Create credits failed for the first: %v", err)
	}

	//Create 2nd Service credit
	newDate2 := currTime.AddDate(0, 0, 60)
	expirationDate2 := timestamppb.New(newDate2)

	_, err = client.Create(ctx, &pb.BillingCredit{CloudAccountId: acct.Id, OriginalAmount: 1001, Reason: 1, Expiration: expirationDate2, CouponCode: "code2"})
	if err != nil {
		t.Fatalf("Create credits failed for the second: %v", err)
	}

	res, err := client.Read(ctx, &pb.BillingCreditFilter{CloudAccountId: acct.Id})
	if err != nil {
		t.Fatalf("error reading client credits: %v", err)
	}

	expirationDate.Nanos, expirationDate2.Nanos = 0, 0
	assert.Equal(t, res.TotalUsedAmount, float64(0))
	assert.Equal(t, res.TotalRemainingAmount, float64(3601))
	/*
		assert functions :-
		1) not working for res.Credits as data from getAcctCredits is not
		in order and thus order of messages also changes.
		2) Please check for create function as although it takes expiration, it doesn't use it
	*/
}

func TestCreditsInvalidCloudAcctId(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditsInvalidCloudAcctId")
	logger.Info("BEGIN")
	defer logger.Info("End")

	client := pb.NewBillingCreditServiceClient(clientConn)

	currTime := time.Now()
	newDate := currTime.AddDate(0, 0, 40)
	expirationDate := timestamppb.New(newDate)

	_, err := client.Create(ctx, &pb.BillingCredit{CloudAccountId: "ABCD", OriginalAmount: 100, Reason: 1, Expiration: expirationDate})
	if err == nil {
		t.Fatalf("create credit did not fail as expected")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got wrong  error code: %v", err)
	}
}

func TestCreditsNotExistingCloudAcctId(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditsNotExistingCloudAcctId")
	logger.Info("BEGIN")
	defer logger.Info("End")

	client := pb.NewBillingCreditServiceClient(clientConn)

	currTime := time.Now()
	newDate := currTime.AddDate(0, 0, 40)
	expirationDate := timestamppb.New(newDate)

	_, err := client.Create(ctx, &pb.BillingCredit{CloudAccountId: "1234", OriginalAmount: 100, Reason: 1, Expiration: expirationDate})
	if err == nil {
		t.Fatalf("create credit did not fail as expected")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got wrong  error code: %v", err)
	}
	if !strings.Contains(err.Error(), InvalidCloudAcct) {
		t.Fatalf("error code does not match")
	}
}

func TestCreditsInvalidAmount(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditsInvalidAmount")
	logger.Info("BEGIN")
	defer logger.Info("End")

	client := pb.NewBillingCreditServiceClient(clientConn)
	user := "premium_" + uuid.NewString() + "example.com"
	acct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
	})

	currTime := time.Now()
	newDate := currTime.AddDate(0, 0, 40)
	expirationDate := timestamppb.New(newDate)

	_, err := client.Create(ctx, &pb.BillingCredit{CloudAccountId: acct.Id, OriginalAmount: 0, Reason: 1, Expiration: expirationDate})
	if err == nil {
		t.Fatalf("create credit did not fail as expected")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got wrong  error code: %v", err)
	}
	if !strings.Contains(err.Error(), InvalidBillingCreditAmount) {
		t.Fatalf("error code does not match")
	}
}

func TestCreditsInvalidExpiration(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditsInvalidExpiration")
	logger.Info("BEGIN")
	defer logger.Info("End")

	client := pb.NewBillingCreditServiceClient(clientConn)
	user := "premium_" + uuid.NewString() + "example.com"
	acct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
	})

	currTime := time.Now()
	newDate := currTime.AddDate(0, 0, -2)
	expirationDate := timestamppb.New(newDate)

	_, err := client.Create(ctx, &pb.BillingCredit{CloudAccountId: acct.Id, OriginalAmount: 100, Reason: 1, Expiration: expirationDate})
	if err == nil {
		t.Fatalf("create credit did not fail as expected")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got wrong  error code: %v", err)
	}
	if !strings.Contains(err.Error(), InvalidBillingCreditExpiration) {
		t.Fatalf("error code does not match")
	}
}

func TestCreditsInvalidReason(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditsInvalidReason")
	logger.Info("BEGIN")
	defer logger.Info("End")

	client := pb.NewBillingCreditServiceClient(clientConn)
	user := "premium_" + uuid.NewString() + "example.com"
	acct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
	})

	currTime := time.Now()
	newDate := currTime.AddDate(0, 0, 40)
	expirationDate := timestamppb.New(newDate)

	_, err := client.Create(ctx, &pb.BillingCredit{CloudAccountId: acct.Id, OriginalAmount: 100, Reason: 20, Expiration: expirationDate})
	if err == nil {
		t.Fatalf("create credit did not fail as expected")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got wrong  error code: %v", err)
	}
	if !strings.Contains(err.Error(), InvalidBillingCreditReason) {
		t.Fatalf("error code does not match")
	}
}
