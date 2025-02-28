// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcredits

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	cloudCreditUsageEventChannel = make(chan bool)
	cloudCreditUsageEventTicker  *time.Ticker
)

// add standardized errors
const (
	CloudCreditUsageEventInvalidCloudAcctError          = "cloud credit usage: invalid cloud account"
	CloudCreditUsageEventGetAllCreditsError             = "cloud credit usage: failed to get all credits"
	CloudCreditUsageEventGetUnappliedCreditBalanceError = "cloud credit usage: failed to get unapplied credit balance"
	CloudCreditUsageEventUpdateCloudAcctHasCreditsError = "cloud credit usage: failed to update cloud account has credits"
	CloudCreditUsageEventUpdateCloudAcctLowCreditsError = "cloud credit usage: failed to update cloud account low credits"
	CloudCreditsAvailableMessage                        = "cloud credits available"
	CloudCreditsThresholdReachedMessage                 = "cloud credits threshold reached"
	CloudCreditsUsedMessage                             = "cloud credits used"
)

type CloudCreditUsageEventScheduler struct {
	notificationGatewayClient  billingCommon.NotificationGatewayClientInterface
	schedulerCloudAccountState *SchedulerCloudAccountState
	cloudAccountClient         *billingCommon.CloudAccountSvcClient
}

func NewCloudCreditUsageEventScheduler(notificationGatewayClient billingCommon.NotificationGatewayClientInterface, schedulerCloudAccountState *SchedulerCloudAccountState, cloudAccountClient *billingCommon.CloudAccountSvcClient) *CloudCreditUsageEventScheduler {
	return &CloudCreditUsageEventScheduler{notificationGatewayClient: notificationGatewayClient, schedulerCloudAccountState: schedulerCloudAccountState, cloudAccountClient: cloudAccountClient}
}

func startCloudCreditUsageEventScheduler(ctx context.Context, cloudCreditUsageEventScheduler CloudCreditUsageEventScheduler) {
	cloudCreditUsageEventTicker = time.NewTicker(time.Duration(config.Cfg.CreditUsageEventSchedulerInterval) * time.Minute)
	go cloudCreditUsageEventLoop(context.Background(), cloudCreditUsageEventScheduler)
}

