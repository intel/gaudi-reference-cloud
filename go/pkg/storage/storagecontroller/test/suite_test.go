// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	log.SetDefaultLogger()
})

func TestStoragecontroller(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storagecontroller Suite")
}
