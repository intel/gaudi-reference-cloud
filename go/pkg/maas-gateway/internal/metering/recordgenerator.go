// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package metering

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RecordGenerator struct {
	logger            logr.Logger
	usageRecordClient pb.UsageRecordServiceClient
	config            *config.Config
}

func NewRecordGenerator(log logr.Logger, config *config.Config, usageRecordClient pb.UsageRecordServiceClient) *RecordGenerator {
	return &RecordGenerator{
		logger:            log.WithName("RecordGenerator"),
		usageRecordClient: usageRecordClient,
		config:            config,
	}
}

// CreateUsageRecord for every successful request we send usage record to usage record service so later
// these request will be billed
func (r *RecordGenerator) CreateUsageRecord(ctx context.Context, requestId, cloudAccountId, productName string, quantity float64, startTime, endTime *timestamppb.Timestamp) error {
	logger := r.logger.WithValues("requestId", requestId).WithValues("cloudAccountId", cloudAccountId)

	usageRecord := &pb.ProductUsageRecordCreate{
		TransactionId:  requestId,
		CloudAccountId: cloudAccountId,
		Region:         r.config.Region,
		Timestamp:      timestamppb.Now(),
		StartTime:      startTime,
		EndTime:        endTime,
		ProductName:    &productName,
		Quantity:       quantity,
		Properties: map[string]string{
			"serviceType":    "ModelAsAService",
			"processingType": "text",
			"modelType":      productName,
		},
	}

	err := retry.Do(func() error {
		_, err := r.usageRecordClient.CreateProductUsageRecord(ctx, usageRecord)
		if err != nil {
			logger.Error(err, "couldn't create usage record")
			return fmt.Errorf("couldn't create usage record")
		}
		return nil
	}, retry.Context(ctx), retry.Attempts(r.config.RetryAttempts))
	if err != nil {
		return fmt.Errorf("failed to create usage record after maximum retries")
	}

	logger.Info("Created usage record", "usageRecord", usageRecord)
	return nil
}
