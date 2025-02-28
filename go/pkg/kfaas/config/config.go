// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

// Application configuration
type Config struct {
	ListenPort    uint16           `koanf:"listenPort"`
	Database      manageddb.Config `koanf:"database"`
	IKSServerAddr string           `koanf:"IKSServerAddr"`
	InstallImage  string           `koanf:"InstallImage"`
}
