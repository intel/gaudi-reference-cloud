// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/go-logr/logr"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	cloudCreditExpiryTicker  *time.Ticker
	cloudCreditExpiryChannel = make(chan bool)
)

func startCloudCreditExpiryScheduler(ctx context.Context, cloudCreditExpiryScheduler CloudCreditExpiryScheduler) {
	cloudCreditExpiryTicker = time.NewTicker((time.Duration(Cfg.CreditExpirySchedulerInterval) * time.Minute))
	go cloudCreditExpiryLoop(context.Background(), cloudCreditExpiryScheduler)
}

func cloudCreditExpiryLoop(ctx context.Context, cloudCreditExpiry CloudCreditExpiryScheduler) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditExpiryScheduler.cloudCreditExpiryLoop").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	for {
		newValue := time.Now().Format(time.RFC3339)
		cloudCreditExpiry.schedulerCloudAccountState.Mutex.Lock()
		cloudCreditExpiry.schedulerCloudAccountState.AccessTimestamp = newValue
		logger.V(1).Info("last access timestamp", "AccessTimestamp", cloudCreditExpiry.schedulerCloudAccountState.AccessTimestamp)
		checkCloudCreditExpiry(ctx, &logger, cloudCreditExpiry)
		cloudCreditExpiry.schedulerCloudAccountState.Mutex.Unlock()
		select {
		case <-cloudCreditExpiryChannel:
			return
		case tm := <-cloudCreditExpiryTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

func stopCloudCreditExpiryScheduler() {
	if cloudCreditExpiryChannel != nil {
		close(cloudCreditExpiryChannel)
		cloudCreditExpiryChannel = nil
	}
}

func checkCloudCreditExpiry(ctx context.Context, logger *logr.Logger, cloudCreditExpiryScheduler CloudCreditExpiryScheduler) {
	logger.Info("check cloud credits expiry")
	err := cloudCreditExpiryScheduler.cloudCreditExpiry(ctx)
	if err != nil {
		logger.Error(err, "failed to handle cloud credit expiry")
	}
}

// add standardized errors
const (
	CloudCreditExpiryInvalidCloudAcctError          = "cloud credit expiry: invalid cloud account"
	CloudCreditExpiryGetAllCreditsError             = "cloud credit expiry: failed to get all credits"
	CloudCreditExpiryUpdateCloudAcctHasCreditsError = "cloud credit expiry: failed to update cloud account has credits"
	CloudCreditExpired                              = "Cloud Credits Expired"
)

type CloudCreditExpiryScheduler struct {
	eventManager               *events.EventManager
	notificationClient         *NotificationClient
	schedulerCloudAccountState *SchedulerCloudAccountState
	cloudAccountClient         *billingCommon.CloudAccountSvcClient
}

func NewCloudCreditExpiryScheduler(eventManager *events.EventManager, notificationClient *NotificationClient,
	schedulerCloudAccountState *SchedulerCloudAccountState, cloudAccountClient *billingCommon.CloudAccountSvcClient) *CloudCreditExpiryScheduler {
	return &CloudCreditExpiryScheduler{eventManager: eventManager, notificationClient: notificationClient, schedulerCloudAccountState: schedulerCloudAccountState, cloudAccountClient: cloudAccountClient}
}

func (cloudCreditExpiryScheduler *CloudCreditExpiryScheduler) cloudCreditExpiry(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditExpiryScheduler.cloudCreditsExpiry").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	cloudAccts, err := cloudCreditExpiryScheduler.cloudAccountClient.GetCloudAccountsWithCredits(ctx)
	if err != nil {
		logger.Error(err, "failed to get cloud accounts with credits", "context", "GetCloudAccountsWithCredits")
		return err
	}
	for _, cloudAcct := range cloudAccts {
		if strings.Contains(cloudAcct.Name, "iks") || strings.Contains(cloudAcct.Name, "validator") {
			continue
		}
		logger.V(9).Info("checking for cloud credit expiry for cloud acct", "cloudAccountId", cloudAcct.GetId(), "cloudAccountType", cloudAcct.Type)
		lastExpiration, err := cloudCreditExpiryScheduler.getLastExpiration(ctx, cloudAcct)
		if err != nil {
			logger.Error(GetSchedulerError(CloudCreditExpiryGetAllCreditsError, err),
				"failed handling cloud credit expiry", "cloudAccountId", cloudAcct.GetId())
			continue
		}
		if time.Now().After(lastExpiration) {
			logger.V(9).Info("cloud credits expired for cloud acct", "cloudAccountId", cloudAcct.GetId(), "cloudAccountType", cloudAcct.Type)
			paidServicesAllowed := cloudAcct.GetPaidServicesAllowed()
			err = UpdateCloudAcctNoCredits(ctx, cloudAcct)
			if err != nil {
				logger.Error(fmt.Errorf("cloud credit expiry:%w", err),
					"failed to update cloud accct for no credits upon expiry", "cloudAccountId", cloudAcct.GetId())
			}
			logger.V(9).Info("sending alert for cloud credit expiry for", "cloudAccountId", cloudAcct.GetId(), "cloudAccountType", cloudAcct.Type)
			// err = cloudCreditExpiryScheduler.eventManager.Create(ctx, events.CreateEvent{
			// 	Status:         events.EventStatus_ACTIVE,
			// 	Type:           events.EventType_ALERT,
			// 	ClientRecordId: uuid.NewString(),
			// })
			// if err != nil {
			// 	logger.Error(err, "failed to send event upon expiry", "cloudAccountId", cloudAcct.GetId(), "cloudAccountType", cloudAcct.Type, "context", "Create")
			// }
			hasCreditCard, err := IsCloudAccountWithCreditCard(ctx, cloudAcct)
			if cloudCreditExpiryScheduler.notificationClient == nil || hasCreditCard || err != nil {
				continue
			}
			if Cfg.GetCreditExpiryEmail() && !cloudAcct.GetTerminatePaidServices() && paidServicesAllowed && Cfg.Features.ServicesTerminationScheduler {
				emailNotificationEnabled := slices.Contains(Cfg.GetCreditExpiryEmailAccountTypes(), cloudAcct.GetType().String())
				logger.V(9).Info("cloud account credit usage", "cloudAccountId", cloudAcct.Id, "creditUsageEmailAccountTypes", Cfg.GetCreditExpiryEmailAccountTypes(), "accountType", cloudAcct.GetType().String(), "emailNotificationEnabled", emailNotificationEnabled)
				if emailNotificationEnabled {
					logger.V(9).Info("sending email for cloud credit expiry for cloud account", "cloudAccountId", cloudAcct.GetId(), "cloudAccountType", cloudAcct.Type)
					if err = cloudCreditExpiryScheduler.notificationClient.SendCloudCreditUsageEmailWithOptions(ctx, CloudCreditExpired, cloudAcct.Owner, Cfg.GetCloudCreditExpiryTemplate()); err != nil {
						logger.Error(err, "couldn't send email", "cloudAccountId", cloudAcct.GetId())
					}
				} else {
					logger.V(9).Info("skipping email for cloud credit expiry for cloud account", "cloudAccountId", cloudAcct.GetId(), "cloudAccountType", cloudAcct.Type)
				}
			}
		} else {
			logger.V(9).Info("cloud credits have not expired for cloud acct", "cloudAccountId", cloudAcct.GetId(), "cloudAccountType", cloudAcct.Type)
			err = UpdateCloudAcctHasCredits(ctx, cloudAcct, false)
			if err != nil {
				logger.Error(GetSchedulerError(CloudCreditExpiryUpdateCloudAcctHasCreditsError, err),
					"failed handling cloud credit expiry", "cloudAccountId", cloudAcct.GetId())
				continue
			}
			const minsInDay = 1440
			if (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM &&
				lastExpiration.Before(time.Now().AddDate(0, 0, int(Cfg.PremiumCloudCreditNotifyBeforeExpiry)/minsInDay))) ||
				// todo: the check for standard has to be different than intel but for now it is ok to be the same.
				(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_STANDARD &&
					lastExpiration.Before(time.Now().AddDate(0, 0, int(Cfg.IntelCloudCreditNotifyBeforeExpiry)/minsInDay))) ||
				(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_INTEL &&
					lastExpiration.Before(time.Now().AddDate(0, 0, int(Cfg.IntelCloudCreditNotifyBeforeExpiry)/minsInDay))) ||
				(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE &&
					lastExpiration.Before(time.Now().AddDate(0, 0, int(Cfg.EnterpriseCloudCreditNotifyBeforeExpiry)/minsInDay))) {
				// todo: integrate with notifications
				logger.V(9).Info("cloud credits are about to expire for cloud account", "cloudAccountId", cloudAcct.GetId(), "cloudAccountType", cloudAcct.Type)
			}
		}

	}
	return nil
}

func (cloudCreditExpiryScheduler *CloudCreditExpiryScheduler) getLastExpiration(ctx context.Context,
	cloudAcct *pb.CloudAccount) (time.Time, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditExpiryScheduler.getLastExpiration").WithValues("cloudAccountId", cloudAcct.GetId(), "cloudAccountType", cloudAcct.GetType()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	driver, err := GetDriver(ctx, cloudAcct.Id)
	if err != nil {
		logger.Error(err, "failed to get driver for cloud account")
		return time.Time{}, GetSchedulerError(CloudCreditExpiryInvalidCloudAcctError, err)
	}
	billingCreditClient, err := driver.billingCredit.ReadInternal(ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.GetId()})
	if err != nil {
		logger.Error(err, "failed to get billing credit client for cloud account", "context", "ReadInternal")
		return time.Time{}, err
	}
	// go back 100 years
	var lastCreditExpiry = time.Now().AddDate(-100, 0, 0)
	for {
		billingCredit, err := billingCreditClient.Recv()
		if errors.Is(err, io.EOF) {
			return lastCreditExpiry, nil
		}
		if err != nil {
			logger.Error(err, "failed to get billing credits for cloud account", "context", "ReadInternal")
			return time.Time{}, err
		}
		if billingCredit.Expiration.AsTime().After(lastCreditExpiry) {
			lastCreditExpiry = billingCredit.Expiration.AsTime()
		}
	}
}
