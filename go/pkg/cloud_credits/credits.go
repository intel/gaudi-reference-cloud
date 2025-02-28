// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcredits

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CloudCreditsCreditService struct {
	pb.CloudCreditsCreditServiceServer
}
type CloudCreditsServiceClient struct {
	CloudCreditsSvcClient pb.CloudCreditsCreditServiceClient
}

func NewCloudCreditsService(ctx context.Context, resolver grpcutil.Resolver) (*CloudCreditsServiceClient, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.NewCloudCreditsServiceClient").Start()
	defer span.End()
	var cloudCreditConn *grpc.ClientConn
	cloudCreditAddr, err := resolver.Resolve(ctx, "cloudcredits")
	if err != nil {
		log.Error(err, "grpc resolver not able to resolve", "addr", cloudCreditAddr)
		return nil, err
	}
	cloudCreditConn, err = grpcConnect(ctx, cloudCreditAddr)
	if err != nil {
		return nil, err
	}
	ca := pb.NewCloudCreditsCreditServiceClient(cloudCreditConn)
	return &CloudCreditsServiceClient{CloudCreditsSvcClient: ca}, nil
}

func (svc *CloudCreditsCreditService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.Ping").Start()
	defer span.End()
	logger.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (s *CloudCreditsCreditService) Create(ctx context.Context, in *pb.Credit) (*empty.Empty, error) {

	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.Create").Start()
	log.WithValues("cloudaccountid", in.CloudAccountId)
	defer span.End()
	log.Info("Executing", "CloudCreditsCreditService.Create for", in.CloudAccountId)

	cloudAcct, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: in.CloudAccountId})
	if err != nil {
		log.Error(err, "error getting cloud acct")
		return nil, err
	}

	// Get the driver type for the cloudaccount
	driver, err := billingCommon.GetDriver(ctx, in.CloudAccountId)
	if err != nil {
		log.Error(err, "unable to find driver", "cloudAccountId", in.CloudAccountId)
		return nil, status.Errorf(codes.NotFound, "error finding driver for cloudaccount (%v): %v", in.CloudAccountId, err)
	}

	if in.Expiration == nil {
		msg := "Expiration time cannot be empty or nil"
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}

	creationTime := timestamppb.Now()
	if in.Expiration.AsTime().Sub(creationTime.AsTime()) < 0 {
		msg := fmt.Sprintf("Expiration time (%v) cannot be smaller than current time (%v)", in.Expiration.AsTime(), creationTime.AsTime())
		err := status.Errorf(codes.InvalidArgument, msg)
		log.Error(err, msg)
		return nil, err
	}

	expirationTime := in.Expiration.AsTime()
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
		CloudAccountId:  in.CloudAccountId,
		OriginalAmount:  in.OriginalAmount,
		RemainingAmount: in.RemainingAmount,
		CouponCode:      in.CouponCode,
	}

	_, err = CreateCredit(ctx, req, driver)
	if err != nil {
		log.Error(err, "error installing credits")
		return nil, err
	}

	// Update installed columns to true in redemeptions table
	updateRedemptionsQuery := `
		update redemptions set installed = $1
		where code = $2 and cloud_account_id = $3
	`
	_, err = db.ExecContext(ctx, updateRedemptionsQuery, true, in.CouponCode, in.CloudAccountId)
	if err != nil {
		log.Error(err, "error updating the redemptions table")
		return nil, err
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
			return nil, err
		}
	}

	log.Info("redeem sucess for cloudAccountId", "cloudAccountId", req.CloudAccountId)

	return &emptypb.Empty{}, nil
}

