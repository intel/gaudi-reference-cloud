// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

// Application configuration
type Config struct {
	ListenPort                            uint16           `koanf:"listenPort"`
	Database                              manageddb.Config `koanf:"database"`
	Region                                string           `koanf:"region"`
	PolicyApplyInterval                   time.Duration    `koanf:"policyApplyInterval"`
	FeatureFlags                          FeatureFlags     `koanf:"featureFlags"`
	ComputeNodePoolForUnknownCloudAccount string           `koanf:"computeNodePoolForUnknownCloudAccount"`
}

type FeatureFlags struct {
}
