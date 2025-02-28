// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// Application configuration
type Config struct {
	ListenPort             uint16                 `koanf:"listenPort"`
	TLS                    conf.TLSConfig         `koanf:"tls"`
	CloudAccountDatabase   manageddb.Config       `koanf:"database"`
	ProductCatalogDatabase manageddb.Config       `koanf:"pcDatabase"`
	DefaultRegions         []pb.DefaultRegionSpec `koanf:"defaultRegions"`
}
