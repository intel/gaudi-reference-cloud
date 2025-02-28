// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

// Application configuration
type Config struct {
	ListenPort               uint16           `koanf:"listenPort"`
	Database                 manageddb.Config `koanf:"database"`
	InsecureSkipVerify       bool             `koanf:"insecureSkipVerify"`
	OpenSearchEndpoint       string           `koanf:"openSearchEndpoint"`
	OpenSearchIndex          string           `koanf:"openSearchIndex"`
	IKSAggregationFieldNames []string         `koanf:"iKSAggregationFieldNames"`
	ClusterRegion            string           `koanf:"clusterRegion"`
	IKSFilterFieldNames      []string         `koanf:"iKSFilterFieldNames"`
	UseProxy                 bool             `koanf:"useProxy"`
}
