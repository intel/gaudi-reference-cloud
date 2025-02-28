// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcreditsworker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudcredits_worker/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	CloudCreditHunderdPercentUsed = "100% Credits Used Notification"
	CloudCreditEightyPercentUsed  = "80% Credits Used Notification"
)

func (w CloudCreditsWorker) HandleThresholdReached(ctx context.Context, message *events.CreateEvent) error {
	logger := log.FromContext(ctx).WithName("CloudCreditsWorker.handleThresholdReached")

	logger.Info("cloud credits threshold reached event", "accountId", message.CloudAccountId)

	// get Cloudaccount
	cloudAcct, err := w.CloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: message.CloudAccountId})
	if err != nil {
		logger.Error(err, "error getting acct details from cloud acct service")
		return nil
	}
	if cloudAcct == nil {
		logger.V(9).Info("cloud account not found", "cloudAcct", cloudAcct)
		return nil
	}
	driver, err := billingCommon.GetDriverByType(cloudAcct.GetType())
	if err != nil {
		logger.Error(err, "error returned from GetDriver")
		return err
	}
	billingUnappliedCreditBalance, err := w.ReadUnappliedCreditBalance(ctx, driver, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
	if err != nil {
		logger.Error(err, "failed to read billing credit")
		return err
	}
	totalInitialCreditAmount, _, err := w.getTotalInitialCreditAmountAndLastCreated(ctx, cloudAcct)
	if err != nil {
		logger.Error(err, "failed to get TotalInitialCreditAmount And LastCreated ")
		return nil
	}
	cloudCreditUsageEventThresholdReached, err :=
		w.IsCreditUsageThresholdReached(ctx, cloudAcct.GetType(),
			totalInitialCreditAmount,
			billingUnappliedCreditBalance.GetUnappliedAmount())
	if err != nil {
		logger.Error(err, "error returned from cloudCreditUsageEventThresholdReached")
		return err
	}
	if !cloudCreditUsageEventThresholdReached {
		logger.Info("returning as cloud credit usage threshold reached")
		return nil
	}
	// get driver

	if !cloudAcct.LowCredits {
		lowCredits := true
		_, err := w.CloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, LowCredits: &lowCredits})
		if err != nil {
			logger.Error(err, "failed to update cloudaccount low credits")
			return err
		}
	}
	hasCreditCard, err := w.BillingCloudAccountClient.IsCloudAccountWithCreditCard(ctx, cloudAcct, driver)
	if hasCreditCard || err != nil || w.NotificationGatewayClient == nil {
		logger.Info("account has credit card added, no further action needed")
		return nil
	}

	emailNotificationEnabled := slices.Contains(config.Cfg.GetCreditUsageEmailAccountTypes(), cloudAcct.GetType().String())
	logger.Info("cloud account credits threshold", "cloudAccountId", cloudAcct.Id, "emailNotificationEnabled", emailNotificationEnabled, "GetCreditUsageEmail", config.Cfg.GetSendCreditUsageEmail())
	if emailNotificationEnabled && config.Cfg.GetSendCreditUsageEmail() {

		if err = w.NotificationGatewayClient.SendCloudCreditUsageEmailWithOptions(ctx, CloudCreditEightyPercentUsed, cloudAcct.Owner, config.Cfg.GetCloudCreditEightyPercentUsedTemplate(), w.Name()); err != nil {
			logger.Error(err, "couldn't send email", "Owner", cloudAcct.Owner)
			return err
		}
	}
	return nil
}

