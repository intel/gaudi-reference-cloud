// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/universe_deployer/hasher:hasher_test" make test-custom
package hasher

import (
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hasher Test Suite")
}

var _ = BeforeSuite(func() {
	log.SetDefaultLogger()
})
