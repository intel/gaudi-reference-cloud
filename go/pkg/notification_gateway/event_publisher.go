// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/snsutil"
)

type EventPublisher struct {
	snsUtil *snsutil.SNSUtil
}

func NewEventPublisher(snsUtil *snsutil.SNSUtil) *EventPublisher {
	return &EventPublisher{snsUtil: snsUtil}
}

func (eventPublisher *EventPublisher) Publish(ctx context.Context, cloudCreditUsageMessage string, topicARN string, messageAttributeName string) (string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventPublisher.PublishCloudCreditUsage").WithValues("topicARN", topicARN).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	msgId, err := eventPublisher.snsUtil.PublishMessageWithOptions(ctx, cloudCreditUsageMessage, topicARN, messageAttributeName)
	return msgId, err
}
