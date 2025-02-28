// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/oidc/server"
	"github.com/zitadel/oidc/example/server/storage"
)

func main() {
	fmt.Println("Starting OIDC Server ...")
	startOIDCServer()
}

func startOIDCServer() {
	ctx := context.Background()
	cfg := server.Config{}

	log.BindFlags()
	configFile := ""
	flag.StringVar(&configFile, "config", "", "config file")
	flag.Parse()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("RunService")

	if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
		logger.Error(err, "error loading config", "file", configFile)
		// Keep going. The caller may decide to provide reasonable defaults.
	}

	// the OpenIDProvider interface needs a Storage interface handling various checks and state manipulations
	// this might be the layer for accessing your database
	// in this example it will be handled in-memory
	userStore := storage.NewUserStore()
	storage := storage.NewStorage(userStore)

	port := strconv.FormatUint(uint64(cfg.ListenPort), 10)
	router := server.SetupServer(ctx, "http://localhost:"+port, storage)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	logger.Info("OIDC server listening on ", "port", port)
	err := server.ListenAndServe()
	if err != nil {
		logger.Error(err, "Error while running OIDC server")
		os.Exit(1)
	}
	<-ctx.Done()
}
