// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Copyright © 2023 Intel Corporation
package ipacmd

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gocarina/gocsv"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type GaudiPCIBus struct {
	ModuleID     string   `csv:"module_id"`
	BusID        string   `csv:"bus_id"`
	MacAddresses []string `csv:"macs"`
}

type LsHwNetwork struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	Claimed       bool   `json:"claimed"`
	Description   string `json:"description"`
	Physid        string `json:"physid"`
	Businfo       string `json:"businfo"`
	Logicalname   string `json:"logicalname"`
	Serial        string `json:"serial"`
	Configuration struct {
		Autonegotiation string `json:"autonegotiation"`
		Broadcast       string `json:"broadcast"`
		Driver          string `json:"driver"`
		Driverversion   string `json:"driverversion"`
		Duplex          string `json:"duplex"`
		Firmware        string `json:"firmware"`
		Link            string `json:"link"`
		Multicast       string `json:"multicast"`
		Port            string `json:"port"`
	} `json:"configuration"`
	Capabilities struct {
		Ethernet        bool   `json:"ethernet"`
		Physical        string `json:"physical"`
		Fibre           string `json:"fibre"`
		Autonegotiation string `json:"autonegotiation"`
	} `json:"capabilities"`
}

const (
	hlSmiCommand = "hl-smi -Q module_id,bus_id -f csv"
	lshwNetwork  = "sudo lshw -C network -json"
)

func (h *ipaCmdHelper) HabanaEthernetBusInfo(ctx context.Context) ([]*GaudiPCIBus, error) {
	log := log.FromContext(ctx).WithName("IpaCmd.HabanaEthernetBusInfo")

	// Connect to the remote host
	client, err := h.sshManager.Dial("tcp", fmt.Sprintf("%s:%d", h.ironicIp, sshConnectPort), h.clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	defer client.Close()

	// Create a new session on the SSH connection
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Run the command and capture the output
	output, err := session.CombinedOutput(hlSmiCommand)
	if err != nil {
		outStr := string(output)
		log.Error(err, "CombinedOutput failed", "hlSmiCommand", hlSmiCommand, "output", outStr)
		return nil, fmt.Errorf("failed to run command: %v", err)
	}

	gaudiBuses := []*GaudiPCIBus{}
	if err := gocsv.UnmarshalBytes(output, &gaudiBuses); err != nil {
		return nil, fmt.Errorf("failed to Unmarshal gaudiBus: %v", err)
	}
	return gaudiBuses, nil
}

func (h *ipaCmdHelper) HabanaEthernetMacAddress(ctx context.Context, gaudiBuses []*GaudiPCIBus) error {
	log := log.FromContext(ctx).WithName("IpaCmd.HabanaEthernetMacAddress")

	// Connect to the remote host
	client, err := h.sshManager.Dial("tcp", fmt.Sprintf("%s:%d", h.ironicIp, sshConnectPort), h.clientConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer client.Close()

	// Create a new session on the SSH connection
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Run the command and capture the output
	output, err := session.CombinedOutput(lshwNetwork)
	if err != nil {
		outStr := string(output)
		log.Error(err, "CombinedOutput failed", "lshwNetwork", lshwNetwork, "output", outStr)
		return fmt.Errorf("failed to run command: %v", err)
	}
	lshwNetworkData := []*LsHwNetwork{}
	if err := json.Unmarshal(output, &lshwNetworkData); err != nil {
		return fmt.Errorf("failed to Unmarshal : %v", err)
	}
	for i, bus := range gaudiBuses {
		for _, item := range lshwNetworkData {
			//fmt.Println(bus.BusID)
			if strings.Contains(strings.TrimSpace(item.Businfo), strings.TrimSpace(bus.BusID)) {
				gaudiBuses[i].MacAddresses = append(gaudiBuses[i].MacAddresses, item.Serial)

			}

		}
		sort.Strings(gaudiBuses[i].MacAddresses)
	}
	return nil
}
