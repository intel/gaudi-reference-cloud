// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing_driver_intel

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type IntelBillingDriverUsageService struct {
	pb.UnimplementedBillingDriverUsageServiceServer
	productServiceClient *billingCommon.ProductClient
	session              *sql.DB
	cloudAccountClient   *billingCommon.CloudAccountSvcClient
}

func (svc *IntelBillingDriverUsageService) ReportUsage(stream pb.BillingDriverUsageService_ReportUsageServer) error {
	ctx := stream.Context()
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("IntelBillingDriverUsageService.ReportUsage").Start()
	defer span.End()
	log.Info("Executing ReportUsage")
	defer log.Info("Returning from ReportUsage")

	var billedIds map[int64]bool
	var startingId int64

	usagesMap := map[string]*billingCommon.CloudAccountToCost{}
	for {
		billingDriverUsage, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Error(err, "error in recev")
			return err
		}
		log.Info("reported usage", "billingDriverUsage", billingDriverUsage)

		if billedIds == nil {
			startingId = billingDriverUsage.UsageId
			billedIds, err = billingCommon.GetAlreadyBilledIds(ctx, svc.session, startingId)
			if err != nil {
				return err
			}
		}

		if billingDriverUsage.UsageId < startingId {
			// The code to avoid charging twice for the same usage id
			// depends on the usages being reported in ascending id
			// order
			panic(fmt.Sprintf("out of order usage %v with id less than %v", billingDriverUsage.UsageId, startingId))
		}

		// check for usage already billed
		if billedIds[billingDriverUsage.UsageId] {
			// Send back the id that's already been billed so its reported flag will
			// be set in the metering database.
			resp := pb.BillingDriverUsageResult{UsageId: billingDriverUsage.UsageId}
			if err := stream.Send(&resp); err != nil {
				log.Error(err, "error sending data")
				return err
			}
			log.Info("Sending back duplicates", "usageId", billingDriverUsage.UsageId)
			continue
		}

		cloudAccountToCost, exist := usagesMap[billingDriverUsage.CloudAccountId]

		// if cloudaccountid doesn't exist add to the map
		if !exist {
			log.Info("Adding to usagesMap", "cloudAccountId", billingDriverUsage.GetCloudAccountId())
			cloudAccountToCost = &billingCommon.CloudAccountToCost{}
			usagesMap[billingDriverUsage.CloudAccountId] = cloudAccountToCost
		}
		rate, err := billingCommon.GetProductRateForAccountType(ctx, billingDriverUsage.ProductId, pb.AccountType_ACCOUNT_TYPE_INTEL, svc.productServiceClient)
		if err != nil {
			log.Error(err, "error in getting rate", "cloudAccountId", billingDriverUsage.GetCloudAccountId(), "productId", billingDriverUsage.GetProductId())
			continue
		}
		rateFloat, err := billingCommon.ParseRate(rate)
		if err != nil {
			log.Error(err, "error in rate", "cloudAccountId", billingDriverUsage.GetCloudAccountId(), "productId", billingDriverUsage.GetProductId())
			continue
		}
		cloudAccountToCost.Cost += billingDriverUsage.Amount * rateFloat

		cloudAccountToCost.Usages = append(cloudAccountToCost.Usages, billingDriverUsage.UsageId)
		log.Info(fmt.Sprintf("calculated usage %+v", cloudAccountToCost), "cloudAccountId", billingDriverUsage.CloudAccountId)

		log.Info("usages value", "usagesMap for cloudAccountId", usagesMap[billingDriverUsage.CloudAccountId])
	}

	log.Info("Finished receiving usages", "usagesMap", usagesMap)

	for cloudAccountId, cloudAccountToCost := range usagesMap {

		// Any new credits installed after this call would be included in next call
		// Get sorted credits
		// Update remainingamount query
		// Applying Greedy Approach for credit consumption wrt credit expiration
		// if credits are installed during this iteration, then such credits will not have the
		// excess amount deducted when new credits are installed and hence can continue having a negative amount.
		// also we need to continue adding unreported cost
		// Remember the usages we processed for this cloudaccount
		err := billingCommon.ProcessCloudAccountCost(ctx, svc.session, cloudAccountId, cloudAccountToCost)
		if err != nil {
			log.Error(err, "error processing usages")
			return err
		}

		log.Info("Sending", "usageIds", cloudAccountToCost.Usages)
		for _, usageId := range cloudAccountToCost.Usages {
			resp := pb.BillingDriverUsageResult{UsageId: usageId}
			if err := stream.Send(&resp); err != nil {
				log.Error(err, "error sending data")
				return err
			}
		}
	}
	return nil
}
