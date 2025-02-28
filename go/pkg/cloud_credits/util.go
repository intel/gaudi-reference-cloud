// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcredits

import (
	"context"
	"fmt"
	"math"
	"sync"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

type SchedulerCloudAccountState struct {
	AccessTimestamp string
	Mutex           sync.Mutex
}

type Credit struct {
	CreditId        uint64
	Code            string
	RemainingAmount float64
}

type CloudAccountLocks struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func GetSchedulerError(schedulerError string, err error) error {
	return fmt.Errorf("scheduler error:%s,error:%w", schedulerError, err)
}

const (
	FailedToReadCloudAccount           = "FAILED_TO_READ_CLOUD_ACCT"
	InvalidCloudAccountId              = "INVALID_CLOUD_ACCT_ID"
	InvalidCouponCode                  = "INVALID_COUPON_CODE"
	InvalidBillingCreditAmount         = "INVALID_BILLING_CREDIT_AMOUNT"
	InvalidBillingCreditExpiration     = "INVALID_BILLING_CREDIT_EXPIRATION"
	InvalidBillingCreditReason         = "INVALID_BILLING_CREDIT_REASON"
	FailedToLookUpCloudAccount         = "FAILED_TO_LOOK_UP_CLOUD_ACCT"
	InvalidCloudAcct                   = "INVALID_CLOUD_ACCOUNT"
	FailedToCreateBillingCredit        = "FAILED_TO_CREATE_BILLING_CREDIT"
	FailedToMigrateCredits             = "FAILED_TO_MIGRATE_CREDITS"
	FailedToReadBillingCredit          = "FAILED_TO_READ_BILLING_CREDIT"
	FailedToReadUnappliedCreditBalance = "FAILED_TO_READ_UNAPPLIED_CREDIT_BALANCE"
	InvalidSchedulerOpsAction          = "INVALID_SCHEDULER_OPS_ACTION"
	FailedToDeleteMigratedCredits      = "FAILED_TO_DELETE_MIGRATED_CREDITS"
	FailedToGetUsages                  = "FAILED_TO_GET_USAGES"
)

func NewCloudAccountLocks(ctx context.Context) *CloudAccountLocks {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("NewCloudAccountLocks").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	return &CloudAccountLocks{
		locks: make(map[string]*sync.Mutex),
	}
}

func GetBillingError(billingApiError string, err error) error {
	return fmt.Errorf("billing api error:%s,service error:%v", billingApiError, err)
}

func GetBillingInternalError(billingInternalError string, err error) error {
	return fmt.Errorf("billing internal error:%s,service error:%v", billingInternalError, err)
}

func GetSortedCredits(ctx context.Context, cloudAccountId string) ([]Credit, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("utils.GetSortedCredits").Start()
	log.WithValues("cloudaccountid", cloudAccountId)
	defer span.End()
	log.Info("Executing ", "getSortedCredits for", cloudAccountId)
	defer log.Info("Returning", "getSortedCredits for", cloudAccountId)

	query := "SELECT id,coupon_code,remaining_amount " +
		"FROM cloud_credits " +
		"WHERE cloud_account_id=$1 AND ( expiry >= NOW() OR remaining_amount < 0 ) " +
		"ORDER BY created_at ASC "

	log.Info("get sorted credit", "query", query, "cloudAccount", cloudAccountId)
	var credits []Credit

	rows, err := db.QueryContext(ctx, query, cloudAccountId)
	if err != nil {
		log.Error(err, "error in getting sorted credits")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		obj := Credit{}
		if err := rows.Scan(&obj.CreditId, &obj.Code, &obj.RemainingAmount); err != nil {
			log.Error(err, "error in getting credits row")
			return nil, err
		}
		log.Info("Credit", "creditId", obj.CreditId, "remainingAmount", obj.RemainingAmount)
		credits = append(credits, obj)
	}

	return credits, nil
}

func ProcessCredit(ctx context.Context, creditList []Credit, cloudAccountId string) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("utils.ProcessCredit").Start()
	log.WithValues("cloudaccountid", cloudAccountId)
	defer span.End()
	log.Info("Executing ProcessCredit")
	defer log.Info("Returning from ProcessCredit")

	lengthOfCreditList := len(creditList)
	if lengthOfCreditList > 1 {
		for i := lengthOfCreditList - 1; i > 0; i-- {
			if creditList[i-1].RemainingAmount < 0 && creditList[lengthOfCreditList-1].RemainingAmount > 0 {
				query := "UPDATE cloud_credits " +
					"SET remaining_amount=$1, updated_at=NOW()::timestamp  " +
					"WHERE cloud_account_id=$2 AND id=$3"

				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					log.Error(err, "error starting db transaction")
					return err
				}
				defer tx.Rollback()
				stmt, err := tx.PrepareContext(ctx, query)
				if err != nil {
					log.Error(err, "error in preparing query for updating credits")
					return err
				}

				defer stmt.Close()

				remainingAmountForLastCredit := creditList[lengthOfCreditList-1].RemainingAmount - math.Abs(creditList[i-1].RemainingAmount)
				log.Info("credit remaining", "cloudAccountId", cloudAccountId, "remainingAmountForLastCredit", remainingAmountForLastCredit)
				_, err = stmt.ExecContext(ctx, 0, cloudAccountId, creditList[i-1].CreditId)
				if err != nil {
					log.Error(err, "error updating last credit with 0 when had negative remaining amount")
					return err
				}

				_, err = stmt.ExecContext(ctx, remainingAmountForLastCredit, cloudAccountId, creditList[lengthOfCreditList-1].CreditId)
				if err != nil {
					log.Error(err, "error updating added credit with remainining when had negative remaining amount")
					return err
				}
				if err := tx.Commit(); err != nil {
					log.Error(err, "error committing db transaction")
					return err
				}
			}
		}
	}
	return nil
}

func (cloudAccountLocks *CloudAccountLocks) Lock(ctx context.Context, cloudAccountID string) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("Lock").WithValues("cloudAccountID", cloudAccountID).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	cloudAccountLocks.mu.Lock()

	if _, exists := cloudAccountLocks.locks[cloudAccountID]; !exists {
		cloudAccountLocks.locks[cloudAccountID] = &sync.Mutex{}
	}
	cloudAccountLocks.mu.Unlock()
	cloudAccountLocks.locks[cloudAccountID].Lock()
}

func (cloudAccountLocks *CloudAccountLocks) Unlock(ctx context.Context, cloudAccountID string) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("Unlock").WithValues("cloudAccountID", cloudAccountID).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	cloudAccountLocks.mu.Lock()
	defer cloudAccountLocks.mu.Unlock()

	if lock, exists := cloudAccountLocks.locks[cloudAccountID]; exists {
		lock.Unlock()
	}
}
