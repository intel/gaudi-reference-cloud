// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"sync"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

var httpServerExitDone = &sync.WaitGroup{}

func CreateWg() *sync.WaitGroup {
	httpServerExitDone.Add(1)
	return httpServerExitDone
}

func GetHttpServerExitDone() *sync.WaitGroup {
	return httpServerExitDone
}
func TestMain(m *testing.M) {
	err := common.Init()
	log.SetDefaultLoggerDebug()
	logger := log.FromContext(context.Background()).WithName("TestMain")
	if err != nil {
		logger.Error(err, "init failed. skipping tests\n")
		return
	}
	if config.Cfg.AriaSystem.AuthKey != "" {
		m.Run()
	}
}
