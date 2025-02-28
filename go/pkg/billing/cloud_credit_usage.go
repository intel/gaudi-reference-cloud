// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/exp/slices"

	"github.com/go-logr/logr"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	cloudCreditUsageChannel = make(chan bool)
	cloudCreditUsageTicker  *time.Ticker
)

func startCloudCreditUsageScheduler(ctx context.Context, cloudCreditUsageScheduler CloudCreditUsageScheduler) {
	cloudCreditUsageTicker = time.NewTicker(time.Duration(Cfg.CreditUsageSchedulerInterval) * time.Minute)
	go cloudCreditUsageLoop(context.Background(), cloudCreditUsageScheduler)
}

func cloudCreditUsageLoop(ctx context.Context, cloudCreditUsageScheduler CloudCreditUsageScheduler) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditUsageScheduler.cloudCreditUsageLoop").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	for {
		cloudCreditUsageScheduler.schedulerCloudAccountState.Mutex.Lock()
		logger.V(1).Info("last access timestamp", "accessTimestamp", cloudCreditUsageScheduler.schedulerCloudAccountState.AccessTimestamp)
		cloudCreditUsages(ctx, &logger, cloudCreditUsageScheduler)
		cloudCreditUsageScheduler.schedulerCloudAccountState.Mutex.Unlock()
		select {
		case <-cloudCreditUsageChannel:
			return
		case tm := <-cloudCreditUsageTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

// todo: this method needs to be upper case
func stopCloudCreditUsageScheduler() {
	if cloudCreditUsageChannel != nil {
		close(cloudCreditUsageChannel)
		cloudCreditUsageChannel = nil
	}
}

func cloudCreditUsages(ctx context.Context, logger *logr.Logger, cloudCreditUsageScheduler CloudCreditUsageScheduler) {
	err := cloudCreditUsageScheduler.cloudCreditUsages(ctx)
	if err != nil {
		logger.Error(err, "failed to handle cloud credit usages", "context", "cloudCreditUsages")
	}
}

// add standardized errors
const (
	CloudCreditUsageInvalidCloudAcctError          = "cloud credit usage: invalid cloud account"
	CloudCreditUsageGetAllCreditsError             = "cloud credit usage: failed to get all credits"
	CloudCreditUsageGetUnappliedCreditBalanceError = "cloud credit usage: failed to get unapplied credit balance"
	CloudCreditUsageUpdateCloudAcctHasCreditsError = "cloud credit usage: failed to update cloud account has credits"
	CloudCreditUsageUpdateCloudAcctLowCreditsError = "cloud credit usage: failed to update cloud account low credits"
	CloudCreditHunderdPercentUsed                  = "100% Credits Used Notification"
	CloudCreditEightyPercentUsed                   = "80% Credits Used Notification"
)

type CloudCreditUsageScheduler struct {
	eventManager               *events.EventManager
	notificationClient         *NotificationClient
	schedulerCloudAccountState *SchedulerCloudAccountState
	cloudAccountClient         *billingCommon.CloudAccountSvcClient
}

func NewCloudCreditUsageScheduler(eventManager *events.EventManager, notificationClient *NotificationClient, schedulerCloudAccountState *SchedulerCloudAccountState, cloudAccountClient *billingCommon.CloudAccountSvcClient) *CloudCreditUsageScheduler {
	return &CloudCreditUsageScheduler{eventManager: eventManager, notificationClient: notificationClient, schedulerCloudAccountState: schedulerCloudAccountState, cloudAccountClient: cloudAccountClient}
}

func (cloudCreditUsageScheduler *CloudCreditUsageScheduler) cloudCreditUsages(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditUsageScheduler.cloudCreditUsages").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	cloudAccts, err := cloudCreditUsageScheduler.cloudAccountClient.GetCloudAccountsWithCredits(ctx)
	if err != nil {
		logger.Error(err, "failed to get cloud accounts with credits", "context", "GetCloudAccountsWithCredits")
		return err
	}
	for _, cloudAcct := range cloudAccts {
		logger.V(9).Info("checking for cloud credit usage for", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
		// get the right driver
		driver, err := GetDriver(ctx, cloudAcct.Id)
		errMsg := fmt.Sprint("failed handling cloud credit usage for cloud account id: ", cloudAcct.GetId())
		if err != nil {
			logger.Error(GetSchedulerError(CloudCreditUsageInvalidCloudAcctError, err), errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "GetDriver")
			continue
		}

		billingUnappliedCreditBalance, err := driver.billingCredit.ReadUnappliedCreditBalance(ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
		if err != nil {
			logger.Error(GetSchedulerError(CloudCreditUsageGetUnappliedCreditBalanceError, err), errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "ReadUnappliedCreditBalance")
			continue
		}
		// todo: this needs to be moved to inside the else statement as the initial credit amount is
		// only needed with the unapplied amount is > 0.
		totalInitialCreditAmount, lastCreditCreated, err := getTotalInitialCreditAmountAndLastCreated(ctx, cloudAcct)
		if err != nil {
			logger.Error(GetSchedulerError(CloudCreditUsageGetAllCreditsError, err), errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "getTotalInitialCreditAmountAndLastCreated")
			continue
		}
		logger.Info("billing unapplied credit balance", "billingUnappliedCreditBalance", billingUnappliedCreditBalance, "totalInitialCreditAmount", totalInitialCreditAmount)
		if billingUnappliedCreditBalance.UnappliedAmount <= 0 {
			logger.V(9).Info("there is no unapplied amount left for cloud account", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
			paidServicesAllowed := cloudAcct.GetPaidServicesAllowed()
			err = UpdateCloudAcctNoCredits(ctx, cloudAcct)
			if err != nil {
				logger.Error(fmt.Errorf("failed to update cloud credit upon comepletely used%w", err), errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "UpdateCloudAcctNoCredits")
			}
			hasCreditCard, err := IsCloudAccountWithCreditCard(ctx, cloudAcct)
			if hasCreditCard || err != nil || cloudCreditUsageScheduler.notificationClient == nil {
				if err != nil {
					logger.Error(err, errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "IsCloudAccountWithCreditCard")
				}
				continue
			}
			if Cfg.GetCreditUsageEmail() && !cloudAcct.GetTerminatePaidServices() && Cfg.Features.ServicesTerminationScheduler && paidServicesAllowed {
				emailNotificationEnabled := slices.Contains(Cfg.GetCreditUsageEmailAccountTypes(), cloudAcct.GetType().String())
				logger.V(9).Info("cloud account credit usage", "cloudAccountId", cloudAcct.Id, "creditUsageEmailAccountTypes", Cfg.GetCreditUsageEmailAccountTypes(), "accountType", cloudAcct.GetType().String(), "emailNotificationEnabled", emailNotificationEnabled)
				if emailNotificationEnabled {
					logger.V(9).Info("sending email for cloud credit usage for cloud account", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
					if err = cloudCreditUsageScheduler.notificationClient.SendCloudCreditUsageEmailWithOptions(ctx, CloudCreditHunderdPercentUsed, cloudAcct.Owner, Cfg.GetCloudCreditHundredPercentUsedTemplate()); err != nil {
						logger.Error(err, "couldn't send email", "cloudAccountId", cloudAcct.GetId(), "context", "SendCloudCreditUsageEmail")
					}
				} else {
					logger.V(9).Info("skipping email for cloud credit usage for cloud account", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
				}
			}
		} else {
			logger.V(9).Info("checking whether cloud acct has credits for cloud acct", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
			isCreditsDepleted := cloudAcct.CreditsDepleted.AsTime().Before(lastCreditCreated)
			logger.V(9).Info("cloud acct credits check", "cloudAccountId", cloudAcct.Id, "creditsDepleted", cloudAcct.CreditsDepleted.AsTime(), "lastCreditCreated", lastCreditCreated, "isCreditsDepleted", isCreditsDepleted)
			if cloudAcct.CreditsDepleted == nil || isCreditsDepleted {
				err = UpdateCloudAcctHasCredits(ctx, cloudAcct, true)
				if err != nil {
					logger.Error(GetSchedulerError(CloudCreditUsageUpdateCloudAcctHasCreditsError, err), errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "UpdateCloudAcctHasCredits")
					continue
				}
			}
			cloudCreditUsageThresholdReached, err :=
				cloudCreditUsageScheduler.isCreditUsageThresholdReached(cloudAcct.GetType(),
					totalInitialCreditAmount,
					billingUnappliedCreditBalance.UnappliedAmount)
			if err != nil {
				logger.Error(err, errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "isCreditUsageThresholdReached")
				continue
			}
			err = cloudCreditUsageScheduler.updateCloudAcctThresholdReached(ctx, cloudAcct, cloudCreditUsageThresholdReached)
			if err != nil {
				logger.Error(GetSchedulerError(CloudCreditUsageUpdateCloudAcctLowCreditsError, err), errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "updateCloudAcctThresholdReached")
			}
		}
	}
	return nil
}

func getTotalInitialCreditAmountAndLastCreated(ctx context.Context, cloudAcct *pb.CloudAccount) (float64, time.Time, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditUsageScheduler.getTotalInitialCreditAmountAndLastCreated").WithValues("cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type).Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	driver, err := GetDriver(ctx, cloudAcct.Id)
	errMsg := fmt.Sprint("failed handling cloud credit usage for cloud account id: ", cloudAcct.GetId())
	if err != nil {
		logger.Error(err, errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "GetDriver")
		return 0, time.Time{}, err
	}
	billingCreditClient, err := driver.billingCredit.ReadInternal(ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
	if err != nil {
		logger.Error(err, errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "ReadInternal")
		return 0, time.Time{}, err
	}
	// Get a total of initial credit amount.
	var totalInitialCreditAmount float64 = 0
	var lastCreated = time.Now().AddDate(-100, 0, 0)
	for {
		billingCredit, err := billingCreditClient.Recv()
		logger.V(9).Info("credit for cloud acct", "cloudAccountId", cloudAcct.Id, "billingCredit", billingCredit)
		if errors.Is(err, io.EOF) {
			return totalInitialCreditAmount, lastCreated, nil
		}
		if err != nil {
			logger.Error(err, errMsg, "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "ReadInternal")
			return 0, time.Time{}, err
		}
		if (billingCredit.RemainingAmount != 0) && billingCredit.Expiration.AsTime().After(time.Now()) {
			totalInitialCreditAmount += billingCredit.OriginalAmount
		}
		if billingCredit.Created.AsTime().After(lastCreated) {
			lastCreated = billingCredit.Created.AsTime()
		}
	}
}

func (cloudCreditUsageScheduler *CloudCreditUsageScheduler) isCreditUsageThresholdReached(cloudAcctType pb.AccountType, initialAmount float64,
	remainingAmount float64) (bool, error) {
	switch cloudAcctType {
	case pb.AccountType_ACCOUNT_TYPE_PREMIUM:
		return remainingAmount < ((float64(100-Cfg.PremiumCloudCreditThreshold) * initialAmount) / 100), nil
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE:
		return remainingAmount < ((float64(100-Cfg.EnterpriseCloudCreditThreshold) * initialAmount) / 100), nil
	// todo: have a dedicated threshold for standard and not necessarily the same as intel.
	case pb.AccountType_ACCOUNT_TYPE_STANDARD:
		return remainingAmount < ((float64(100-Cfg.IntelCloudCreditThreshold) * initialAmount) / 100), nil
	case pb.AccountType_ACCOUNT_TYPE_INTEL:
		return remainingAmount < ((float64(100-Cfg.IntelCloudCreditThreshold) * initialAmount) / 100), nil
	default:
		return false, fmt.Errorf("invalid account type %v", cloudAcctType)
	}
}

func (cloudCreditUsageScheduler *CloudCreditUsageScheduler) updateCloudAcctThresholdReached(ctx context.Context, cloudAcct *pb.CloudAccount, cloudCreditUsageThresholdReached bool) error {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditUsageScheduler.updateCloudAcctThresholdReached").WithValues("cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if cloudAcct.LowCredits && !cloudCreditUsageThresholdReached {
		lowCredits := false
		_, err := cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, LowCredits: &lowCredits})
		if err != nil {
			logger.Error(err, "cloud account client update ", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "lowCredits", cloudAcct.LowCredits, "cloudCreditUsageThresholdReached", cloudCreditUsageThresholdReached, "context", "Update")
			return err
		}
	}

	if !cloudAcct.LowCredits && cloudCreditUsageThresholdReached {
		lowCredits := true
		_, err := cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, LowCredits: &lowCredits})
		if err != nil {
			logger.Error(err, "cloud account client update ", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "lowCredits", cloudAcct.LowCredits, "cloudCreditUsageThresholdReached", cloudCreditUsageThresholdReached, "context", "Update")
			return err
		}
	}
	cloudAccountLowCredits := cloudAcct.GetLowCredits()
	if cloudCreditUsageThresholdReached {
		hasCreditCard, err := IsCloudAccountWithCreditCard(ctx, cloudAcct)
		if Cfg.GetCreditUsageEmail() && !cloudAccountLowCredits && Cfg.Features.ServicesTerminationScheduler {
			emailNotificationEnabled := slices.Contains(Cfg.GetCreditUsageEmailAccountTypes(), cloudAcct.GetType().String())
			logger.V(9).Info("cloud account credit usage", "cloudAccountId", cloudAcct.Id, "creditUsageEmailAccountTypes", Cfg.GetCreditUsageEmailAccountTypes(), "accountType", cloudAcct.GetType().String(), "emailNotificationEnabled", emailNotificationEnabled)
			if emailNotificationEnabled {
				if cloudCreditUsageScheduler.notificationClient != nil && (!hasCreditCard || err == nil) {
					if err = cloudCreditUsageScheduler.notificationClient.SendCloudCreditUsageEmail(ctx, CloudCreditEightyPercentUsed, cloudAcct.Owner, Cfg.GetCloudCreditEightyPercentUsedTemplate()); err != nil {
						logger.Error(err, "couldn't send email", "cloudAccountId", cloudAcct.Id, "context", "SendCloudCreditUsageEmail")
						return err
					}
				}
			} else {
				logger.V(9).Info("skipping email for credit threshold reached for cloud account", "cloudAccountId", cloudAcct.Id)
			}
		}
	}
	return nil
}
