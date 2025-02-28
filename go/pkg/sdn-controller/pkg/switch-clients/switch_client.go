// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package switchclients

import (
	"context"
	"fmt"

	idcv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
)

type UpdateVlanRequest struct {
	Vlan       int32
	SwitchFQDN string
	PortName   string
	Env        string
	UpdateLLDP bool
}

type UpdateTrunkGroupsRequest struct {
	TrunkGroups []string
	SwitchFQDN  string
	PortName    string
}

type UpdateNativeVlanRequest struct {
	NativeVlan int32
	SwitchFQDN string
	PortName   string
}

type UpdateModeRequest struct {
	Mode       string
	SwitchFQDN string
	PortName   string
	//Env        string
}

type UpdateDescriptionRequest struct {
	Description string
	SwitchFQDN  string
	PortName    string
}

type UpdateBGPCommunityRequest struct {
	BGPCommunityIncomingGroupName string
	BGPCommunity                  int32
}

type GetBGPCommunityRequest struct {
	BGPCommunityIncomingGroupName string
}

type GetSwitchPortsRequest struct {
	SwitchFQDN string
	Env        string
}

type ListMacAddressTableRequest struct {
	SwitchFQDN string
	Env        string
}

type ListVlansParamsRequest struct {
	SwitchFQDN string
}

type ListPortParamsRequest struct {
	SwitchFQDN string
}

type PortParamsRequest struct {
	SwitchFQDN  string
	SwitchPort  string
	PortChannel int
}

type ParamsRequest struct {
	SwitchFQDN string
}

/* Port-Channel related */
type GetPortChannelsRequest struct {
	SwitchFQDN string
}

type CreatePortChannelRequest struct {
	PortChannel int64
}

type UpdatePortChannelVlanRequest struct {
	Vlan        int32
	PortChannel int32
	Description string
}

type UpdatePortChannelModeRequest struct {
	Mode        string
	PortChannel int32
}

type UpdatePortChannelTrunkGroupRequest struct {
	TrunkGroups []string
	NativeVlan  int32
	SwitchFQDN  string
	PortName    string
	Description string
}

type DeletePortChannelRequest struct {
	TargetPortChannel int64
}

type AssignSwitchPortToPortChannelRequest struct {
	SwitchPort        string
	TargetPortChannel int64
}

type RemoveSwitchPortFromPortChannelRequest struct {
	SwitchPort string
}

func convertSwitchPortStatusToResPortInfo(interf *idcv1alpha1.SwitchPortStatus) ResPortInfo {
	entry := ResPortInfo{}

	entry.InterfaceName = interf.Name
	entry.Port = utils.ConvertStandardPortNameToRavenFormat(interf.Name)
	entry.Mode = interf.Mode
	if interf.PortChannel != 0 {
		entry.PortChannel = fmt.Sprintf("%d", interf.PortChannel)
		entry.Mode = "portchannel"
	}
	entry.VlanId = int(interf.VlanId)
	entry.LinkStatus = interf.LinkStatus
	entry.Bandwidth = int64(interf.Bandwidth)
	entry.Duplex = interf.Duplex
	entry.TrunkGroups = interf.TrunkGroups
	if interf.Mode == "trunk" {
		entry.NativeVlan = int(interf.NativeVlan)
		entry.UntaggedVlan = int(interf.NativeVlan)
	}
	if interf.Mode == "access" {
		entry.UntaggedVlan = entry.VlanId
	}
	entry.Description = interf.Description
	entry.SwitchSideLastStatusChangeTimestamp = interf.SwitchSideLastStatusChangeTimestamp

	return entry
}

type SwitchClient interface {
	/* SwitchPort related */
	// UpdateVlan update vlan for a switch port
	UpdateVlan(context.Context, UpdateVlanRequest) error
	// UpdateMode changes the mode of a switchport to either "access" or "trunk"
	UpdateMode(context.Context, UpdateModeRequest) error
	// UpdateDescription updates description for a switch port
	UpdateDescription(context.Context, UpdateDescriptionRequest) error
	// UpdateTrunkGroups changes the trunkGroups or associated with a switch port or portchannel.
	UpdateTrunkGroups(context.Context, UpdateTrunkGroupsRequest) error
	// UpdateNativeVlan changes the nativeVlan associated with a switch port or portchannel.
	UpdateNativeVlan(context.Context, UpdateNativeVlanRequest) error
	// GetSwitchPorts return a list of port for a switch. key: port name (eg, "Ethernet27/1")
	GetSwitchPorts(context.Context, GetSwitchPortsRequest) (map[string]*idcv1alpha1.SwitchPortStatus, error)

	GetMacAddressTable(ctx context.Context, req ListMacAddressTableRequest) ([]ResMacAddressTableEntry, error)
	GetLLDPNeighbors(ctx context.Context, req PortParamsRequest) ([]ResLLDPNeighbors, error)
	GetLLDPPortNeighbors(ctx context.Context, req PortParamsRequest) ([]ResLLDPPortNeighbors, error)
	SaveConfig(ctx context.Context, fqdn string) (string, error)
	GetPortRunningConfig(ctx context.Context, req PortParamsRequest) ([]string, error)
	GetPortDetails(ctx context.Context, req PortParamsRequest) (ResPortInfo, error)
	ListPortsDetails(ctx context.Context, req ListPortParamsRequest) ([]ResPortInfo, error)
	ListVlans(ctx context.Context, req ListVlansParamsRequest) ([]VlanWithTrunkGroups, error)
	GetIpMacInfo(ctx context.Context, req ParamsRequest) ([]ResIpMacInfo, error)
	ClearMacAddressTable(ctx context.Context, fqdn string) (string, error)

	/* Port-Channel related */
	// AssignSwitchPortToPortChannel assigns a switch port to a specific port-channel
	AssignSwitchPortToPortChannel(context.Context, AssignSwitchPortToPortChannelRequest) error
	// RemoveSwitchPortFromPortChannel removes a switch port from a port-channel
	RemoveSwitchPortFromPortChannel(context.Context, RemoveSwitchPortFromPortChannelRequest) error

	// GetPortChannels
	GetPortChannels(context.Context, GetPortChannelsRequest) (map[string]PortChannel, error)
	// CreatePortChannel
	CreatePortChannel(context.Context, CreatePortChannelRequest) error
	// DeletePortChannel
	DeletePortChannel(context.Context, DeletePortChannelRequest) error

	//// UpdatePortChannelMode
	//UpdatePortChannelMode(context.Context, UpdatePortChannelModeRequest) error
	// Not used anywhere - we can control port channels exactly the same as regular ports (ie. using UpdateVlan, UpdateTrunkGroup, etc. above)
	// UpdatePortChannelVlan
	//UpdatePortChannelVlan(context.Context, UpdatePortChannelVlanRequest) error
	//// UpdatePortChannelTrunkGroup
	//UpdatePortChannelTrunkGroup(context.Context, UpdatePortChannelTrunkGroupRequest) error

	/* BGP related */
	// UpdateBGPCommunity update the BGP community for the switch
	UpdateBGPCommunity(context.Context, UpdateBGPCommunityRequest) error
	// GetBGPCommunity get the BGP community for the switch
	GetBGPCommunity(context.Context, GetBGPCommunityRequest) (int, error) // TODO: Should response be idcnetworkv1alpha1.BGPCommunityStatus or similar, for consistency with GetSwitchPorts?

	// RefreshConnection try to get the latest credential to recreate the connection
	RefreshConnection() error
	// ValidateConnection checks that we can connect and run basic commands. Returns nil on success, or error.
	ValidateConnection() error

	GetHost() (string, error)
}
