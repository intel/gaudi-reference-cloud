// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"regexp"
	"time"

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
)

var TestScheduler bool

type BillingCouponService struct {
	pb.UnimplementedBillingCouponServiceServer
	db                           *sql.DB
	creditsExpiryMinimumInterval uint16
}

func NewBillingCouponService(db *sql.DB, creditExpiryMinInterval uint16) (*BillingCouponService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is requied")
	}
	return &BillingCouponService{
		db:                           db,
		creditsExpiryMinimumInterval: creditExpiryMinInterval,
	}, nil
}

func (s *BillingCouponService) Create(ctx context.Context, req *pb.BillingCouponCreate) (*pb.BillingCoupon, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCouponService.Create").WithValues("req", req).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	bytes := make([]byte, couponLengthWithoutDashes)

	if req.NumUses < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "number of uses should be atleast 1")
	}

	if req.Amount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "amount should be greater than 0")
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
	for i := range bytes {
		randomInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowedCharacters))))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate random integer: %v", err)
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
		return nil, status.Errorf(codes.InvalidArgument, "expiration time (%v) cannot be smaller than current time (%v)", req.Expires.AsTime(), creationTime.AsTime())
	}
	if req.Expires.AsTime().Sub(req.Start.AsTime()) < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "expiration time (%v) cannot be smaller than start time (%v)", req.Expires.AsTime(), req.Start.AsTime())
	}

	var isStandard = false
	if req.IsStandard != nil {
		isStandard = *req.IsStandard
	}

	if isStandard {
		if req.NumUses > uint32(Cfg.CouponNumberOfUsesThresholdStandard) {
			msg := fmt.Sprintf("Coupon number of uses %d exceeds the limit of %d for standard coupons", req.NumUses, Cfg.CouponNumberOfUsesThresholdStandard)
			err := status.Errorf(codes.InvalidArgument, msg)
			log.Error(err, msg)
			return nil, err
		}
	} else {
		if req.NumUses > uint32(Cfg.CouponNumberOfUsesThresholdNonStandard) {
			msg := fmt.Sprintf("Coupon number of uses %d exceeds the limit of %d for non-standard coupons", req.NumUses, Cfg.CouponNumberOfUsesThresholdNonStandard)
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
	if _, err := s.db.ExecContext(ctx, query, coupon, req.Amount, req.GetCreator(), start_timestamp, creation_timestamp,
		expiration_timestamp, nil, req.NumUses, 0, isStandard); err != nil {
		log.Error(err, "error creating coupon record in db")
		return nil, status.Errorf(codes.Internal, "coupon record insertion failed")
	}

	return &pb.BillingCoupon{
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

func (s *BillingCouponService) Read(ctx context.Context, req *pb.BillingCouponFilter) (*pb.BillingCouponResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCouponService.Read").WithValues("req", req).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")
	obj := pb.BillingCouponResponse{}
	if (req.Code != nil && *req.Code != "") || ((req.Code != nil && *req.Code == "") && (req.GetRedeemed() == nil)) || (req.Code == nil && req.Redeemed == nil) {
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

			rows, err = s.db.QueryContext(ctx, query, filterParams.GetValues()...)
			if err != nil {
				log.Error(err, "error executing db query")
				return err
			}

			defer rows.Close()

			foundRows := false
			for rows.Next() {
				foundRows = true
				resp := pb.BillingCoupon{}
				var creator sql.NullString
				var startTime time.Time
				var creationTime time.Time
				var expirationTime time.Time
				var disabledTime *time.Time

				if err := rows.Scan(&resp.Code, &resp.Amount, &creator, &startTime, &creationTime,
					&expirationTime, &disabledTime, &resp.NumUses, &resp.NumRedeemed, &resp.IsStandard); err != nil {
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

				var redemptions []*pb.BillingCouponRedemption

				queryRedemptions := "select code, cloud_account_id, redeemed, installed from redemptions " +
					"where code = $1"

				rowsInRedemptions, err := s.db.QueryContext(ctx, queryRedemptions, &resp.Code)
				if err != nil {
					return err
				}
				defer rowsInRedemptions.Close()

				for rowsInRedemptions.Next() {
					var code string
					var cloudaccountId string
					var redemptionTime time.Time
					var installed bool

					if err := rowsInRedemptions.Scan(&code, &cloudaccountId, &redemptionTime, &installed); err != nil {
						return err
					}
					redemption := &pb.BillingCouponRedemption{
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
					return fmt.Errorf("no coupons found")
				}
				return fmt.Errorf("coupon code (%v) does not exists", *req.Code)
			}
			return nil
		}()
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
	} else if req.GetRedeemed() != nil {
		if err := req.Redeemed.CheckValid(); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid redeemed date or format")
		}
		redemptions, err := s.getRedemptions(ctx, s.db, req.Redeemed.AsTime(), req.GetCreator())
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

func (s *BillingCouponService) Redeem(ctx context.Context, req *pb.BillingCouponRedeem) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCouponService.Redeem").WithValues("code", req.Code, "cloudAccountId", req.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	if req.Code == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing coupon code")
	}

	cloudAcct, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: req.CloudAccountId})
	if err != nil {
		log.Error(err, "unable to look up cloud account", "cloudAccountId", req.CloudAccountId, "context", "GetById")
		return nil, status.Errorf(codes.Internal, "error looking up cloud account (%v): %v", req.CloudAccountId, err)
	}

	// Get the driver type for the cloudaccount
	driver, err := GetDriver(ctx, req.CloudAccountId)
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
	if err := s.db.QueryRowContext(ctx, query, req.Code, req.CloudAccountId).Scan(&couponExistsInDb, &couponStartTime, &couponExpirationTime, &couponDisabledTime, &couponIsStandard, &couponAlreadyReedemed); err != nil {
		if err != sql.ErrNoRows {
			log.Error(err, "unable to look up coupon redemptions with code ", "cloudAccountId", req.CloudAccountId, "code", req.Code, "context", "QueryRowContext")
			return nil, status.Errorf(codes.Internal, "%v", err)
		}
	}

	// Check whether coupon exists in DB
	if !couponExistsInDb {
		return nil, status.Errorf(codes.NotFound, "the provided coupon code (%v) does not exists. please provide a valid coupon code", req.Code)
	}
	if couponDisabledTime != nil {
		if couponDisabledTime.Sub(timestamppb.Now().AsTime()) < 0 {
			return nil, status.Errorf(codes.Internal, "coupon %v has been disabled", req.Code)
		}
	}
	if couponStartTime.Sub(timestamppb.Now().AsTime()) > 0 {
		return nil, status.Errorf(codes.Internal, "coupon %v cannot be redeemed before %v", req.Code, couponStartTime)
	}
	if couponExpirationTime.Sub(timestamppb.Now().AsTime()) < 0 {
		return nil, status.Errorf(codes.Internal, "cannot redeem coupon %v as it has expired on %v", req.Code, couponExpirationTime)
	}
	if couponAlreadyReedemed {
		return nil, status.Errorf(codes.AlreadyExists, "the coupon code (%v) has already been redeemed", req.Code)
	}
	if (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_STANDARD) && !couponIsStandard {
		return nil, status.Errorf(codes.Internal, "cannot redeem coupon %v for standard customers", req.Code)
	}
	if (cloudAcct.Type != pb.AccountType_ACCOUNT_TYPE_STANDARD) && couponIsStandard {
		return nil, status.Errorf(codes.Internal, "coupon %v can be redeemed only for standard customers", req.Code)
	}

	// Begin the transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error(err, "unable to begin transactions")
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
		log.Error(err, "error updating number of redemption in the coupons table", "cloudAccountId", req.CloudAccountId, "code", req.Code, "context", "ExecContext")
		return nil, status.Errorf(codes.Internal, "failed to create redeem record: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error(err, "error fetching affected rows", "cloudAccountId", req.CloudAccountId, "context", "RowsAffected")
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if rowsAffected < 1 {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to redeem the coupon (%v)", req.Code)
	}

	// Create a redemption record in the DB
	queryToCreateRedeemRecord := `
		insert into redemptions (code, cloud_account_id, redeemed, installed)
		values ($1, $2, $3, $4)
	`
	redemptionTime := timestamppb.Now().AsTime().UTC().Format(time.RFC3339)
	_, err = tx.ExecContext(ctx, queryToCreateRedeemRecord, req.Code, req.CloudAccountId, redemptionTime, false)
	if err != nil {
		log.Error(err, "error inserting redeem record in db", "cloudAccountId", req.CloudAccountId, "code", req.Code, "redemptionTime", redemptionTime, "context", "ExecContext")
		return nil, status.Errorf(codes.Internal, "redeem record insertion failed: %v", err)
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		log.Error(err, "error committing transaction")
		return nil, status.Errorf(codes.Internal, "redeem record insertion failed: %v", err)
	}

	if !TestScheduler {
		// Create a new sync credit install scheduler
		creditsInstallSched := CreditsInstallScheduler{
			db: s.db,
		}

		// Install the Coupon
		installErr := creditsInstallSched.InstallCoupon(ctx, req.Code, req.CloudAccountId, driver, s.creditsExpiryMinimumInterval)
		if installErr != nil {
			// Delete the redemption record from the database
			queryDelete := `
			delete from redemptions where code = $1 and cloud_account_id = $2 and installed = $3
			`
			result, err := s.db.ExecContext(ctx, queryDelete, req.Code, req.CloudAccountId, false)
			if err != nil {
				log.Error(err, "error encountered while deleting the record from the redemption", "cloudAccountId", req.CloudAccountId, "code", req.Code, "context", "ExecContext")
				return nil, status.Errorf(codes.Internal, "error encountered while installing coupon: %v", err)
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				log.Error(err, "error fetching affected rows", "cloudAccountId", req.CloudAccountId, "context", "RowsAffected")
				return nil, status.Errorf(codes.Internal, "%v", err)
			}
			if rowsAffected < 1 {
				return nil, status.Errorf(codes.NotFound, "record with coupon (%v), cloud_account_id (%v) and"+
					"installed (%v) not found", req.Code, req.CloudAccountId, false)
			}

			// Revert back the number of redemptions in the coupons table
			queryRevertRedemptionsNumber := `
			update coupons set num_redeemed = num_redeemed - 1
			where code = $1
			`
			result, err = s.db.ExecContext(ctx, queryRevertRedemptionsNumber, req.Code)
			if err != nil {
				log.Error(err, "error encountered while reverting the number of redemptions in the coupons table", "code", req.Code)
				return nil, status.Errorf(codes.Internal, "error encountered while installing coupon: %v", err)
			}
			rowsAffected, err = result.RowsAffected()
			if err != nil {
				log.Error(err, "error fetching affected rows", "cloudAccountId", req.CloudAccountId, "context", "RowsAffected")
				return nil, status.Errorf(codes.Internal, "%v", err)
			}
			if rowsAffected < 1 {
				return nil, status.Errorf(codes.NotFound, "record with coupon (%v) not found", req.Code)
			}

			return nil, status.Errorf(codes.Internal, "error encountered while applying coupon: %v", installErr)
		}
	}

	log.Info("redeem sucess for cloudAccountId", "cloudAccountId", req.CloudAccountId)

	return &emptypb.Empty{}, nil
}

