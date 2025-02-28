// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	DefaultCloudCreditAmount    = float64(1000)
	DefaultCreditCoupon         = "test_coupon"
	DefaultCreditExpirationDays = 35
	DefaultCreditReason         = pb.BillingCreditReason_CREDIT_INITIAL
)

func CreateAndGetCloudAccount(t *testing.T, ctx context.Context, user string, acctType pb.AccountType) *pb.CloudAccount {
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	acctCreate := pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  acctType,
	}
	id, err := cloudAcctClient.Create(ctx, &acctCreate)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	acctOut, err := cloudAcctClient.GetById(context.Background(), &pb.CloudAccountId{Id: id.GetId()})
	if err != nil {
		t.Fatalf("failed to read account: %v", err)
	}
	return acctOut
}

func CreateAndGetAccount(t *testing.T, acctCreate *pb.CloudAccountCreate) *pb.CloudAccount {
	ctx := context.Background()
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	id, err := cloudAcctClient.Create(ctx, acctCreate)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	acctOut, err := cloudAcctClient.GetById(context.Background(), &pb.CloudAccountId{Id: id.GetId()})
	if err != nil {
		t.Fatalf("read account: %v", err)
	}
	// todo: there are two issues here - low credits should be true until credits are assigned,
	// paid services allowed should be false until set to true (credits installed, etc.)
	if acctOut.Delinquent || acctOut.LowCredits || acctOut.TerminateMessageQueued ||
		acctOut.TerminatePaidServices { //|| acctOut.PaidServicesAllowed {
		t.Fatalf("acct created with the wrong flags")
	}

	billingClient := pb.NewBillingAccountServiceClient(clientConn)
	_, err = billingClient.Create(ctx, &pb.BillingAccount{
		CloudAccountId: id.GetId(),
	})
	if err != nil {
		t.Fatalf("create billing account: %v", err)
	}

	return acctOut
}

func CreateBillingCredit(t *testing.T, ctx context.Context, cloudAcct *pb.CloudAccount, billingAcct *pb.BillingAccount) {
	billingCreditClient := pb.NewBillingCreditServiceClient(clientConn)
	currTime := time.Now()
	newDate := currTime.AddDate(0, 0, DefaultCreditExpirationDays)
	expirationDate := timestamppb.New(newDate)
	billingCredit := &pb.BillingCredit{
		CloudAccountId:  billingAcct.CloudAccountId,
		Created:         timestamppb.New(time.Now()),
		OriginalAmount:  DefaultCloudCreditAmount,
		RemainingAmount: DefaultCloudCreditAmount,
		Reason:          DefaultCreditReason,
		CouponCode:      DefaultCreditCoupon,
		Expiration:      expirationDate}
	_, err := billingCreditClient.Create(context.Background(), billingCredit)
	if err != nil {
		t.Fatalf("failed to create driver credits: %v", err)
	}
}

func GetBillingCredits(t *testing.T, ctx context.Context, billingAcct *pb.BillingAccount) []*pb.BillingCredit {
	billingClient := pb.NewBillingCreditServiceClient(clientConn)
	billingCreditClient, err := billingClient.ReadInternal(ctx, &pb.BillingAccount{CloudAccountId: billingAcct.CloudAccountId})
	if err != nil {
		t.Fatalf("failed to create billing credit client: %v", err)
	}
	var billingCredits []*pb.BillingCredit
	for {
		billingCredit, err := billingCreditClient.Recv()
		if errors.Is(err, io.EOF) {
			return billingCredits
		}
		if err != nil {
			t.Fatalf("failed to read billing credits: %v", err)
		}
		billingCredits = append(billingCredits, billingCredit)
	}
}

func VerifyCloudAcctHasCreditsForExpiry(t *testing.T, cloudAcctId string) {
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAcct, err := cloudAcctClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId})
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
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAcct, err := cloudAcctClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId})
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
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAcct, err := cloudAcctClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId})
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
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAcct, err := cloudAcctClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId})
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

func VerifyCloudAcctServicesNoTermination(t *testing.T, cloudAcctId string) {
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAcct, err := cloudAcctClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId})
	if err != nil {
		t.Fatalf("failed to get cloud account: %v", err)
	}
	if cloudAcct.TerminatePaidServices {
		t.Fatalf("cloud acct paid services should not be terminated")
	}
}

func VerifyCloudAcctServicesTermination(t *testing.T, cloudAcctId string) {
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAcct, err := cloudAcctClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId})
	if err != nil {
		t.Fatalf("failed to get cloud account: %v", err)
	}
	if !cloudAcct.TerminatePaidServices {
		t.Fatalf("cloud acct paid services should be terminated")
	}
	if cloudAcct.PaidServicesAllowed {
		t.Fatalf("cloud acct paid services should not be allowed")
	}
}
