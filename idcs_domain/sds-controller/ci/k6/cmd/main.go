// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	_ "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/ci/k6"
	k6cmd "go.k6.io/k6/cmd"
)

func main() {
	k6cmd.Execute()
}
