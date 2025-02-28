// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quick_connect/api_server/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config Tests", func() {
	It("Should load the TLS config", func() {
		var cfg api.Config
		err := conf.LoadConfigFile(context.Background(), "config_test.yaml", &cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.ListenPort).Should(Equal(uint16(30410)))
		Expect(cfg.ClientCertificate.CommonName).Should(Equal("dev.quick-connect.us-dev-1a.cloud.intel.com.kind.local"))
		Expect(cfg.ClientCertificate.TTL).Should(Equal(2 * time.Minute))
	})
})
