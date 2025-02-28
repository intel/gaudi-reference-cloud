// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	DefaultCloudCreditAmount    = float64(1000)
	DefaultCreditCoupon         = "test_coupon"
	DefaultCreditExpirationDays = 35
	DefaultCreditReason         = pb.BillingCreditReason_CREDIT_INITIAL
)

func CreateAndGetAccount(t *testing.T, acctCreate *pb.CloudAccountCreate) *pb.CloudAccount {
	ctx := context.Background()
	id, err := TestCloudAccountSvcClient.CloudAccountClient.Create(ctx, acctCreate)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	acctOut, err := TestCloudAccountSvcClient.CloudAccountClient.GetById(context.Background(), &pb.CloudAccountId{Id: id.GetId()})
	if err != nil {
		t.Fatalf("read account: %v", err)
	}
	// todo: there are two issues here - low credits should be true until credits are assigned,
	// paid services allowed should be false until set to true (credits installed, etc.)
	if acctOut.Delinquent || acctOut.LowCredits || acctOut.TerminateMessageQueued ||
		acctOut.TerminatePaidServices { //|| acctOut.PaidServicesAllowed {
		t.Fatalf("acct created with the wrong flags")
	}
	if acctOut.GetType() == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		driver, err := billingCommon.GetDriver(ctx, id.GetId())
		if err != nil {
			t.Fatalf("failed to get billing driver: %v", err)
		}

		_, err = driver.BillingAcct.Create(ctx, &pb.BillingAccount{
			CloudAccountId: id.GetId(),
		})
		if err != nil {
			t.Fatalf("create billing account: %v", err)
		}
	}
	return acctOut
}

func GetBillingCredits(t *testing.T, ctx context.Context, billingAcct *pb.BillingAccount) []*pb.BillingCredit {
	driver, err := billingCommon.GetDriver(ctx, billingAcct.CloudAccountId)
	if err != nil {
		t.Fatalf("failed to get billing driver: %v", err)
	}
	billingCreditClient, err := driver.BillingCredit.ReadInternal(ctx, &pb.BillingAccount{CloudAccountId: billingAcct.CloudAccountId})
	if err != nil {
		t.Fatalf("failed to create billing credit client: %v", err)
	}
	var billingCredits []*pb.BillingCredit
	for {
		BillingCredit, err := billingCreditClient.Recv()
		if errors.Is(err, io.EOF) {
			return billingCredits
		}
		if err != nil {
			t.Fatalf("failed to read billing credits: %v", err)
		}
		billingCredits = append(billingCredits, BillingCredit)
	}
}

func CreateBillingCredit(t *testing.T, ctx context.Context, cloudAcct *pb.CloudAccount, billingAcct *pb.BillingAccount) {
	driver, err := billingCommon.GetDriver(ctx, billingAcct.CloudAccountId)
	if err != nil {
		t.Fatalf("failed to get billing driver: %v", err)
	}
	currTime := time.Now()
	newDate := currTime.AddDate(0, 0, DefaultCreditExpirationDays)
	expirationDate := timestamppb.New(newDate)
	BillingCredit := &pb.BillingCredit{
		CloudAccountId:  billingAcct.CloudAccountId,
		Created:         timestamppb.New(time.Now()),
		OriginalAmount:  DefaultCloudCreditAmount,
		RemainingAmount: DefaultCloudCreditAmount,
		Reason:          DefaultCreditReason,
		CouponCode:      DefaultCreditCoupon,
		Expiration:      expirationDate}
	_, err = driver.BillingCredit.Create(context.Background(), BillingCredit)
	if err != nil {
		t.Fatalf("failed to create driver credits: %v", err)
	}
}

func VerifyCloudAcctHasCreditsForExpiry(t *testing.T, cloudAcctId string) {
	cloudAcct, err := TestCloudAccountSvcClient.CloudAccountClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId})
	if err != nil {
		t.Fatalf("failed to get cloud account: %v", err)
	}
	if cloudAcct.LowCredits {
		t.Fatalf("cloud acct should not have low credits")
	}
	if cloudAcct.CreditsDepleted.AsTime().Unix() != 0 {
		t.Fatalf("cloud acct credits should not have depleted")
	}
	if cloudAcct.TerminatePaidServices {
		t.Fatalf("cloud acct paid services need not be terminated")
	}
	if cloudAcct.TerminateMessageQueued {
		t.Fatalf("cloud acct paid services terminate message should not be queued")
	}
}

func VerifyCloudAcctHasCredits(t *testing.T, cloudAcctId string) {
	cloudAcct, err := TestCloudAccountSvcClient.CloudAccountClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId})
	if err != nil {
		t.Fatalf("failed to get cloud account: %v", err)
	}
	if cloudAcct.LowCredits {
		t.Fatalf("cloud acct should not have low credits")
	}
	if cloudAcct.CreditsDepleted.AsTime().Unix() != 0 {
		t.Fatalf("cloud acct credits should not have depleted")
	}
	if !cloudAcct.PaidServicesAllowed {
		t.Fatalf("cloud acct paid services need to be allowed")
	}
	if cloudAcct.TerminatePaidServices {
		t.Fatalf("cloud acct paid services need not be terminated")
	}
	if cloudAcct.TerminateMessageQueued {
		t.Fatalf("cloud acct paid services terminate message should not be queued")
	}
}

func VerifyCloudAcctHasLowCredits(t *testing.T, cloudAcctId string) {
	cloudAcct, err := TestCloudAccountSvcClient.CloudAccountClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId})
	if err != nil {
		t.Fatalf("failed to get cloud account: %v", err)
	}
	if !cloudAcct.LowCredits {
		t.Fatalf("cloud acct should have low credits")
	}
	if cloudAcct.CreditsDepleted.AsTime().Unix() != 0 {
		t.Fatalf("cloud acct credits should not have depleted")
	}
	if !cloudAcct.PaidServicesAllowed {
		t.Fatalf("cloud acct paid services need to be allowed")
	}
	if cloudAcct.TerminatePaidServices {
		t.Fatalf("cloud acct paid services need not be terminated")
	}
	if cloudAcct.TerminateMessageQueued {
		t.Fatalf("cloud acct paid services terminate message should not be queued")
	}
}

func VerifyCloudAcctHasNoCredits(t *testing.T, cloudAcctId string) {
	cloudAcct, err := TestCloudAccountSvcClient.CloudAccountClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId})
	if err != nil {
		t.Fatalf("failed to get cloud account: %v", err)
	}
	if cloudAcct.CreditsDepleted.AsTime().Unix() == 0 {
		t.Fatalf("cloud acct credits should have depleted")
	}
	if cloudAcct.PaidServicesAllowed {
		t.Fatalf("cloud acct paid services should not be allowed")
	}
}
