// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package builder

import (
	"fmt"

	kubevirtv1 "kubevirt.io/api/core/v1"
)

const (
	NetworkInterfaceTypeBridge     = "bridge"
	NetworkInterfaceTypeMasquerade = "masquerade"

	LabelKeyNetworkType    = "network.harvesterhci.io/type"
	LabelKeyClusterNetwork = "network.harvesterhci.io/clusternetwork"

	NetworkTypeVLAN           = "L2VlanNetwork"
	NetworkTypeCustom         = "Custom"
	networkVLANConfigTemplate = `{"cniVersion":"0.3.1","name":"%s","mtu":%d,"type":"bridge","bridge":"%s-br","promiscMode":true,"vlan":%d,"ipam":{}}`
)

func GetNetworkAttachmentDefinitionName(vlanId int) string {
	return fmt.Sprintf("%d", vlanId)
}

func CreateNetworkAttachmentDefinitionSpecStr(vlanId int, vmClusterNetwork string, mtu int) string {
	return fmt.Sprintf(networkVLANConfigTemplate, GetNetworkAttachmentDefinitionName(vlanId), mtu, vmClusterNetwork, vlanId)
}

func (v *VMBuilder) NetworkInterface(interfaceName, interfaceModel, interfaceMACAddress, interfaceType, networkName string) *VMBuilder {
	v.Interface(interfaceName, interfaceModel, interfaceMACAddress, interfaceType)
	v.Network(interfaceName, networkName)
	return v
}

func (v *VMBuilder) Network(interfaceName, networkName string) *VMBuilder {
	networks := v.VirtualMachine.Spec.Template.Spec.Networks
	network := kubevirtv1.Network{
		Name: interfaceName,
	}
	if networkName != "" {
		network.NetworkSource = kubevirtv1.NetworkSource{
			Multus: &kubevirtv1.MultusNetwork{
				NetworkName: networkName,
				Default:     false,
			},
		}
	} else {
		network.NetworkSource = kubevirtv1.NetworkSource{
			Pod: &kubevirtv1.PodNetwork{},
		}
	}
	networks = append(networks, network)
	v.VirtualMachine.Spec.Template.Spec.Networks = networks
	return v
}

func (v *VMBuilder) Interface(interfaceName, interfaceModel, interfaceMACAddress string, interfaceType string) *VMBuilder {
	interfaces := v.VirtualMachine.Spec.Template.Spec.Domain.Devices.Interfaces
	networkInterface := kubevirtv1.Interface{
		Name:       interfaceName,
		Model:      interfaceModel,
		MacAddress: interfaceMACAddress,
		InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
			Bridge: &kubevirtv1.InterfaceBridge{},
		},
	}
	switch interfaceType {
	case NetworkInterfaceTypeBridge:
		networkInterface.InterfaceBindingMethod = kubevirtv1.InterfaceBindingMethod{
			Bridge: &kubevirtv1.InterfaceBridge{},
		}
	default:
		networkInterface.InterfaceBindingMethod = kubevirtv1.InterfaceBindingMethod{
			Masquerade: &kubevirtv1.InterfaceMasquerade{},
		}
	}
	interfaces = append(interfaces, networkInterface)
	v.VirtualMachine.Spec.Template.Spec.Domain.Devices.Interfaces = interfaces
	return v
}

func (v *VMBuilder) WaitForLease(interfaceName string) *VMBuilder {
	v.WaitForLeaseInterfaceNames = append(v.WaitForLeaseInterfaceNames, interfaceName)
	return v
}
