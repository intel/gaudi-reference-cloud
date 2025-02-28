// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type NotificationGatewayClientInterface interface {
	PublishEvent(ctx context.Context, req *pb.PublishEventRequest) (*pb.PublishEventResponse, error)
	SubscribeEvents(ctx context.Context, req *pb.SubscribeEventRequest) (*pb.SubscribeEventResponse, error)
	ReceiveEvents(ctx context.Context, req *pb.ReceiveEventRequest) (*pb.ReceiveEventResponse, error)
}
