// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer_config

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
)

// Type for parsing deployment/universe_deployer/config.yaml.
type Config struct {
	// TODO: Use EnvConfig instead.
	HelmRepositories map[string]HelmRepository `yaml:"helmRepositories"`
}

type HelmRepository struct {
	Registry string `yaml:"registry"`
}

func NewConfigFromFile(ctx context.Context, filename string) (*Config, error) {
	var cfg Config
	if err := conf.LoadConfigFile(ctx, filename, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
