// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/apiservice"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/apiservice/config"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	configFile := "/config.json"
	config, err := config.GetConfig(configFile)
	if err != nil {
		log.Error(err, "Failed to get the config\nError: %s \nExiting...")
		os.Exit(1)
	}
	var router *gin.Engine
	router, err = api.NewRouter(ctx, config)
	if err != nil {
		log.Error(err, "Failed to start the Router\nError: %s \nExiting...")
		os.Exit(1)
	}
	err = router.Run(":8080")
	if err != nil {
		log.Error(err, "Failed to run the Router\nError: %s \nExiting...")
		os.Exit(1)
	}
}
