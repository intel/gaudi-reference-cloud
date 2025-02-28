// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"io"
	"time"

	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var planActivationSyncTicker *time.Ticker

func startSyncingPlanActivationToCloudAcct(ctx context.Context, planController PlanController) {
	planActivationSyncTicker = time.NewTicker(time.Duration(config.Cfg.AcctPlanActiveInterval) * time.Second)
	go syncingPlanActivationToCloudAcctLoop(context.Background(), planController)
}

func syncingPlanActivationToCloudAcctLoop(ctx context.Context, planController PlanController) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("SyncingPlanActivationToCloudAcctLoop").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	for {
		syncingPlanActivationToCloudAcct(&logger, planController)
		tm := <-planActivationSyncTicker.C
		if tm.IsZero() {
			return
		}
	}
}

func syncingPlanActivationToCloudAcct(logger *logr.Logger, planController PlanController) {
	logger.Info("syncing plan activation to cloud account", "prefix", config.Cfg.ClientIdPrefix)
}

type PlanController struct {
	cloudAccountClient pb.CloudAccountServiceClient
	ariaAccountClient  *client.AriaAccountClient
}

func NewPlanController(cloudAccountClient pb.CloudAccountServiceClient, ariaAccountClient *client.AriaAccountClient) *PlanController {
	return &PlanController{cloudAccountClient: cloudAccountClient, ariaAccountClient: ariaAccountClient}
}

func (planController *PlanController) syncPlanActivationToCloudAccount(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PlanController.syncPlanActivationToCloudAccount").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("syncing plan activation status to cloud account status")
	acctTypePremium := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	cloudAccountSearchClientForPremium, err :=
		planController.cloudAccountClient.Search(ctx, &pb.CloudAccountFilter{Type: &acctTypePremium})
	if err != nil {
		return err
	}
	acctTypeEnterprise := pb.AccountType_ACCOUNT_TYPE_ENTERPRISE
	cloudAccountSearchClientForEnterprise, err :=
		planController.cloudAccountClient.Search(ctx, &pb.CloudAccountFilter{Type: &acctTypeEnterprise})
	if err != nil {
		return err
	}

	err = planController.syncPlanActivation(ctx, cloudAccountSearchClientForPremium)
	if err != nil {
		logger.Error(err, "failed to sync plan activation for premium accounts")
	}
	err = planController.syncPlanActivation(ctx, cloudAccountSearchClientForEnterprise)
	if err != nil {
		logger.Error(err, "failed to sync plan activation for enterprise accounts")
	}

	return nil
}

func (planController *PlanController) syncPlanActivation(ctx context.Context, cloudAccountSearchClient pb.CloudAccountService_SearchClient) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PlanController.syncPlanActivation").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	for {
		cloudAccount, err := cloudAccountSearchClient.Recv()
		if err == io.EOF {
			if cerr := cloudAccountSearchClient.CloseSend(); cerr != nil {
				logger.Error(cerr, "Failed to terminate client stream")
				return cerr
			}
			return nil
		}
		if err != nil {
			return err
		}
		planInstancesActive, err := planController.checkPlanInstancesActive(ctx, cloudAccount.Id)
		if err != nil {
			logger.Error(err, "failed to get if plans are active for cloud account", "cloudAccountId", cloudAccount.Id)
			continue
		}
		terminatePaidServices := false
		delinquent := false
		if planInstancesActive && (cloudAccount.Delinquent || cloudAccount.TerminatePaidServices) {
			delinquent = false
			terminatePaidServices = false
		} else if !planInstancesActive && (!cloudAccount.Delinquent || !cloudAccount.TerminatePaidServices) {
			delinquent = true
			terminatePaidServices = true
		} else {
			continue
		}
		_, err = planController.cloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{
			Id:                    cloudAccount.Id,
			Delinquent:            &delinquent,
			TerminatePaidServices: &terminatePaidServices,
		})
		if err != nil {
			logger.Error(err, "failed to update termination status for cloud account", "cloudAccountId", cloudAccount.Id)
		}
	}
}

func (planController *PlanController) checkPlanInstancesActive(ctx context.Context, cloudAccountId string) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PlanController.checkPlanInstancesActive").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getAcctPlans, err := planController.ariaAccountClient.GetAcctPlans(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		return false, err
	}
	for _, plan := range getAcctPlans.AcctPlansM {
		const planIsActive int64 = 1
		if plan.PlanInstanceStatusCd != planIsActive {
			return false, nil
		}
	}
	return true, nil
}
