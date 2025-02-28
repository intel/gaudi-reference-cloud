// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package conf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Sample configuration for testing.
type sampleConfig struct {
	ListenPort uint16          `koanf:"listenPort"`
	TLS        sampleTlsConfig `koanf:"tls"`
}

type sampleTlsConfig struct {
	UseTLS   bool   `koanf:"useTLS"`
	CertFile string `koanf:"certFile"`
	KeyFile  string `koanf:"keyFile"`
}

func TestConfig(t *testing.T) {
	var cfg sampleConfig
	if err := LoadConfigFile(context.Background(), "conf_test.yaml", &cfg); err != nil {
		assert.FailNow(t, "failed to load the configuration")
	}
	assert.Equal(t, uint16(8080), cfg.ListenPort)
	assert.Equal(t, true, cfg.TLS.UseTLS)
	assert.Equal(t, "server.crt", cfg.TLS.CertFile)
	assert.Equal(t, "server.key", cfg.TLS.KeyFile)
}
