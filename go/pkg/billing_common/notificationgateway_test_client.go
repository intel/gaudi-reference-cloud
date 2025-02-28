// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type NotificationGatewayTestClient struct {
}

func NewNotificationGatewayTestClient(ctx context.Context) (*NotificationGatewayTestClient, error) {
	return &NotificationGatewayTestClient{}, nil
}

func (notificationGatewayTestClient *NotificationGatewayTestClient) PublishEvent(ctx context.Context, req *pb.PublishEventRequest) (*pb.PublishEventResponse, error) {
	logger := log.FromContext(ctx).WithName("NotificationGatewayTestClient.PublishEvent")
	logger.Info("publish event")

	return &pb.PublishEventResponse{}, nil
}

func (notificationGatewayTestClient *NotificationGatewayTestClient) SubscribeEvents(ctx context.Context, req *pb.SubscribeEventRequest) (*pb.SubscribeEventResponse, error) {
	logger := log.FromContext(ctx).WithName("NotificationGatewayTestClient.Subscribe")
	logger.Info("Subscribe events")

	return &pb.SubscribeEventResponse{}, nil
}

func (notificationGatewayTestClient *NotificationGatewayTestClient) ReceiveEvents(ctx context.Context, req *pb.ReceiveEventRequest) (*pb.ReceiveEventResponse, error) {
	logger := log.FromContext(ctx).WithName("NotificationGatewayClient.ReceiveEvents")
	logger.Info("Receive events")

	return &pb.ReceiveEventResponse{}, nil
}
