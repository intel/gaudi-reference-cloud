// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func UpdateStandardCreditUsed(t *testing.T, ctx context.Context, billingAcct *pb.BillingAccount, standardCredits []*pb.BillingCredit) {
	log := log.FromContext(ctx).WithName("UpdateStandardCreditUsed")
	log.Info("BEGIN")
	defer log.Info("END")
	updateStandardCreditsSql := "UPDATE cloud_credits SET remaining_amount = 0 WHERE cloud_account_id = $1"
	tx, err := GetDriverDb().BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("transaction begin failed for update standard credits to used: %v", err)
	}
	defer tx.Rollback()
	for _, standardCredit := range standardCredits {
		_, err = tx.ExecContext(ctx, updateStandardCreditsSql, standardCredit.CloudAccountId)
		if err != nil {
			t.Fatalf("failed to update standard credits to used: %v", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit transaction for update standard credits to used: %v", err)
	}
}

// Debug test func
func GetStandardCredits(t *testing.T, ctx context.Context, standardCredits []*pb.BillingCredit) ([]*pb.BillingCredit, error) {
	log := log.FromContext(ctx).WithName("GetStandardCredits")
	log.Info("BEGIN")
	defer log.Info("END")
	query :=
		"SELECT coupon_code, cloud_account_id, created_at, expiry," +
			"original_amount, remaining_amount, updated_at " +
			"FROM cloud_credits WHERE cloud_account_id=$1"
	var billingCredits []*pb.BillingCredit
	for _, standardCredit := range standardCredits {
		log.Info("get billing credits", "standardCredits", standardCredits)
		rows, err := GetDriverDb().Query(query, standardCredit.CloudAccountId)
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

func DeleteStandardCredits(t *testing.T, ctx context.Context) {
	deleteStandardCreditsSql := "DELETE FROM cloud_credits"
	_, err := GetDriverDb().ExecContext(ctx, deleteStandardCreditsSql)
	if err != nil {
		t.Fatalf("failed to delete standard credits: %v", err)
	}
}

func UpdateStandardCreditWithRemainingAmount(t *testing.T, ctx context.Context, billingAcct *pb.BillingAccount, cloudAcctId string, remainingAmount float64) {
	updateStandardCreditsSql := "UPDATE cloud_credits SET remaining_amount = $1 WHERE cloud_account_id = $2"
	_, err := GetDriverDb().ExecContext(ctx, updateStandardCreditsSql, remainingAmount, cloudAcctId)
	if err != nil {
		t.Fatalf("failed to update standard credits with usage amount: %v", err)
	}
}

func CreateStanardCloudAcct(t *testing.T, ctx context.Context) *pb.CloudAccount {
	intelUser := "intel_" + uuid.NewString() + "@test.com"

	acct := &pb.CloudAccountCreate{
		Name:  intelUser,
		Owner: intelUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}

	cloudAcctId, err := StandardService.cloudAccountClient.CloudAccountClient.Create(ctx, acct)

	if err != nil {
		t.Fatalf("failed to create cloud account: %v", err)
	}

	acctOut, err := StandardService.cloudAccountClient.CloudAccountClient.GetById(context.Background(), &pb.CloudAccountId{Id: cloudAcctId.Id})
	if err != nil {
		t.Fatalf("failed to read cloud account: %v", err)
	}

	return acctOut
}