func (w CloudCreditsWorker) HandleCreditsUsed(ctx context.Context, message *events.CreateEvent, isExpired bool) error {
	logger := log.FromContext(ctx).WithName("CloudCreditsWorker.HandleCreditsUsed")
	logger.Info("cloud credits used", "accountId", message.CloudAccountId)

	isCreditsAvailable, err := w.IsCreditsAvailable(ctx, message.CloudAccountId)
	if err != nil {
		logger.Error(err, "error cannot apply credit event due to failed IsCreditsAvailable call", "cloudAccount", message.CloudAccountId)
		return err
	}
	if isCreditsAvailable {
		logger.Info("ignoring message credits available for", "cloudAccountId", message.CloudAccountId, "isCreditsAvailable ", isCreditsAvailable)
		return nil
	}
	// get Cloudaccount
	cloudAcct, err := w.CloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: message.CloudAccountId})
	if err != nil {
		logger.Error(err, "error getting acct details from cloud acct service")
		return nil
	}
	if cloudAcct == nil {
		logger.V(9).Info("cloud account not found", "cloudAcct", cloudAcct)
		return nil
	}

	// get driver
	driver, err := billingCommon.GetDriverByType(cloudAcct.GetType())
	if err != nil {
		logger.Error(err, "error returned from GetDriver")
		return err
	}

	// Update flags
	paidServicesAllowed := cloudAcct.GetPaidServicesAllowed()
	err = w.BillingCloudAccountClient.UpdateCloudAcctNoCredits(ctx, cloudAcct, driver)
	if err != nil {
		logger.Error(err, "failed to update cloud credit upon complete usage")
	}

	hasCreditCard, err := w.BillingCloudAccountClient.IsCloudAccountWithCreditCard(ctx, cloudAcct, driver)
	if hasCreditCard || err != nil || w.NotificationGatewayClient == nil {
		logger.Info("account has credit card added, no further action needed")
		return nil
	}

	logger.Info("cloud account credit used", "cloudAccountId", cloudAcct.Id, "GetCreditExpiryEmail", config.Cfg.GetSendCreditExpiryEmail(), "GetCreditUsageEmail", config.Cfg.GetSendCreditUsageEmail(), "isExpired", isExpired)

	// Send email Notification
	if paidServicesAllowed {
		if isExpired && config.Cfg.GetSendCreditExpiryEmail() {

			emailNotificationEnabled := slices.Contains(config.Cfg.GetCreditExpiryEmailAccountTypes(), cloudAcct.GetType().String())
			logger.V(9).Info("cloud account credit expiry email", "cloudAccountId", cloudAcct.Id, "emailNotificationEnabled", emailNotificationEnabled)
			if emailNotificationEnabled {
				logger.Info("sending email for cloud credit expiry for cloud account", "id", cloudAcct.Id)
				if err = w.NotificationGatewayClient.SendCloudCreditUsageEmailWithOptions(ctx, CloudCreditExpired, cloudAcct.Owner, config.Cfg.GetCloudCreditExpiryTemplate(), w.Name()); err != nil {
					logger.Error(err, "couldn't send email", "Owner", cloudAcct.Owner)
				}
			}
		} else if !isExpired && config.Cfg.GetSendCreditUsageEmail() {

			emailNotificationEnabled := slices.Contains(config.Cfg.GetCreditUsageEmailAccountTypes(), cloudAcct.GetType().String())
			logger.V(9).Info("cloud account credit usage", "emailNotificationEnabled", emailNotificationEnabled)
			if emailNotificationEnabled {
				logger.Info("sending email for cloud credit usage for cloud account", "id", cloudAcct.Id)
				if err = w.NotificationGatewayClient.SendCloudCreditUsageEmailWithOptions(ctx, CloudCreditHunderdPercentUsed, cloudAcct.Owner, config.Cfg.GetCloudCreditHundredPercentUsedTemplate(), w.Name()); err != nil {
					logger.Error(err, "couldn't send email", "Owner", cloudAcct.Owner)
				}
			}
		}
	}
	return nil
}

func (w CloudCreditsWorker) HandleCreditsAvailable(ctx context.Context, message *events.CreateEvent) error {
	logger := log.FromContext(ctx).WithName("CloudCreditsWorker.handleCreditsAvailable")

	logger.Info("cloud credits available event", "accountId", message.CloudAccountId)
	// get Cloudaccount
	cloudAcct, err := w.CloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: message.CloudAccountId})
	if err != nil {
		logger.Error(err, "error getting acct details from cloud acct service")
		return nil
	}
	if cloudAcct == nil {
		logger.V(9).Info("cloud account not found", "cloudAcct", cloudAcct)
		return nil
	}

	if err := w.BillingCloudAccountClient.UpdateCloudAcctHasCredits(ctx, cloudAcct); err != nil {
		logger.Error(err, "failed handling cloud credits available", "cloud account id", message.CloudAccountId)
	}

	return nil
}

