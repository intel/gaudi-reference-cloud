// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package switchclients

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	idcv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
)

type MockSwitchClient struct {
	sync.Mutex
	SwitchFQDN string

	// PortSet stores all the ports for a switch. Key is the port name
	PortSet              map[string]*idcv1alpha1.SwitchPort
	BGPCommunity         int
	AllowedVlanIds       []int
	AllowedNativeVlanIds []int
}

func NewMockSwitchClient(allSwitches map[string]SwitchClient, newSwitchFQDN string, allowedVlanIds []int, allowedNativeVlanIds []int) (*MockSwitchClient, error) {
	client := &MockSwitchClient{
		SwitchFQDN:           newSwitchFQDN,
		AllowedVlanIds:       allowedVlanIds,
		AllowedNativeVlanIds: allowedNativeVlanIds,
	}

	// simply generate 52 ports for this switch. from 1/1 to 52/1
	ports := make(map[string]*idcv1alpha1.SwitchPort)
	for i := 1; i <= 64; i++ {
		for j := 1; j <= 9; j++ {
			portName := fmt.Sprintf("Ethernet%d/%d", i, j)
			vlan := 4008
			// assume that all accel switches are with -zas prefix, and their default vlan should be 100.
			if strings.Contains(newSwitchFQDN, "-zas") {
				vlan = 100
			}
			sp := utils.NewSwitchPortTemplate(newSwitchFQDN, portName, int64(vlan), int64(-1), nil)
			ports[portName] = sp
		}
	}

	client.PortSet = ports
	time.Sleep(800 * time.Millisecond)
	return client, nil
}

func (a *MockSwitchClient) GetHost() (string, error) {
	return a.SwitchFQDN, nil
}

func (a *MockSwitchClient) RefreshConnection() error {
	return nil
}

func (a *MockSwitchClient) UpdateMode(ctx context.Context, req UpdateModeRequest) error {
	logger := log.FromContext(ctx).WithName("MockSwitchClient.UpdateMode").WithValues(utils.LogFieldMode, req.Mode)

	time.Sleep(300 * time.Millisecond)
	actualPort, found := a.PortSet[req.PortName]
	if !found {
		return fmt.Errorf("port not found (mockSwitchClient)")
	}
	actualPort.Spec.Mode = req.Mode

	logger.V(1).Info("UpdateMode success! (not really- mocked)")
	return nil
}

func (a *MockSwitchClient) UpdateDescription(ctx context.Context, req UpdateDescriptionRequest) error {
	logger := log.FromContext(ctx).WithName("MockSwitchClient.UpdateDescription").WithValues(utils.LogFieldDescription, req.Description)

	time.Sleep(300 * time.Millisecond)
	actualPort, found := a.PortSet[req.PortName]
	if !found {
		return fmt.Errorf("port not found (mockSwitchClient)")
	}
	actualPort.Spec.Description = req.Description

	logger.V(1).Info("UpdateDescription success! (not really- mocked)")
	return nil
}

func (a *MockSwitchClient) UpdateVlan(ctx context.Context, req UpdateVlanRequest) error {
	logger := log.FromContext(ctx).WithName("MockSwitchClient.UpdateVlan")

	err := utils.ValidatePortValue(req.PortName)
	if err != nil {
		return fmt.Errorf("ValidatePortValue failed, error: %v", err)
	}

	err = utils.ValidateVlanValue(int(req.Vlan), a.AllowedVlanIds)
	if err != nil {
		return fmt.Errorf("ValidateVlanValue failed, error: %v", err)
	}

	time.Sleep(400 * time.Millisecond)
	a.Lock()
	defer a.Unlock()
	actualPort, found := a.PortSet[req.PortName]
	if !found {
		return fmt.Errorf("port not found (mockSwitchClient)")
	}
	actualPort.Spec.VlanId = int64(req.Vlan)
	a.PortSet[req.PortName] = actualPort

	logger.V(1).Info("UpdateVlan success! (not really- mocked)")
	return nil
}

func (a *MockSwitchClient) UpdateTrunkGroups(ctx context.Context, req UpdateTrunkGroupsRequest) error {
	logger := log.FromContext(ctx).WithName("MockSwitchClient.UpdateTrunkGroups").WithValues(utils.LogFieldTrunkGroups, req.TrunkGroups)

	time.Sleep(300 * time.Millisecond)
	actualPort, found := a.PortSet[req.PortName]
	if !found {
		return fmt.Errorf("port not found (mockSwitchClient)")
	}
	actualPort.Spec.TrunkGroups = &req.TrunkGroups

	logger.V(1).Info("UpdateTrunkGroups success! (not really- mocked)")
	return nil
}

func (a *MockSwitchClient) UpdateNativeVlan(ctx context.Context, req UpdateNativeVlanRequest) error {
	logger := log.FromContext(ctx).WithName("MockSwitchClient.UpdateNativeVlan").WithValues(utils.LogFieldNativeVlan, req.NativeVlan)

	time.Sleep(300 * time.Millisecond)

	err := utils.ValidateVlanValue(int(req.NativeVlan), a.AllowedNativeVlanIds)
	if err != nil {
		return fmt.Errorf("ValidateNativeVlanValue failed, error: %v", err)
	}

	actualPort, found := a.PortSet[req.PortName]
	if !found {
		return fmt.Errorf("port not found (mockSwitchClient)")
	}
	actualPort.Spec.NativeVlan = int64(req.NativeVlan)

	logger.V(1).Info("UpdateNativeVlan success! (not really- mocked)")
	return nil
}

