// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing_driver_intel

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"time"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var fieldOpts []protodb.FieldOptions = []protodb.FieldOptions{
	{Name: "", StoreEmptyStringAsNull: false},
}

type IntelBillingCreditService struct {
	pb.UnimplementedBillingCreditServiceServer
	session *sql.DB
}

func (svc *IntelBillingCreditService) Create(ctx context.Context, obj *pb.BillingCredit) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("IntelBillingCreditService.Create").WithValues("cloudAccountId", obj.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")
	log.Info("Executing", "IntelBillingCreditService.Create for", obj.CloudAccountId)

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

	// creditList, err := billingCommon.GetSortedCredits(ctx, svc.session, obj.CloudAccountId)
	// if err != nil {
	// 	log.Error(err, "error getting credit from db")
	// 	return nil, err
	// }

	// lengthOfCreditList := len(creditList)
	// span.SetAttributes(attribute.String("creditListLen", strconv.Itoa(lengthOfCreditList)))
	// log.Info("credit list", "cloudAccountId", obj.CloudAccountId, "creditListLen", lengthOfCreditList)

	return &emptypb.Empty{}, nil
}

func (svc *IntelBillingCreditService) ReadInternal(billingAccount *pb.BillingAccount, outStream pb.BillingCreditService_ReadInternalServer) error {
	ctx := outStream.Context()
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("IntelBillingCreditService.ReadInternal").WithValues("cloudAccountId", billingAccount.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	filterParams := protodb.NewProtoToSql(billingAccount)
	obj := pb.BillingCredit{}

	query := "SELECT coupon_code, cloud_account_id," +
		"created_at, expiry, original_amount ,remaining_amount " +
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
			log.Error(err, "error on scan rows")
			return err
		}
		obj.Created = timestamppb.New(creation)
		obj.Expiration = timestamppb.New(expiration)

		amountUsed := float64(obj.GetOriginalAmount() - obj.GetRemainingAmount())
		obj.AmountUsed = amountUsed

		if err := outStream.Send(&obj); err != nil {
			log.Error(err, "error sending Read records")
		}
	}

	return nil
}

func (svc *IntelBillingCreditService) ReadUnappliedCreditBalance(ctx context.Context, in *pb.BillingAccount) (*pb.BillingUnappliedCreditBalance, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("IntelBillingCreditService.ReadUnappliedCreditBalance").WithValues("cloudAccountId", in.CloudAccountId).Start()
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

	return &obj, nil
}

func (svc *IntelBillingCreditService) Read(ctx context.Context, in *pb.BillingCreditFilter) (*pb.BillingCreditResponse, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("IntelBillingCreditService.Read").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	response := pb.BillingCreditResponse{}

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
		creditObj := pb.BillingCredit{}
		if err := rows.Scan(&creditObj.CouponCode, &creditObj.CloudAccountId, &creation, &expiration,
			&creditObj.OriginalAmount, &creditObj.RemainingAmount, &updated_at,
		); err != nil {
			log.Error(err, "error reading result row.")
			return nil, err
		}
		log.V(9).Info("credit for cloud account", "cloudAccountId", in.CloudAccountId, "creditObj", &creditObj)
		if lastExpirationDate.Before(expiration) {
			lastExpirationDate = expiration
		}
		if lastUpdated.Before(updated_at) {
			lastUpdated = updated_at
		}
		creditObj.Created = timestamppb.New(creation)
		creditObj.Expiration = timestamppb.New(expiration)
		creditObj.Reason = pb.BillingCreditReason_CREDIT_INITIAL
		creditObj.AmountUsed = creditObj.GetOriginalAmount() - creditObj.GetRemainingAmount()

		if creditObj.RemainingAmount < 0 {
			creditObj.AmountUsed = creditObj.GetOriginalAmount()
			totalUnApplied += math.Abs(creditObj.GetRemainingAmount())

			span.SetAttributes(attribute.String("totalUnapplied", strconv.FormatFloat(totalUnApplied, 'f', -1, 64)))
			log.V(9).Info("remaining credit", "cloudAccountId", in.CloudAccountId, "totalUnapplied", totalUnApplied)
			creditObj.RemainingAmount = 0
		}

		if in.GetHistory() {
			response.Credits = append(response.Credits, &creditObj)
		}

		// calculate total used and remaining amt
		if creditObj.Expiration.AsTime().After(currentTime) {
			total_ra += creditObj.RemainingAmount
		}
		total_ua += creditObj.AmountUsed
	}

	// lastUpdated is current time if there are no credits
	if lastUpdated.IsZero() {
		lastUpdated = time.Now().UTC()
	}
	response.ExpirationDate = timestamppb.New(lastExpirationDate)
	response.LastUpdated = timestamppb.New(lastUpdated)

	response.TotalRemainingAmount = total_ra
	response.TotalUsedAmount = total_ua
	response.TotalUnAppliedAmount = totalUnApplied

	return &response, nil
}
