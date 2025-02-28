// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package event

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestMain(m *testing.M) {
	log.SetDefaultLogger()
	ctx := context.Background()
	EmbedService(ctx)
	grpcutil.StartTestServices(ctx)
	defer grpcutil.StopTestServices()
	m.Run()
}

func TestCreateError(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreateError")
	logger.Info("BEGIN")
	defer logger.Info("End")
	notificationClient := pb.NewNotificationGatewayServiceClient(test.clientConn)

	eventSeverity := pb.EventSeverity_LOW
	serviceName := pb.ServiceName_BILLING
	message := "this is a test message"
	cloudAcctId := "1234"
	userId := uuid.NewString()
	region := "us-west-2"
	properties := map[string]string{
		"key": "value",
	}

	_, err := notificationClient.Create(ctx, &pb.CreateEvent{
		Status:         pb.EventStatus_ACTIVE,
		Type:           pb.EventType_ERROR,
		Severity:       &eventSeverity,
		ServiceName:    &serviceName,
		Message:        &message,
		CloudAccountId: &cloudAcctId,
		UserId:         &userId,
		EventSubType:   "",
		Properties:     properties,
		ClientRecordId: uuid.NewString(),
		Region:         &region,
	})

	if err != nil {
		t.Fatalf("failed to create error: %v", err)
	}
}
