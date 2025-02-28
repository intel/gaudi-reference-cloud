// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

import (
	billingcommonconfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

type Config struct {
	ListenConfig grpcutil.ListenConfig            `koanf:"listenConfig"`
	Database     manageddb.Config                 `koanf:"database"`
	CommonConfig billingcommonconfig.CommonConfig `koanf:"commonConfig"`
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
		return &Config{CommonConfig: billingcommonconfig.CommonConfig{MaxDefaultHistory: 12, InstanceSearchWindow: 2}}
	}
	return Cfg
}
