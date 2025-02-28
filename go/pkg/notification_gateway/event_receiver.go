// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/snsutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sqsutil"
)

type EventReceiver struct {
	snsUtil *snsutil.SNSUtil
	sqsUtil *sqsutil.SQSUtil
}

func NewEventReceiver(snsUtil *snsutil.SNSUtil, sqsUtil *sqsutil.SQSUtil) *EventReceiver {
	return &EventReceiver{snsUtil: snsUtil, sqsUtil: sqsUtil}
}

func (eventReceiver *EventReceiver) Receive(ctx context.Context, queueURL string, timeout int32, maxNumberOfMessages int32, messageAttribute string) ([]*pb.MessageResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventReceiver.Receive").WithValues("messageAttribute", messageAttribute).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	receivedMessages, err := eventReceiver.sqsUtil.ReceiveMessageWithOptions(ctx, queueURL, timeout, maxNumberOfMessages, messageAttribute)
	msges := []*pb.MessageResponse{}
	if len(receivedMessages) > 0 {
		for _, receivedMessage := range receivedMessages {
			messageId, body, attributes, err := eventReceiver.parseMessage(ctx, receivedMessage)
			if err != nil {
				logger.Error(err, "error parsing message", "messageId", messageId)
				continue
			}
			if messageId == "" || body == "" {
				continue
			}
			msgAttribues := eventReceiver.getAttributes(attributes)
			logger.V(1).Info("mapped message atttribues", "msgAttribues", msgAttribues)
			msg := &pb.MessageResponse{MessageId: messageId, Body: body, Attributes: msgAttribues, ReceiptHandle: *receivedMessage.ReceiptHandle}
			msges = append(msges, msg)
		}
	}
	return msges, err
}

func (eventReceiver *EventReceiver) Delete(ctx context.Context, queueURL string, messageId string, reciptHandle string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventReceiver.Delete").WithValues("messageId", messageId).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	err := eventReceiver.sqsUtil.DeleteMessageWithReceiptHandle(ctx, queueURL, reciptHandle)
	return err
}

func (eventReceiver *EventReceiver) parseMessage(ctx context.Context, message types.Message) (string, string, map[string]types.MessageAttributeValue, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventReceiver.parseMessageBody").Start()
	defer span.End()
	messageId := *message.MessageId
	messageBody := *message.Body
	messageAttribute := message.MessageAttributes
	logger.V(1).Info("received message id", "message id", messageId)
	logger.V(1).Info("received message body", "message body", messageBody)
	logger.V(1).Info("received message attribue", "message attribue", messageAttribute)
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(*message.Body), &data); err != nil {
		logger.Error(err, "error parsing message body")
		return messageId, messageBody, messageAttribute, err
	}
	logger.V(1).Info("parsed message data", "data", data)
	return messageId, messageBody, messageAttribute, nil
}

func (eventReceiver *EventReceiver) getAttributes(messageAttributes map[string]types.MessageAttributeValue) map[string]string {
	messageAttributesMap := make(map[string]string)
	for key, _ := range messageAttributes {
		messageTypeAttr, ok := messageAttributes[key]
		if !ok || messageTypeAttr.StringValue == nil {
			continue
		}
		messageAttributesMap[key] = *messageTypeAttr.StringValue
	}
	return messageAttributesMap
}

func (eventReceiver *EventReceiver) Subscribe(ctx context.Context, queueUrl string, topicARN string) (string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventReceiver.Subscribe").WithValues("queueUrl", queueUrl).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	subscriptionArn, err := eventReceiver.snsUtil.Subscribe(ctx, queueUrl, topicARN)
	if err != nil {
		logger.Error(err, "error subscribe event")
	}
	return subscriptionArn, err
}
