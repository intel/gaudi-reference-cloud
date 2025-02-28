// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"time"
)

func (c *Config) CreatePrometheusSecretFetcher(log logr.Logger) SecretFetcher {
	return func() (username, password string, err error) {
		username, password, err = ReadSecret(c.MetricsConfig.PromCredentialsPath, log)
		return username, password, errors.Wrap(err, "failed to fetch credentials")
	}
}

func NewConfig(log logr.Logger) *Config {
	logger := log.WithName("Config")
	config := Config{}
	config.MetricsConfig = MetricsConfig{
		GetCredentials: config.CreatePrometheusSecretFetcher(logger),
	}

	return &config
}

type SecretFetcher func() (username, password string, err error)

type MetricsConfig struct {
	MetricsPort         uint16        `koanf:"metricsPort"`
	PromCredentialsPath string        `koanf:"promCredentialsPath"`
	Enabled             bool          `koanf:"enabled"`
	GetCredentials      SecretFetcher `json:"-"`
}

type Config struct {
	Region                          string            `koanf:"region"`
	ListenPort                      uint16            `koanf:"listenPort"`
	UsageServerAddr                 string            `koanf:"usageServerAddr"`
	UsageServerTimeout              time.Duration     `koanf:"usageServerTimeout"`
	DispatcherServerAddr            string            `koanf:"dispatcherServerAddr"`
	ProductCatalogServerAddr        string            `koanf:"productCatalogServerAddr"`
	GrpcHealthCheckExecutionPeriod  time.Duration     `koanf:"grpcHealthCheckExecutionPeriod"`
	GrpcHealthCheckExecutionTimeout time.Duration     `koanf:"grpcHealthCheckExecutionTimeout"`
	ModelsProductIds                map[string]string `koanf:"modelsProductIds"`
	RetryAttempts                   uint              `koanf:"retryAttempts"`
	FamilyId                        string            `koanf:"familyId"`
	MetricsConfig                   MetricsConfig     `koanf:"metrics"`
}
