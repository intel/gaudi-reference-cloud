// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing_driver_intel

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func sendUsages(t *testing.T, wg *sync.WaitGroup, stream pb.BillingDriverUsageService_ReportUsageClient, usages []*pb.BillingDriverUsage) {
	defer wg.Done()
	defer stream.CloseSend()
	for _, usage := range usages {
		if err := stream.Send(usage); err != nil {
			t.Error(err)
			return
		}
	}
}

/**func ReportUsage(t *testing.T, ctx context.Context, cloudAccountId string, amount float64) {

	intelReportUsageClient := pb.NewBillingDriverUsageServiceClient(intelDriverConn)
	reportUsageStream, err := intelReportUsageClient.ReportUsage(ctx)

	if err != nil {
		t.Fatalf("failed to create channel for reporting usages: %v", err)
	}

	defer reportUsageStream.CloseSend()

	billingDriverUsage := &pb.BillingDriverUsage{
		CloudAccountId: cloudAccountId,
		ProductId:      uuid.NewString(),
		Amount:         amount,
		UsageId:        int64(rand.Uint64()),
	}

	err = reportUsageStream.Send(billingDriverUsage)

	if err != nil {
		t.Fatalf("failed to report usage: %v", err)
	}
}**/

func CreateIntelCloudAcct(t *testing.T, ctx context.Context) *pb.CloudAccount {
	intelUser := "intel_" + uuid.NewString() + "@intel.com"

	acct := &pb.CloudAccountCreate{
		Name:  intelUser,
		Owner: intelUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_INTEL,
	}

	cloudAcctId, err := IntelService.cloudAccountClient.CloudAccountClient.Create(ctx, acct)

	if err != nil {
		t.Fatalf("failed to create cloud account: %v", err)
	}

	acctOut, err := IntelService.cloudAccountClient.CloudAccountClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId.Id})
	if err != nil {
		t.Fatalf("failed to read cloud account: %v", err)
	}

	return acctOut
}

func UpdateIntelCreditUsed(t *testing.T, ctx context.Context, billingAcct *pb.BillingAccount, intelCredits []*pb.BillingCredit) {
	log := log.FromContext(ctx).WithName("UpdateIntelCreditUsed")
	log.Info("BEGIN")
	defer log.Info("END")
	updateIntelCreditsSql := "UPDATE cloud_credits SET remaining_amount = 0 WHERE cloud_account_id = $1"
	tx, err := GetDriverDb().BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("transaction begin failed for update intel credits to used: %v", err)
	}
	defer tx.Rollback()
	for _, intelCredit := range intelCredits {
		_, err = tx.ExecContext(ctx, updateIntelCreditsSql, intelCredit.CloudAccountId)
		if err != nil {
			t.Fatalf("failed to update intel credits to used: %v", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit transaction for update intel credits to used: %v", err)
	}
}

// Debug test func
func GetIntelCredits(t *testing.T, ctx context.Context, intelCredits []*pb.BillingCredit) ([]*pb.BillingCredit, error) {
	log := log.FromContext(ctx).WithName("GetIntelCredits")
	log.Info("BEGIN")
	defer log.Info("END")
	query :=
		"SELECT coupon_code, cloud_account_id, created_at, expiry," +
			"original_amount, remaining_amount, updated_at " +
			"FROM cloud_credits WHERE cloud_account_id=$1"
	var billingCredits []*pb.BillingCredit
	for _, intelCredit := range intelCredits {
		log.Info("get billing credits", "intelCredits", intelCredits)
		rows, err := GetDriverDb().Query(query, intelCredit.CloudAccountId)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var creation, expiration, updated_at, lastExpirationDate, lastUpdated time.Time
		for rows.Next() {
			log.Info("reading row")
			billingCredit := &pb.BillingCredit{}
			if err := rows.Scan(&billingCredit.CouponCode, &billingCredit.CloudAccountId, &creation, &expiration,
				&billingCredit.OriginalAmount, &billingCredit.RemainingAmount, &updated_at,
			); err != nil {
				return nil, err
			}
			if lastExpirationDate.Before(expiration) {
				lastExpirationDate = expiration
			}
			if lastUpdated.Before(updated_at) {
				lastUpdated = updated_at
			}
			billingCredit.Created = timestamppb.New(creation)
			billingCredit.Expiration = timestamppb.New(expiration)
			billingCredit.Reason = pb.BillingCreditReason_CREDIT_INITIAL
			billingCredit.AmountUsed = billingCredit.GetOriginalAmount() - billingCredit.GetRemainingAmount()
			log.Info("billingCredit ", "billingCredit", billingCredit)
			billingCredits = append(billingCredits, billingCredit)
		}
	}
	return billingCredits, nil
}

func DeleteIntelCredits(t *testing.T, ctx context.Context) {
	deleteIntelCreditsSql := "DELETE FROM cloud_credits"
	_, err := GetDriverDb().ExecContext(ctx, deleteIntelCreditsSql)
	if err != nil {
		t.Fatalf("failed to delete intel credits: %v", err)
	}
}

func UpdateIntelCreditWithRemainingAmount(t *testing.T, ctx context.Context, billingAcct *pb.BillingAccount, cloudAcctId string, remainingAmount float64) {
	updateIntelCreditsSql := "UPDATE cloud_credits SET remaining_amount = $1 WHERE cloud_account_id = $2"
	_, err := GetDriverDb().ExecContext(ctx, updateIntelCreditsSql, remainingAmount, cloudAcctId)
	if err != nil {
		t.Fatalf("failed to update intel credits with usage amount: %v", err)
	}
}
