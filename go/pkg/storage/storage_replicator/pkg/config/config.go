// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

type Config struct {
	SchedulerInterval uint16 `koanf:"schedulerInterval"`
	IDCServiceConfig  struct {
		StorageAPIGrpcEndpoint string `koanf:"storageApiServerAddr"`
		VASTEnabled            bool   `koanf:"generalPurposeVASTEnabled"`
	} `koanf:"idcServiceConfig"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		cfg := &Config{}
		cfg.SchedulerInterval = 2
		cfg.IDCServiceConfig.VASTEnabled = false
		Cfg = cfg
	}
	return Cfg
}
