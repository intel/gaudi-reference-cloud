// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing_driver_intel

import (
	"context"
	"sync"
	"sync/atomic"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Used for testing
var SyncWait = atomic.Pointer[sync.WaitGroup]{}

type IntelBillingProductCatalogSyncService struct {
	pb.UnimplementedBillingProductCatalogSyncServiceServer
}

func (svc *IntelBillingProductCatalogSyncService) Sync(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("IntelBillingProductCatalogSyncService.Sync").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	wg := SyncWait.Load()
	if wg != nil {
		wg.Done()
	}
	// Intel driver does no syncing. Successful NOOP
	return &emptypb.Empty{}, nil
}
