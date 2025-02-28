// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package sqsutil

import (
	"context"
	"os"
	"strconv"

	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func TestSQSEnqueue(t *testing.T) {

	sqsUtil := SQSUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("sqsProcessor.TestSQSEnqueue")
	logger.Info("TestSQSEnqueue BEGIN")
	defer logger.Info("TestSQSEnqueue END")
	awsQueueURL := os.Getenv("AWS_QUEUE_URL")
	if len(awsQueueURL) == 0 {
		t.Skip("AWS envionment variable is not set")
	}
	if err := sqsUtil.Init(ctx, "us-west-2", awsQueueURL, ""); err != nil {
		logger.Error(err, "couldn't init sqsProcessor")
		return
	}

	logger.Info("Going to call SendMessage")
	messageAttributes := map[string]types.MessageAttributeValue{
		"Billing": {
			DataType:    aws.String("String"),
			StringValue: aws.String("Yes"),
		},
	}
	for i := 1; i < 3; i++ {
		if err := sqsUtil.SendMessage(ctx, strconv.Itoa(i), messageAttributes); err != nil {
			logger.Error(err, "couldn't send message")
			return
		}
	}

	logger.Info("Going to call ReceiveMessage")
	messageAttributeNames := "Billing"
	if messages, err := sqsUtil.ReceiveMessage(ctx, 10, 10, messageAttributeNames); err != nil {
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

func TestSQSLongPoll(t *testing.T) {
	sqsUtil := SQSUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("sqsProcessor.TestSQSEnqueue")
	logger.Info("TestSQSEnqueue BEGIN")
	defer logger.Info("TestSQSEnqueue END")
	awsQueueURL := os.Getenv("AWS_QUEUE_URL")
	if len(awsQueueURL) == 0 {
		t.Skip("AWS envionment variable is not set")
	}
	region := "us-west-2"
	credentialsFile := ""
	messageCount := 5
	dataChannel := make(chan types.Message, 10)

	if err := sqsUtil.Init(ctx, region, awsQueueURL, credentialsFile); err != nil {
		logger.Error(err, "couldn't init sqsProcessor")
		return
	}

	logger.Info("Going to call SendMessage")
	messageAttributes := map[string]types.MessageAttributeValue{
		"CloudCredits": {
			DataType:    aws.String("String"),
			StringValue: aws.String("Yes"),
		},
	}
	for i := 1; i < messageCount; i++ {
		if err := sqsUtil.SendMessage(ctx, strconv.Itoa(i), messageAttributes); err != nil {
			logger.Error(err, "couldn't send message")
			return
		}
	}

	logger.Info("Going to call LongPoll")
	messageAttributeNames := "CloudCredits"

	go sqsUtil.LongPoll(ctx, dataChannel, 10, 10, messageAttributeNames)

	time.Sleep(100 * time.Millisecond)
	count := 0

	for {
		data := <-dataChannel
		count++
		logger.Info("No of Message", "No of Message", count)
		logger.Info("Message Received", "Message Received", data)
		if count == messageCount {
			logger.Info("All messages received")
			break
		}
	}
}

func TestSQSCloudCreditsSendReceive(t *testing.T) {
	sqsUtil := SQSUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("sqsProcessor.TestSQSCloudCreditsSendReceive")
	logger.Info("TestSQSCloudCreditsSendReceive BEGIN")
	defer logger.Info("TestSQSCloudCreditsSendReceive END")
	awsQueueURL := os.Getenv("AWS_SNS_QUEUE_URL")
	if len(awsQueueURL) == 0 {
		t.Skip("AWS envionment variable is not set")
	}

	if err := sqsUtil.Init(ctx, "us-west-2", awsQueueURL, ""); err != nil {
		logger.Error(err, "couldn't init sqsProcessor")
		return
	}

	logger.Info("Going to call SendMessage")
	messageAttributes := map[string]types.MessageAttributeValue{
		"CloudCreditsAvailable": {
			DataType:    aws.String("String"),
			StringValue: aws.String("Yes"),
		},
	}
	for i := 1; i < 3; i++ {
		if err := sqsUtil.SendMessage(ctx, strconv.Itoa(i), messageAttributes); err != nil {
			logger.Error(err, "couldn't send message")
			return
		}
	}

	logger.Info("Going to call ReceiveMessage")
	messageAttributeNames := "CloudCreditsAvailable"
	if messages, err := sqsUtil.ReceiveMessage(ctx, 10, 10, messageAttributeNames); err != nil {
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

func TestSQSCloudCreditsReceive(t *testing.T) {
	sqsUtil := SQSUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("sqsProcessor.TestSQSCloudCreditsReceive")
	logger.Info("TestSQSCloudCreditsReceive BEGIN")
	defer logger.Info("TestSQSCloudCreditsReceive END")
	awsQueueURL := os.Getenv("AWS_SNS_QUEUE_URL")
	if len(awsQueueURL) == 0 {
		t.Skip("AWS envionment variable is not set")
	}
	if err := sqsUtil.Init(ctx, "us-west-2", awsQueueURL, ""); err != nil {
		logger.Error(err, "couldn't init sqsProcessor")
		return
	}

	logger.Info("Going to call ReceiveMessage")
	messageAttributeNames := "CloudCreditsExpired"
	if messages, err := sqsUtil.ReceiveMessageWithOptions(ctx, awsQueueURL, 1, 1, messageAttributeNames); err != nil {
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