func (w CloudCreditsWorker) IsCreditsAvailable(ctx context.Context, cloudAccountId string) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsWorker.IsCreditsAvailable").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()

	in := &pb.CreditFilter{
		CloudAccountId: cloudAccountId,
	}
	resp, err := CloudCreditsClient.CloudCreditsSvcClient.ReadCredits(ctx, in)
	if err != nil {
		logger.Error(err, "error calling ReadCredits")
		return false, err
	}
	if resp != nil && resp.TotalRemainingAmount > 0 {
		logger.Info("credits available", "cloudAccountId", cloudAccountId, "totalRemainingAmount", resp.TotalRemainingAmount)
		return true, nil
	}
	return false, nil
}
func (w CloudCreditsWorker) getTotalInitialCreditAmountAndLastCreated(ctx context.Context, cloudAcct *pb.CloudAccount) (float64, time.Time, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditsWorker.getTotalInitialCreditAmountAndLastCreated").WithValues("cloudAccountId", cloudAcct.Id).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	driver, err := billingCommon.GetDriverByType(cloudAcct.GetType())
	if err != nil {
		return 0, time.Time{}, err
	}
	billingCreditClient, err := driver.BillingCredit.ReadInternal(ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})

	if err != nil {
		return 0, time.Time{}, err
	}
	// Get a total of initial credit amount.
	var totalInitialCreditAmount float64 = 0
	var lastCreated = time.Now().AddDate(-100, 0, 0)
	for {
		BillingCredit, err := billingCreditClient.Recv()
		if errors.Is(err, io.EOF) {
			return totalInitialCreditAmount, lastCreated, nil
		}
		if err != nil {
			return 0, time.Time{}, err
		}
		if (BillingCredit.RemainingAmount != 0) && BillingCredit.Expiration.AsTime().After(time.Now()) {
			totalInitialCreditAmount += BillingCredit.OriginalAmount
		}
		if BillingCredit.Created.AsTime().After(lastCreated) {
			lastCreated = BillingCredit.Created.AsTime()
		}
	}
}

func (w CloudCreditsWorker) ReadUnappliedCreditBalance(ctx context.Context, driver *billingCommon.BillingDriverClients, in *pb.BillingAccount) (*pb.BillingUnappliedCreditBalance, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsWorker.ReadUnappliedCreditBalance").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	res, err := driver.BillingCredit.ReadUnappliedCreditBalance(ctx, in)

	if err != nil {
		logger.Error(err, "failed to read unapplied cloud credit balance")
		return nil, status.Errorf(codes.Internal, billingCommon.GetBillingError("FailedToReadUnappliedCreditBalance", err).Error())
	}
	return res, err
}
func (w CloudCreditsWorker) IsCreditUsageThresholdReached(ctx context.Context, cloudAcctType pb.AccountType, initialAmount float64,
	remainingAmount float64) (bool, error) {
	switch cloudAcctType {
	case pb.AccountType_ACCOUNT_TYPE_PREMIUM:
		return remainingAmount < ((float64(100-config.Cfg.PremiumCloudCreditThreshold) * initialAmount) / 100), nil
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE:
		return remainingAmount < ((float64(100-config.Cfg.EnterpriseCloudCreditThreshold) * initialAmount) / 100), nil
	// todo: have a dedicated threshold for standard and not necessarily the same as intel.
	case pb.AccountType_ACCOUNT_TYPE_STANDARD:
		return remainingAmount < ((float64(100-config.Cfg.IntelCloudCreditThreshold) * initialAmount) / 100), nil
	case pb.AccountType_ACCOUNT_TYPE_INTEL:
		return remainingAmount < ((float64(100-config.Cfg.IntelCloudCreditThreshold) * initialAmount) / 100), nil
	default:
		return false, fmt.Errorf("invalid account type %v", cloudAcctType)
	}
}
