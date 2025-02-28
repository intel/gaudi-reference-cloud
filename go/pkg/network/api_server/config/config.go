// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

// Application configuration
type Config struct {
	ListenPort             uint16           `koanf:"listenPort"`
	Database               manageddb.Config `koanf:"database"`
	PrometheusListenPort   uint16           `koanf:"prometheusListenPort"`
	Region                 string           `koanf:"region"`
	AvailabilityZones      []string         `koanf:"availabilityZones"`
	FeatureFlags           FeatureFlags     `koanf:"featureFlags"`
	CloudaccountServerAddr string           `koanf:"cloudaccountServerAddr"`
}

type FeatureFlags struct {
}
