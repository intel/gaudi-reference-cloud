// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

type Config struct {
	ListenConfig grpcutil.ListenConfig `koanf:"listenConfig"`
	Database     manageddb.Config      `koanf:"database"`
	TestProfile  bool
	DisableEmail bool
	Authz        struct {
		Enabled bool `koanf:"enabled"`
	} `koanf:"authz"`
	Notifications struct {
		InviteLink            string `koanf:"inviteLink"`
		SenderEmail           string `koanf:"senderEmail"`
		InviteTemplate        string `koanf:"inviteTemplate"`
		InviteAcceptTemplate  string `koanf:"inviteAcceptTemplate"`
		InviteExpiredTemplate string `koanf:"inviteExpiredTemplate"`
		OTPTemplate           string `koanf:"otpTemplate"`
	} `koanf:"notifications"`
	RunSchedulers                       bool  `koanf:"runSchedulers"`
	InvitationsExpirySchedulerTime      int8  `koanf:"invitationsExpirySchedulerTime"`
	InvitationsExpirySchedulerBatchSize int   `koanf:"invitationsExpirySchedulerBatchSize"`
	IntelMemberInvitationLimit          int32 `koanf:"intelMemberInvitationLimit"`
	PremiumMemberInvitationLimit        int32 `koanf:"premiumMemberInvitationLimit"`
	EnterpriseMemberInvitationLimit     int32 `koanf:"enterpriseMemberInvitationLimit"`
	OtpRetryLimit                       int32 `koanf:"otpRetryLimit"`
	OtpRetryLimitIntervalDuration       int32 `koanf:"otpRetryLimitIntervalDuration"`
	Features                            struct {
		InvitationsExpiryScheduler bool `koanf:"invitationsExpiryScheduler"`
		InvitationsExpiryEmail     bool `koanf:"invitationsExpiryEmail"`
	} `koanf:"features"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		Cfg = &Config{}
	}
	return Cfg
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenConfig.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenConfig.ListenPort = port
}

func (config *Config) InitTestConfig() {
	// setting default values for testing
	config.IntelMemberInvitationLimit = 15
	config.PremiumMemberInvitationLimit = 10
	config.EnterpriseMemberInvitationLimit = 50
	config.OtpRetryLimit = 3
	config.OtpRetryLimitIntervalDuration = 1
	config.TestProfile = true
	config.Authz.Enabled = true
}

func (config *Config) GetInviteLink() string {
	return config.Notifications.InviteLink
}

func (config *Config) GetSenderEmail() string {
	return config.Notifications.SenderEmail
}
func (config *Config) GetInviteTemplate() string {
	return config.Notifications.InviteTemplate
}

func (config *Config) GetInviteExpiredTemplate() string {
	return config.Notifications.InviteExpiredTemplate
}

func (config *Config) GetOTPTemplate() string {
	return config.Notifications.OTPTemplate
}

func (config *Config) GetInviteAcceptTemplate() string {
	return config.Notifications.InviteAcceptTemplate
}

func (config *Config) GetInviteExpiryEmail() bool {
	return config.Features.InvitationsExpiryEmail
}
