// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package sqsutil

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

const (
	attributeKey          = "Timestamp"
	attributeValueAll     = "All"
	defaultMaxMessages    = 10
	defaultWaitTimeSecond = 20
)

type SQSUtil struct {
	SqsClient *sqs.Client
	QueueURL  string
}

func (sqsUtil *SQSUtil) Init(ctx context.Context, region string, queueURL string, credentialsFile string) error {
	logger := log.FromContext(ctx).WithName("SQSUtil")
	logger.Info("Init BEGIN")
	defer logger.Info("Init END")
	logger.Info("sqs", "region", region, "queueURL", queueURL, "credentialsFile", credentialsFile)
	var cfg aws.Config
	var err error
	localStackEnabled := os.Getenv("LOCALSTACK_ENABLED")
	localStackEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	if credentialsFile != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithSharedCredentialsFiles([]string{credentialsFile}),
		)
	} else if localStackEnabled == "true" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           localStackEndpoint,
				SigningRegion: region,
			}, nil
		})
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummyaccesskey", "dummysecretkey", "")),
			config.WithEndpointResolverWithOptions(customResolver))
	} else {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
	}
	if err != nil {
		logger.Error(err, "failed to load defaultConfig")
		return err
	}
	sqsUtil.SqsClient = sqs.NewFromConfig(cfg)
	sqsUtil.QueueURL = queueURL
	return nil
}

func (sqsUtil *SQSUtil) SendMessage(ctx context.Context, message string, messageAttributes map[string]types.MessageAttributeValue) error {
	logger := log.FromContext(ctx).WithName("sqsUtil.SendMessage")
	logger.Info("SendMessage BEGIN")
	defer logger.Info("SendMessage END")

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(sqsUtil.QueueURL),
		MessageBody: aws.String(message),
	}

	if len(messageAttributes) > 0 {
		input.MessageAttributes = messageAttributes
		logger.Info("SendMessage messageAttributes", "messageAttributes", messageAttributes)
	}

	sendMessageOutput, err := sqsUtil.SqsClient.SendMessage(ctx, input)
	if err != nil {
		logger.Error(err, "failed to send message")
		return err
	}
	logger.Info("message send successfully", "sendMessageOutput", sendMessageOutput)
	return nil
}

func (sqsUtil *SQSUtil) ReceiveMessageWithOptions(ctx context.Context, queueURL string, timeout int32, maxNumberOfMessages int32, messageAttribute string) ([]types.Message, error) {
	logger := log.FromContext(ctx).WithName("SqsUtil.ReceiveMessageWithOptions")
	logger.Info("ReceiveMessageWithOptions BEGIN")
	defer logger.Info("ReceiveMessageWithOptions END")

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: maxNumberOfMessages, // Maximum number of messages to retrieve (adjust as needed)
		WaitTimeSeconds:     timeout,             // Wait for messages in seconds (adjust as needed)
		MessageAttributeNames: []string{
			"All",
		},
	}
	if len(messageAttribute) > 0 {
		logger.Info("messageAttribute", "messageAttribute", messageAttribute)
		input.MessageAttributeNames = []string{attributeKey}
	}
	if maxNumberOfMessages <= 0 {
		input.MaxNumberOfMessages = defaultMaxMessages
	}
	if timeout <= 0 {
		input.WaitTimeSeconds = defaultWaitTimeSecond
	}
	logger.V(1).Info("message input", "input", input)
	// Check if the message has the expected message type attribute
	return sqsUtil.ReceiveMessageWithInput(ctx, input, messageAttribute)
}

func (sqsUtil *SQSUtil) ReceiveMessage(ctx context.Context, timeout int32, maxNumberOfMessages int32, messageAttribute string) ([]types.Message, error) {
	logger := log.FromContext(ctx).WithName("SqsUtil.ReceiveMessage")
	logger.Info("ReceiveMessage BEGIN")
	defer logger.Info("ReceiveMessage END")

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            &sqsUtil.QueueURL,
		MaxNumberOfMessages: maxNumberOfMessages, // Maximum number of messages to retrieve (adjust as needed)
		WaitTimeSeconds:     timeout,             // Wait for messages in seconds (adjust as needed)
		MessageAttributeNames: []string{
			"All",
		},
	}

	if len(messageAttribute) > 0 {
		logger.Info("messageAttribute", "messageAttribute", messageAttribute)
		input.MessageAttributeNames = []string{attributeKey}
	}

	// Check if the message has the expected message type attribute
	return sqsUtil.ReceiveAllMessages(ctx, input, messageAttribute)
}

