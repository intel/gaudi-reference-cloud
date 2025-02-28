// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var cloudmonitorClient pb.CloudMonitorServiceClient
var test Test

func TestMain(m *testing.M) {
	test = Test{}
	test.Init()
	defer test.Done()

	// Single client used for testing the APIs
	cloudmonitorClient = pb.NewCloudMonitorServiceClient(test.clientConn)
	m.Run()
}
