// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcredits

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"regexp"
	"time"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	couponLengthWithoutDashes = 12
	allowedCharacters         = "ABCDEFGHJKLMNPQRSTUVWXYZ0123456789"
	BillLagDays               = 7
)

type CloudCreditsCouponService struct {
	pb.UnimplementedCloudCreditsCouponServiceServer
}

func (svc *CloudCreditsCouponService) Create(ctx context.Context, req *pb.CouponCreate) (*pb.Coupon, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCouponService.Create").Start()
	defer span.End()
	log.Info("BEGIN")
	log.WithValues("Req", req)
	defer log.Info("END")

	bytes := make([]byte, couponLengthWithoutDashes)

	if req.NumUses < 1 {
		msg := "number of uses should be atleast 1"
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}
	if req.Amount <= 0 {
		msg := "amount should be greater than 0"
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}
	creatorEmail := req.GetCreator()
	if creatorEmail == "" {
		msg := "creator email is required"
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}
	if !isValidEmail(creatorEmail) {
		msg := fmt.Sprintf("invalid email format: %s", creatorEmail)
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}

	requesterEmail, err := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.EmailClaim)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if requesterEmail != req.Creator {
		msg := "User is not authorized to perform this action"
		err := status.Errorf(codes.PermissionDenied, msg)
		log.Error(err, msg)
		return nil, err
	}

	for i := range bytes {
		randomInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowedCharacters))))
		if err != nil {
			msg := fmt.Sprintf("failed to generate random integer: %v", err)
			log.Error(err, msg)
			return nil, err
		}
		bytes[i] = allowedCharacters[randomInt.Int64()]
	}
	groups := string(bytes)
	coupon := fmt.Sprintf("%s-%s-%s", groups[:4], groups[4:8], groups[8:])

	creationTime := timestamppb.Now()
	if req.Start == nil {
		req.Start = creationTime
	}
	if req.Expires == nil {
		ninetyDaysFromCreation := creationTime.AsTime().AddDate(0, 0, 90)
		req.Expires = timestamppb.New(ninetyDaysFromCreation)
	}
	if req.Expires.AsTime().Sub(creationTime.AsTime()) < 0 {
		msg := fmt.Sprintf("Expiration time (%v) cannot be smaller than current time (%v)", req.Expires.AsTime(), creationTime.AsTime())
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}
	if req.Expires.AsTime().Sub(req.Start.AsTime()) < 0 {
		msg := fmt.Sprintf("Expiration time (%v) cannot be smaller than start time (%v)", req.Expires.AsTime(), req.Start.AsTime())
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}

	var isStandard = false
	if req.IsStandard != nil {
		isStandard = *req.IsStandard
	}

	if isStandard {
		if req.NumUses > uint32(config.Cfg.CouponNumberOfUsesThresholdStandard) {
			msg := fmt.Sprintf("Coupon number of uses %d exceeds the limit of %d for standard coupons", req.NumUses, config.Cfg.CouponNumberOfUsesThresholdStandard)
			err := status.Errorf(codes.InvalidArgument, msg)
			log.Error(err, msg)
			return nil, err
		}
	} else {
		if req.NumUses > uint32(config.Cfg.CouponNumberOfUsesThresholdNonStandard) {
			msg := fmt.Sprintf("Coupon number of uses %d exceeds the limit of %d for non-standard coupons", req.NumUses, config.Cfg.CouponNumberOfUsesThresholdNonStandard)
			err := status.Errorf(codes.InvalidArgument, msg)
			log.Error(err, msg)
			return nil, err
		}
	}

	start_timestamp := req.Start.AsTime().UTC().Format(time.RFC3339)
	creation_timestamp := creationTime.AsTime().UTC().Format(time.RFC3339)
	expiration_timestamp := req.Expires.AsTime().UTC().Format(time.RFC3339)

	query := `
		insert into coupons (code, amount, creator, start, created, expires, disabled, num_uses, num_redeemed, is_standard)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	if _, err := db.ExecContext(ctx, query, coupon, req.Amount, req.GetCreator(), start_timestamp, creation_timestamp,
		expiration_timestamp, nil, req.NumUses, 0, isStandard); err != nil {
		log.Error(err, "error creating coupon record in db")
		return nil, status.Errorf(codes.Internal, "coupon record insertion failed")
	}

	return &pb.Coupon{
		Code:        coupon,
		Amount:      req.Amount,
		Creator:     req.GetCreator(),
		Created:     creationTime,
		Start:       req.Start,
		Expires:     req.Expires,
		Disabled:    nil,
		NumUses:     req.NumUses,
		NumRedeemed: 0,
		IsStandard:  req.IsStandard,
	}, nil
}

func (s *CloudCreditsCouponService) Redeem(ctx context.Context, req *pb.CloudCreditsCouponRedeem) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCouponService.Redeem").Start()

	log.WithValues("Code", req.Code, "CloudAccountId", req.CloudAccountId)
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	if req.Code == "" {
		msg := "missing coupon code"
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}

	cloudAcct, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: req.CloudAccountId})
	if err != nil {
		log.Error(err, "unable to look up cloud account", "cloudAccountId", req.CloudAccountId)
		return nil, status.Errorf(codes.Internal, "error looking up cloud account (%v): %v", req.CloudAccountId, err)
	}

	// Get the driver type for the cloudaccount
	driver, err := billingCommon.GetDriver(ctx, req.CloudAccountId)
	if err != nil {
		log.Error(err, "unable to find driver", "cloudAccountId", req.CloudAccountId)
		return nil, status.Errorf(codes.NotFound, "error finding driver for cloudaccount (%v): %v", req.CloudAccountId, err)
	}

	var couponStartTime time.Time
	var couponExpirationTime time.Time
	var couponDisabledTime *time.Time
	var couponExistsInDb bool
	var couponAlreadyReedemed bool
	var couponIsStandard bool
	couponPattern := `^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$`
	couponPatternre := regexp.MustCompile(couponPattern)

	if !couponPatternre.MatchString(req.Code) {
		err := status.Errorf(codes.Internal, "Invalid coupon code format")
		log.Error(err, "Invalid coupon code format")
		return nil, err
	}

	query := `
		select
			exists(select code from coupons where code = $1) AS coupon_exists,
			start,
			expires,
			disabled,
			is_standard,
			exists(SELECT (code, cloud_account_id) from redemptions where code = $1 AND cloud_account_id = $2) AS redemption_exists
		from coupons
		where code = $1;
	`
	if err := db.QueryRowContext(ctx, query, req.Code, req.CloudAccountId).Scan(&couponExistsInDb, &couponStartTime, &couponExpirationTime, &couponDisabledTime, &couponIsStandard, &couponAlreadyReedemed); err != nil {
		if err != sql.ErrNoRows {
			err := status.Errorf(codes.Internal, "Error looking up coupon in db")
			log.Error(err, "Error looking up coupon in db")
			return nil, err
		}
	}

	// Check whether coupon exists in DB
	if !couponExistsInDb {
		msg := fmt.Sprintf("the provided coupon code (%v) does not exists. please provide a valid coupon code", req.Code)
		err := status.Errorf(codes.NotFound, msg)
		log.Error(err, msg)
		return nil, err
	}
	if couponDisabledTime != nil {
		if couponDisabledTime.Sub(timestamppb.Now().AsTime()) < 0 {
			msg := fmt.Sprintf("coupon %v has been disabled", req.Code)
			err := status.Errorf(codes.Internal, msg)
			log.Error(err, msg)
			return nil, err
		}
	}
	if couponStartTime.Sub(timestamppb.Now().AsTime()) > 0 {
		msg := fmt.Sprintf("coupon %v cannot be redeemed before %v", req.Code, couponStartTime)
		err := status.Errorf(codes.Internal, msg)
		log.Error(err, msg)
		return nil, err
	}
	if couponExpirationTime.Sub(timestamppb.Now().AsTime()) < 0 {
		msg := fmt.Sprintf("cannot redeem coupon %v as it has expired on %v", req.Code, couponExpirationTime)
		err := status.Errorf(codes.Internal, msg)
		log.Error(err, msg)
		return nil, err
	}
	if couponAlreadyReedemed {
		msg := fmt.Sprintf("the coupon code (%v) has already been redeemed", req.Code)
		err := status.Errorf(codes.AlreadyExists, msg)
		log.Error(err, msg)
		return nil, err
	}
	if (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_STANDARD) && !couponIsStandard {
		msg := fmt.Sprintf("cannot redeem coupon %v for standard customers", req.Code)
		err := status.Errorf(codes.Internal, msg)
		log.Error(err, msg)
		return nil, err
	}
	if (cloudAcct.Type != pb.AccountType_ACCOUNT_TYPE_STANDARD) && couponIsStandard {
		msg := fmt.Sprintf("coupon %v can be redeemed only for standard customers", req.Code)
		err := status.Errorf(codes.Internal, msg)
		log.Error(err, msg)
		return nil, err
	}

	// Begin the transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Error(err, "Error in BeginTx")
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	// update the number of redemptions in coupons table
	queryToUpdateCouponsTable := `
		update coupons set num_redeemed = num_redeemed + 1
		where code = $1 and num_redeemed < num_uses and disabled is null
	`
	result, err := tx.ExecContext(ctx, queryToUpdateCouponsTable, req.Code)
	if err != nil {
		log.Error(err, "error updating number of redemption in the coupons table")
		return nil, status.Errorf(codes.Internal, "failed to create redeem record: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error(err, "Error in rowsaffected")
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if rowsAffected < 1 {
		msg := fmt.Sprintf("failed to redeem the coupon (%v)", req.Code)
		err := status.Errorf(codes.FailedPrecondition, msg)
		log.Error(err, msg)
		return nil, err
	}

	// Create a redemption record in the DB
	queryToCreateRedeemRecord := `
		insert into redemptions (code, cloud_account_id, redeemed, installed)
		values ($1, $2, $3, $4)
	`
	redemption_time := timestamppb.Now().AsTime().UTC().Format(time.RFC3339)
	_, err = tx.ExecContext(ctx, queryToCreateRedeemRecord, req.Code, req.CloudAccountId, redemption_time, false)
	if err != nil {
		log.Error(err, "error inserting redeem record in db")
		return nil, status.Errorf(codes.Internal, "redeem record insertion failed: %v", err)
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		log.Error(err, "error committing transaction")
		return nil, status.Errorf(codes.Internal, "redeem record insertion failed: %v", err)
	}

	// Install the Coupon
	installErr := InstallCoupon(ctx, req.Code, req.CloudAccountId, driver)
	if installErr != nil {
		// Delete the redemption record from the database
		queryDelete := `
		delete from redemptions where code = $1 and cloud_account_id = $2 and installed = $3
		`
		result, err := db.ExecContext(ctx, queryDelete, req.Code, req.CloudAccountId, false)
		if err != nil {
			log.Error(err, "error encountered while deleting the record from the redemption")
			return nil, status.Errorf(codes.Internal, "error encountered while installing coupon: %v", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Error(err, "Error in rowsaffected")
			return nil, status.Errorf(codes.Internal, "%v", err)
		}
		if rowsAffected < 1 {
			msg := fmt.Sprintf("record with coupon (%v), cloud_account_id (%v) and installed (%v) not found", req.Code, req.CloudAccountId, false)
			err := status.Errorf(codes.NotFound, msg)
			return nil, err
		}

		// Revert back the number of redemptions in the coupons table
		queryRevertRedemptionsNumber := `
		update coupons set num_redeemed = num_redeemed - 1
		where code = $1
		`
		result, err = db.ExecContext(ctx, queryRevertRedemptionsNumber, req.Code)
		if err != nil {
			log.Error(err, "error encountered while reverting the number of redemptions in the coupons table")
			return nil, status.Errorf(codes.Internal, "error encountered while installing coupon: %v", err)
		}
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			log.Error(err, "Error in rowsaffected")
			return nil, status.Errorf(codes.Internal, "%v", err)
		}
		if rowsAffected < 1 {
			msg := fmt.Sprintf("record with coupon (%v) not found", req.Code)
			err := status.Errorf(codes.NotFound, msg)
			log.Error(err, msg)
			return nil, pb.ErrorValidationError{}
		}
		msg := fmt.Sprintf("error encountered while applying coupon: %v", installErr)
		err = status.Errorf(codes.Internal, msg)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CloudCreditsCouponService) Read(ctx context.Context, req *pb.CouponFilter) (*pb.CouponResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCouponService.Read").Start()
	log.WithValues("Req", req)
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")
	obj := pb.CouponResponse{}
	if (req.Code != nil && *req.Code != "") || ((req.Code != nil && *req.Code == "") && (req.Redeemed == nil)) || (req.Code == nil && req.Redeemed == nil) {
		req.Redeemed = nil
		err := func() error {
			var rows *sql.Rows
			var err error

			query := "select code, amount, creator, start, created, expires, disabled, num_uses, num_redeemed, is_standard from coupons"
			filterParams := protodb.NewProtoToSql(req)
			filter := filterParams.GetFilter()
			if filter != "" {
				query += " " + filter
			}
			rows, err = db.QueryContext(ctx, query, filterParams.GetValues()...)
			if err != nil {
				log.Error(err, "error executing db query")
				return err
			}

			defer rows.Close()

			foundRows := false
			for rows.Next() {
				foundRows = true
				resp := pb.Coupon{}
				var creator sql.NullString
				var startTime time.Time
				var creationTime time.Time
				var expirationTime time.Time
				var disabledTime *time.Time

				if err := rows.Scan(&resp.Code, &resp.Amount, &creator, &startTime, &creationTime,
					&expirationTime, &disabledTime, &resp.NumUses, &resp.NumRedeemed, &resp.IsStandard); err != nil {
					log.Error(err, "error scanning row")
					return err
				}
				if creator.Valid {
					resp.Creator = creator.String
				} else {
					resp.Creator = ""
				}

				resp.Start = timestamppb.New(startTime)
				resp.Created = timestamppb.New(creationTime)
				resp.Expires = timestamppb.New(expirationTime)

				if disabledTime == nil {
					resp.Disabled = nil
				} else {
					resp.Disabled = timestamppb.New(*disabledTime)
				}

				var redemptions []*pb.CouponRedemption

				queryRedemptions := "select code, cloud_account_id, redeemed, installed from redemptions " +
					"where code = $1"

				rowsInRedemptions, err := db.QueryContext(ctx, queryRedemptions, &resp.Code)
				if err != nil {
					log.Error(err, "error executing db query")
					return err
				}
				defer rowsInRedemptions.Close()

				for rowsInRedemptions.Next() {
					var code string
					var cloudaccountId string
					var redemptionTime time.Time
					var installed bool

					if err := rowsInRedemptions.Scan(&code, &cloudaccountId, &redemptionTime, &installed); err != nil {
						log.Error(err, "error scanning row")
						return err
					}
					redemption := &pb.CouponRedemption{
						Code:           code,
						CloudAccountId: cloudaccountId,
						Redeemed:       timestamppb.New(redemptionTime),
						Installed:      installed,
					}
					redemptions = append(redemptions, redemption)
				}
				resp.Redemptions = redemptions
				obj.Coupons = append(obj.Coupons, &resp)
			}

			if !foundRows {
				if req.Code == nil {
					err := fmt.Errorf("no coupons found")
					log.Error(err, "no coupons found")
					return err
				}
				err := fmt.Errorf("coupon code (%v) does not exists", *req.Code)
				log.Error(err, "coupons  code not exists")
				return err
			}
			return nil
		}()
		if err != nil {
			log.Error(err, "error encountered while reading coupon")
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
	} else if req.GetRedeemed() != nil {
		if err := req.Redeemed.CheckValid(); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid redeemed date or format")
		}
		redemptions, err := GetRedemptions(ctx, db, req.Redeemed.AsTime(), req.GetCreator())
		if err != nil {
			if err.Error() == "no coupon redemptions found for the redeemed date" && req.Code != nil {
				return nil, status.Errorf(codes.NotFound, "coupon redemptions does not exist for the redeemed date")
			}
			return nil, status.Errorf(codes.Internal, "%v", err)
		}
		obj.Redemptions = redemptions
	}
	return &obj, nil
}

func (s *CloudCreditsCouponService) Disable(ctx context.Context, req *pb.CouponDisable) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCouponService.Disable").Start()
	log.WithValues("Code", req.Code)
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	if req.Code == "" {
		msg := "missing coupon code"
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}

	val, err := s.Read(ctx, &pb.CouponFilter{Code: &req.Code})
	if err != nil {
		log.Error(err, "error encountered while reading coupon")
		return nil, status.Errorf(codes.Internal, "error encountered while reading coupon: %v", err)
	}

	if val.Coupons[0].Disabled == nil {
		req.Disabled = timestamppb.Now()
	} else {
		msg := "coupon already disabled"
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}

	query := "update coupons set disabled = $1 where code = $2"
	result, err := db.ExecContext(ctx, query, req.Disabled.AsTime().UTC().Format(time.RFC3339), req.Code)
	if err != nil {
		log.Error(err, "error encountered while disabling coupon")
		return nil, status.Errorf(codes.Internal, "disabling coupon failed: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error(err, "error in RowsAffected")
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if rowsAffected < 1 {
		msg := fmt.Sprintf("the provided coupon code (%v) does not exist. please provide a valid coupon code", req.Code)
		err := status.Errorf(codes.NotFound, msg)
		log.Error(err, msg)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func GetRedemptions(ctx context.Context, db *sql.DB, redeemedDate time.Time, creator string) ([]*pb.CouponRedemption, error) {
	query := "SELECT r.code, r.cloud_account_id, r.redeemed, r.installed, c.creator FROM redemptions r JOIN coupons c ON r.code = c.code WHERE r.redeemed >= $1"
	args := []interface{}{redeemedDate.Format(time.RFC3339)}
	if creator != "" {
		query += " AND creator = $2"
		args = append(args, creator)
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing db query: %w", err)
	}
	defer rows.Close()

	var redemptions []*pb.CouponRedemption
	for rows.Next() {
		var code string
		var cloudAccountId string
		var redemptionTime time.Time
		var installed bool
		var creator string

		if err := rows.Scan(&code, &cloudAccountId, &redemptionTime, &installed, &creator); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		redemption := &pb.CouponRedemption{
			Code:           code,
			CloudAccountId: cloudAccountId,
			Redeemed:       timestamppb.New(redemptionTime),
			Installed:      installed,
			Creator:        creator,
		}
		redemptions = append(redemptions, redemption)
	}

	if len(redemptions) == 0 {
		return nil, fmt.Errorf("no coupon redemptions found for the redeemed date")
	}

	return redemptions, nil
}
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}
