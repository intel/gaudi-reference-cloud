// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

// Application configuration
type Config struct {
	ListenPort          uint16           `koanf:"listenPort"`
	Database            manageddb.Config `koanf:"database"`
	VictoriaMetricsAddr string           `koanf:"victoriaMetricsAddr"`
	RemoteWriteIKSAddr  string           `koanf:"remoteWriteIKSAddr"`
	RemoteWriteBMAddr   string           `koanf:"remoteWriteBMAddr"`
	InsecureSkipVerify  bool             `koanf:"insecureSkipVerify"`
	VMClusterName       string           `koanf:"vMClusterName"`
	ClusterEndpoint     string           `koanf:"clusterEndpoint"`
	AwsVMClusterRegion  string           `koanf:"region"`
	IamRole             string           `koanf:"iamRole"`
	EnableMetricsBM     bool             `koanf:"enableMetricsBM"`
	RemoteReadAddrBM    string           `koanf:"remoteReadAddrBM"`
}
