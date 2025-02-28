// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/universe_deployer/multi_version_manifests_generator:multi_version_manifests_generator_test" make test-custom
package multi_version_manifests_generator

import (
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multi Version Manifests Generator Test Suite")
}

var _ = BeforeSuite(func() {
	log.SetDefaultLogger()
})