func (svc *CloudCreditsCreditService) ReadCredits(ctx context.Context, in *pb.CreditFilter) (*pb.CreditResponse, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.ReadCredits").Start()
	log.WithValues("cloudaccountid", in.CloudAccountId)
	defer span.End()
	log.Info("Executing", "CloudCreditsCreditService.ReadCredits for", in.CloudAccountId)

	obj := pb.CreditResponse{}
	driver, err := billingCommon.GetDriverAll(ctx, in.CloudAccountId)
	if err != nil {
		return nil, err
	}

	req := &pb.BillingCreditFilter{
		CloudAccountId: in.CloudAccountId,
		History:        in.History,
	}

	res, err := driver.BillingCredit.Read(ctx, req)
	if err != nil {
		log.Error(err, "failed to read billing credit")
		return nil, status.Errorf(codes.Internal, GetBillingError(FailedToReadBillingCredit, err).Error())
	}
	obj.TotalRemainingAmount = res.TotalRemainingAmount
	obj.TotalUsedAmount = res.TotalUsedAmount
	obj.TotalUnAppliedAmount = res.TotalUnAppliedAmount
	obj.LastUpdated = res.LastUpdated
	obj.ExpirationDate = res.ExpirationDate

	for _, cred := range res.Credits {
		creditDetails := &pb.Credit{
			Created:         cred.Created,
			Expiration:      cred.Expiration,
			CloudAccountId:  cred.CloudAccountId,
			OriginalAmount:  cred.OriginalAmount,
			RemainingAmount: cred.RemainingAmount,
			CouponCode:      cred.CouponCode,
			AmountUsed:      cred.AmountUsed,
			Reason:          pb.CloudCreditReason(cred.Reason),
		}

		obj.Credits = append(obj.Credits, creditDetails)
	}
	return &obj, err
}

func (svc *CloudCreditsCreditService) ReadUnappliedCreditBalance(ctx context.Context, in *pb.Account) (*pb.UnappliedCreditBalance, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.ReadUnappliedCreditBalance").Start()
	log.WithValues("Cloudaccountid", in.CloudAccountId)
	defer span.End()
	log.Info("Executing CloudCreditsCreditService.ReadUnappliedCreditBalance")

	obj := pb.UnappliedCreditBalance{}

	cloudAcct, err := cloudacctClient.GetById(ctx,
		&pb.CloudAccountId{Id: in.CloudAccountId})
	if err != nil {
		log.Error(err, "error getting cloud account")
		return nil, GetBillingInternalError(FailedToReadCloudAccount, err)
	}

	accountType := cloudAcct.GetType()
	if accountType != pb.AccountType_ACCOUNT_TYPE_STANDARD && accountType != pb.AccountType_ACCOUNT_TYPE_INTEL {
		driver, err := billingCommon.GetDriverAll(ctx, in.CloudAccountId)
		if err != nil {
			log.Error(err, "unable to find driver", "cloudAccountId", in.CloudAccountId)
			return nil, status.Errorf(codes.InvalidArgument, GetBillingError(InvalidCloudAcct, err).Error())
		}

		req := &pb.BillingAccount{
			CloudAccountId: in.CloudAccountId,
		}

		res, err := driver.BillingCredit.ReadUnappliedCreditBalance(ctx, req)
		if err != nil {
			log.Error(err, "failed to read billing credit")
			return nil, status.Errorf(codes.Internal, GetBillingError(FailedToReadBillingCredit, err).Error())
		}

		obj.UnappliedAmount = res.UnappliedAmount
		return &obj, err
	}

	filterParams := protodb.NewProtoToSql(in, fieldOpts...)
	vals := filterParams.GetValues()

	readParams := protodb.NewSqlToProto(&obj, fieldOpts...)

	// Executing summation query
	query := fmt.Sprintf("SELECT COALESCE(SUM(remaining_amount),0)  AS %v FROM cloud_credits WHERE %v=$1 AND expiry > NOW()",
		readParams.GetNamesString(), filterParams.GetNamesString())
	log.Info(query)

	rows, err := db.QueryContext(ctx, query, vals...)
	if err != nil {
		log.Error(err, "error reading result row")
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&obj.UnappliedAmount); err != nil {
			log.Error(err, "error reading result row.")
			return nil, status.Error(codes.Internal, "error reading rows")
		}
	}

	log.Info("Execution completed CloudCreditsCreditService.ReadUnappliedCreditBalance")
	return &obj, nil
}

