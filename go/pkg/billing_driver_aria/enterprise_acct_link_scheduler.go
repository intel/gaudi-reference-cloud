// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	enterpriseAcctLinkChannel = make(chan bool)
	enterpriseAcctLinkTicker  *time.Ticker
)

func StartEnterpriseAcctLinkScheduler(ctx context.Context, enterpriseAcctLinkScheduler *EnterpriseAcctLinkScheduler) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("StartEnterpriseAcctLinkScheduler").Start()
	defer span.End()
	logger.Info("BEGIN")
	logger.Info("cfg", "EnterpriseAcctLinkSchedulerInterval", config.Cfg.EnterpriseAcctLinkSchedulerInterval)
	defer logger.Info("END")
	enterpriseAcctLinkTicker = time.NewTicker(time.Duration(config.Cfg.EnterpriseAcctLinkSchedulerInterval) * time.Second)
	go enterpriseAcctLinkLoop(context.Background(), enterpriseAcctLinkScheduler)
}

func StopEnterpriseAcctLinkScheduler() {
	if enterpriseAcctLinkChannel != nil {
		close(enterpriseAcctLinkChannel)
		enterpriseAcctLinkChannel = nil
	}
}

func enterpriseAcctLinkLoop(ctx context.Context, enterpriseAcctLinkScheduler *EnterpriseAcctLinkScheduler) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EnterpriseAcctLinkScheduler.enterpriseAcctLinkLoop").Start()
	defer span.End()
	for {
		err := enterpriseAcctLinkScheduler.handleLinkedEnterpriseAcct(ctx)
		if err != nil {
			logger.Error(err, "failed to handle linked enterprise account")
		}
		select {
		case <-enterpriseAcctLinkChannel:
			return
		case tm := <-enterpriseAcctLinkTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

type EnterpriseAcctLinkScheduler struct {
	cloudAccountClient      pb.CloudAccountServiceClient
	ariaAccountClient       *client.AriaAccountClient
	ariaServiceCreditClient *client.ServiceCreditClient
	eventManager            *events.EventManager
}

func NewEnterpriseAcctLinkScheduler(ctx context.Context, eventManager *events.EventManager, cloudAccountClient pb.CloudAccountServiceClient, ariaAccountClient *client.AriaAccountClient, ariaServiceCreditClient *client.ServiceCreditClient) *EnterpriseAcctLinkScheduler {
	logger := log.FromContext(ctx).WithName("EnterpriseAcctLinkScheduler.NewEnterpriseAcctLinkScheduler")
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("cfg", "enterpriseAcctLinkSchedulerInterval", config.Cfg.EnterpriseAcctLinkSchedulerInterval)
	return &EnterpriseAcctLinkScheduler{
		eventManager:            eventManager,
		cloudAccountClient:      cloudAccountClient,
		ariaAccountClient:       ariaAccountClient,
		ariaServiceCreditClient: ariaServiceCreditClient,
	}
}

func (enterpriseAcctLinkScheduler *EnterpriseAcctLinkScheduler) handleLinkedEnterpriseAcct(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EnterpriseAcctLinkScheduler.handleLinkedEnterpriseAcct").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	acctType := pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING
	cloudAccountSearchClient, err :=
		enterpriseAcctLinkScheduler.cloudAccountClient.Search(ctx, &pb.CloudAccountFilter{Type: &acctType})
	if err != nil {
		return err
	}

	for {
		cloudAccount, err := cloudAccountSearchClient.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		// Disabling adding default credits #TWC4726-1391
		// addedDefaultCreditsIfLinked, err := enterpriseAcctLinkScheduler.addDefaultCreditsIfLinked(ctx, cloudAccount.Id)
		// if err != nil {
		// 	logger.Error(err, "failed to add default credits if linked cloud account", "id", cloudAccount.Id)
		// 	continue
		// }

		isLinked, err := enterpriseAcctLinkScheduler.isAccountLinked(ctx, cloudAccount.Id)
		if err != nil {
			logger.Error(err, "failed to get if cloud account isLinked", "cloudAccountId", cloudAccount.Id)
			continue
		}
		if isLinked {
			acctType := pb.AccountType_ACCOUNT_TYPE_ENTERPRISE
			paidServicesAllowed := true
			_, err = enterpriseAcctLinkScheduler.cloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{
				Id:                  cloudAccount.Id,
				Type:                &acctType,
				PaidServicesAllowed: &paidServicesAllowed,
			})
			if err != nil {
				logger.Error(err, "failed to update cloud account to enterprise for", "cloudAccount", cloudAccount.Id)
			}
		}
	}
}

// not used until we re-enable adding default-credits
func (enterpriseAcctLinkScheduler *EnterpriseAcctLinkScheduler) addDefaultCreditsIfLinked(ctx context.Context, cloudAccountId string) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EnterpriseAcctLinkScheduler.addDefaultCreditsIfLinked").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	getAcctNoResponse, err := enterpriseAcctLinkScheduler.ariaAccountClient.GetAccountNoFromUserId(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		logger.Error(err, "failed to get account number from cloud account", "cloudAccount", cloudAccountId)
		return false, err
	}

	getAcctHierarchyResponse, err := enterpriseAcctLinkScheduler.ariaAccountClient.GetAccountHierarchyDetails(ctx, getAcctNoResponse.AcctNo)
	if err != nil {
		logger.Error(err, "failed to get acct hierarchy details", "cloudAccountId", cloudAccountId)
		return false, err
	}

	for _, acctHierarchy := range getAcctHierarchyResponse.AcctHierarchyDtls {
		const nilSeniorAcctNo = 0
		if acctHierarchy.Userid == client.GetAccountClientId(cloudAccountId) && acctHierarchy.SeniorAcctNo != nilSeniorAcctNo {
			logger.V(9).Info("parent account found", "cloudAccountId", cloudAccountId, "ariaUserid", acctHierarchy.Userid, "parentAccount", acctHierarchy.SeniorAcctNo)
			acctCredits, err := enterpriseAcctLinkScheduler.ariaAccountClient.GetAccountCredits(ctx, getAcctNoResponse.AcctNo)
			if err != nil {
				logger.Error(err, "failed to get account credits", "cloudAccountId", cloudAccountId, "ariaUserid", acctHierarchy.Userid, "parentAccount", acctHierarchy.SeniorAcctNo)
				return false, err
			}
			for _, acctCredit := range acctCredits.AllCredits {
				//TODO: use a specific reason code for enterprise credit assignment
				creditDetails, err := enterpriseAcctLinkScheduler.ariaAccountClient.GetAccountCreditDetails(ctx, client.GetAccountClientId(cloudAccountId), acctCredit.CreditNo)
				if err != nil {
					logger.Error(err, "failed to get AccountCreditDetails", "cloudAccountId", cloudAccountId, "ariaUserid", acctHierarchy.Userid, "parentAccount", acctHierarchy.SeniorAcctNo)
					return false, err
				}

				if comments := creditDetails.Comments; comments != "" && comments == defaultCreditAssignmentComment {
					logger.V(9).Info("default cloud credit has been assigned to enterprise account", "cloudAccountId", cloudAccountId)
					return true, nil
				}
			}
			currentDate := time.Now()
			newDate := currentDate.AddDate(0, 0, config.Cfg.EntDefaultCreditExpirationDays)
			expirationDate := fmt.Sprintf("%04d-%02d-%02d", newDate.Year(), newDate.Month(), newDate.Day())
			_, err = enterpriseAcctLinkScheduler.ariaServiceCreditClient.CreateServiceCredits(ctx, client.GetAccountClientId(cloudAccountId), config.Cfg.EntDefaultCreditAmount, kFixMeWeHaventImplementedReasonCodeYet, expirationDate, defaultCreditAssignmentComment)
			if err != nil {
				logger.Error(err, "failed to assign default cloud credit to enterprise account", "cloudAccountId", cloudAccountId)
				return false, err
			}
			logger.Info("assigned default cloud credit to enterprise account", "cloudAccountId", cloudAccountId)
			// notify user when default cloud credits are added to account
			err = enterpriseAcctLinkScheduler.eventManager.Create(ctx, events.CreateEvent{
				Status:         events.EventStatus_ACTIVE,
				Type:           events.EventType_NOTIFICATION,
				Severity:       events.EventSeverity_LOW,
				ServiceName:    "billing",
				Message:        "default cloud credits added",
				CloudAccountId: cloudAccountId,
				EventSubType:   "ASSIGNED_DEFAULT_CLOUD_CREDIT",
				ClientRecordId: uuid.NewString(),
			})

			if err != nil {
				logger.Error(err, "error notifying", "cloudAccountId", cloudAccountId)
			}
			return true, nil
		} else {
			logger.Info("parent account not found", "cloudAccountId", cloudAccountId, "ariaUserid", acctHierarchy.Userid)
		}
	}

	return false, nil
}