func cloudCreditUsageEventLoop(ctx context.Context, cloudCreditUsageEventScheduler CloudCreditUsageEventScheduler) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditUsageEventScheduler.cloudCreditUsageEventLoop").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	for {
		CloudCreditUsageEvent(ctx, &logger, cloudCreditUsageEventScheduler)
		select {
		case <-cloudCreditUsageEventChannel:
			return
		case tm := <-cloudCreditUsageEventTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

// todo: this method needs to be upper case
func stopCloudCreditUsageEventScheduler() {
	if cloudCreditUsageEventChannel != nil {
		close(cloudCreditUsageEventChannel)
		cloudCreditUsageEventChannel = nil
	}
}

func CloudCreditUsageEvent(ctx context.Context, logger *logr.Logger, cloudCreditUsageEventScheduler CloudCreditUsageEventScheduler) {
	err := cloudCreditUsageEventScheduler.CloudCreditUsageEvent(ctx)
	if err != nil {
		logger.Error(err, "failed to handle cloud credit usages")
	}
}

func (cloudCreditUsageEventScheduler *CloudCreditUsageEventScheduler) CloudCreditUsageEvent(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditUsageEventScheduler.CloudCreditUsageEvent").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	cloudAccts, err := cloudCreditUsageEventScheduler.cloudAccountClient.GetCloudAccountsWithCredits(ctx)
	if err != nil {
		logger.Error(err, "failed to get cloud accounts with credits")
		return err
	}
	for _, cloudAcct := range cloudAccts {
		logger.Info("checking for cloud credit usage for", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
		// get the right driver
		driver, err := billingCommon.GetDriver(ctx, cloudAcct.Id)
		errMsg := fmt.Sprint("failed handling cloud credit usage for cloud account id: ", cloudAcct.GetId())
		if err != nil {
			logger.Error(GetSchedulerError(CloudCreditUsageEventInvalidCloudAcctError, err), errMsg)
			continue
		}

		billingUnappliedCreditBalance, err := driver.BillingCredit.ReadUnappliedCreditBalance(ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
		if err != nil {
			logger.Error(GetSchedulerError(CloudCreditUsageEventGetUnappliedCreditBalanceError, err), errMsg)
			continue
		}
		// todo: this needs to be moved to inside the else statement as the initial credit amount is
		// only needed with the unapplied amount is > 0.
		totalInitialCreditAmount, lastCreditCreated, err := cloudCreditUsageEventScheduler.getTotalInitialCreditAmountAndLastCreated(ctx, cloudAcct)
		if err != nil {
			logger.Error(GetSchedulerError(CloudCreditUsageEventGetAllCreditsError, err), errMsg)
			continue
		}
		logger.Info("billing unapplied credit balance", "billingUnappliedCreditBalance", billingUnappliedCreditBalance, "totalInitialCreditAmount", totalInitialCreditAmount)
		if billingUnappliedCreditBalance.UnappliedAmount <= 0 {
			logger.Info("there is no unapplied amount left for", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
			hasCreditCard, err := cloudCreditUsageEventScheduler.cloudAccountClient.IsCloudAccountWithCreditCard(ctx, cloudAcct, driver)
			if err != nil {
				logger.Error(err, "couldn't get hascreditcard", "cloudAccountId", cloudAcct.Id)
				continue
			}
			if (!cloudAcct.GetPaidServicesAllowed() && cloudAcct.GetLowCredits()) || hasCreditCard {
				logger.Info("Skip publishing credit usage event", "cloudAccountId", cloudAcct.Id)
				continue
			}
			logger.Info("sending cloud credit used event for", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
			properties := map[string]string{}
			cloudCreditsUsedRequest := cloudCreditUsageEventScheduler.getPublishCloudCreditsUsageEventRequest(cloudAcct.GetId(), billingCommon.CloudCreditsUsed, billingCommon.CloudCreditsUsedEventSubType, CloudCreditsUsedMessage, properties)
			resp, err := cloudCreditUsageEventScheduler.notificationGatewayClient.PublishEvent(ctx, cloudCreditsUsedRequest)
			if err != nil {
				logger.Error(err, "couldn't publish cloud credit used event", "cloudAccountId", cloudAcct.Id)
			}
			logger.Info("publishing event for cloud credit used for", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "resp", resp)

		} else {
			logger.Info("updating cloud acct has credits for cloud acct", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
			isNewCreditCreated := cloudAcct.CreditsDepleted.AsTime().Before(lastCreditCreated)
			logger.V(1).Info("cloud acct credits check", "cloudAccountId", cloudAcct.Id, "creditsDepleted", cloudAcct.CreditsDepleted.AsTime(), "lastCreditCreated", lastCreditCreated, "isNewCreditCreated", isNewCreditCreated)

			cloudCreditUsageEventThresholdReached, err :=
				cloudCreditUsageEventScheduler.IsCreditUsageThresholdReached(ctx, cloudAcct.GetType(),
					totalInitialCreditAmount,
					billingUnappliedCreditBalance.UnappliedAmount)
			if err != nil {
				logger.Error(err, errMsg)
				continue
			}
			hasCreditCard, err := cloudCreditUsageEventScheduler.cloudAccountClient.IsCloudAccountWithCreditCard(ctx, cloudAcct, driver)
			if err != nil {
				logger.Error(err, "couldn't get hascreditcard", "cloudAccountId", cloudAcct.Id)
				continue
			}
			if cloudCreditUsageEventThresholdReached {
				if !cloudAcct.GetLowCredits() && !hasCreditCard {
					err = cloudCreditUsageEventScheduler.notifyCloudCreditThresholdReached(ctx, cloudAcct)
					if err != nil {
						logger.Error(GetSchedulerError(CloudCreditUsageEventUpdateCloudAcctLowCreditsError, err), errMsg)
					}
				} else if hasCreditCard {
					logger.Info("Skip publishing credit threshold event since user has credit card added", "cloudAccountId", cloudAcct.Id)
				} else {
					logger.Info("Skip publishing duplicate credit threshold event", "cloudAccountId", cloudAcct.Id)
				}
				continue
			}

			if cloudAcct.CreditsDepleted == nil || isNewCreditCreated {
				err = cloudCreditUsageEventScheduler.notifyCloudCreditsAvailable(ctx, cloudAcct)
				if err != nil {
					logger.Error(GetSchedulerError(CloudCreditUsageEventUpdateCloudAcctLowCreditsError, err), errMsg)
				}
			}
		}
	}
	return nil
}

func (cloudCreditUsageEventScheduler *CloudCreditUsageEventScheduler) getPublishCloudCreditsUsageEventRequest(cloudAccountId string, eventName string, eventSubType string, message string, properties map[string]string) *pb.PublishEventRequest {
	eventSeverity := pb.EventSeverity_LOW
	serviceName := pb.ServiceName_CREDIT
	userId := uuid.NewString()
	clientRecordId := uuid.NewString()
	CloudCreditUsageEvent := &pb.CreateEvent{
		EventName:      eventName,
		Status:         pb.EventStatus_ACTIVE,
		Type:           pb.EventType_OPERATION,
		Severity:       &eventSeverity,
		ServiceName:    &serviceName,
		Message:        &message,
		UserId:         &userId,
		EventSubType:   eventSubType,
		Properties:     properties,
		ClientRecordId: clientRecordId,
		CloudAccountId: &cloudAccountId,
	}
	cloudCreditUsageEventRequest := &pb.PublishEventRequest{
		TopicName:   TopicName,
		Subject:     eventSubType,
		CreateEvent: CloudCreditUsageEvent,
	}
	return cloudCreditUsageEventRequest
}

func (cloudCreditUsageEventScheduler *CloudCreditUsageEventScheduler) getTotalInitialCreditAmountAndLastCreated(ctx context.Context, cloudAcct *pb.CloudAccount) (float64, time.Time, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageEventScheduler.getTotalInitialCreditAmountAndLastCreated").Start()
	logger.WithValues("cloudaccountid", cloudAcct.Id)
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("getting total initial credit amount for cloud acct", "id", cloudAcct.Id, "type", cloudAcct.Type)
	driver, err := billingCommon.GetDriver(ctx, cloudAcct.Id)
	if err != nil {
		logger.Error(err, "failed to get driver")
		return 0, time.Time{}, err
	}
	billingCreditClient, err := driver.BillingCredit.ReadInternal(ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
	if err != nil {
		logger.Error(err, "failed to get billing credit")
		return 0, time.Time{}, err
	}
	// Get a total of initial credit amount.
	var totalInitialCreditAmount float64 = 0
	var lastCreated = time.Now().AddDate(-100, 0, 0)
	for {
		BillingCredit, err := billingCreditClient.Recv()
		logger.Info("credit for cloud acct", "cloudAccountId", cloudAcct.Id, "billingCredit", BillingCredit)
		if errors.Is(err, io.EOF) {
			return totalInitialCreditAmount, lastCreated, nil
		}
		if err != nil {
			logger.Error(err, "failed to get billing credit")
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

func (cloudCreditUsageEventScheduler *CloudCreditUsageEventScheduler) IsCreditUsageThresholdReached(ctx context.Context, cloudAcctType pb.AccountType, initialAmount float64,
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

func (cloudCreditUsageEventScheduler *CloudCreditUsageEventScheduler) notifyCloudCreditThresholdReached(ctx context.Context, cloudAcct *pb.CloudAccount) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditUsageEventScheduler.notifyCloudCreditThresholdReached").WithValues("cloudAccountId", cloudAcct.Id).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	properties := map[string]string{}
	cloudCreditsThresholdReachedRequest := cloudCreditUsageEventScheduler.getPublishCloudCreditsUsageEventRequest(cloudAcct.GetId(), billingCommon.CloudCreditsThresholdReached, billingCommon.CloudCreditsThresholdReachedEventSubType, CloudCreditsThresholdReachedMessage, properties)
	if cloudCreditUsageEventScheduler.notificationGatewayClient == nil {
		logger.Error(errors.New("notification client nil"), "couldn't publish cloud credit event", "cloudAccountId", cloudAcct.Id)
		return nil
	}
	resp, err := cloudCreditUsageEventScheduler.notificationGatewayClient.PublishEvent(ctx, cloudCreditsThresholdReachedRequest)
	if err != nil {
		logger.Error(err, "couldn't publish cloud credit used event", "cloudAccountId", cloudAcct.Id)
	}
	logger.Info("publishing event for cloud credit used for", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "resp", resp)
	return nil
}

func (cloudCreditUsageEventScheduler *CloudCreditUsageEventScheduler) notifyCloudCreditsAvailable(ctx context.Context, cloudAcct *pb.CloudAccount) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditUsageEventScheduler.notifyCloudCreditsAvailable").WithValues("cloudAccountId", cloudAcct.Id).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	properties := map[string]string{}
	cloudCreditsAvailableRequest := cloudCreditUsageEventScheduler.getPublishCloudCreditsUsageEventRequest(cloudAcct.GetId(), billingCommon.CloudCreditsAvailable, billingCommon.CloudCreditsAvailableEventSubType, CloudCreditsAvailableMessage, properties)
	if cloudCreditUsageEventScheduler.notificationGatewayClient == nil {
		logger.Error(errors.New("notification client nil"), "couldn't publish cloud credit event", "cloudAccountId", cloudAcct.Id)
		return nil
	}
	resp, err := cloudCreditUsageEventScheduler.notificationGatewayClient.PublishEvent(ctx, cloudCreditsAvailableRequest)
	if err != nil {
		logger.Error(err, "couldn't publish cloud credit available event", "cloudAccountId", cloudAcct.Id)
	}
	logger.Info("publishing event for cloud credit available for", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "resp", resp)
	return nil
}
