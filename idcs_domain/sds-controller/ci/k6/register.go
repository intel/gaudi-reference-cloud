// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k6

import (
	weka "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/ci/k6/modules"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/intel_internal/weka/v4", weka.WekaV4())
}
