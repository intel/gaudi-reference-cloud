// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

type NetworkProviderConfig struct {
	MaxConcurrentReconciles int    `koanf:"maxConcurrentReconciles"`
	NetworkAPIServerAddr    string `koanf:"networkApiServerAddr"`
	SDNServerAddr           string `koanf:"sdnServerAddr"`
	Region                  string `koanf:"region"`
}
