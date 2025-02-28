// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"context"
	"log"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

// Application configuration
type Config struct {
	ListenPort        uint16            `koanf:"listenPort"`
	Database          manageddb.Config  `koanf:"database"`
	DefaultRegistry   OCIRegistryConfig `koanf:"defaultRegistry"`
	GrpcAPIServerAddr string            `koanf:"grpcAPIServerAddr"`
	Tls               TlsConfig         `koanf:"tls"`
	Dns               DnsConfig         `koanf:"dns"`
	Encryption        EncryptionConfig  `koanf:"encryption"`
}

type OCIRegistryConfig struct {
	Host         string `koanf:"host"`
	Username     string `koanf:"username"`
	PasswordFile string `koanf:"passwordFile"`
}

type EncryptionConfig struct {
	KeyFile string `koanf:"keyFile"`
}

type TlsConfig struct {
	HwSSLProfileID int `koanf:"hwSSLProfileID"`
}

type DnsConfig struct {
	Realm              string `koanf:"realm"`
	DnsIpamApiEndpoint string `koanf:"dnsIpamApiEndpoint"`
}

type IksConfig struct {
	IKSServerAddr string `koanf:"IKSServerAddr"`
}

func ReadConfig() (Config, error) {
	log.Println("Reading the config file")

	var cfg Config
	if err := conf.LoadConfigFile(context.Background(), "/config.yaml", &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// database:
//   host: localhost
//   port: 5432
//   dbname: dpai
//   username: postgres
//   passwordEnvKey: PG_PASSWORD
// registry:
//   dockerPersonal:
//     url: ghcr.io
//     username: venkadeshwarank
//     passwordEnvKey: GITHUB_TOKEN
//   docker:
//     url: amr-idc-registry-pre.infra-host.com
//     username: robotdpai+backend-api-dev
//     passwordEnvKey: HARBOR_TOKEN
//   helm:
//     reponame: dpai
//     url: https://amr-idc-registry-pre.infra-host.com/chartrepo/dpai
//     username: robotdpai+backend-api-dev
//     passwordEnvKey: HARBOR_TOKEN
