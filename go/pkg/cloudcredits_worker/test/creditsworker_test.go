// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	worker "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudcredits_worker"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	// cc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"gotest.tools/v3/assert"
)

type CloudAcccountClientMock struct {
}

func (m *CloudAcccountClientMock) GetById(ctx context.Context, in *pb.CloudAccountId, opts ...grpc.CallOption) (*pb.CloudAccount, error) {
	return &pb.CloudAccount{}, nil
}

func (m *CloudAcccountClientMock) Update(ctx context.Context, in *pb.CloudAccountUpdate, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func TestValidateAndParseMessage(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestValidateAndParseMessage")
	logger.Info("TestValidateAndParseMessage starts")
	defer logger.Info("TestValidateAndParseMessage ends")
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)
	standardUser := "standard_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  standardUser,
		Owner: standardUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	})
	workerIns := &worker.CloudCreditsWorker{CloudAccountClient: TestCloudAccountSvcClient.CloudAccountClient, BillingCloudAccountClient: TestCloudAccountSvcClient}
	attributes := map[string]string{"Timestamp": "2024-08-08T15:04:05Z"}
	body := "{\"CloudAccountId\": \"" + cloudAcct.Id + "\"}"
	msg := pb.MessageResponse{Body: body, MessageId: "123", Attributes: attributes}
	event, err := workerIns.ValidateAndParseMessage(ctx, &msg)
	if err != nil {
		logger.Error(err, "Error in ValidateAndParseMessage")
	}
	assert.Equal(t, event.CloudAccountId, cloudAcct.Id)
}

func TestOnNotifyCloudCredits(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestOnNotifyCloudCredits")
	logger.Info("TestOnNotifyCloudCredits starts")
	defer logger.Info("TestOnNotifyCloudCredits ends")
	standardUser := "standard2_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  standardUser,
		Owner: standardUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	})
	workerIns := &worker.CloudCreditsWorker{CloudAccountClient: TestCloudAccountSvcClient.CloudAccountClient, BillingCloudAccountClient: TestCloudAccountSvcClient, NotificationGatewayClient: TestNotificationGatewayClient}
	msg := events.CreateEvent{Type: "operation", EventSubType: "CLOUD_CREDITS_AVAILABLE", CloudAccountId: cloudAcct.Id}
	msgId := "123"
	rhandle := "123"
	err := workerIns.OnNotifyCloudCredits(ctx, &msg, msgId, rhandle, timestamppb.Now())
	if err != nil {
		logger.Error(err, "Error in OnNotifyCloudCredits")
	}
}

func CreateAndGetAccount(t *testing.T, acctCreate *pb.CloudAccountCreate) *pb.CloudAccount {
	ctx := context.Background()
	id, err := TestCloudAccountSvcClient.CloudAccountClient.Create(ctx, acctCreate)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	acctOut, err := TestCloudAccountSvcClient.CloudAccountClient.GetById(context.Background(), &pb.CloudAccountId{Id: id.GetId()})
	if err != nil {
		t.Fatalf("read account: %v", err)
	}
	return acctOut
}