func (sqsUtil *SQSUtil) ReceiveMessageWithInput(ctx context.Context, input *sqs.ReceiveMessageInput, messageAttribute string) ([]types.Message, error) {
	logger := log.FromContext(ctx).WithName("SqsUtil.ReceiveMessageWithInput")
	logger.Info("ReceiveMessageWithInput BEGIN")
	defer logger.Info("ReceiveMessageWithInput END")
	messagesRet := []types.Message{}
	messages, err := sqsUtil.receiveMessages(ctx, input, messageAttribute)
	if err != nil {
		logger.Error(err, "failed to receive message")
		return nil, err
	}
	messagesRet = append(messagesRet, messages...)

	return messagesRet, nil
}

func (sqsUtil *SQSUtil) ReceiveAllMessages(ctx context.Context, input *sqs.ReceiveMessageInput, messageAttribute string) ([]types.Message, error) {
	logger := log.FromContext(ctx).WithName("SqsUtil.ReceiveAllMessages")
	logger.Info("ReceiveAllMessages BEGIN")
	defer logger.Info("ReceiveAllMessages END")
	messagesRet := []types.Message{}
	for {
		messages, err := sqsUtil.receiveMessages(ctx, input, messageAttribute)
		if err != nil {
			logger.Error(err, "failed to receive message")
			return nil, err
		}
		if len(messages) == 0 {
			break
		}
		messagesRet = append(messagesRet, messages...)

	}
	return messagesRet, nil
}

func (sqsUtil *SQSUtil) receiveMessages(ctx context.Context, input *sqs.ReceiveMessageInput, messageAttribute string) ([]types.Message, error) {
	logger := log.FromContext(ctx).WithName("SqsUtil.receiveMessages")
	logger.Info("receiveMessages BEGIN")
	defer logger.Info("receiveMessages END")
	messages := []types.Message{}
	receiveMessageOutput, err := sqsUtil.SqsClient.ReceiveMessage(ctx, input)
	if err != nil {
		logger.Error(err, "failed to receive message")
		return nil, err
	}
	if len(receiveMessageOutput.Messages) == 0 {
		logger.Info("No more messages in the queue.")
		return messages, nil
	}

	for _, message := range receiveMessageOutput.Messages {
		if len(messageAttribute) > 0 && messageAttribute != attributeValueAll {
			logger.V(1).Info("message filter", "messageAttribute", messageAttribute, "MessageAttributes", message.MessageAttributes)
			messageTypeAttr, ok := message.MessageAttributes[attributeKey]
			if !ok || messageTypeAttr.StringValue == nil {
				logger.Info(" Skip due to messagefilter", "Id", message.MessageId, "receiptHandle", message.ReceiptHandle, "body", *message.Body)
				continue
			}
			if *messageTypeAttr.StringValue != messageAttribute {
				logger.Info(" Skip due to message attribue filter", "Id", message.MessageId, "receiptHandle", message.ReceiptHandle, "body", *message.Body)
				continue
			}
		}
		messages = append(messages, message)
		logger.Info("message", "Id", message.MessageId, "receiptHandle", message.ReceiptHandle, "body", *message.Body)
	}
	return messages, nil
}

func (sqsUtil *SQSUtil) DeleteMessage(ctx context.Context, message types.Message) error {
	logger := log.FromContext(ctx).WithName("sqsUtil.DeleteMessage")
	logger.Info("DeleteMessage BEGIN")
	defer logger.Info("DeleteMessage END")

	_, err := sqsUtil.SqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &sqsUtil.QueueURL,
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		logger.Error(err, "failed to delete message from SQS")
		return err
	}
	return nil
}

func (sqsUtil *SQSUtil) DeleteMessageWithReceiptHandle(ctx context.Context, queueUrl string, receiptHandle string) error {
	logger := log.FromContext(ctx).WithName("sqsUtil.DeleteMessageWithReceiptHandle")
	logger.Info("DeleteMessageWithReceiptHandle BEGIN")
	defer logger.Info("DeleteMessageWithReceiptHandle END")

	_, err := sqsUtil.SqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		logger.Error(err, "failed to delete message from SQS")
		return err
	}
	return nil
}

func (sqsUtil *SQSUtil) LongPoll(ctx context.Context, messageChannel chan types.Message, timeout int32, maxNumberOfMessages int32, messageAttribute string) error {
	logger := log.FromContext(ctx).WithName("sqsUtil.LongPolling")
	logger.Info("LongPolling BEGIN")

	for {
		msg, err := sqsUtil.ReceiveMessage(ctx, timeout, maxNumberOfMessages, messageAttribute)
		if err != nil {
			logger.Error(err, "failed to receive message from SQS")
			return err
		}
		for _, str := range msg {
			messageChannel <- str
		}
		time.Sleep(1 * time.Second)
	}
}
