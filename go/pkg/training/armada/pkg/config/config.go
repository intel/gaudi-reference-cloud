// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

type Config struct {
	Database      manageddb.Config             `koanf:"database"`
	ClusterConfig SchedulerGlobalConfig        `koanf:"globalConfig"`
	CloudConfig   SchedulerCloudProviderConfig `koanf:"cloudConfig"`
}

type SchedulerGlobalConfig struct {
	SchedulerInterval uint16 `koanf:"schedulerIntervalSeconds"`
	MaxClusters       uint16 `koanf:"maxClustersPerAccount"`
	MaxNodes          uint16 `koanf:"maxNodesPerAccount"`
	MaxVNets          uint16 `koanf:"maxVnetsPerAccount"`
	MaxStorage        uint16 `koanf:"maxStorageSizePerAccount"`
}

type SchedulerCloudProviderConfig struct {
	IDC struct {
		AuthAPIEndpoint           string `koanf:"authAPIEndpoint"`
		ComputeAPIEndpoint        string `koanf:"computeAPIEndpoint"`
		ComputeGrpcAPIEndpoint    string `koanf:"computeGrpcAPIEndpoint"`
		FoundationGrpcAPIEndpoint string `koanf:"foundationGrpcAPIEndpoint"`
		FoundationAPIEndpoint     string `koanf:"foundationAPIEndpoint"`
		StorageGrpcAPIEndpoint    string `koanf:"storageGrpcAPIEndpoint"`
		AvailabilityZone          string `koanf:"availabilityZone"`
		Region                    string `koanf:"region"`
	} `koanf:"idc"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		// FUTURE IMPROVEMENT: Make scheduler values set from environments instead of here
		clusterConfig := SchedulerGlobalConfig{
			SchedulerInterval: 6000,
			MaxNodes:          0,
			MaxVNets:          0,
			MaxStorage:        0,
		}
		return &Config{
			ClusterConfig: clusterConfig,
			CloudConfig:   SchedulerCloudProviderConfig{},
		}
	}
	return Cfg
}
