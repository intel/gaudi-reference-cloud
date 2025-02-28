// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestSchedulerOps(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestSchedulerOps")
	logger.Info("BEGIN")
	defer logger.Info("End")
	client := pb.NewBillingOpsActionServiceClient(clientConn)
	client.Create(ctx, &pb.SchedulerAction{Action: pb.SchedulerActionType_START_CREDIT_EXPIRY_SCHEDULER})
}
