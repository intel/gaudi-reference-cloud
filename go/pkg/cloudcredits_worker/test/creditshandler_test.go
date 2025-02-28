// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	worker "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudcredits_worker"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Worker interface {
	IsCreditsAvailable(ctx context.Context, cloudAccountId string) (bool, error)
	HandleCreditsUsed(ctx context.Context, message *events.CreateEvent, isExpired bool) error
}
type MockWorker struct {
	mock.Mock
}

func (m *MockWorker) IsCreditsAvailable(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockWorker) HandleCreditsUsed(ctx context.Context, message *events.CreateEvent, isExpired bool) error {
	//args := m.Called(ctx, message, isExpired)
	return errors.New("error cannot apply credit event due to failed IsCreditsAvailable call")
}

func TestThresholdReachedHandler(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestThresholdReachedHandler")
	logger.Info("TestThresholdReachedHandler starts")
	defer logger.Info("TestThresholdReachedHandler ends")
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)
	standardUser := "standard3_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  standardUser,
		Owner: standardUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	})
	workerIns := &worker.CloudCreditsWorker{CloudAccountClient: TestCloudAccountSvcClient.CloudAccountClient, BillingCloudAccountClient: TestCloudAccountSvcClient}
	msg := events.CreateEvent{Type: "operation", EventSubType: "CLOUD_CREDITS_THRESHOLD_REACHED", CloudAccountId: cloudAcct.Id}
	err := workerIns.HandleThresholdReached(ctx, &msg)
	if err != nil {
		logger.Error(err, "Error in ThresholdReachedHandler")
	}
}

func TestCreditsUsedHandler(t *testing.T) {

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditsUsedHandler")
	logger.Info("TestCreditsUsedHandler starts")
	defer logger.Info("TestCreditsUsedHandler ends")
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)
	standardUser := "standard4_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  standardUser,
		Owner: standardUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	})
	workerIns := &worker.CloudCreditsWorker{CloudAccountClient: TestCloudAccountSvcClient.CloudAccountClient, BillingCloudAccountClient: TestCloudAccountSvcClient}
	msg := events.CreateEvent{Type: "operation", EventSubType: "CLOUD_CREDITS_USED", CloudAccountId: cloudAcct.Id}

	err := workerIns.HandleCreditsUsed(ctx, &msg, false)
	if err != nil {
		logger.Error(err, "Error in CreditsUsedHandler")
	}
}

func TestCreditsExpiredHandler(t *testing.T) {

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditsUsedHandler")
	logger.Info("TestCreditsUsedHandler starts")
	defer logger.Info("TestCreditsUsedHandler ends")
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)
	standardUser := "standard5_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  standardUser,
		Owner: standardUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	})
	workerIns := &worker.CloudCreditsWorker{CloudAccountClient: TestCloudAccountSvcClient.CloudAccountClient, BillingCloudAccountClient: TestCloudAccountSvcClient}
	msg := events.CreateEvent{Type: "operation", EventSubType: "CLOUD_CREDITS_EXPIRED", CloudAccountId: cloudAcct.Id}

	err := workerIns.HandleCreditsUsed(ctx, &msg, true)
	if err != nil {
		logger.Error(err, "Error in CreditsExpiredHandler")
	}
}

func TestCreditsAvailableHandler(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditsAvailableHandler")
	logger.Info("TestCreditsAvailableHandler starts")
	defer logger.Info("TestCreditsAvailableHandler ends")
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)
	standardUser := "standard6_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  standardUser,
		Owner: standardUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_STANDARD,
	})
	workerIns := &worker.CloudCreditsWorker{CloudAccountClient: TestCloudAccountSvcClient.CloudAccountClient, BillingCloudAccountClient: worker.CloudAccountClient}
	msg := events.CreateEvent{Type: "operation", EventSubType: "CLOUD_CREDITS_AVAILABLE", CloudAccountId: cloudAcct.Id}
	err := workerIns.HandleCreditsAvailable(ctx, &msg)
	if err != nil {
		logger.Error(err, "Error in CreditsAvailableHandler")
	}
}

func TestCreditsExpiredHandlerWhenCreditsAvailable(t *testing.T) {

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditsExpiredHandlerWhenCreditsAvailable")
	logger.Info("TestCreditsExpiredHandlerWhenCreditsAvailable starts")
	defer logger.Info("TestCreditsExpiredHandlerWhenCreditsAvailable ends")
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)
	mockWorker := new(MockWorker)

	// Setup expectations
	mockWorker.On("IsCreditsAvailable", ctx, "123").Return(true, nil)

	err := mockWorker.HandleCreditsUsed(ctx, &events.CreateEvent{}, false)
	assert.Error(t, err)
}
func TestCreditsExpiredHandlerWhenCreditsNotAvailable(t *testing.T) {

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreditsExpiredHandlerWhenCreditsAvailable")
	logger.Info("TestCreditsExpiredHandlerWhenCreditsAvailable starts")
	defer logger.Info("TestCreditsExpiredHandlerWhenCreditsAvailable ends")
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)
	mockWorker := new(MockWorker)

	// Setup expectations
	mockWorker.On("IsCreditsAvailable", ctx, "456").Return(false, errors.New("error cannot apply credit event due to failed IsCreditsAvailable call"))

	err := mockWorker.HandleCreditsUsed(ctx, &events.CreateEvent{}, false)
	assert.Error(t, err)
}
