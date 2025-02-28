// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var syncChan = make(chan int, 1)

type AriaBillingProductCatalogSyncService struct {
	pb.UnimplementedBillingProductCatalogSyncServiceServer
}

// Used for testing
var SyncWait = atomic.Pointer[sync.WaitGroup]{}

func StartTestSync() {

}

func (svc *AriaBillingProductCatalogSyncService) Sync(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Sync").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	select {
	case syncChan <- 1:
		logger.Info("scheduled product sync")
	default:
		logger.Info("sync already scheduled")
	}

	return &emptypb.Empty{}, nil
}

type Scheduler struct {
	syncTicker        *time.Ticker
	productController *ProductController
}

func NewScheduler(productController *ProductController) *Scheduler {
	return &Scheduler{
		syncTicker:        time.NewTicker(time.Duration(config.Cfg.SyncInterval) * time.Second),
		productController: productController,
	}
}

func (s *Scheduler) StartSync(ctx context.Context) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("StartSync").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("start product scheduled product sync")
	go s.SyncLoop(ctx)
}

func (s *Scheduler) StopSync() {
	close(syncChan)
	syncChan = nil
}

func (s *Scheduler) SyncLoop(ctx context.Context) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("SyncLoop").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("syncLoop product sync")
	for {
		s.SyncProductCatalog(ctx, &logger)
		select {
		case ii := <-syncChan:
			if ii == 0 {
				return
			}
		case tm := <-s.syncTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

func (s *Scheduler) SyncProductCatalog(ctx context.Context, logger *logr.Logger) {
	logger.Info("SyncProductCatalog ..")
	// TODO: synchronize product catalog with aria
	//
	// Use cfg.ClientIdPrefix for client ids configured in Aria.
	// cfg.ClientIdPrefix is "idc" for CI/CD deployments
	//
	// For kind deployments, cfg.ClientIdPrefix is the username
	// of the user running the deployment.
	//
	// When running "go test", cfg.ClientIdPrefix is "test". For
	// "go test" we will need to have some sort of mock of the
	// aria service. We may want to consider being able to deploy
	// the mock in kind because not everyone on the team has
	// credentials for connecting to Aria and it would be good
	// to have some level of premium/enterprise testing available
	// without Aria. Engineers who do have credentials for Aria will
	// want to run against the actual Aria "IDC Dev" environment
	// using their username (IDSID).
	//
	// For testing outside of kind, use your username, which should
	// be your Intel IDSID

	if err := s.productController.SyncProducts(ctx); err != nil {
		logger.Error(err, "error in sync of aria plan")
	}
	wg := SyncWait.Load()
	if wg != nil {
		wg.Done()
	}
	logger.Info("syncing product catalog with aria", "prefix", config.Cfg.ClientIdPrefix)
}
