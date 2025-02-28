// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

const (
	AWS_ACCOUNTID_DEFAULT    = "accountid"
	AWS_REGION_DEFAULT       = "us-west-2"
	AWS_TOPIC_NAME_DEFAULT   = "idc-staging-cloud-credits-topic"
	AWS_QUEUE_NAME_DEFAULT   = "idc-staging-cloud-credits-queue"
	AWS_ACCOUNTID_EXPRESSION = `(?m)^aws_account_id\s*=\s*(.*)$`
)

type Config struct {
	ListenPort    uint16           `koanf:"listenPort"`
	Database      manageddb.Config `koanf:"database"`
	TestProfile   bool
	Notifications struct {
		AWS struct {
			Region          string `koanf:"region"`
			CredentialsFile string `koanf:"credentialsFile"`
			AccountId       string `koanf:"accountId"`
			AccountIdFile   string `koanf:"accountIdFile"`
			SQS             struct {
				QueueUrl string `koanf:"queueUrl"`
				QueueArn string `koanf:"queueArn"`
			} `koanf:"sqs"`
			SNS struct {
				TopicArn string `koanf:"topicArn"`
			} `koanf:"sns"`
		} `koanf:"aws"`
	} `koanf:"notifications"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		Cfg = &Config{
			ListenPort: 8443,
		}
	}
	return Cfg
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (config *Config) InitTestConfig() {
	cfg := NewDefaultConfig()
	cfg.TestProfile = true
	Cfg = cfg
}

func (config *Config) Construct() {
	accountId, err := config.GetAWSAccountID()
	if err == nil {
		config.SetAWSAccountId(string(accountId))
	}
}

func (config *Config) GetAWSAccountID() (string, error) {
	logger := log.FromContext(context.Background()).WithName("Config.GetAWSAccountID")
	credentialsContent, err := os.ReadFile(config.GetAWSAccountIdFile())
	if err != nil {
		logger.Error(err, "failed to read credentials file")
		return "", err
	}
	re := regexp.MustCompile(AWS_ACCOUNTID_EXPRESSION)
	matchAccountId := re.FindStringSubmatch(string(credentialsContent))
	if len(matchAccountId) < 2 {
		logger.Error(errors.New("AWS account ID not found"), "credentialsContent", credentialsContent)
		return "", err
	}
	return matchAccountId[1], nil
}

func (config *Config) GetAWSSQSDefaultQueueUrl() string {
	if !strings.Contains(config.Notifications.AWS.SQS.QueueUrl, config.Notifications.AWS.Region) {
		config.Notifications.AWS.SQS.QueueUrl = strings.Replace(config.Notifications.AWS.SQS.QueueUrl, AWS_REGION_DEFAULT, config.Notifications.AWS.Region, -1)
	}
	if !strings.Contains(config.Notifications.AWS.SQS.QueueUrl, config.Notifications.AWS.AccountId) {
		config.Notifications.AWS.SQS.QueueUrl = strings.Replace(config.Notifications.AWS.SQS.QueueUrl, AWS_ACCOUNTID_DEFAULT, config.Notifications.AWS.AccountId, -1)
	}
	return config.Notifications.AWS.SQS.QueueUrl
}

func (config *Config) GetAWSSQSQueueArn() string {
	if !strings.Contains(config.Notifications.AWS.SQS.QueueArn, config.Notifications.AWS.Region) {
		config.Notifications.AWS.SQS.QueueArn = strings.Replace(config.Notifications.AWS.SQS.QueueArn, AWS_REGION_DEFAULT, config.Notifications.AWS.Region, -1)
	}
	if !strings.Contains(config.Notifications.AWS.SQS.QueueArn, config.Notifications.AWS.AccountId) {
		config.Notifications.AWS.SQS.QueueArn = strings.Replace(config.Notifications.AWS.SQS.QueueArn, AWS_ACCOUNTID_DEFAULT, config.Notifications.AWS.AccountId, -1)
	}
	return config.Notifications.AWS.SQS.QueueArn
}

func (config *Config) GetAWSSNSDefaultTopicArn() string {
	if !strings.Contains(config.Notifications.AWS.SNS.TopicArn, config.Notifications.AWS.Region) {
		config.Notifications.AWS.SNS.TopicArn = strings.Replace(config.Notifications.AWS.SNS.TopicArn, AWS_REGION_DEFAULT, config.Notifications.AWS.Region, -1)
	}
	if !strings.Contains(config.Notifications.AWS.SNS.TopicArn, config.Notifications.AWS.AccountId) {
		config.Notifications.AWS.SNS.TopicArn = strings.Replace(config.Notifications.AWS.SNS.TopicArn, AWS_ACCOUNTID_DEFAULT, config.Notifications.AWS.AccountId, -1)
	}
	return config.Notifications.AWS.SNS.TopicArn
}
func (config *Config) GetAWSSNSTopicArn(topicName string) string {
	if !strings.Contains(config.Notifications.AWS.SNS.TopicArn, config.Notifications.AWS.Region) {
		config.Notifications.AWS.SNS.TopicArn = strings.Replace(config.Notifications.AWS.SNS.TopicArn, AWS_REGION_DEFAULT, config.Notifications.AWS.Region, -1)
	}
	if !strings.Contains(config.Notifications.AWS.SNS.TopicArn, config.Notifications.AWS.AccountId) {
		config.Notifications.AWS.SNS.TopicArn = strings.Replace(config.Notifications.AWS.SNS.TopicArn, AWS_ACCOUNTID_DEFAULT, config.Notifications.AWS.AccountId, -1)
	}
	if !strings.Contains(config.Notifications.AWS.SNS.TopicArn, topicName) {
		config.Notifications.AWS.SNS.TopicArn = strings.Replace(config.Notifications.AWS.SNS.TopicArn, AWS_TOPIC_NAME_DEFAULT, topicName, -1)
	}
	return config.Notifications.AWS.SNS.TopicArn
}

func (config *Config) GetAWSSQSQueueUrl(queueName string) string {
	if !strings.Contains(config.Notifications.AWS.SQS.QueueUrl, config.Notifications.AWS.Region) {
		config.Notifications.AWS.SQS.QueueUrl = strings.Replace(config.Notifications.AWS.SQS.QueueUrl, AWS_REGION_DEFAULT, config.Notifications.AWS.Region, -1)
	}
	if !strings.Contains(config.Notifications.AWS.SQS.QueueUrl, config.Notifications.AWS.AccountId) {
		config.Notifications.AWS.SQS.QueueUrl = strings.Replace(config.Notifications.AWS.SQS.QueueUrl, AWS_ACCOUNTID_DEFAULT, config.Notifications.AWS.AccountId, -1)
	}
	if !strings.Contains(config.Notifications.AWS.SQS.QueueUrl, queueName) {
		config.Notifications.AWS.SNS.TopicArn = strings.Replace(config.Notifications.AWS.SQS.QueueUrl, AWS_QUEUE_NAME_DEFAULT, queueName, -1)
	}
	return config.Notifications.AWS.SQS.QueueUrl
}

func (config *Config) GetAWSCredentialsFile() string {
	awsCredentials, err := os.ReadFile(config.Notifications.AWS.CredentialsFile)
	if err == nil && !strings.EqualFold(string(awsCredentials), "") {
		return config.Notifications.AWS.CredentialsFile
	}
	return ""
}

func (config *Config) GetAWSSESRegion() string {
	return config.Notifications.AWS.Region
}

func (config *Config) GetAWSAccountId() string {
	return config.Notifications.AWS.AccountId
}

func (config *Config) GetAWSAccountIdFile() string {
	return config.Notifications.AWS.AccountIdFile
}

func (config *Config) SetAWSAccountId(accountId string) {
	config.Notifications.AWS.AccountId = accountId
}
