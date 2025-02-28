// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package snsutil

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	tsqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	tsqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sqsutil"
)

type MockSNSClient struct {
}

func (m *MockSNSClient) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	messageId := "1234567890"
	return &sns.PublishOutput{MessageId: &messageId}, nil
}

func (m *MockSNSClient) Subscribe(ctx context.Context, params *sns.SubscribeInput, optFns ...func(*sns.Options)) (*sns.SubscribeOutput, error) {
	return &sns.SubscribeOutput{SubscriptionArn: aws.String("testsubscription")}, nil
}

type MockSQSClient struct {
	message string
}

func (m *MockSQSClient) ReceiveMessage(ctx context.Context, params *tsqs.ReceiveMessageInput, optFns ...func(*tsqs.Options)) (*tsqs.ReceiveMessageOutput, error) {
	return &tsqs.ReceiveMessageOutput{Messages: []tsqstypes.Message{{
		Body: &m.message,
	}}}, nil
}

func TestInit(t *testing.T) {
	snsUtil := &SNSUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("ses.TestSNSInit")
	logger.Info("TestSNSInit BEGIN")
	defer logger.Info("TestSNSInit END")

	snsTopicARN := "arn:aws:sns:us-west-2:00000000000:snstopic"
	region := "us-west-2"
	credentialsFile := ""
	if err := snsUtil.Init(ctx, region, snsTopicARN, credentialsFile); err != nil {
		logger.Error(err, "Unable to init the sns")
	}
}

func TestSNSPublishMessage(t *testing.T) {
	snsUtil := &SNSUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("ses.TestSNSPublishMessage")
	logger.Info("TestSNSPublishMessage BEGIN")
	defer logger.Info("TestSNSPublishMessage END")

	snsTopicARN := "arn:aws:sns:us-west-2:00000000000:snstopic"
	region := "us-west-2"
	credentialsFile := ""
	if err := snsUtil.Init(ctx, region, snsTopicARN, credentialsFile); err != nil {
		logger.Error(err, "Unable to init the sns")
	}
	mockSNS := &MockSNSClient{}
	snsUtil.snsClient = mockSNS
	if _, err := snsUtil.PublishMessage(ctx, "Test Message"); err != nil {
		logger.Error(err, "Unable to publish message")
	}

	messageAttributes := map[string]types.MessageAttributeValue{
		"String": {
			DataType:    aws.String("String"),
			StringValue: aws.String("Yes"),
		},
	}
	optionsParams := &PublishMessageOptionalParams{
		topicArn:          "arn:aws:sns:us-west-2:00000000000:newsnstopic",
		messageStructure:  "json",
		messageAttributes: messageAttributes,
	}

	if _, err := snsUtil.PublishMessageWithOpts(ctx, "Test Message", optionsParams); err != nil {
		logger.Error(err, "Unable to publish message")
	}
}

func TestSubscribeTopic(t *testing.T) {
	snsUtil := &SNSUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("ses.TestSNSSubscribe")
	logger.Info("TestSNSSubscribe BEGIN")
	defer logger.Info("TestSNSSubscribe END")

	snsTopicARN := "arn:aws:sns:us-west-2:00000000000:snstopic"
	region := "us-west-2"
	credentialsFile := ""
	if err := snsUtil.Init(ctx, region, snsTopicARN, credentialsFile); err != nil {
		logger.Error(err, "Unable to init the sns")
	}
	mockSNS := &MockSNSClient{}
	snsUtil.snsClient = mockSNS
	sqsARN := "arn:aws:sqs:us-west-2:00000000000:sqsqueue"
	sqsURL := "https://sqs.region.amazonaws.com/accountid/sqsqueue"
	if _, err := snsUtil.Subscribe(ctx, sqsURL, sqsARN); err != nil {
		logger.Error(err, "Unable to Subscribe")
	}
}

func TestSNSPublishAndSQSReceive(t *testing.T) {
	//Integration test

	snsUtil := &SNSUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("ses.TestSNSPublishAndSQSReceive")
	logger.Info("TestSNSPublishAndSQSReceive BEGIN")
	defer logger.Info("TestSNSPublishAndSQSReceive END")

	snsTopicARN := "arn:aws:sns:us-west-2:000000000000:idc-staging-notifications-topic"
	region := "us-west-2"
	credentialsFile := ""
	if err := snsUtil.Init(ctx, region, snsTopicARN, credentialsFile); err != nil {
		logger.Error(err, "Unable to init the sns")
	}

	mockSNS := &MockSNSClient{}
	snsUtil.snsClient = mockSNS

	sqsARN := "arn:aws:sqs:us-west-2:00000000000:sqsqueue"
	sqsURL := "https://sqs.region.amazonaws.com/accountid/sqsqueue"
	if _, err := snsUtil.Subscribe(ctx, sqsURL, sqsARN); err != nil {
		logger.Error(err, "Unable to Subscribe")
	}

	message := "Test Message"
	if _, err := snsUtil.PublishMessage(ctx, message); err != nil {
		logger.Error(err, "Unable to publish message")
	}

	logger.Info("Listening to sqs queue")
	sqsUtil := sqsutil.SQSUtil{}

	awsQueueURL := "https://sqs.us-west-2.amazonaws.com/000000000000/idc-staging-notifications-queue"
	credentialsFile = ""
	if err := sqsUtil.Init(ctx, region, awsQueueURL, credentialsFile); err != nil {
		logger.Error(err, "couldn't init sqsProcessor")
		return
	}

	logger.Info("Going to call ReceiveMessage")
	messageAttributeNames := ""
	if messages, err := sqsUtil.ReceiveMessage(ctx, 10, 1, messageAttributeNames); err != nil {
		logger.Error(err, "couldn't receive message from queue")
		return
	} else {
		logger.Info("No of Message", "No of Message", len(messages))
		for _, message := range messages {
			logger.Info("Message", "Going to delete Message", *message.Body)
			if err := sqsUtil.DeleteMessage(ctx, message); err != nil {
				logger.Error(err, "couldn't delete message from the queue")
				return
			}
		}
	}
}

