// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

const FileStorageServiceType = "FileStorageAsAService"
const ObjectStorageServiceType = "ObjectStorageAsAService"
const StorageMetricUnitType = "TB"
const StorageTimeMetricUnitType = "hour"

type Config struct {
	ListenPort                      uint16           `koanf:"listenPort"`
	Database                        manageddb.Config `koanf:"database"`
	UsageSchedulerInterval          uint16           `koanf:"usageSchedulerInterval"`
	ProductUsageSchedulerInterval   uint16           `koanf:"productUsageSchedulerInterval"`
	MigrationResourcePaginationSize int              `koanf:"migrationResourcePaginationSize"`
	TestProfile                     bool
	StorageServiceTypes             []string `koanf:"storageServiceTypes"`
	StorageMetricUnitType           string   `koanf:"storageMetricUnitType"`
	StorageTimeMetricUnitType       string   `koanf:"storageTimeMetricUnitType"`
	Features                        struct {
		UsageScheduler        bool `koanf:"usageScheduler"`
		ProductUsageScheduler bool `koanf:"productUsageScheduler"`
	} `koanf:"features"`
}

var Cfg *Config

func NewDefaultConfig() *Config {

	if Cfg == nil {
		Cfg = &Config{
			UsageSchedulerInterval:          3600,
			ProductUsageSchedulerInterval:   3600,
			MigrationResourcePaginationSize: 40,
			StorageServiceTypes:             []string{FileStorageServiceType, ObjectStorageServiceType},
			StorageMetricUnitType:           StorageMetricUnitType,
			StorageTimeMetricUnitType:       StorageTimeMetricUnitType,
		}
	}
	return Cfg
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (config *Config) InitTestConfig() {
	cfg := NewDefaultConfig()
	cfg.UsageSchedulerInterval = 15
	cfg.TestProfile = true
}
