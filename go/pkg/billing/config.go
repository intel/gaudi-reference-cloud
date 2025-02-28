// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

type Config struct {
	ListenConfig                            grpcutil.ListenConfig `koanf:"listenConfig"`
	Database                                manageddb.Config      `koanf:"database"`
	CreditsInstallSchedulerInterval         uint16                `koanf:"creditsInstallSchedulerInterval"`
	ReportUsageSchedulerInterval            uint16                `koanf:"reportUsageSchedulerInterval"`
	CloudCreditUsageReportSchedulerInterval uint16                `koanf:"cloudCreditUsageReportSchedulerInterval"`
	CreditUsageSchedulerInterval            uint16                `koanf:"creditUsageSchedulerInterval"`
	CreditExpirySchedulerInterval           uint16                `koanf:"creditExpirySchedulerInterval"`
	PremiumCloudCreditThreshold             uint16                `koanf:"premiumCloudCreditThreshold"`
	IntelCloudCreditThreshold               uint16                `koanf:"intelCloudCreditThreshold"`
	EnterpriseCloudCreditThreshold          uint16                `koanf:"enterpriseCloudCreditThreshold"`
	PremiumCloudCreditNotifyBeforeExpiry    uint16                `koanf:"premiumCloudCreditNotifyBeforeExpiry"`
	IntelCloudCreditNotifyBeforeExpiry      uint16                `koanf:"intelCloudCreditNotifyBeforeExpiry"`
	EnterpriseCloudCreditNotifyBeforeExpiry uint16                `koanf:"enterpriseCloudCreditNotifyBeforeExpiry"`
	ServicesTerminationSchedulerInterval    uint16                `koanf:"servicesTerminationSchedulerInterval"`
	ServicesTerminationInterval             uint16                `koanf:"servicesTerminationInterval"`
	CreditsExpiryMinimumInterval            uint16                `koanf:"creditsExpiryMinimumInterval"`
	EventExpirySchedulerInterval            uint16                `koanf:"eventExpirySchedulerInterval"`
	RunSchedulers                           bool                  `koanf:"runSchedulers"`
	InstanceTerminationMeteringDuration     int64                 `koanf:"instanceTerminationMeteringDuration"`
	CouponNumberOfUsesThresholdStandard     uint16                `koanf:"couponNumberOfUsesThresholdStandard"`
	CouponNumberOfUsesThresholdNonStandard  uint16                `koanf:"couponNumberOfUsesThresholdNonStandard"`
	TestProfile                             bool
	Features                                struct {
		CreditInstallScheduler          bool     `koanf:"creditInstallScheduler"`
		ReportUsageScheduler            bool     `koanf:"reportUsageScheduler"`
		CreditUsageScheduler            bool     `koanf:"creditUsageScheduler"`
		CreditExpiryScheduler           bool     `koanf:"creditExpiryScheduler"`
		ServicesTerminationScheduler    bool     `koanf:"servicesTerminationScheduler"`
		EventExpiryScheduler            bool     `koanf:"eventExpiryScheduler"`
		CloudCreditUsageReportScheduler bool     `koanf:"cloudCreditUsageReportScheduler"`
		CreditUsageEmail                bool     `koanf:"creditUsageEmail"`
		CreditExpiryEmail               bool     `koanf:"creditExpiryEmail"`
		BillingUsageMetrics             bool     `koanf:"billingUsageMetrics"`
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
		PDLEmail                              string `koanf:"pdlEmail"`
	} `koanf:"notifications"`
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenConfig.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenConfig.ListenPort = port
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		Cfg = &Config{
			CreditsInstallSchedulerInterval:         3600,
			ReportUsageSchedulerInterval:            1800,
			CloudCreditUsageReportSchedulerInterval: 1800,
			CreditUsageSchedulerInterval:            240,
			CreditExpirySchedulerInterval:           240,
			PremiumCloudCreditThreshold:             80,
			IntelCloudCreditThreshold:               80,
			EnterpriseCloudCreditThreshold:          80,
			PremiumCloudCreditNotifyBeforeExpiry:    4320,
			IntelCloudCreditNotifyBeforeExpiry:      4320,
			EnterpriseCloudCreditNotifyBeforeExpiry: 4320,
			ServicesTerminationSchedulerInterval:    240,
			ServicesTerminationInterval:             1440,
			InstanceTerminationMeteringDuration:     2,
			CreditsExpiryMinimumInterval:            31,
			EventExpirySchedulerInterval:            1440,
			RunSchedulers:                           false,
			CouponNumberOfUsesThresholdStandard:     500,
			CouponNumberOfUsesThresholdNonStandard:  50,
		}
		Cfg.Features.CreditInstallScheduler = true
		Cfg.Features.CreditUsageScheduler = true
		Cfg.Features.CreditExpiryScheduler = true
		Cfg.Features.ServicesTerminationScheduler = true
		Cfg.Features.EventExpiryScheduler = true
		Cfg.Features.CloudCreditUsageReportScheduler = true
		Cfg.Features.CreditUsageEmail = true
		Cfg.Features.CreditExpiryEmail = true
		Cfg.Features.BillingUsageMetrics = false
	}
	return Cfg
}

