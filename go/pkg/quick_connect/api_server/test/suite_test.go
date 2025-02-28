// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quick_connect/api_server/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	ctx    context.Context
	cancel context.CancelFunc
	router *gin.Engine
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

var _ = BeforeSuite(func() {
	log.SetDefaultLogger()
	ctx, cancel = context.WithCancel(context.TODO())
	log := log.FromContext(ctx)
	_ = log

	cfg := api.Config{}

	var err error
	router, err = api.NewRouter(ctx, &cfg, nil, nil)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	cancel()
})
