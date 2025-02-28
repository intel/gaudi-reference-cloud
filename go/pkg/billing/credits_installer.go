// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CreditsInstallScheduler struct {
	syncTicker                   *time.Ticker
	db                           *sql.DB
	creditsExpiryMinimumInterval uint16
}

var (
	creditsInstallChan = make(chan bool)
	mutex              = &sync.Mutex{}
)

func NewCreditsInstallScheduler(db *sql.DB, tickerDuration uint16, creditExpiryMinInterval uint16) (*CreditsInstallScheduler, error) {
	if db == nil {
		return nil, fmt.Errorf("db is requied")
	}
	return &CreditsInstallScheduler{
		syncTicker:                   time.NewTicker(time.Duration(tickerDuration) * time.Second),
		db:                           db,
		creditsExpiryMinimumInterval: creditExpiryMinInterval,
	}, nil
}

func (s *CreditsInstallScheduler) StartCreditsInstallScheduler(ctx context.Context) {
	log := log.FromContext(ctx).WithName("CreditsInstallScheduler.StartCreditsInstallScheduler")
	log.Info("start credits installer scheduler")
	go s.CreditsInstallLoop(ctx)
}

func (s *CreditsInstallScheduler) StopCreditsInstallScheduler() {
	if creditsInstallChan != nil {
		close(creditsInstallChan)
		creditsInstallChan = nil
	}
}

func (s *CreditsInstallScheduler) CreditsInstallLoop(ctx context.Context) {
	log := log.FromContext(ctx).WithName("CreditsInstallScheduler.CreditsInstallLoop")
	log.Info("Install Credits")
	defer log.Info("END")
	for {
		err := s.CreditsInstall(ctx)
		if err != nil {
			log.Error(err, "error encountered in installing credits")
		}
		select {
		case <-creditsInstallChan:
			return
		case tm := <-s.syncTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

func (s *CreditsInstallScheduler) CreditsInstall(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CreditsInstallScheduler.CreditsInstall").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	query := `
		select code, cloud_account_id, redeemed, installed from redemptions
		where installed = $1
	`
	rows, err := s.db.QueryContext(ctx, query, false)
	if err != nil {
		logger.Error(err, "error querying the redemptions table")
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var code string
		var cloudaccountId string
		var redemptionTime time.Time
		var installed bool

		if err := rows.Scan(&code, &cloudaccountId, &redemptionTime, &installed); err != nil {
			logger.Error(err, "error scanning the row from redemptions table")
			return err
		}

		// Get the driver type for each cloudaccount
		driver, err := GetDriver(ctx, cloudaccountId)
		if err != nil {
			logger.Error(err, "unable to find driver", "cloudAccountId", cloudaccountId)
			return err
		}

		in := &pb.BillingCreditFilter{
			CloudAccountId: cloudaccountId,
		}

		// Read all the coupons installed for the cloudaccount
		resp, err := driver.billingCredit.Read(ctx, in)
		if err != nil {
			logger.Error(err, "error calling Read", "cloudAccountId", cloudaccountId)
			return err
		}

		foundCoupon := false
		for _, billingCredit := range resp.Credits {
			if billingCredit.CouponCode == code {
				foundCoupon = true
				break
			}
		}

		// Install the Coupon
		if !foundCoupon {
			logger = logger.WithValues("coupon", code, "cloudAccountId", cloudaccountId)
			ctx = log.IntoContext(ctx, logger)
			err := s.InstallCoupon(ctx, code, cloudaccountId, driver, s.creditsExpiryMinimumInterval)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *CreditsInstallScheduler) InstallCoupon(ctx context.Context, code string, cloudAccountId string, driver *BillingDriverClients, creditsExpiryMinimumInterval uint16) error {
	mutex.Lock()
	defer mutex.Unlock()
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CreditsInstallScheduler.InstallCoupon").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	// todo: for this method - return errors or assimilate errors
	cloudAcct, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil {
		logger.Error(err, "error getting cloud account")
		return err
	}

	var expirationTime time.Time
	var amount float64
	queryCoupons := `
		select	amount, expires
		from coupons
		where code = $1 and disabled is null
	`
	var row *sql.Row
	if row = s.db.QueryRowContext(ctx, queryCoupons, code); row.Err() != nil {
		logger.Error(row.Err(), "error executing query in coupons table")
		return row.Err()
	}
	if err := row.Scan(&amount, &expirationTime); err != nil {
		logger.Error(err, "error scan the row from coupons table")
		return err
	}

	// TODO:  Aria driver needs to extend the coupon to the end of the customer's billing cycle within which the expiration falls.
	if cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		redemptionTime := timestamppb.Now().AsTime()
		diffBetweenCouponExpirationAndRedemptionInDays := uint16(expirationTime.Sub(redemptionTime).Hours() / 24)
		logger.V(9).Info("coupon expiration duration", "expirationTime", expirationTime, "expirationDurationDiff", diffBetweenCouponExpirationAndRedemptionInDays, "creationTime", cloudAcct.Created.AsTime())
		if diffBetweenCouponExpirationAndRedemptionInDays < creditsExpiryMinimumInterval {
			expirationTime = redemptionTime.AddDate(0, 0, int(creditsExpiryMinimumInterval))
		} else {
			creationTime := cloudAcct.Created.AsTime()
			if expirationTime.Day() != creationTime.Day()+BillLagDays {
				// change expiratiin if falls between month
				expirationTime = expirationTime.AddDate(0, 0, int(creditsExpiryMinimumInterval))
			}
			logger.V(9).Info("coupon expiration", "expirationTime", expirationTime, "creationTime", creationTime)
		}
	}
	req := &pb.BillingCredit{
		Created:         timestamppb.New(timestamppb.Now().AsTime().UTC()),
		Expiration:      timestamppb.New(expirationTime),
		CloudAccountId:  cloudAccountId,
		OriginalAmount:  amount,
		RemainingAmount: amount,
		CouponCode:      code,
		Reason:          pb.BillingCreditReason_CREDIT_COUPON,
	}

	_, err = driver.billingCredit.Create(ctx, req)
	if err != nil {
		logger.Error(err, "error installing credits")
		return err
	}

	// Update installed columns to true in redemeptions table
	updateRedemptionsQuery := `
		update redemptions set installed = $1
		where code = $2 and cloud_account_id = $3
	`
	_, err = s.db.ExecContext(ctx, updateRedemptionsQuery, true, code, cloudAccountId)
	if err != nil {
		logger.Error(err, "error updating the redemptions table")
		return err
	}

	in := &pb.BillingCreditFilter{
		CloudAccountId: cloudAccountId,
	}
	billingCreditResponse, err := driver.billingCredit.Read(ctx, in)
	if err != nil {
		logger.Error(err, "error calling Read")
		return err
	}
	logger.V(9).Info("billing credit", "cloudAccountId", cloudAccountId, "totalRemainingAmount", billingCreditResponse.GetTotalRemainingAmount(), "totalUsedAmount", billingCreditResponse.GetTotalUsedAmount())
	if !cloudAcct.PaidServicesAllowed && billingCreditResponse.GetTotalRemainingAmount() > 0 {
		paidServicesAllowed := true
		enrolled := true
		terminatePaidServices := false
		lowCredits := false
		_, err := cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, PaidServicesAllowed: &paidServicesAllowed, Enrolled: &enrolled, TerminatePaidServices: &terminatePaidServices, LowCredits: &lowCredits})
		if err != nil {
			logger.Error(err, "failed to update cloud account paid services allowed")
			return err
		}
	}

	return nil
}
