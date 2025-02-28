// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/server"
)

func main() {
	server.StartMeteringService()
}
