// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

import (
	"context"
	"sync"
	"sync/atomic"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type StandardBillingProductCatalogSyncService struct {
	pb.UnimplementedBillingProductCatalogSyncServiceServer
}

// Used for testing
var SyncWait = atomic.Pointer[sync.WaitGroup]{}

func (svc *StandardBillingProductCatalogSyncService) Sync(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("StandardBillingProductCatalogSyncService.Sync").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	wg := SyncWait.Load()
	if wg != nil {
		wg.Done()
	}
	// Standard driver does no syncing. Successful NOOP
	return &emptypb.Empty{}, nil
}
