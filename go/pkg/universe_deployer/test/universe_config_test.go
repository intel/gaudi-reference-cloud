// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Universe Config Tests", func() {
	It("Trimmed", func() {
		ctx := context.Background()
		universeConfig, err := universe_config.NewUniverseConfigFromFile(ctx, "../testdata/universe_config_1.json")
		Expect(err).Should(Succeed())
		expected, err := universe_config.NewUniverseConfigFromFile(ctx, "../testdata/universe_config_1_trimmed_b966d5f3a3aa2762532d82a06e5ea0435fbc89d4.json")
		Expect(err).Should(Succeed())
		trimmed := universeConfig.Trimmed(ctx, "b966d5f3a3aa2762532d82a06e5ea0435fbc89d4")
		Expect(trimmed).Should(Equal(expected))
	})
})
