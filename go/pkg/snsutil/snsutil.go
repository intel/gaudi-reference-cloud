// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package snsutil

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type SNSClientInterface interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
	Subscribe(ctx context.Context, params *sns.SubscribeInput, optFns ...func(*sns.Options)) (*sns.SubscribeOutput, error)
}

const (
	attributeKey          = "eventName"
	attributeTimeStampKey = "Timestamp"
)

type SNSUtil struct {
	snsClient SNSClientInterface
	topicARN  string
}

func (snsUtil *SNSUtil) Init(ctx context.Context, region string, topicARN string, credentialsFile string) error {
	logger := log.FromContext(ctx).WithName("SNSUtil.Init")
	logger.Info("Init BEGIN")
	defer logger.Info("Init END")
	logger.Info("sns", "region", region, "TopicARN", topicARN, "credentialsFile", credentialsFile)
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
	snsUtil.snsClient = sns.NewFromConfig(cfg)
	snsUtil.topicARN = topicARN
	return nil
}

type PublishMessageOptionalParams struct {
	topicArn          string
	messageStructure  string
	messageAttributes map[string]types.MessageAttributeValue
}

func (snsUtil *SNSUtil) PublishMessage(ctx context.Context, message string) (string, error) {
	logger := log.FromContext(ctx).WithName("snsUtil.PublishMessage")
	logger.Info("PublishMessage BEGIN")
	defer logger.Info("PublishMessage END")
	return snsUtil.PublishMessageWithOpts(ctx, message, nil)
}

func (snsUtil *SNSUtil) PublishMessageWithOptions(ctx context.Context, message string, topicARN string, messageAttribute string) (string, error) {
	logger := log.FromContext(ctx).WithName("snsUtil.PublishMessage")
	logger.Info("PublishMessage BEGIN")
	defer logger.Info("PublishMessage END")
	timestamp := time.Now().Format(time.RFC3339)
	messageAttributes := map[string]types.MessageAttributeValue{
		attributeKey: {
			DataType:    aws.String("String"),
			StringValue: aws.String(messageAttribute),
		},
		attributeTimeStampKey: {
			DataType:    aws.String("String"),
			StringValue: aws.String(timestamp),
		},
	}
	opts := &PublishMessageOptionalParams{topicArn: topicARN}
	opts.messageAttributes = messageAttributes
	return snsUtil.PublishMessageWithOpts(ctx, message, opts)
}

func (snsUtil *SNSUtil) PublishMessageWithOpts(ctx context.Context, message string, opts *PublishMessageOptionalParams) (string, error) {
	logger := log.FromContext(ctx).WithName("snsUtil.PublishMessageWithOpts")
	logger.Info("PublishMessageWithOpts BEGIN")
	defer logger.Info("PublishMessageWithOpts END")

	publishMessage := &sns.PublishInput{
		Message:          aws.String(message),
		TopicArn:         aws.String(snsUtil.topicARN),
		MessageStructure: aws.String("json"),
	}

	if opts != nil {
		if len(opts.messageAttributes) > 0 {
			publishMessage.MessageAttributes = opts.messageAttributes
			logger.Info("PublishMessage messageAttributes", "messageAttributes", opts.messageAttributes)
		}

		if len(opts.messageStructure) > 0 {
			publishMessage.MessageStructure = aws.String(opts.messageStructure)
		}
		if len(opts.topicArn) > 0 {
			publishMessage.TopicArn = aws.String(opts.topicArn)
		}
	}

	logger.V(1).Info("publishing message", "publishMessage", publishMessage)
	publishMessageOutput, err := snsUtil.snsClient.Publish(ctx, publishMessage)

	if err != nil {
		logger.Error(err, "failed to publish message")
		return "", err
	}

	logger.Info("Message published successfully", "Message id: ", *publishMessageOutput.MessageId)
	return *publishMessageOutput.MessageId, nil
}

func (snsUtil *SNSUtil) Subscribe(ctx context.Context, endpoint string, topicArn string) (string, error) {
	// endpoint is sqsARN
	logger := log.FromContext(ctx).WithName("SNSUtil.Subscribe")
	logger.Info("Subscribe BEGIN", "endpoint", endpoint, "topicArn", topicArn)
	defer logger.Info("Subscribe END")
	protocol := "sqs"
	topicARN := snsUtil.topicARN
	if len(topicArn) != 0 {
		topicARN = topicArn
	}
	subscribeOutput, err := snsUtil.snsClient.Subscribe(ctx, &sns.SubscribeInput{
		Protocol: &protocol,
		Endpoint: &endpoint,
		TopicArn: &topicARN,
	})
	if err != nil {
		logger.Error(err, "Couldn't subscribe")
		return "", err
	}
	return *subscribeOutput.SubscriptionArn, nil
}
