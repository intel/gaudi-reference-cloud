// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vulns

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/common"
)

type Scanner interface {
	Init(common.ScannerConfig) error
	ScanImage(context.Context, string) (common.VulnerabilityData, error)
}
