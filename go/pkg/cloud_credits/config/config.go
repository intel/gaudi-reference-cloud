// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

type Config struct {
	ListenConfig                            grpcutil.ListenConfig `koanf:"listenConfig"`
	Database                                manageddb.Config      `koanf:"database"`
	CreditUsageEventSchedulerInterval       uint16                `koanf:"creditUsageEventSchedulerInterval"`
	CreditExpiryEventSchedulerInterval      uint16                `koanf:"creditExpiryEventSchedulerInterval"`
	CreditUsageReportSchedulerInterval      uint16                `koanf:"creditUsageReportSchedulerInterval"`
	PremiumCloudCreditThreshold             uint16                `koanf:"premiumCloudCreditThreshold"`
	IntelCloudCreditThreshold               uint16                `koanf:"intelCloudCreditThreshold"`
	EnterpriseCloudCreditThreshold          uint16                `koanf:"enterpriseCloudCreditThreshold"`
	PremiumCloudCreditNotifyBeforeExpiry    uint16                `koanf:"premiumCloudCreditNotifyBeforeExpiry"`
	IntelCloudCreditNotifyBeforeExpiry      uint16                `koanf:"intelCloudCreditNotifyBeforeExpiry"`
	EnterpriseCloudCreditNotifyBeforeExpiry uint16                `koanf:"enterpriseCloudCreditNotifyBeforeExpiry"`
	CreditsExpiryMinimumInterval            uint16                `koanf:"creditsExpiryMinimumInterval"`
	RunCreditEventSchedulers                bool                  `koanf:"runCreditEventSchedulers"`
	CouponNumberOfUsesThresholdStandard     uint16                `koanf:"couponNumberOfUsesThresholdStandard"`
	CouponNumberOfUsesThresholdNonStandard  uint16                `koanf:"couponNumberOfUsesThresholdNonStandard"`
	TestProfile                             bool
	Features                                struct {
		CreditUsageEventScheduler  bool `koanf:"creditUsageEventScheduler"`
		CreditExpiryEventScheduler bool `koanf:"creditExpiryEventScheduler"`
		CreditUsageReportScheduler bool `koanf:"CreditUsageReportScheduler"`
	} `koanf:"features"`
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
			CreditUsageEventSchedulerInterval:       240,
			CreditExpiryEventSchedulerInterval:      240,
			CreditUsageReportSchedulerInterval:      240,
			PremiumCloudCreditThreshold:             80,
			IntelCloudCreditThreshold:               80,
			EnterpriseCloudCreditThreshold:          80,
			PremiumCloudCreditNotifyBeforeExpiry:    4320,
			IntelCloudCreditNotifyBeforeExpiry:      4320,
			EnterpriseCloudCreditNotifyBeforeExpiry: 4320,
			CreditsExpiryMinimumInterval:            31,
			CouponNumberOfUsesThresholdStandard:     500,
			CouponNumberOfUsesThresholdNonStandard:  50,
			RunCreditEventSchedulers:                false,
		}
		Cfg.Features.CreditUsageEventScheduler = true
		Cfg.Features.CreditExpiryEventScheduler = true
		Cfg.Features.CreditUsageReportScheduler = true
	}
	return Cfg
}

func (config *Config) InitTestConfig() {
	config.CreditUsageEventSchedulerInterval = 240
	config.CreditExpiryEventSchedulerInterval = 240
	config.PremiumCloudCreditThreshold = 80
	config.IntelCloudCreditThreshold = 80
	config.EnterpriseCloudCreditThreshold = 80
	config.PremiumCloudCreditNotifyBeforeExpiry = 4320
	config.IntelCloudCreditNotifyBeforeExpiry = 4320
	config.EnterpriseCloudCreditNotifyBeforeExpiry = 4320
	config.RunCreditEventSchedulers = false
	config.CouponNumberOfUsesThresholdStandard = 500
	config.CouponNumberOfUsesThresholdNonStandard = 50
	config.Features.CreditUsageEventScheduler = true
	config.Features.CreditExpiryEventScheduler = true
	config.TestProfile = true
}

func (config *Config) GetCreditUsageEventSchedulerInterval() uint16 {
	return config.CreditUsageEventSchedulerInterval
}

func (config *Config) GetCreditExpiryEventSchedulerInterval() uint16 {
	return config.CreditExpiryEventSchedulerInterval
}

func (config *Config) GetPremiumCloudCreditThreshold() uint16 {
	return config.PremiumCloudCreditThreshold
}
func (config *Config) GetIntelCloudCreditThreshold() uint16 {
	return config.IntelCloudCreditThreshold
}

func (config *Config) GetEnterpriseCloudCreditThreshold() uint16 {
	return config.EnterpriseCloudCreditThreshold
}

func (config *Config) GetPremiumCloudCreditNotifyBeforeExpiry() uint16 {
	return config.PremiumCloudCreditNotifyBeforeExpiry
}

func (config *Config) GetIntelCloudCreditNotifyBeforeExpiry() uint16 {
	return config.IntelCloudCreditNotifyBeforeExpiry
}

func (config *Config) GetEnterpriseCloudCreditNotifyBeforeExpiry() uint16 {
	return config.EnterpriseCloudCreditNotifyBeforeExpiry
}

func (config *Config) GetRunCreditEventSchedulers() bool {
	return config.RunCreditEventSchedulers
}

func (config *Config) GetCreditUsageReportSchedulerInterval() uint16 {
	return config.CreditUsageReportSchedulerInterval
}

func (config *Config) GetUsageReportScheduler() bool {
	return config.Features.CreditUsageReportScheduler
}

func (config *Config) GetFeatureCreditUsageEventScheduler() bool {
	return config.Features.CreditUsageEventScheduler
}

func (config *Config) GetCreditExpiryEventScheduler() bool {
	return config.Features.CreditExpiryEventScheduler
}
