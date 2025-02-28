// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/security-scanner/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/security-scanner/pkg/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

func main() {
	ctx := context.Background()
	// Initialize logger.
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("security-scanner")
	logger.Info("kube security scanner scheduler starts")

	listDir()
	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	scanService := server.SecurityScanService{}
	cfg := config.NewDefaultConfig()
	if err := loadConfig(ctx, cfg); err != nil {
		logger.Error(err, "error loading config", "err", err)
		// Keep going. The caller may decide to provide reasonable defaults.
	}
	token, err := os.ReadFile(cfg.GithubKey)
	if err != nil {
		logger.Info("unable to read GithubKey file %s: %v", cfg.GithubKey, err)
	}

	logger.Info("Run", "cfg", cfg)
	if err := os.Setenv("GITHUB_TOKEN", string(token)); err != nil {
		logger.Error(err, "failed to set env GITHUB_TOKEN")
	}
	if err := scanService.Init(ctx, cfg); err != nil {
		logger.Error(err, "failed to initialize security scanner service")
	}
}

func loadConfig(ctx context.Context, cfg *config.Config) error {
	log.BindFlags()
	configFile := ""
	flag.StringVar(&configFile, "config", "", "config file")
	flag.Parse()
	if configFile == "" {
		return fmt.Errorf("config flag can't be an empty string")
	}
	if err := conf.LoadConfigFile(ctx, configFile, cfg); err != nil {
		return fmt.Errorf("error loading config file (%s): %w", configFile, err)
	}
	return nil
}

func listDir() {
	rootDir := "/" // Start from the current directory

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() {
			fmt.Printf("Directory: %s\n", path)
		} else {
			fmt.Printf("File: %s (Size: %d bytes)\n", path, info.Size())
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", rootDir, err)
	}
}