func (config *Config) InitTestConfig() {
	config.CreditsInstallSchedulerInterval = 15
	config.ReportUsageSchedulerInterval = 1800
	config.CloudCreditUsageReportSchedulerInterval = 1800
	config.CreditUsageSchedulerInterval = 240
	config.CreditExpirySchedulerInterval = 240
	config.PremiumCloudCreditThreshold = 80
	config.IntelCloudCreditThreshold = 80
	config.EnterpriseCloudCreditThreshold = 80
	config.PremiumCloudCreditNotifyBeforeExpiry = 4320
	config.IntelCloudCreditNotifyBeforeExpiry = 4320
	config.EnterpriseCloudCreditNotifyBeforeExpiry = 4320
	config.ServicesTerminationSchedulerInterval = 240
	config.ServicesTerminationInterval = 1440
	config.InstanceTerminationMeteringDuration = 2
	config.CreditsExpiryMinimumInterval = 31
	config.CouponNumberOfUsesThresholdStandard = 500
	config.CouponNumberOfUsesThresholdNonStandard = 50
	config.EventExpirySchedulerInterval = 1440
	config.RunSchedulers = false
	config.Features.CreditInstallScheduler = true
	config.Features.CreditUsageScheduler = true
	config.Features.CreditExpiryScheduler = true
	config.Features.ServicesTerminationScheduler = true
	config.Features.EventExpiryScheduler = true
	config.Features.CloudCreditUsageReportScheduler = true
	config.Features.BillingUsageMetrics = true
	Cfg = config
	Cfg.Features.ServicesTerminationAccountTypes = []string{"ACCOUNT_TYPE_STANDARD", "ACCOUNT_TYPE_PREMIUM"}
	Cfg.Features.CreditExpiryEmailAccountTypes = []string{"ACCOUNT_TYPE_STANDARD", "ACCOUNT_TYPE_PREMIUM"}
	Cfg.Features.CreditUsageEmailAccountTypes = []string{"ACCOUNT_TYPE_STANDARD", "ACCOUNT_TYPE_PREMIUM"}
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

func (config *Config) GetCreditUsageEmail() bool {
	return config.Features.CreditUsageEmail
}

func (config *Config) GetCreditExpiryEmail() bool {
	return config.Features.CreditExpiryEmail
}

func (config *Config) GetCloudCreditExpiryTemplate() string {
	return config.Notifications.CloudCreditExpiryTemplate
}

func (config *Config) GetPDLEmail() string {
	return config.Notifications.PDLEmail
}

func (config *Config) GetFeaturesBillingUsageMetrics() bool {
	return config.Features.BillingUsageMetrics
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

func (config *Config) GetinstanceTerminationMeteringDuration() int64 {
	return config.InstanceTerminationMeteringDuration
}
