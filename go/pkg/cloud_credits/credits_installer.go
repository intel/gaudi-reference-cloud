// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcredits

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var fieldOpts []protodb.FieldOptions = []protodb.FieldOptions{
	{Name: "", StoreEmptyStringAsNull: false},
}

func InstallCoupon(ctx context.Context, code string, cloudAccountId string, driver *billingCommon.BillingDriverClients) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CreditsInstaller.InstallCoupon").Start()
	defer span.End()
	log.Info("Install Coupon")

	cloudAcct, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil {
		log.Error(err, "error getting cloud acct")
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
	if row = db.QueryRowContext(ctx, queryCoupons, code); row.Err() != nil {
		log.Error(row.Err(), "error executing query in coupons table")
		return row.Err()
	}
	if err := row.Scan(&amount, &expirationTime); err != nil {
		log.Error(err, "error scan the row from coupons table")
		return err
	}

	// TODO: Remove installation of coupon to aria once DB from Aria is migrated to cloud_Credits db
	if cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		redemptionTime := timestamppb.Now().AsTime()
		diffBetweenCouponExpirationAndRedemptionInDays := uint16(expirationTime.Sub(redemptionTime).Hours() / 24)
		log.Info("coupon expiration duration", "expirationTime", expirationTime, "ExpirationDurationDiff", diffBetweenCouponExpirationAndRedemptionInDays, "creationTime", cloudAcct.Created.AsTime())
		if diffBetweenCouponExpirationAndRedemptionInDays < conf.CreditsExpiryMinimumInterval {
			expirationTime = redemptionTime.AddDate(0, 0, int(conf.CreditsExpiryMinimumInterval))
		} else {
			creationTime := cloudAcct.Created.AsTime()
			if expirationTime.Day() != creationTime.Day()+BillLagDays {
				// change expiratiin if falls between month
				expirationTime = expirationTime.AddDate(0, 0, int(conf.CreditsExpiryMinimumInterval))
			}
			log.V(1).Info("coupon expiration", "expirationTime", expirationTime, "creationTime", creationTime)
		}
	}

	req := &pb.CreditInstall{
		CreatedAt:       timestamppb.New(timestamppb.Now().AsTime().UTC()),
		Expiry:          timestamppb.New(expirationTime),
		CloudAccountId:  cloudAccountId,
		OriginalAmount:  amount,
		RemainingAmount: amount,
		CouponCode:      code,
		Reason:          pb.CloudCreditReason_CREDITS_COUPON,
	}

	_, err = CreateCredit(ctx, req, driver)
	if err != nil {
		log.Error(err, "error installing credits")
		return err
	}

	// Update installed columns to true in redemeptions table
	updateRedemptionsQuery := `
		update redemptions set installed = $1
		where code = $2 and cloud_account_id = $3
	`
	_, err = db.ExecContext(ctx, updateRedemptionsQuery, true, code, cloudAccountId)
	if err != nil {
		log.Error(err, "error updating the redemptions table")
		return err
	}

	// cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM
	if !cloudAcct.PaidServicesAllowed {
		paidServicesAllowed := true
		enrolled := true
		terminatePaidServices := false
		lowCredits := false
		_, err := cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, PaidServicesAllowed: &paidServicesAllowed, Enrolled: &enrolled, TerminatePaidServices: &terminatePaidServices, LowCredits: &lowCredits})
		if err != nil {
			log.Error(err, "failed to update cloud account paid services allowed")
			return err
		}
	}

	return nil
}

func CreateCredit(ctx context.Context, obj *pb.CreditInstall, driver *billingCommon.BillingDriverClients) (*emptypb.Empty, error) {

	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CreditsInstaller.CreateCredit").Start()
	defer span.End()
	log.Info("Executing", "CreditsInstaller.CreateCredit for", obj.CloudAccountId)

	if obj.OriginalAmount <= 0 {
		log.Error(errors.New(InvalidBillingCreditAmount), "invalid billing credit amount", "originalAmount", obj.GetOriginalAmount())
		return nil, status.Errorf(codes.InvalidArgument, InvalidBillingCreditAmount)
	}
	if obj.Expiry == nil {
		log.Error(errors.New(InvalidBillingCreditExpiration), "invalid billing credit expiration", "expiration", obj.GetExpiry())
		return nil, status.Errorf(codes.InvalidArgument, InvalidBillingCreditExpiration)
	}
	currentTime := timestamppb.Now().AsTime()
	if obj.Expiry.AsTime().Sub(currentTime) < 0 {
		log.Error(errors.New(InvalidBillingCreditExpiration), "expiration time cannot be lesser than current", "expiration time", obj.Expiry.AsTime())
		return nil, status.Errorf(codes.InvalidArgument, InvalidBillingCreditExpiration)
	}

	cloudAcct, err := cloudacctClient.GetById(ctx,
		&pb.CloudAccountId{Id: obj.CloudAccountId})
	if err != nil {
		log.Error(err, "error getting cloud acct")
		return nil, GetBillingInternalError(FailedToReadCloudAccount, err)
	}

	accountType := cloudAcct.GetType()
	if accountType == pb.AccountType_ACCOUNT_TYPE_PREMIUM || accountType == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE || accountType == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING {

		req := &pb.BillingCredit{
			Created:         obj.CreatedAt,
			Expiration:      obj.Expiry,
			CloudAccountId:  obj.CloudAccountId,
			OriginalAmount:  obj.OriginalAmount,
			RemainingAmount: obj.RemainingAmount,
			CouponCode:      obj.CouponCode,
			Reason:          pb.BillingCreditReason(obj.Reason),
		}

		_, err = driver.BillingCredit.Create(ctx, req)
		if err != nil {
			log.Error(err, "error installing credits")
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}

	params := protodb.NewProtoToSql(obj, fieldOpts...)
	vals := params.GetValues()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Error(err, "error starting db transaction")
		return nil, err
	}
	defer tx.Rollback()

	// Execute insert query
	query := fmt.Sprintf("INSERT INTO cloud_credits (%v) VALUES(%v)",
		params.GetNamesString(), params.GetParamsString())
	if _, err = tx.ExecContext(ctx, query, vals...); err != nil {
		log.Error(err, "error inserting account into db", "query", query)
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		log.Error(err, "error committing db transaction")
		return nil, err
	}
	creditList, err := GetSortedCredits(ctx, obj.CloudAccountId)
	if err != nil {
		log.Error(err, "error getting credit from db")
		return nil, err
	}
	log.Info("credit list", "cloudAccountId", obj.CloudAccountId, "creditListLen", len(creditList))
	err = ProcessCredit(ctx, creditList, obj.CloudAccountId)
	if err != nil {
		log.Error(err, "error processing credit")
		return nil, err
	}

	log.Info("Execution completed CreditsInstaller.CreateCredit")
	return &emptypb.Empty{}, nil
}
