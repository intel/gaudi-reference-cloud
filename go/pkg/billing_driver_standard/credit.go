// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

import (
	"context"
	"database/sql"
	"fmt"
	"math"
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

type StandardBillingCreditService struct {
	pb.UnimplementedBillingCreditServiceServer
	session *sql.DB
}

func (svc *StandardBillingCreditService) Create(ctx context.Context, obj *pb.BillingCredit) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("StandardBillingCreditService.Create").WithValues("cloudAccountId", obj.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	params := protodb.NewProtoToSql(obj, fieldOpts...)
	vals := params.GetValues()

	tx, err := svc.session.BeginTx(ctx, nil)
	if err != nil {
		log.Error(err, "error starting db transaction")
		return nil, err
	}
	defer tx.Rollback()

	// Execute insert query
	query := fmt.Sprintf("INSERT INTO cloud_credits_intel (%v) VALUES(%v)",
		params.GetNamesString(), params.GetParamsString())
	if _, err = tx.ExecContext(ctx, query, vals...); err != nil {
		log.Error(err, "error inserting account into db", "query", query)
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		log.Error(err, "error committing db transaction")
		return nil, err
	}
	creditList, err := billingCommon.GetSortedCredits(ctx, svc.session, obj.CloudAccountId)
	if err != nil {
		log.Error(err, "error getting credit from db")
		return nil, err
	}
	log.Info("credit list", "cloudAccountId", obj.CloudAccountId, "creditListLen", len(creditList))
	err = billingCommon.ProcessCredit(ctx, svc.session, creditList, obj.CloudAccountId)
	if err != nil {
		log.Error(err, "error processing credit")
		return nil, err
	}

	log.Info("Execution completed StandardBillingCreditService.Create")
	return &emptypb.Empty{}, nil
}

func (svc *StandardBillingCreditService) ReadInternal(billingAccount *pb.BillingAccount, outStream pb.BillingCreditService_ReadInternalServer) error {
	ctx := outStream.Context()
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("StandardBillingCreditService.ReadInternal").WithValues("cloudAccountId", billingAccount.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	filterParams := protodb.NewProtoToSql(billingAccount)
	obj := pb.BillingCredit{}

	query := "SELECT coupon_code, cloud_account_id," +
		"created_at, expiry, original_amount, remaining_amount " +
		"FROM cloud_credits"

	names := filterParams.GetNamesString()
	if len(names) > 0 {
		query += fmt.Sprintf(" WHERE (%v) IN (%v)", names,
			filterParams.GetParamsString())
	}
	rows, err := svc.session.Query(query, filterParams.GetValues()...)
	if err != nil {
		log.Error(err, "error executing query")
		return err
	}
	defer rows.Close()
	for rows.Next() {
		creation := time.Time{}
		expiration := time.Time{}
		if err := rows.Scan(&obj.CouponCode, &obj.CloudAccountId, &creation, &expiration,
			&obj.OriginalAmount, &obj.RemainingAmount,
		); err != nil {
			log.Error(err, "error in scan")
			return err
		}
		obj.Created = timestamppb.New(creation)
		obj.Expiration = timestamppb.New(expiration)
		if obj.RemainingAmount < 0 {
			obj.RemainingAmount = 0
		}
		amountUsed := float64(obj.GetOriginalAmount() - obj.GetRemainingAmount())
		obj.AmountUsed = amountUsed

		if err := outStream.Send(&obj); err != nil {
			log.Error(err, "error sending Read records")
		}
	}

	return nil
}

func (svc *StandardBillingCreditService) ReadUnappliedCreditBalance(ctx context.Context, in *pb.BillingAccount) (*pb.BillingUnappliedCreditBalance, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("StandardBillingCreditService.ReadUnappliedCreditBalance").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	filterParams := protodb.NewProtoToSql(in, fieldOpts...)
	vals := filterParams.GetValues()

	obj := pb.BillingUnappliedCreditBalance{}
	readParams := protodb.NewSqlToProto(&obj, fieldOpts...)

	// Executing summation query
	query := fmt.Sprintf("SELECT COALESCE(SUM(remaining_amount),0)  AS %v FROM cloud_credits WHERE %v=$1 AND expiry > NOW()",
		readParams.GetNamesString(), filterParams.GetNamesString())
	log.Info(query)

	rows, err := svc.session.QueryContext(ctx, query, vals...)
	if err != nil {
		log.Error(err, "error executing query")
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&obj.UnappliedAmount); err != nil {
			log.Error(err, "error reading result row.")
			return nil, status.Error(codes.Internal, "error reading rows")
		}
	}

	log.Info("Execution completed StandardBillingCreditService.ReadUnappliedCreditBalance")
	return &obj, nil
}

func (svc *StandardBillingCreditService) Read(ctx context.Context, in *pb.BillingCreditFilter) (*pb.BillingCreditResponse, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("StandardBillingCreditService.Read").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	obj := pb.BillingCreditResponse{}

	// Read Query
	query :=
		"SELECT cc.coupon_code, cc.cloud_account_id, cc.created_at, cc.expiry," +
			"cc.original_amount, cc.remaining_amount, cc.updated_at " +
			"FROM  	( SELECT * FROM cloud_credits WHERE cloud_account_id = $1 ) cc " +
			"JOIN   redemptions r " +
			"ON    	cc.coupon_code = r.code AND cc.cloud_account_id = r.cloud_account_id " +
			"WHERE 	r.installed = true"

	rows, err := svc.session.Query(query, in.CloudAccountId)
	if err != nil {
		log.Error(err, "error executing query")
		return nil, err
	}
	defer rows.Close()
	var total_ra, total_ua float64
	totalUnApplied := float64(0)
	var creation, expiration, updated_at, lastExpirationDate, lastUpdated time.Time
	currentTime := time.Now().UTC()

	for rows.Next() {
		log.Info("reading row")
		credit_obj := pb.BillingCredit{}
		if err := rows.Scan(&credit_obj.CouponCode, &credit_obj.CloudAccountId, &creation, &expiration,
			&credit_obj.OriginalAmount, &credit_obj.RemainingAmount, &updated_at,
		); err != nil {
			log.Error(err, "error reading result row.")
			return nil, err
		}
		log.Info("credit for cloud account", "cloudAccountId", in.CloudAccountId, "creditObj", &credit_obj)
		if lastExpirationDate.Before(expiration) {
			lastExpirationDate = expiration
		}
		if lastUpdated.Before(updated_at) {
			lastUpdated = updated_at
		}
		credit_obj.Created = timestamppb.New(creation)
		credit_obj.Expiration = timestamppb.New(expiration)
		credit_obj.Reason = pb.BillingCreditReason_CREDIT_INITIAL
		credit_obj.AmountUsed = credit_obj.GetOriginalAmount() - credit_obj.GetRemainingAmount()
		if credit_obj.RemainingAmount < 0 {
			credit_obj.AmountUsed = credit_obj.GetOriginalAmount()
			totalUnApplied += math.Abs(credit_obj.GetRemainingAmount())
			log.Info("remaining credit", "cloudAccountId", in.CloudAccountId, "totalUnApplied", totalUnApplied)
			credit_obj.RemainingAmount = 0
		}
		if in.GetHistory() {
			obj.Credits = append(obj.Credits, &credit_obj)
		}

		// calculate total used and remaining amt
		if credit_obj.Expiration.AsTime().After(currentTime) {
			total_ra += credit_obj.RemainingAmount
		}
		total_ua += credit_obj.AmountUsed
	}

	// lastUpdated is current time if there are no credits
	if lastUpdated.IsZero() {
		lastUpdated = time.Now().UTC()
	}
	obj.ExpirationDate = timestamppb.New(lastExpirationDate)
	obj.LastUpdated = timestamppb.New(lastUpdated)

	obj.TotalRemainingAmount = total_ra
	obj.TotalUsedAmount = total_ua
	obj.TotalUnAppliedAmount = totalUnApplied
	log.Info("Execution completed StandardBillingCreditService.Read")
	return &obj, nil
}

func (svc *StandardBillingCreditService) DeleteMigratedCredit(ctx context.Context, in *pb.BillingMigratedCredit) (*emptypb.Empty, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("StandardBillingCreditService.DeleteAppliedCreditBalance").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	query := "UPDATE cloud_credits_intel SET remaining_amount = $1 WHERE cloud_account_id = $2"
	log.Info(query)

	// Updating the remaining_amount to 0 in standard cloudaccount after migration to premium
	rows, err := svc.session.Query(query, 0, in.CloudAccountId)
	if err != nil {
		log.Error(err, "error executing query")
		return nil, err
	}
	defer rows.Close()

	return &emptypb.Empty{}, nil
}
