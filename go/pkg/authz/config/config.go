// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

type AuditLogging struct {
	Enabled                 bool `koanf:"enabled"`
	CleanupSchedulerEnabled bool `koanf:"cleanupSchedulerEnabled"`
	SchedulerTime           int8 `koanf:"schedulerTime"`
	RetentionPeriodInDays   int8 `koanf:"retentionPeriodInDays"`
}

type Features struct {
	PoliciesStartupSync bool         `koanf:"policiesStartupSync"`
	Watcher             bool         `koanf:"watcher"`
	AuditLogging        AuditLogging `koanf:"auditLogging"`
}

type Limits struct {
	MaxCloudAccountRoles int `koanf:"maxCloudAccountRoles"`
	MaxPermissions       int `koanf:"maxPermissions"`
}

type Config struct {
	Database          manageddb.Config      `koanf:"database"`
	ListenConfig      grpcutil.ListenConfig `koanf:"listenConfig"`
	Features          Features              `koanf:"features"`
	Limits            Limits                `koanf:"limits"`
	ResourcesFilePath string                `koanf:"resourcesFilePath"`
	ModelFilePath     string                `koanf:"modelFilePath"`
	PolicyFilePath    string                `koanf:"policyFilePath"`
	GroupFilePath     string                `koanf:"groupFilePath"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		Cfg = &Config{ListenConfig: grpcutil.ListenConfig{}}
	}
	return Cfg
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenConfig.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenConfig.ListenPort = port
}
