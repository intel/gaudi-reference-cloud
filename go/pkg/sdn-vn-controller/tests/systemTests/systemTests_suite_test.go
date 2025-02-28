// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package systemTests

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var configFile string

func init() {
	// The default config file is "config_local.json" but you can override it with "--config=config.yaml"
	flag.StringVar(&configFile, "config", "config_local.json", "Path to configuration file")
}

func TestSystemTests(t *testing.T) {
	RegisterFailHandler(Fail)
	// Configure the suite to run serially
	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.ParallelTotal = 1         // Ensure tests run serially
	suiteConfig.RandomizeAllSpecs = false // Disable randomization
	RunSpecs(t, "SystemTests Suite", suiteConfig, reporterConfig)
}

var _ = BeforeSuite(func() {
	var err error
	config, err = loadConfig(configFile) // Load configuration and create namespaces
	Expect(err).NotTo(HaveOccurred(), "Failed to load network configuration")
})

var _ = AfterSuite(func() {
	Expect(cleanupConfig()).To(Succeed(), "Failed to clean up namespaces")
})
