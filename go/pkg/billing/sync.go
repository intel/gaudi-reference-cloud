// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type BillingProductCatalogSyncService struct {
	pb.UnimplementedBillingProductCatalogSyncServiceServer
}

func (svc *BillingProductCatalogSyncService) Sync(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	// syncing the product catalog is a fire and forget operation.
	// The caller does not need to handle any errors. The sync itself
	// logs any errors, which lead to alerts so the IDC team can
	// respond.
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BillingProductCatalogSyncService.Sync").Start()
	defer span.End()
	logger.Info("billing product catalog sync ")
	for driver, conn := range driverConnections {
		driver := driver
		conn := conn
		go func() {
			client := pb.NewBillingProductCatalogSyncServiceClient(conn)
			_, err := client.Sync(context.Background(), &emptypb.Empty{})
			if err != nil {
				logger := log.FromContext(ctx)
				logger.Error(err, "product catalog sync failed", "driver", driver)
			}
		}()
	}
	return &emptypb.Empty{}, nil
}