func (a *MockSwitchClient) UpdateBGPCommunity(ctx context.Context, req UpdateBGPCommunityRequest) error {
	logger := log.FromContext(ctx).WithName("MockSwitchClient.UpdateBGPCommunity")

	err := utils.ValidateBGPCommunityValue(req.BGPCommunity)
	if err != nil {
		return fmt.Errorf("ValidateBGPCommunityValue failed, error: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	a.BGPCommunity = int(req.BGPCommunity)

	logger.V(1).Info("UpdateBGPCommunity success! (not really- mocked)")
	return nil
}

func (a *MockSwitchClient) GetBGPCommunity(ctx context.Context, req GetBGPCommunityRequest) (int, error) {
	time.Sleep(300 * time.Millisecond)
	return a.BGPCommunity, nil
}

func (a *MockSwitchClient) GetMacAddressTable(ctx context.Context, req ListMacAddressTableRequest) ([]ResMacAddressTableEntry, error) {
	var ret = make([]ResMacAddressTableEntry, 0)
	return ret, nil
}

func (a *MockSwitchClient) GetLLDPNeighbors(ctx context.Context, req PortParamsRequest) ([]ResLLDPNeighbors, error) {
	var ret = make([]ResLLDPNeighbors, 0)
	return ret, nil
}

func (a *MockSwitchClient) GetLLDPPortNeighbors(ctx context.Context, req PortParamsRequest) ([]ResLLDPPortNeighbors, error) {
	var ret = make([]ResLLDPPortNeighbors, 0)
	return ret, nil
}

func (a *MockSwitchClient) SaveConfig(ctx context.Context, fqdn string) (string, error) {
	return "", fmt.Errorf("SaveConfig not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) GetPortRunningConfig(ctx context.Context, req PortParamsRequest) ([]string, error) {
	var ret = make([]string, 0)
	return ret, nil
}

func (a *MockSwitchClient) GetPortDetails(ctx context.Context, req PortParamsRequest) (ResPortInfo, error) {
	return ResPortInfo{}, nil
}

func (a *MockSwitchClient) GetIpMacInfo(ctx context.Context, req ParamsRequest) ([]ResIpMacInfo, error) {
	var ret = make([]ResIpMacInfo, 0)
	return ret, nil
}

func (a *MockSwitchClient) ClearMacAddressTable(ctx context.Context, fqdn string) (string, error) {
	return "", fmt.Errorf("ClearMacAddressTable not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) GetSwitchPorts(ctx context.Context, req GetSwitchPortsRequest) (map[string]*idcv1alpha1.SwitchPortStatus, error) {
	time.Sleep(1100 * time.Millisecond)
	a.Lock()
	defer a.Unlock()
	res := make(map[string]*idcv1alpha1.SwitchPortStatus)
	for portName, port := range a.PortSet {
		res[portName] = &idcv1alpha1.SwitchPortStatus{
			Name:   port.Spec.Name,
			VlanId: port.Spec.VlanId,
			Mode:   "access",
		}
	}
	return res, nil
}

func (a *MockSwitchClient) showCurrentSwitches() {
	fmt.Printf("=================== Switches ===================== \n")
	fmt.Printf("== Switch: name: %v, BGPCommunity: %d \n", a.SwitchFQDN, a.BGPCommunity)
	fmt.Printf("=================== Switch Ports ===================== \n")
	for _, v := range a.PortSet {
		fmt.Printf("==== port name: %v, vlan: %v \n", v.Spec.Name, v.Spec.VlanId)
	}
	fmt.Printf("===================================================== \n")
}

func (a *MockSwitchClient) ValidateConnection() error {
	return nil
}

func (a *MockSwitchClient) ListPortsDetails(ctx context.Context, req ListPortParamsRequest) ([]ResPortInfo, error) {
	return nil, fmt.Errorf("ListPortsDetails not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) ListVlans(ctx context.Context, req ListVlansParamsRequest) ([]VlanWithTrunkGroups, error) {
	return nil, fmt.Errorf("ListVlans not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) performPortChannelModeUpdate(ctx context.Context, portChannel int32, mode string) error {
	return fmt.Errorf("performPortChannelModeUpdate not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) UpdatePortChannelMode(ctx context.Context, req UpdatePortChannelModeRequest) error {
	return fmt.Errorf("UpdatePortChannelMode not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) performVlanUpdate(ctx context.Context, portChannel int32, vlan int32, description string) error {
	return fmt.Errorf("performVlanUpdate not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) UpdatePortChannelVlan(ctx context.Context, req UpdatePortChannelVlanRequest) error {
	return fmt.Errorf("UpdatePortChannelVlan not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) DeletePortChannel(ctx context.Context, req DeletePortChannelRequest) error {
	return fmt.Errorf("DeletePortChannel not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) CreatePortChannel(ctx context.Context, req CreatePortChannelRequest) error {
	return fmt.Errorf("DeletePortChannel not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) GetPortChannels(ctx context.Context, req GetPortChannelsRequest) (map[string]PortChannel, error) {
	return nil, fmt.Errorf("DeletePortChannel not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) AssignSwitchPortToPortChannel(ctx context.Context, req AssignSwitchPortToPortChannelRequest) error {
	return fmt.Errorf("AssignSwitchPortToPortChannel not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) RemoveSwitchPortFromPortChannel(ctx context.Context, req RemoveSwitchPortFromPortChannelRequest) error {
	return fmt.Errorf("RemoveSwitchPortFromPortChannel not implemented by MockSwitchClient")
}

func (a *MockSwitchClient) UpdatePortChannelTrunkGroup(ctx context.Context, req UpdatePortChannelTrunkGroupRequest) error {
	return fmt.Errorf("UpdatePortChannelTrunkGroup not implemented by MockSwitchClient")
}
