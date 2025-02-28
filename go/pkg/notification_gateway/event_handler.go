// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/encoding/protojson"
)

type EventHandler struct {
	eventData      *EventData
	eventPublisher *EventPublisher
	eventReceiver  *EventReceiver
}

type Message struct {
	Default string `json:"default"`
}

type DeleteMessageError struct {
	MessageId    string `json:"messageId"`
	MessageError string `json:"messageError"`
}

func NewEventHandler(eventData *EventData, eventPublisher *EventPublisher, eventReceiver *EventReceiver) *EventHandler {
	return &EventHandler{
		eventData:      eventData,
		eventPublisher: eventPublisher,
		eventReceiver:  eventReceiver,
	}
}

func (eventHandler *EventHandler) HandlePublishEvent(ctx context.Context, cloudAccountId string, event *pb.CreateEvent, topicName string) (string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventHandler.HandlePublishEvent").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	topicARN := config.Cfg.GetAWSSNSTopicArn(topicName)
	logger.V(1).Info("topic ARN", "topicARN", topicARN)
	messageId, err := eventHandler.ProcessPublishEvent(ctx, cloudAccountId, event, topicARN)
	return messageId, err
}

func (eventHandler *EventHandler) ProcessPublishEvent(ctx context.Context, cloudAccountId string, event *pb.CreateEvent, topicARN string) (string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventHandler.ProcessPublishEvent").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	err := eventHandler.Create(ctx, cloudAccountId, event)
	if err != nil {
		return "", err
	}
	eventMesage, err := getEventMessage(ctx, event)
	if err != nil {
		return "", err
	}
	return eventHandler.eventPublisher.Publish(ctx, eventMesage, topicARN, event.EventName)
}

func getEventMessage(ctx context.Context, event *pb.CreateEvent) (string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventHandler.getEventMessage").Start()
	defer span.End()
	jsonData, err := protojson.Marshal(event)
	if err != nil {
		logger.Error(err, "failed to marshal message to json")
		return "", err
	}
	eventMessage, err := addDefaultToEventMessage(ctx, jsonData)
	if err != nil {
		logger.Error(err, "failed to add default to event message")
		return "", err
	}
	logger.V(1).Info("event message", "eventMessage", eventMessage)
	return eventMessage, nil
}

func addDefaultToEventMessage(ctx context.Context, jsonData []byte) (string, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventHandler.AddDefaultToEventMessage").Start()
	defer span.End()

	messageData := string(jsonData)
	message := Message{
		Default: string(messageData),
	}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.Error(err, "failed to marshal message default")
		return "", err
	}
	eventMessage := string(messageBytes)
	return string(eventMessage), nil
}

func (eh *EventHandler) Create(ctx context.Context, cloudAccountId string, createEvent *pb.CreateEvent) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("EventHandler.create").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	eventId, err := NewId()

	if err != nil {
		logger.Error(err, "failed to generate id")
	}

	eventBase := EventBase{
		Id:             eventId,
		CloudAccountId: cloudAccountId,
		UserId:         createEvent.GetUserId(),
		Creation:       time.Now(),
		Expiration:     time.Now().AddDate(defaultExpirationOfEventYears, 1, defaultExpirationOfEventDays),
		Status:         createEvent.GetStatus().String(),
		Severity:       createEvent.GetSeverity().String(),
		Properties:     createEvent.Properties,
		ClientRecordId: createEvent.ClientRecordId,
	}
	cloudCreditsEvent := CloudCreditsEvent{
		CreditEvent: eventBase,
		Message:     createEvent.GetMessage(),
		ServiceName: createEvent.GetServiceName().String(),
	}
	if err = eh.eventData.storeCloudCreditsEvent(ctx, cloudCreditsEvent); err != nil {
		logger.Error(err, "failed to store cloud credits event")
		return err
	}
	return nil

}

func (eventHandler *EventHandler) HandleReceiveEvents(ctx context.Context, queueName string, timeout int32, maxNumberOfMessages int32, messageAttribute string) ([]*pb.MessageResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventHandler.HandleReceiveEvents").WithValues("messageAttribute", messageAttribute).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	queueUrl := config.Cfg.GetAWSSQSQueueUrl(queueName)
	messages, err := eventHandler.ProcessReciveEvents(ctx, queueUrl, timeout, maxNumberOfMessages, messageAttribute)
	return messages, err
}

func (eventHandler *EventHandler) HandleDeleteEvents(ctx context.Context, queueName string, deleteEventRequest []*pb.DeleteEventRequest) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventHandler.HandleDeleteEvents").WithValues("queueName", queueName).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	queueUrl := config.Cfg.GetAWSSQSQueueUrl(queueName)
	err := eventHandler.ProcessDeleteEvents(ctx, queueUrl, deleteEventRequest)
	if err != nil {
		logger.Error(err, "error in process delete event")
	}
	return err
}

func (eventHandler *EventHandler) HandleSubscribeEvents(ctx context.Context, topicName string, queueName string) (string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventHandler.HandleSubscribeEvents").WithValues("queueName", queueName).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	topicARN := config.Cfg.GetAWSSNSTopicArn(topicName)
	queueUrl := config.Cfg.GetAWSSQSQueueUrl(queueName)
	subscriptionArn, err := eventHandler.eventReceiver.Subscribe(ctx, queueUrl, topicARN)
	return subscriptionArn, err
}

func (eventHandler *EventHandler) ProcessReciveEvents(ctx context.Context, queueURL string, timeout int32, maxNumberOfMessages int32, messageAttribute string) ([]*pb.MessageResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventHandler.ProcessReciveEvents").WithValues("messageAttribute", messageAttribute).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	messages, err := eventHandler.eventReceiver.Receive(ctx, queueURL, timeout, maxNumberOfMessages, messageAttribute)
	if err != nil {
		logger.Error(err, "error in process receive event")
	}
	return messages, err
}

func (eventHandler *EventHandler) ProcessDeleteEvents(ctx context.Context, queueURL string, deleteEventRequests []*pb.DeleteEventRequest) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventHandler.ProcessDeleteEvents").WithValues("queueURL", queueURL).Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	var deleteMessageErrors []DeleteMessageError
	for _, deleteEventRequest := range deleteEventRequests {
		logger.Info("deleting message", "messageId", deleteEventRequest.MessageId)
		err := eventHandler.eventReceiver.Delete(ctx, queueURL, deleteEventRequest.MessageId, deleteEventRequest.ReceiptHandle)
		if err != nil {
			logger.Error(err, "error deleting message", "messageId", deleteEventRequest.MessageId)
			deleteMessageErrors = append(deleteMessageErrors, DeleteMessageError{MessageId: deleteEventRequest.MessageId, MessageError: err.Error()})
		}
	}
	var errRet error
	if len(deleteMessageErrors) > 0 {
		errData, err := json.Marshal(deleteMessageErrors)
		if err != nil {
			logger.Error(err, "error in marshaling delete event error message")
		}
		errRet = fmt.Errorf("error in process delete event %v", errData)
	}
	return errRet
}
