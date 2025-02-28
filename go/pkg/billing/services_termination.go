// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	servicesTerminationTicker *time.Ticker
	serviceTerminationChannel = make(chan bool)
)

type ServicesTerminationScheduler struct {
	schedulerCloudAccountState *SchedulerCloudAccountState
	cloudAccountClient         *billingCommon.CloudAccountSvcClient
	meteringServiceClient      *billingCommon.MeteringClient
}

func NewServicesTerminationScheduler(schedulerCloudAccountState *SchedulerCloudAccountState, cloudAccountClient *billingCommon.CloudAccountSvcClient, meteringClient *billingCommon.MeteringClient) *ServicesTerminationScheduler {
	return &ServicesTerminationScheduler{schedulerCloudAccountState: schedulerCloudAccountState, cloudAccountClient: cloudAccountClient, meteringServiceClient: meteringClient}
}

func startServicesTerminationScheduler(ctx context.Context, servicesTerminationScheduler ServicesTerminationScheduler) {
	servicesTerminationTicker = time.NewTicker((time.Duration(Cfg.ServicesTerminationSchedulerInterval) * time.Minute))
	go servicesTerminationLoop(context.Background(), servicesTerminationScheduler)
}

func servicesTerminationLoop(ctx context.Context, servicesTerminationScheduler ServicesTerminationScheduler) {
	ctx, logger, _ := obs.LogAndSpanFromContext(ctx).WithName("ServicesTerminationScheduler.servicesTerminationLoop").Start()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	for {
		servicesTerminationScheduler.schedulerCloudAccountState.Mutex.Lock()
		logger.V(9).Info("last access timestamp", "accessTimestamp", servicesTerminationScheduler.schedulerCloudAccountState.AccessTimestamp)
		checkServicesTermination(ctx, &logger, servicesTerminationScheduler)
		servicesTerminationScheduler.schedulerCloudAccountState.Mutex.Unlock()
		select {
		case <-serviceTerminationChannel:
			return
		case tm := <-servicesTerminationTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

func stopServicesTerminationScheduler() {
	if serviceTerminationChannel != nil {
		close(serviceTerminationChannel)
		serviceTerminationChannel = nil
	}
}

func checkServicesTermination(ctx context.Context, logger *logr.Logger, servicesTerminationScheduler ServicesTerminationScheduler) {
	logger.Info("check services termination")
	err := servicesTerminationScheduler.servicesTermination(ctx)
	if err != nil {
		logger.Error(err, "failed to handle services termination")
	}
}

// add standardized errors
const (
	ServicesTerminationExpiryInvalidCloudAcctError           string = "services termination: invalid cloud account"
	ServicesTerminationUpdateCloudAcctTerminateServicesError string = "services termination: failed to update cloud acct for service termination"
)

func (servicesTerminationScheduler *ServicesTerminationScheduler) servicesTermination(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ServicesTerminationScheduler.servicesTermination").Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")
	servicesTerminationAccountTypes := Cfg.GetServicesTerminationAccountTypes()
	logger.V(2).Info("configured service termination account types", "serviceTerminationAccountTypes", servicesTerminationAccountTypes)
	accountTypes := GetAccountTypes(ctx, servicesTerminationAccountTypes)
	if len(accountTypes) == 0 {
		return fmt.Errorf("no account types enabled for service termination %v", servicesTerminationAccountTypes)
	}
	logger.V(2).Info("service termination account types", "accountTypes", accountTypes)
	paidServicesAllowed := false
	terminatePaidServices := false
	filter := &pb.CloudAccountFilter{
		PaidServicesAllowed:   &paidServicesAllowed,
		TerminatePaidServices: &terminatePaidServices,
	}
	cloudAccts, err := servicesTerminationScheduler.cloudAccountClient.GetCloudAcctsOfTypesWithFilter(ctx, accountTypes, filter)
	if err != nil {
		logger.Error(err, "failed to get cloud accounts for checking termination of paid instances", "accountTypes", accountTypes)
		return err
	}
	logger.V(2).Info("cloud account for service termination", "NumberOfcloudAccounts", len(cloudAccts))
	for _, cloudAcct := range cloudAccts {
		// check for running instances
		filter := &pb.MeteringAvailableFilter{CloudAccountId: cloudAcct.Id, MeteringDuration: Cfg.InstanceTerminationMeteringDuration}
		hasRunningInstances, err := servicesTerminationScheduler.meteringServiceClient.IsMeteringRecordAvailable(ctx, filter)
		if err != nil {
			logger.Error(GetSchedulerError(ServicesTerminationUpdateCloudAcctTerminateServicesError, err),
				"failed to get metering data", "cloudAccountId", cloudAcct.GetId())
		}
		if !hasRunningInstances {
			logger.Info("no running instance found", "cloudAcct.Id", cloudAcct.Id)
			continue
		}
		logger.V(9).Info("running instance found", "cloudAcct.Id", cloudAcct.Id)
		terminatePaidServices := true
		updateCloudAcct := func() {
			_, err := cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{
				Id:                    cloudAcct.Id,
				TerminatePaidServices: &terminatePaidServices,
			})
			if err != nil {
				logger.Error(GetSchedulerError(ServicesTerminationUpdateCloudAcctTerminateServicesError, err),
					"failed handling services termination", "cloudAccountId", cloudAcct.GetId())
			}
		}
		if cloudAcct.CreditsDepleted != nil && cloudAcct.CreditsDepleted.AsTime().Unix() != 0 {
			logger.V(9).Info("service termination ", "cloudAccountId", cloudAcct.GetId())
			terminationCheck := cloudAcct.CreditsDepleted.AsTime().AddDate(0, 0, int(Cfg.ServicesTerminationInterval/1440))
			terminationCheckFlag := time.Now().After(terminationCheck)
			logger.V(9).Info("service termination check ", "creditsDepleted", cloudAcct.CreditsDepleted.AsTime(), "terminationCheck", terminationCheck, "currentTime", time.Now(), "terminatePaidServices", cloudAcct.TerminatePaidServices, "paidServicesAllowed", cloudAcct.PaidServicesAllowed, "terminationCheckFlag", terminationCheckFlag)
			if terminationCheckFlag && !cloudAcct.TerminatePaidServices && !cloudAcct.PaidServicesAllowed {
				switch cloudAcct.Type {
				case pb.AccountType_ACCOUNT_TYPE_STANDARD, pb.AccountType_ACCOUNT_TYPE_INTEL:
					logger.V(9).Info("update cloud account for service termination ", "cloudAccountId", cloudAcct.GetId())
					updateCloudAcct()
				case pb.AccountType_ACCOUNT_TYPE_PREMIUM:
					billingOptions, err := GetBillingOptions(ctx, cloudAcct)
					if err != nil {
						logger.Error(GetSchedulerError(GetBillingOptionsError, err), "service termination scheduler error")
						continue
					}
					logger.V(9).Info("premium account service termination ", "billingOptions", billingOptions)
					if !CheckIfOptionsHasCreditCard(ctx, billingOptions) {
						logger.Info("update cloud account for service termination ", "cloudAccountId", cloudAcct.GetId())
						updateCloudAcct()
					}
				}
			}
		}
	}
	return nil
}