func (svc *CloudCreditsCreditService) CreditMigrate(ctx context.Context, in *pb.UnappliedCredits) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.CreditMigrate").Start()
	logger.WithValues("CloudAccountId", in.CloudAccountId)
	res := &emptypb.Empty{}
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("End")

	currAccountObj, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: in.CloudAccountId})
	if err != nil {
		logger.Error(err, "failed to get cloudaccount details.")
		return nil, status.Error(codes.Internal, "failed to get cloudaccount details.")
	}

	if currAccountObj.UpgradedToPremium == pb.UpgradeStatus_UPGRADE_PENDING_CC {
		// TODO: Handle this to Enterprise as well when Upgrade to enterprise is supported
		logger.Info("Cannot complete upgrade since credit card not added successfully", "cloudAccountId", in.CloudAccountId)
		return nil, status.Error(codes.Internal, "credit card addition not successful")
	}

	if currAccountObj.UpgradedToPremium == pb.UpgradeStatus_UPGRADE_NOT_INITIATED || currAccountObj.UpgradedToPremium == pb.UpgradeStatus_UPGRADE_COMPLETE {
		// nothing to be done in this case, simply return
		logger.Info("No credit migration needed", "cloudAccountId", in.CloudAccountId)
		return res, nil
	}

	cloudAccountLocks.Lock(ctx, in.CloudAccountId)
	defer cloudAccountLocks.Unlock(ctx, in.CloudAccountId)

	driver, err := billingCommon.GetDriverAll(ctx, in.CloudAccountId)
	if err != nil {
		logger.Error(err, "unable to find driver", "cloudAccountId", in.CloudAccountId)
		return nil, status.Errorf(codes.InvalidArgument, GetBillingError(InvalidCloudAcct, err).Error())
	}

	standardDriver, err := billingCommon.GetDriverAllByType(pb.AccountType_ACCOUNT_TYPE_STANDARD)
	if err != nil {
		logger.Error(err, "unable to find standard driver", "cloudAccountId", in.CloudAccountId)
		return nil, status.Errorf(codes.InvalidArgument, GetBillingError(InvalidCloudAcct, err).Error())
	}

	credHistory := true
	creditDetails, err := standardDriver.BillingCredit.Read(ctx, &pb.BillingCreditFilter{CloudAccountId: in.CloudAccountId, History: &credHistory})
	if err != nil {
		logger.Error(err, "unable to read standard accounts credit details", "cloudAccountId", in.CloudAccountId)
		return nil, status.Errorf(codes.InvalidArgument, GetBillingError(FailedToReadBillingCredit, err).Error())
	}

	if creditDetails.GetTotalRemainingAmount() > 0 {
		creditPremDetailsOne, err := driver.BillingCredit.Read(ctx, &pb.BillingCreditFilter{CloudAccountId: in.CloudAccountId, History: &credHistory})
		if err != nil {
			logger.Error(err, "unable to read premium account credit details", "cloudAccountId", in.CloudAccountId)
			return nil, status.Errorf(codes.InvalidArgument, GetBillingError(FailedToReadBillingCredit, err).Error())
		}

		res, err = driver.BillingCredit.CreditMigrate(ctx, &pb.BillingUnappliedCredit{
			CloudAccountId:  in.CloudAccountId,
			RemainingAmount: creditDetails.TotalRemainingAmount,
			Expiration:      creditDetails.ExpirationDate,
		})
		if err != nil {
			logger.Error(err, "failed to migrate credits")

			creditPremDetailsTwo, err := driver.BillingCredit.Read(ctx, &pb.BillingCreditFilter{CloudAccountId: in.CloudAccountId, History: &credHistory})
			if err != nil {
				logger.Error(err, "unable to read premium account credit details", "cloudAccountId", in.CloudAccountId)
				return nil, status.Errorf(codes.InvalidArgument, GetBillingError(FailedToReadBillingCredit, err).Error())
			}

			// In case credit service in aria was created with standard unapplied amount and failed in the subsequent
			// steps, we need to take care of deletion of those migrated credits.
			creditValue := creditPremDetailsTwo.TotalRemainingAmount - creditPremDetailsOne.TotalRemainingAmount
			if creditValue == creditDetails.TotalRemainingAmount {
				_, err = svc.DeleteMigratedCredit(ctx, &pb.MigratedCredits{
					CloudAccountId: in.CloudAccountId,
				})
				if err != nil {
					logger.Error(err, "failed to delete migrated credits")
					return nil, status.Errorf(codes.Internal, GetBillingError(FailedToDeleteMigratedCredits, err).Error())
				}
			}

			return nil, status.Errorf(codes.Internal, GetBillingError(FailedToMigrateCredits, err).Error())
		}

		// Deletion of credits in standard CA which is upgraded to premium
		_, err = svc.DeleteMigratedCredit(ctx, &pb.MigratedCredits{
			CloudAccountId: in.CloudAccountId,
		})
		if err != nil {
			logger.Error(err, "failed to delete migrated credits")
			return nil, status.Errorf(codes.Internal, GetBillingError(FailedToDeleteMigratedCredits, err).Error())
		}
	}
	// TODO: Handle this to Enterprise as well when Upgrade to enterprise is supported
	_, err = cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{
		Id:                in.CloudAccountId,
		UpgradedToPremium: pb.UpgradeStatus_UPGRADE_COMPLETE.Enum(),
	})
	if err != nil {
		logger.Error(err, "cloudaccount update failed", "cloudAccountId", in.CloudAccountId)
		return nil, err
	}

	return res, nil
}

