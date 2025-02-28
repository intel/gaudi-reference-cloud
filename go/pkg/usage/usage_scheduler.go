// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

var usageChan = make(chan bool)

// Used for testing
var SyncWait = atomic.Pointer[sync.WaitGroup]{}

type UsageScheduler struct {
	syncTicker      *time.Ticker
	usageController *UsageController
}

func NewUsageScheduler(ctx context.Context, usageController *UsageController, cfg *Config) *UsageScheduler {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageScheduler.NewUsageScheduler").Start()
	defer span.End()
	logger.Info("cfg", "usageSchedulerInterval", cfg.UsageSchedulerInterval)
	return &UsageScheduler{
		syncTicker:      time.NewTicker(time.Duration(cfg.UsageSchedulerInterval) * time.Second),
		usageController: usageController,
	}
}

func (usageScheduler *UsageScheduler) StartCalculatingUsage(ctx context.Context) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageScheduler.StartCalculatingUsage").Start()
	defer span.End()
	logger.Info("start usage calculation")
	go usageScheduler.CalculateUsageLoop(ctx)
}

func (usageScheduler *UsageScheduler) StopUsageCalculation() {
	if usageChan != nil {
		close(usageChan)
		usageChan = nil
	}
}

func (usageScheduler *UsageScheduler) CalculateUsageLoop(ctx context.Context) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageScheduler.CalculateUsageLoop").Start()
	defer span.End()
	logger.Info("calculating usages")
	for {
		usageScheduler.usageController.CalculateUsages(ctx)
		select {
		case <-usageChan:
			return
		case tm := <-usageScheduler.syncTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}