func (s *BillingCouponService) Disable(ctx context.Context, req *pb.BillingCouponDisable) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCouponService.Disable").WithValues("code", req.Code).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	if req.Code == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing coupon code")
	}
	if req.Disabled == nil {
		req.Disabled = timestamppb.Now()
	}

	query := "update coupons set disabled = $1 where code = $2"
	result, err := s.db.ExecContext(ctx, query, req.Disabled.AsTime().UTC().Format(time.RFC3339), req.Code)
	if err != nil {
		log.Error(err, "error encountered while disabling coupon")
		return nil, status.Errorf(codes.Internal, "disabling coupon failed: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	if rowsAffected < 1 {
		return nil, status.Errorf(codes.NotFound, "the provided coupon code (%v) does not exist. please provide a valid coupon code", req.Code)
	}
	return &emptypb.Empty{}, nil
}

func CheckCouponCode(code string) error {
	regex := regexp.MustCompile(`^[ABCDEFGHJKLMNPQRSTUVWXYZ1234567890]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ1234567890]{4}-[ABCDEFGHJKLMNPQRSTUVWXYZ1234567890]{4}$`)
	if !regex.MatchString(code) {
		return fmt.Errorf("provided coupon code (%v) is not in correct format. Expected format is 1FJN-35Z2-95XB where the characters are capital letters and digits except letters I and O", code)
	}
	return nil
}

func (s *BillingCouponService) getRedemptions(ctx context.Context, db *sql.DB, redeemedDate time.Time, creator string) ([]*pb.BillingCouponRedemption, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCouponService.getRedemptions").WithValues("redeemedDate", redeemedDate).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

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

	var redemptions []*pb.BillingCouponRedemption
	for rows.Next() {
		var code string
		var cloudAccountId string
		var redemptionTime time.Time
		var installed bool
		var creator string

		if err := rows.Scan(&code, &cloudAccountId, &redemptionTime, &installed, &creator); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		redemption := &pb.BillingCouponRedemption{
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
