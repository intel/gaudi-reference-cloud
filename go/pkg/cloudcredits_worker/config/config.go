// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
)

type Config struct {
	ListenConfig grpcutil.ListenConfig `koanf:"listenConfig"`
	AWS          struct {
		SQS struct {
			QueueName           string `koanf:"queueName"`
			WaitTimeSeconds     int64  `koanf:"waitTimeSeconds"`
			MaxNumberOfMessages int32  `koanf:"maxNumberofMessages"`
		} `koanf:"sqs"`
	} `koanf:"aws"`
	Features struct {
		SendCreditUsageEmail            bool     `koanf:"sendCreditUsageEmail"`
		SendCreditExpiryEmail           bool     `koanf:"sendCreditExpiryEmail"`
		ServicesTerminationScheduler    bool     `koanf:"servicesTerminationScheduler"`
		ServicesTerminationAccountTypes []string `koanf:"servicesTerminationAccountTypes"`
		CreditUsageEmailAccountTypes    []string `koanf:"creditUsageEmailAccountTypes"`
		CreditExpiryEmailAccountTypes   []string `koanf:"creditExpiryEmailAccountTypes"`
	} `koanf:"features"`
	Notifications struct {
		SenderEmail                           string `koanf:"senderEmail"`
		ConsoleUrl                            string `koanf:"consoleUrl"`
		PaymentUrl                            string `koanf:"paymentUrl"`
		CloudCreditHundredPercentUsedTemplate string `koanf:"cloudCreditHundredPercentUsedTemplate"`
		CloudCreditEightyPercentUsedTemplate  string `koanf:"cloudCreditEightyPercentUsedTemplate"`
		CloudCreditExpiryTemplate             string `koanf:"cloudCreditExpiryTemplate"`
		OpsPDL                                string `koanf:"opsPDL"`
	} `koanf:"notifications"`
	WorkerConfig struct {
		InitSleepTimeSeconds int64 `koanf:"initSleepTimeSeconds"`
		MaxSleepTimeSeconds  int64 `koanf:"maxSleepTimeSeconds"`
		ChannelCapacity      int32 `koanf:"channelCapacity"`
	} `koanf:"workerConfig"`
	TestProfile                    bool
	PremiumCloudCreditThreshold    uint16 `koanf:"premiumCloudCreditThreshold"`
	IntelCloudCreditThreshold      uint16 `koanf:"intelCloudCreditThreshold"`
	EnterpriseCloudCreditThreshold uint16 `koanf:"enterpriseCloudCreditThreshold"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		Cfg = &Config{
			PremiumCloudCreditThreshold:    80,
			IntelCloudCreditThreshold:      80,
			EnterpriseCloudCreditThreshold: 80,
		}
	}
	return Cfg
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenConfig.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenConfig.ListenPort = port
}

func (config *Config) GetAWSSQSQueueName() string {
	return config.AWS.SQS.QueueName
}

func (config *Config) GetSenderEmail() string {
	return config.Notifications.SenderEmail
}

func (config *Config) GetConsoleUrl() string {
	return config.Notifications.ConsoleUrl
}

func (config *Config) GetPaymentUrl() string {
	return config.Notifications.PaymentUrl
}

func (config *Config) GetCloudCreditHundredPercentUsedTemplate() string {
	return config.Notifications.CloudCreditHundredPercentUsedTemplate
}

func (config *Config) GetCloudCreditEightyPercentUsedTemplate() string {
	return config.Notifications.CloudCreditEightyPercentUsedTemplate
}

func (config *Config) GetSendCreditUsageEmail() bool {
	return config.Features.SendCreditUsageEmail
}

func (config *Config) GetCloudCreditExpiryTemplate() string {
	return config.Notifications.CloudCreditExpiryTemplate
}

func (config *Config) GetWaitTimeSeconds() int64 {
	return config.AWS.SQS.WaitTimeSeconds
}

func (config *Config) GetAWSSQSMaxNumberOfMessages() int32 {
	return config.AWS.SQS.MaxNumberOfMessages
}

func (config *Config) GetInitSleepTimeSeconds() int64 {
	return config.WorkerConfig.InitSleepTimeSeconds
}

func (config *Config) GetMaxSleepTimeSeconds() int64 {
	return config.WorkerConfig.MaxSleepTimeSeconds
}

func (config *Config) GetChannelCapacity() int32 {
	return config.WorkerConfig.ChannelCapacity
}

func (config *Config) GetSendCreditExpiryEmail() bool {
	return config.Features.SendCreditExpiryEmail
}

func (config *Config) GetServicesTerminationAccountTypes() []string {
	return config.Features.ServicesTerminationAccountTypes
}

func (config *Config) GetCreditUsageEmailAccountTypes() []string {
	return config.Features.CreditUsageEmailAccountTypes
}

func (config *Config) GetCreditExpiryEmailAccountTypes() []string {
	return config.Features.CreditExpiryEmailAccountTypes
}

func (config *Config) SetServicesTerminationAccountTypes(servicesTerminationAccountTypes []string) {
	config.Features.ServicesTerminationAccountTypes = servicesTerminationAccountTypes
}

func (config *Config) SetCreditUsageEmailAccountTypes(creditUsageEmailAccountTypes []string) {
	config.Features.CreditUsageEmailAccountTypes = creditUsageEmailAccountTypes
}

func (config *Config) SetCreditExpiryEmailAccountTypes(creditExpiryEmailAccountTypes []string) {
	config.Features.CreditExpiryEmailAccountTypes = creditExpiryEmailAccountTypes
}

func (config *Config) GetOpsPDL() string {
	return config.Notifications.OpsPDL
}
