// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	idcCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/trade_scanner/pkg/config"
	tradecheck "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tradecheck/tradecheckintel"
)

type TradeComplianceScanScheduler struct {
	syncTicker         *time.Ticker
	CloudAccountClient *idcCommon.CloudAccountSvcClient
	Cfg                *config.Config
	GTSClient          *tradecheck.GTSclient
}

func NewTradeComplianceScheduler(cloudaccountClient *idcCommon.CloudAccountSvcClient, gtsClient *tradecheck.GTSclient,
	cfg *config.Config) (*TradeComplianceScanScheduler, error) {
	if cloudaccountClient == nil {
		return nil, fmt.Errorf("cloudaccount service is requied")
	}

	return &TradeComplianceScanScheduler{
		syncTicker:         time.NewTicker(time.Duration(cfg.SchedulerInterval) * time.Second),
		Cfg:                cfg,
		CloudAccountClient: cloudaccountClient,
		GTSClient:          gtsClient,
	}, nil
}

func (tradeSchd *TradeComplianceScanScheduler) StartComlianceScanScheduler(ctx context.Context) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TradeComplianceScanScheduler.StartComlianceScanScheduler").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	tradeSchd.ScanTradeComplianceLoop(ctx)
}

func (tradeSchd *TradeComplianceScanScheduler) ScanTradeComplianceLoop(ctx context.Context) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TradeComplianceScanScheduler.ScanTradeComplianceLoop").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	for {
		tradeSchd.ComplianceScan(ctx)
		tm := <-tradeSchd.syncTicker.C
		if tm.IsZero() {
			return
		}
	}
}

func (tradeSchd *TradeComplianceScanScheduler) ComplianceScan(ctx context.Context) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TradeComplianceScanScheduler.ComplianceScan").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	accounts, err := tradeSchd.CloudAccountClient.GetAllCloudAccount(ctx)
	if err != nil {
		logger.Error(err, "error reading cloud accounts")
		return
	}

	for idx, account := range accounts {
		if account.Type == pb.AccountType_ACCOUNT_TYPE_INTEL {
			logger.Info("Skipping check for intel user account ", "cloudAccountId", account.Id)
			continue
		}
		tradeRestriced := account.GetTradeRestricted()
		logger.Info("processing cloudaccount for trade compliance", "idx", idx, "cloudAccountId", account.Id)
		partner := tradecheck.BusinessPartnerRequest{
			EnterpriseID: account.GetPersonId(),
			Name:         account.GetName(),
			Country:      account.GetCountryCode(),
		}
		screenReq := tradecheck.ScreenRequest{}
		screenReq.Partners = append(screenReq.Partners, partner)

		complianceResp, err := tradeSchd.GTSClient.ScreenBusinessPartner(ctx, screenReq)
		if err != nil {
			logger.Error(err, "gts trade compliance error", "cloudAccountId", account.Id)
			continue
		}
		jsonResp, err := json.Marshal(complianceResp)
		if err != nil {
			logger.Info("error decoding compliance report")
			//no action required
			jsonResp = []byte{}
		}

		if complianceResp.BusinessPartner.Status.SPLStatus == "FAIL" ||
			complianceResp.BusinessPartner.Status.EmbargoStatus == "FAIL" {
			logger.Error(err, "gts trade compliance failure", "cloudAccountId", account.Id, "report", string(jsonResp))
			if !tradeRestriced {
				tradeSchd.updateTradeRestricted(ctx, account.Id, true)
			}
			continue
		}
		if tradeRestriced {
			tradeSchd.updateTradeRestricted(ctx, account.Id, false)
		}
		logger.Info("gts trade compliance passed", "cloudAccountId", account.Id, "report", string(jsonResp))
	}
}

func (tradeSchd *TradeComplianceScanScheduler) updateTradeRestricted(ctx context.Context, id string, tradeRestriced bool) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TradeComplianceScanScheduler.updateTradeRestricted").WithValues("cloudAccountId", id).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	_, err := tradeSchd.CloudAccountClient.CloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{
		Id:              id,
		TradeRestricted: &tradeRestriced,
	})
	if err != nil {
		logger.Error(err, "error in updating trade restricted field for cloudaccount")
	}
}
