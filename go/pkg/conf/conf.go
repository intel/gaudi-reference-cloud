// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package conf

import (
	"context"
	"fmt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type TLSConfig struct {
	UseTLS   bool   `koanf:"useTLS"`
	CertFile string `koanf:"certFile"`
	KeyFile  string `koanf:"keyFile"`
}

type ConfigConstructor interface {
	Construct()
}

// Load the configuration from the provided yaml file.
//
// Example:
//
//   var cfg grpc_rest_gateway.Config
//   if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
// 	   return err
//   }

func LoadConfigFile(ctx context.Context, filePath string, cfg interface{}) error {
	var k = koanf.New(".")
	// Read and parse YAML file.
	if filePath != "" {
		if err := k.Load(file.Provider(filePath), yaml.Parser()); err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}
	// Unmarshal into the configuration object.
	if err := k.Unmarshal("", &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal the configuration: %w", err)
	}
	if constructor, ok := cfg.(ConfigConstructor); ok {
		constructor.Construct()
	}
	return nil
}

func GetKubeRestConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}