func (enterpriseAcctLinkScheduler *EnterpriseAcctLinkScheduler) isAccountLinked(ctx context.Context, cloudAccountId string) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EnterpriseAcctLinkScheduler.isAccountLinked").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	getAcctNoResponse, err := enterpriseAcctLinkScheduler.ariaAccountClient.GetAccountNoFromUserId(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		logger.Error(err, "failed to get account number from cloud account", "id", cloudAccountId)
		return false, err
	}

	getAcctHierarchyResponse, err := enterpriseAcctLinkScheduler.ariaAccountClient.GetAccountHierarchyDetails(ctx, getAcctNoResponse.AcctNo)
	if err != nil {
		logger.Error(err, "failed to get acct hierarchy details", "cloudAccountId", cloudAccountId)
		return false, err
	}

	for _, acctHierarchy := range getAcctHierarchyResponse.AcctHierarchyDtls {
		const nilSeniorAcctNo = 0
		if acctHierarchy.Userid == client.GetAccountClientId(cloudAccountId) && acctHierarchy.SeniorAcctNo != nilSeniorAcctNo {
			logger.Info("parent account found", "cloudAccountId", cloudAccountId, "ariaUserid", acctHierarchy.Userid, "parentAccount", acctHierarchy.SeniorAcctNo)
			return true, nil
		} else {
			logger.Info("parent account not found", "cloudAccountId", cloudAccountId, "ariaUserid", acctHierarchy.Userid)
		}
	}

	return false, nil
}
