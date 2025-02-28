// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	test             Test
	client           pb.IksClient
	computeclient    pb.InstanceTypeServiceClient
	reconcilerClient pb.IksPrivateReconcilerClient
)

func TestMain(m *testing.M) {
	test = Test{}
	test.Init()
	defer test.Done()

	// Single client used for testing the APIs
	client = pb.NewIksClient(test.clientConn)
	reconcilerClient = pb.NewIksPrivateReconcilerClient(test.clientConn)
	m.Run()
}
