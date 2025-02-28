// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package pbconvert

import (
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func baseline() (instancePrivateFull *pb.InstancePrivate, instancePrivate *pb.InstancePrivate, instance *pb.Instance) {
	// Identical to instancePrivate except that it has values for all fields.
	instancePrivateFull = &pb.InstancePrivate{
		Metadata: &pb.InstanceMetadataPrivate{
			CloudAccountId:    "CloudAccountId1",
			Name:              "Name1",
			ResourceId:        "ResourceId1",
			ResourceVersion:   "ResourceVersion1",
			CreationTimestamp: timestamppb.New(time.Unix(1600000000, 0).UTC()),
			DeletionTimestamp: timestamppb.New(time.Unix(1700000000, 0).UTC()),
		},
		Spec: &pb.InstanceSpecPrivate{
			AvailabilityZone:  "AvailabilityZone1",
			InstanceType:      "InstanceType1",
			MachineImage:      "MachineImage1",
			RunStrategy:       pb.RunStrategy_Halted,
			SshPublicKeyNames: []string{"SshPublicKeyName1", "SshPublicKeyName2"},
			Interfaces: []*pb.NetworkInterfacePrivate{{
				Name:    "Interface1",
				VNet:    "VNet1",
				DnsName: "DnsName1",
			}},
			InstanceTypeSpec: &pb.InstanceTypeSpec{
				Name:             "InstanceTypeSpecName1",
				InstanceCategory: pb.InstanceCategory_VirtualMachine,
				Cpu: &pb.CpuSpec{
					Cores:     4,
					Sockets:   1,
					Threads:   2,
					Id:        "0x806F2",
					ModelName: "ModelName1",
				},
				Description: "Description1",
				Disks: []*pb.DiskSpec{
					{Size: "100Gi"},
				},
				DisplayName: "Tiny VM",
				Memory: &pb.MemorySpec{
					DimmCount: 2,
					Speed:     3200,
					DimmSize:  "8Gi",
					Size:      "16Gi",
				},
			},
			MachineImageSpec: &pb.MachineImageSpec{},
			SshPublicKeySpecs: []*pb.SshPublicKeySpec{
				{
					SshPublicKey: "SshPublicKey1",
				},
			},
			ClusterGroupId: "ClusterGroupId1",
			ClusterId:      "ClusterId1",
		},
		Status: &pb.InstanceStatusPrivate{
			Phase:   pb.InstancePhase_Ready,
			Message: "Message1",
			Interfaces: []*pb.InstanceInterfaceStatusPrivate{
				{
					Name:         "InterfaceName1",
					VNet:         "VNet1",
					DnsName:      "DnsName1",
					PrefixLength: 24,
					Addresses:    []string{"1.2.3.4"},
					Subnet:       "1.2.3.0",
					Gateway:      "1.2.3.1",
					VlanId:       1001,
				},
			},
			SshProxy: &pb.SshProxyTunnelStatus{
				ProxyUser:    "ProxyUser1",
				ProxyAddress: "ProxyAddress1",
				ProxyPort:    2222,
			},
		},
	}
	// Equivalent to instance.
	instancePrivate = &pb.InstancePrivate{
		Metadata: &pb.InstanceMetadataPrivate{
			CloudAccountId:    "CloudAccountId1",
			Name:              "Name1",
			ResourceId:        "ResourceId1",
			ResourceVersion:   "ResourceVersion1",
			CreationTimestamp: timestamppb.New(time.Unix(1600000000, 0).UTC()),
			DeletionTimestamp: timestamppb.New(time.Unix(1700000000, 0).UTC()),
		},
		Spec: &pb.InstanceSpecPrivate{
			AvailabilityZone:  "AvailabilityZone1",
			InstanceType:      "InstanceType1",
			MachineImage:      "MachineImage1",
			RunStrategy:       pb.RunStrategy_Halted,
			SshPublicKeyNames: []string{"SshPublicKeyName1", "SshPublicKeyName2"},
			Interfaces: []*pb.NetworkInterfacePrivate{{
				Name: "Interface1",
				VNet: "VNet1",
			}},
		},
		Status: &pb.InstanceStatusPrivate{
			Phase:   pb.InstancePhase_Ready,
			Message: "Message1",
			Interfaces: []*pb.InstanceInterfaceStatusPrivate{
				{
					Name:         "InterfaceName1",
					VNet:         "VNet1",
					DnsName:      "DnsName1",
					PrefixLength: 24,
					Addresses:    []string{"1.2.3.4"},
					Subnet:       "1.2.3.0",
					Gateway:      "1.2.3.1",
				},
			},
			SshProxy: &pb.SshProxyTunnelStatus{
				ProxyUser:    "ProxyUser1",
				ProxyAddress: "ProxyAddress1",
				ProxyPort:    2222,
			},
		},
	}
	// Equivalent to instancePrivate.
	instance = &pb.Instance{
		Metadata: &pb.InstanceMetadata{
			CloudAccountId:    "CloudAccountId1",
			Name:              "Name1",
			ResourceId:        "ResourceId1",
			ResourceVersion:   "ResourceVersion1",
			CreationTimestamp: timestamppb.New(time.Unix(1600000000, 0).UTC()),
			DeletionTimestamp: timestamppb.New(time.Unix(1700000000, 0).UTC()),
		},
		Spec: &pb.InstanceSpec{
			AvailabilityZone:  "AvailabilityZone1",
			InstanceType:      "InstanceType1",
			MachineImage:      "MachineImage1",
			RunStrategy:       pb.RunStrategy_Halted,
			SshPublicKeyNames: []string{"SshPublicKeyName1", "SshPublicKeyName2"},
			Interfaces: []*pb.NetworkInterface{{
				Name: "Interface1",
				VNet: "VNet1",
			}},
		},
		Status: &pb.InstanceStatus{
			Phase:   pb.InstancePhase_Ready,
			Message: "Message1",
			Interfaces: []*pb.InstanceInterfaceStatus{
				{
					Name:         "InterfaceName1",
					VNet:         "VNet1",
					DnsName:      "DnsName1",
					PrefixLength: 24,
					Addresses:    []string{"1.2.3.4"},
					Subnet:       "1.2.3.0",
					Gateway:      "1.2.3.1",
				},
			},
			SshProxy: &pb.SshProxyTunnelStatus{
				ProxyUser:    "ProxyUser1",
				ProxyAddress: "ProxyAddress1",
				ProxyPort:    2222,
			},
		},
	}
	return
}

var _ = Describe("InstancePrivate to Instance", func() {
	It("instancePrivateFull to Instance should succeed", func() {
		converter := NewPbConverter()
		instancePrivateFull, _, instanceExpected := baseline()
		instanceActual := &pb.Instance{}
		Expect(converter.Transcode(instancePrivateFull, instanceActual)).Should(Succeed())
		diff := cmp.Diff(instanceActual, instanceExpected, protocmp.Transform())
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})

	It("instancePrivate to Instance should succeed", func() {
		converter := NewPbConverter()
		_, instancePrivate, instanceExpected := baseline()
		instanceActual := &pb.Instance{}
		Expect(converter.Transcode(instancePrivate, instanceActual)).Should(Succeed())
		diff := cmp.Diff(instanceActual, instanceExpected, protocmp.Transform())
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})
})

var _ = Describe("Instance to InstancePrivate", func() {
	It("instance to InstancePrivate should succeed", func() {
		converter := NewPbConverter()
		_, instancePrivateExpected, instance := baseline()
		instancePrivateActual := &pb.InstancePrivate{}
		Expect(converter.Transcode(instance, instancePrivateActual)).Should(Succeed())
		diff := cmp.Diff(instancePrivateActual, instancePrivateExpected, protocmp.Transform())
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})
})