// Integration test
func TestPublishAndSQSReceive(t *testing.T) {
	snsUtil := &SNSUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("ses.TestPublishAndSQSReceive")
	logger.Info("TestPublishAndSQSReceive BEGIN")
	defer logger.Info("TestPublishAndSQSReceive END")

	snsTopicARN := os.Getenv("AWS_SNS_TOPIC_ARN")
	snsQueueURL := os.Getenv("AWS_SNS_QUEUE_URL")
	snsQueueARN := os.Getenv("AWS_SNS_QUEUE_ARN")

	if len(snsTopicARN) == 0 || len(snsQueueURL) == 0 || len(snsQueueARN) == 0 {
		t.Skip("AWS envionment variable is not set")
	}
	logger.Info("public topic arn", "snsTopicARN", snsTopicARN)
	logger.Info("subscribe queue url", "snsQueueURL", snsQueueURL)
	region := "us-west-2"
	credentialsFile := ""
	if err := snsUtil.Init(ctx, region, snsTopicARN, credentialsFile); err != nil {
		logger.Error(err, "Unable to init the sns")
	}

	message := GetTestEventMessage(ctx)
	logger.Info("test message", "message", message)
	messageAttributes := map[string]types.MessageAttributeValue{
		"eventName": {
			DataType:    aws.String("String"),
			StringValue: aws.String("CloudCreditsExpiryTest"),
		},
	}
	optionsParams := &PublishMessageOptionalParams{
		topicArn:          snsTopicARN,
		messageStructure:  "json",
		messageAttributes: messageAttributes,
	}
	if _, err := snsUtil.PublishMessageWithOpts(ctx, message, optionsParams); err != nil {
		logger.Error(err, "Unable to publish message")
	}

	logger.Info("Listening to sqs queue")

	logger.Info("subscribe queue url", "snsQueueARN", snsQueueARN)
	if _, err := snsUtil.Subscribe(ctx, snsQueueARN, snsTopicARN); err != nil {
		logger.Error(err, "Unable to Subscribe")
	}

	awsQueueURL := os.Getenv("AWS_SNS_QUEUE_URL")
	logger.Info("receive queue url", "awsQueueURL", awsQueueURL)
	sqsUtil := sqsutil.SQSUtil{}
	if err := sqsUtil.Init(ctx, region, awsQueueURL, credentialsFile); err != nil {
		logger.Error(err, "couldn't init sqsProcessor")
		return
	}
	messageAttributeName := "CloudCreditsExpiryTest"
	if messages, err := sqsUtil.ReceiveMessageWithOptions(ctx, awsQueueURL, 10, 10, messageAttributeName); err != nil {
		logger.Error(err, "couldn't receive message from queue")
		return
	} else {
		logger.Info("No of Message", "No of Message", len(messages))
		for _, message := range messages {
			logger.Info("Message", "Going to delete Message", *message.Body)
			if err := sqsUtil.DeleteMessageWithReceiptHandle(ctx, awsQueueURL, *message.ReceiptHandle); err != nil {
				logger.Error(err, "couldn't delete message from the queue")
				return
			}
		}
	}

}

func GetTestEventMessage(ctx context.Context) string {
	logger := log.FromContext(ctx).WithName("ses.GetTestEventMessage")
	messageData := map[string]interface{}{
		"cloudAccountId": "498293553112",
		"topicName":      "idc-staging-cloud-credits-topic",
		"subject":        "CloudCreditsExpiryTest",
		"status":         "ACTIVE",
		"type":           "NOTIFICATION",
		"severity":       "MEDIUM",
		"serviceName":    "CREDIT",
		"eventState":     "New",
		"userId":         "testuserid",
		"eventSubType":   "CloudCreditsExpiryTest",
		"clientRecordId": "94d6bb67-e741-4336-8510-aadb03a8d026",
		"eventName":      "CloudCreditsExpiryTest",
		"message":        "abc",
	}
	messageJSON, err := json.Marshal(messageData)
	if err != nil {
		logger.Error(err, "failed to marshal message")
	}

	testMessage := string(messageJSON)

	type Message struct {
		Default string `json:"default"`
	}
	message := Message{
		Default: string(testMessage),
	}
	messageBytes, _ := json.Marshal(message)
	messageStr := string(messageBytes)
	return messageStr
}
