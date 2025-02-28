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
	cloudCreditExpiryEventTicker  *time.Ticker
	cloudCreditExpiryEventChannel = make(chan bool)
)

const (
	CloudCreditsExpiredMessage       = "cloud credits expired"
	CloudCreditsAboutToExpireMessage = "cloud credits about to expire"
)

func startCloudCreditExpiryEventScheduler(ctx context.Context, cloudCreditExpiryEventScheduler CloudCreditExpiryEventScheduler) {
	cloudCreditExpiryEventTicker = time.NewTicker((time.Duration(config.Cfg.CreditExpiryEventSchedulerInterval) * time.Minute))
	go cloudCreditExpiryEventLoop(context.Background(), cloudCreditExpiryEventScheduler)
}

func cloudCreditExpiryEventLoop(ctx context.Context, cloudCreditExpiryEvent CloudCreditExpiryEventScheduler) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("cloudCreditExpiryEvent.cloudCreditExpiryEventLoop").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	for {
		checkCloudCreditExpiryEvent(ctx, &logger, cloudCreditExpiryEvent)
		select {
		case <-cloudCreditExpiryEventChannel:
			return
		case tm := <-cloudCreditExpiryEventTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

func stopCloudCreditExpiryEventScheduler() {
	if cloudCreditExpiryEventChannel != nil {
		close(cloudCreditExpiryEventChannel)
		cloudCreditExpiryEventChannel = nil
	}
}

func checkCloudCreditExpiryEvent(ctx context.Context, logger *logr.Logger, cloudCreditExpiryEventScheduler CloudCreditExpiryEventScheduler) {
	logger.Info("check cloud credits expiry")
	err := cloudCreditExpiryEventScheduler.PublishCloudCreditExpiryEvent(ctx)
	if err != nil {
		logger.Error(err, "failed to handle cloud credit expiry")
	}
}

// add standardized errors
const (
	CloudCreditExpiryEventInvalidCloudAcctError          = "cloud credit expiry: invalid cloud account"
	CloudCreditExpiryEventGetAllCreditsError             = "cloud credit expiry: failed to get all credits"
	CloudCreditExpiryEventUpdateCloudAcctHasCreditsError = "cloud credit expiry: failed to update cloud account has credits"
	CloudCreditExpiredEvent                              = "Cloud Credits Expired"
	TopicName                                            = "idc-staging-cloud-credits-topic"
)

type CloudCreditExpiryEventScheduler struct {
	notificationGatewayClient  billingCommon.NotificationGatewayClientInterface
	schedulerCloudAccountState *SchedulerCloudAccountState
	cloudAccountClient         *billingCommon.CloudAccountSvcClient
}

func NewCloudCreditExpiryEventScheduler(notificationGatewayClient billingCommon.NotificationGatewayClientInterface, schedulerCloudAccountState *SchedulerCloudAccountState, cloudAccountClient *billingCommon.CloudAccountSvcClient) *CloudCreditExpiryEventScheduler {
	return &CloudCreditExpiryEventScheduler{notificationGatewayClient: notificationGatewayClient, schedulerCloudAccountState: schedulerCloudAccountState, cloudAccountClient: cloudAccountClient}
}

func (cloudCreditExpiryEventScheduler *CloudCreditExpiryEventScheduler) PublishCloudCreditExpiryEvent(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditExpiryEventScheduler.PublishCloudCreditExpiryEvent").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	cloudAccts, err := cloudCreditExpiryEventScheduler.cloudAccountClient.GetCloudAccountsWithCredits(ctx)
	if err != nil {
		logger.Error(err, "failed to get cloud accounts with credits")
		return err
	}

	for _, cloudAcct := range cloudAccts {
		logger.Info("checking for cloud credit expiry for cloud acct", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)

		lastExpiration, err := cloudCreditExpiryEventScheduler.getLastExpiration(ctx, cloudAcct)
		if err != nil {
			logger.Error(GetSchedulerError(CloudCreditExpiryEventGetAllCreditsError, err),
				"failed handling cloud credit expiry", "cloud account id", cloudAcct.GetId())
			continue
		}
		if time.Now().After(lastExpiration) {
			logger.Info("sending alert for cloud credit expiry for", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)

			if cloudCreditExpiryEventScheduler.notificationGatewayClient == nil {
				logger.Error(errors.New("notification client nil"), "couldn't publish cloud credit event", "cloudAccountId", cloudAcct.Id)
				continue
			}
			driver, err := billingCommon.GetDriver(ctx, cloudAcct.Id)
			errMsg := fmt.Sprint("failed handling cloud credit expiry for cloud account id: ", cloudAcct.GetId())
			if err != nil {
				logger.Error(GetSchedulerError(CloudCreditExpiryEventInvalidCloudAcctError, err), errMsg)
				continue
			}
			hasCreditCard, err := cloudCreditExpiryEventScheduler.cloudAccountClient.IsCloudAccountWithCreditCard(ctx, cloudAcct, driver)
			if err != nil {
				logger.Error(err, "couldn't get hascreditcard", "cloudAccountId", cloudAcct.Id)
				continue
			}
			if (!cloudAcct.GetPaidServicesAllowed() && cloudAcct.GetLowCredits()) || hasCreditCard {
				logger.Info("Skip publishing expiry event", "cloudAccountId", cloudAcct.Id)
				continue
			}
			properties := map[string]string{}
			cloudCreditExpiredRequest := cloudCreditExpiryEventScheduler.getPublishCloudCreditsEventRequest(cloudAcct.GetId(), billingCommon.CloudCreditsExpired, billingCommon.CloudCreditsExpiredEventSubType, CloudCreditsExpiredMessage, properties)
			resp, err := cloudCreditExpiryEventScheduler.notificationGatewayClient.PublishEvent(ctx, cloudCreditExpiredRequest)
			if err != nil {
				logger.Error(err, "couldn't publish cloud credit expiry event", "cloudAccountId", cloudAcct.Id)
			}
			logger.Info("publishing event for cloud credit expiry for", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "resp", resp)
		} else {
			logger.Info("cloud credits have not expired for cloud acct", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
			const minsInDay = 1440
			if (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM &&
				lastExpiration.Before(time.Now().AddDate(0, 0, int(config.Cfg.PremiumCloudCreditNotifyBeforeExpiry)/minsInDay))) ||
				// todo: the check for standard has to be different than intel but for now it is ok to be the same.
				(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_STANDARD &&
					lastExpiration.Before(time.Now().AddDate(0, 0, int(config.Cfg.IntelCloudCreditNotifyBeforeExpiry)/minsInDay))) ||
				(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_INTEL &&
					lastExpiration.Before(time.Now().AddDate(0, 0, int(config.Cfg.IntelCloudCreditNotifyBeforeExpiry)/minsInDay))) ||
				(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE &&
					lastExpiration.Before(time.Now().AddDate(0, 0, int(config.Cfg.EnterpriseCloudCreditNotifyBeforeExpiry)/minsInDay))) {
				logger.Info("cloud credits are about to expire for cloud account", "cloudAccountId", cloudAcct.Id)
				properties := map[string]string{}
				cloudCreditAboutToExpireRequest := cloudCreditExpiryEventScheduler.getPublishCloudCreditsEventRequest(cloudAcct.GetId(), billingCommon.CloudCreditsAboutToExpire, billingCommon.CloudCreditsAboutToExpireEventSubType, CloudCreditsAboutToExpireMessage, properties)
				resp, err := cloudCreditExpiryEventScheduler.notificationGatewayClient.PublishEvent(ctx, cloudCreditAboutToExpireRequest)
				if err != nil {
					logger.Error(err, "couldn't publish cloud credit expiry event", "cloudAccountId", cloudAcct.Id)
				}
				logger.Info("publishing event for cloud credit expiry for", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "resp", resp)
			}
		}

	}
	return nil
}

func (cloudCreditExpiryEventScheduler *CloudCreditExpiryEventScheduler) getPublishCloudCreditsEventRequest(cloudAccountId string, eventName string, eventSubType string, message string, properties map[string]string) *pb.PublishEventRequest {
	eventSeverity := pb.EventSeverity_LOW
	serviceName := pb.ServiceName_CREDIT
	userId := uuid.NewString()
	clientRecordId := uuid.NewString()
	cloudcreditsEvent := &pb.CreateEvent{
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
	cloudCreditEventRequest := &pb.PublishEventRequest{
		TopicName:   TopicName,
		Subject:     eventSubType,
		CreateEvent: cloudcreditsEvent,
	}
	return cloudCreditEventRequest
}
func (cloudCreditExpiryEventScheduler *CloudCreditExpiryEventScheduler) getLastExpiration(ctx context.Context,
	cloudAcct *pb.CloudAccount) (time.Time, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditExpiryEventScheduler.getLastExpiration").Start()
	logger.WithValues("cloudaccountid", cloudAcct.Id)
	defer span.End()

	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("checking for cloud credit expiry for cloud acct", "id", cloudAcct.Id, "type", cloudAcct.Type)
	driver, err := billingCommon.GetDriver(ctx, cloudAcct.Id)
	if err != nil {
		return time.Time{}, GetSchedulerError(CloudCreditExpiryEventInvalidCloudAcctError, err)
	}
	billingCreditClient, err := driver.BillingCredit.ReadInternal(ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
	if err != nil {
		logger.Error(err, "failed to get billing credit")
		return time.Time{}, err
	}
	// go back 100 years
	var lastCreditExpiry = time.Now().AddDate(-100, 0, 0)
	for {
		BillingCredit, err := billingCreditClient.Recv()
		if errors.Is(err, io.EOF) {
			return lastCreditExpiry, nil
		}
		if err != nil {
			logger.Error(err, "failed to get billing credit")
			return time.Time{}, err
		}
		if BillingCredit.Expiration.AsTime().After(lastCreditExpiry) {
			lastCreditExpiry = BillingCredit.Expiration.AsTime()
		}
	}
}