func (svc *CloudCreditsCreditService) DeleteMigratedCredit(ctx context.Context, in *pb.MigratedCredits) (*emptypb.Empty, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.DeleteAppliedCreditBalance").Start()
	log.WithValues("cloudaccountid", in.CloudAccountId)
	defer span.End()
	log.Info("Executing CloudCreditsCreditService.DeleteAppliedCreditBalance")

	query := "UPDATE cloud_credits SET remaining_amount = $1 WHERE cloud_account_id = $2"
	log.Info(query)

	// Updating the remaining_amount to 0 in standard cloudaccount after migration to premium
	rows, err := db.Query(query, 0, in.CloudAccountId)
	if err != nil {
		log.Error(err, "failed to update remaining_amount")
		return nil, err
	}
	defer rows.Close()

	return &emptypb.Empty{}, nil
}

func (s *CloudCreditsCreditService) ReadInternal(in *pb.Account, outStream pb.CloudCreditsCreditService_ReadInternalServer) error {
	ctx := outStream.Context()
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.ReadInternal").WithValues("CloudAccountId", in.CloudAccountId).Start()
	log.WithValues("CloudAccountId", in.CloudAccountId)
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	cloudAcct, err := cloudacctClient.GetById(ctx,
		&pb.CloudAccountId{Id: in.CloudAccountId})
	if err != nil {
		log.Error(err, "failed to read cloud account")
		return GetBillingInternalError(FailedToReadCloudAccount, err)
	}

	accountType := cloudAcct.GetType()
	if accountType != pb.AccountType_ACCOUNT_TYPE_STANDARD && accountType != pb.AccountType_ACCOUNT_TYPE_INTEL {
		driver, err := billingCommon.GetDriverAll(ctx, in.CloudAccountId)
		if err != nil {
			log.Error(err, "unable to find driver", "cloudAccountId", in.CloudAccountId)
			return status.Errorf(codes.InvalidArgument, GetBillingError(InvalidCloudAcct, err).Error())
		}

		req := &pb.BillingAccount{
			CloudAccountId: in.CloudAccountId,
		}

		res, err := driver.BillingCredit.ReadInternal(ctx, req)
		if err != nil {
			log.Error(err, "error installing credits")
			return err
		}

		for {
			out, err := res.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				log.Error(err, "failed to read billing credit")
				return status.Errorf(codes.Internal, GetBillingError(FailedToReadBillingCredit, err).Error())
			}

			obj := &pb.Credit{
				Created:         out.Created,
				Expiration:      out.Expiration,
				CloudAccountId:  out.CloudAccountId,
				OriginalAmount:  out.OriginalAmount,
				RemainingAmount: out.RemainingAmount,
				CouponCode:      out.CouponCode,
				AmountUsed:      out.AmountUsed,
				Reason:          pb.CloudCreditReason(out.Reason),
			}

			if err := outStream.Send(obj); err != nil {
				log.Error(err, "error sending read details")
				return err
			}
		}
	} else {
		filterParams := protodb.NewProtoToSql(in)
		obj := pb.Credit{}

		query := "SELECT coupon_code,cloud_account_id," +
			"created_at,expiry,original_amount,remaining_amount " +
			"FROM cloud_credits"

		names := filterParams.GetNamesString()
		if len(names) > 0 {
			query += fmt.Sprintf(" WHERE (%v) IN (%v)", names,
				filterParams.GetParamsString())
		}
		rows, err := db.Query(query, filterParams.GetValues()...)
		if err != nil {
			log.Error(err, "failed to read cloud credits")
			return err
		}
		defer rows.Close()
		for rows.Next() {
			creation := time.Time{}
			expiration := time.Time{}
			if err := rows.Scan(&obj.CouponCode, &obj.CloudAccountId, &creation, &expiration,
				&obj.OriginalAmount, &obj.RemainingAmount,
			); err != nil {
				log.Error(err, "failed to read cloud credits")
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
	}

	return nil
}

func (s *CloudCreditsCreditService) CreateCreditStateLog(ctx context.Context, in *pb.CreditsState) (*empty.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.CreateCreditStateLog").WithValues("CloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	log.Info("log new credit state", "state", in.State.Enum().String())

	if in.CloudAccountId == "" || in.State == pb.CloudCreditsState_CLOUD_CREDITS_STATE_UNSPECIFIED || in.EventAt == nil {
		err := status.Errorf(codes.InvalidArgument, "invalid input arguments")
		log.Error(err, "invalid input arguments, ignoring record creation")
		return nil, err
	}
	event_timestamp := in.EventAt.AsTime().UTC().Format(time.RFC3339)

	query := "INSERT INTO credits_state_log (cloud_account_id, state, event_at) VALUES ($1, $2, $3)"

	if _, err := db.ExecContext(ctx, query, in.CloudAccountId, in.State, event_timestamp); err != nil {
		log.Error(err, "error creating credit_state_log record")
		return nil, status.Errorf(codes.Internal, "credit_state_log insertion failed")
	}
	log.Info("created credit_state_log record", "state", in.State.Enum().String())
	return &emptypb.Empty{}, nil
}

func (s *CloudCreditsCreditService) ReadCreditStateLog(ctx context.Context, filter *pb.CreditsStateFilter) (*pb.CreditsStateResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.ReadCreditStateLog").WithValues("CloudAccountId", filter.CloudAccountId).Start()
	defer span.End()
	log.Info("begin reading latest credit state")

	resp := pb.CreditsStateResponse{}
	if filter.CloudAccountId == "" {
		err := status.Errorf(codes.InvalidArgument, "invalid input arguments")
		log.Error(err, "invalid input arguments, CloudAccountId")
		return &resp, err
	}

	query := "SELECT state,event_at,updated_at " +
		"FROM credits_state_log " +
		"WHERE cloud_account_id = $1 " +
		"ORDER BY event_at DESC " +
		"LIMIT 1;"
	var newState int64
	var eventTime, lastUpdateTime time.Time
	err := db.QueryRowContext(ctx, query, filter.CloudAccountId).Scan(&newState, &eventTime, &lastUpdateTime)
	if err != nil {
		log.Error(err, "error finding latest record")
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "credit_state_log record not found")
		}
		return nil, status.Errorf(codes.Internal, "credit_state_log query failed")
	}
	resp.State = pb.CloudCreditsState(newState)
	resp.EventAt = timestamppb.New(eventTime)
	resp.UpdatedAt = timestamppb.New(lastUpdateTime)
	return &resp, nil
}
