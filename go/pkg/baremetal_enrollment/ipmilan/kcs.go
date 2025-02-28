// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Copyright © 2023 Intel Corporation
package ipmilan

import (
	"context"
	"fmt"

	ipmi "github.com/bougou/go-ipmi"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type KCSPolicyControlModeType uint8

const (
	// Reference: https://cdrdv2.intel.com/v1/dl/getContent/598340
	// You need “RDC Privileged User Internal” access from AGS for the document.

	// 3, Provisioning, Allow all commands through KCS interface
	AllowAll KCSPolicyControlModeType = 3
	// 4, Provisioned, Allowed IPMI commands from KCS channel are
	// defined in the Access Control List, once BIOS POST completes
	Restricted KCSPolicyControlModeType = 4
	// 5, Provisioned, After CORE-BIOS-DONE, no commands are allowed
	// through the KCS interface
	DenyAll KCSPolicyControlModeType = 5
	// iDRAC Allow all
	IDRACAllow KCSPolicyControlModeType = 1
	// iDRAC Deny all
	IDRACDisable KCSPolicyControlModeType = 0
	// Quanta Allow all
	QuantaAllow KCSPolicyControlModeType = 3
	// Quanta Deny all
	QuantaDisable KCSPolicyControlModeType = 5
)

// Map the enum to String
func (kcsPolicy KCSPolicyControlModeType) String() string {
	switch kcsPolicy {
	case AllowAll:
		return "Allow All"
	case Restricted:
		return "Restricted"
	case DenyAll:
		return "Deny All"
	}
	return "Unknown"
}

func (h *ipmiLanHelper) VerifyKCSMode(ctx context.Context, expectedKcsMode KCSPolicyControlModeType, netFn ipmi.NetFn, cmd uint8) error {
	log := log.FromContext(ctx).WithName("IpmiLan.VerifySecuredKCS")
	log.Info(fmt.Sprintf("Verifying KCS Policy Control Mode is '%s'", expectedKcsMode))

	// Read back the current KCS Policy Control Mode and check if it matches the desired Mode

	// Send IPMI command
	rawResults, err := h.client.RawCommand(netFn, cmd, nil, "Intel General Application: Get KCS Policy Control Mode")
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Check KCS Policy Control Mode Results: '%s'", rawResults.Format()))

	if len(rawResults.Response) < 1 {
		return fmt.Errorf("Check KCS Policy Control mode: Response too short.")
	}

	if rawResults.Response[0] != byte(expectedKcsMode) {
		return fmt.Errorf("KCS Policy Control Mode is NOT set to '%s'", expectedKcsMode)
	}

	log.Info(fmt.Sprintf("KCS Policy Control Mode verified as: '%s'", expectedKcsMode))

	return nil
}
