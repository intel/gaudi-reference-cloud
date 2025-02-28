// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"time"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

var productUsageChan = make(chan bool)

type ProductUsageScheduler struct {
	syncTicker      *time.Ticker
	usageController *UsageController
}

func NewProductUsageScheduler(ctx context.Context, usageController *UsageController, cfg *Config) *ProductUsageScheduler {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ProductUsageScheduler.NewProductUsageScheduler").Start()
	defer span.End()
	logger.Info("cfg", "productUsageSchedulerInterval", cfg.ProductUsageSchedulerInterval)
	return &ProductUsageScheduler{
		syncTicker:      time.NewTicker(time.Duration(cfg.ProductUsageSchedulerInterval) * time.Second),
		usageController: usageController,
	}
}

func (productUsageScheduler *ProductUsageScheduler) StartCalculatingProductUsage(ctx context.Context) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ProductUsageScheduler.StartCalculatingProductUsage").Start()
	defer span.End()
	logger.Info("start product usage calculation")
	go productUsageScheduler.CalculateProductUsageLoop(ctx)
}

func (productUsageScheduler *ProductUsageScheduler) StopProductUsageCalculation() {
	if productUsageChan != nil {
		close(productUsageChan)
		productUsageChan = nil
	}
}

func (productUsageScheduler *ProductUsageScheduler) CalculateProductUsageLoop(ctx context.Context) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ProductUsageScheduler.CalculateProductUsageLoop").Start()
	defer span.End()
	logger.Info("calculating product usages")
	for {
		productUsageScheduler.usageController.CalculateProductUsages(ctx)
		select {
		case <-productUsageChan:
			return
		case tm := <-productUsageScheduler.syncTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}
