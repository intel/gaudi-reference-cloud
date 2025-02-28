// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bmc

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

var _ Interface = (*VirtualBMC)(nil)

// VirtualBMC provides APIs for Virtual Redfish BMC
type VirtualBMC struct {
	BMC
}

func (c *VirtualBMC) SanitizeBMCBootOrder(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("VirtualBMC")
	log.Info("Virtual BMC, BMC Boot Order not supported.")
	return nil
}

func (c *VirtualBMC) ConfigureNTP(ctx context.Context) (err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.ConfigureNTP")
	log.Info("ConfigureNTP Not Supported for Virtual BMC")

	return nil
}
