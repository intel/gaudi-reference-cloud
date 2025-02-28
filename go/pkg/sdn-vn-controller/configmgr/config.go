// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package configmgr

import (
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/handlers"
)

type Config struct {
	OvnCentralCfg  OvnConfig          `koanf:"ovnCentral"`
	GrpcServerCfg  GrpcServerConfig   `koanf:"grpcServer"`
	DbCfg          DbConfig           `koanf:"database"`
	Database       manageddb.Config   `koanf:"database"`
	ServiceAreaCfg ServiceAreaConfig  `koanf:"serviceArea"`
	SecurityCfg    SecurityConfig     `koanf:"security"`
	GatewaysCfg    []handlers.Gateway `koanf:"gateways"`
}

func IsYes(input string) bool {
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
