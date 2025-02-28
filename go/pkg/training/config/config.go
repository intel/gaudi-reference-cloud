// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

type BatchServiceConfig struct {
	IDC                            IdcConfig `koanf:"idc"`
	SlurmBatchServiceEndpoint      string    `koanf:"slurmBatchService"`
	SlurmJupyterhubServiceEndpoint string    `koanf:"slurmJupyterhubService"`
	SlurmSSHServiceEndpoint        string    `koanf:"slurmSSHService"`
}

type IdcConfig struct {
	ComputeGrpcAPIEndpoint    string `koanf:"computeGrpcAPIEndpoint"`
	FoundationGrpcAPIEndpoint string `koanf:"foundationGrpcAPIEndpoint"`
	StorageGrpcAPIEndpoint    string `koanf:"storageGrpcAPIEndpoint"`
	AvailabilityZone          string `koanf:"availabilityZone"`
	Region                    string `koanf:"region"`
}

var Cfg *BatchServiceConfig

func NewDefaultConfig() *BatchServiceConfig {
	if Cfg == nil {
		return &BatchServiceConfig{
			IDC: IdcConfig{},
		}
	}
	return Cfg
}
