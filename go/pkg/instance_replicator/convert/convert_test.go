// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package convert

import (
	"time"

	"github.com/google/go-cmp/cmp"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func baseline() (*pb.InstancePrivate, *privatecloudv1alpha1.Instance) {
	creationTimestamp := time.Unix(1600000000, 0).UTC()
	deletionTimestamp := time.Unix(1700000000, 0).UTC()
	k8sDeletionTimestamp := metav1.NewTime(deletionTimestamp)
	pbInstance := &pb.InstancePrivate{
		Metadata: &pb.InstanceMetadataPrivate{
			CloudAccountId:  "CloudAccountId1",
			Name:            "Name1",
			ResourceId:      "ResourceId1",
			ResourceVersion: "ResourceVersion1",
			Labels: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			CreationTimestamp: timestamppb.New(creationTimestamp),
			DeletionTimestamp: timestamppb.New(deletionTimestamp),
		},
		Spec: &pb.InstanceSpecPrivate{
			AvailabilityZone:  "AvailabilityZone1",
			InstanceType:      "InstanceType1",
			MachineImage:      "MachineImage1",
			RunStrategy:       pb.RunStrategy_Halted,
			SshPublicKeyNames: []string{"SshPublicKeyName1", "SshPublicKeyName2"},
			Interfaces: []*pb.NetworkInterfacePrivate{{
				Name:        "Interface1",
				VNet:        "VNet1",
				DnsName:     "DnsName1",
				Nameservers: []string{"nameserver1"},
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
				Gpu: &pb.GpuSpec{
					Count:     0,
					ModelName: "",
				},
				HbmMode:     "",
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
			SuperComputeGroupId: "SuperComputeGroupId1",
			ClusterGroupId:      "ClusterGroupId1",
			ClusterId:           "ClusterId1",
			Region:              "Region1",
			NodeId:              "NodeId1",
			ServiceType:         pb.InstanceServiceType_ComputeAsAService,
			TopologySpreadConstraints: []*pb.TopologySpreadConstraints{
				{
					LabelSelector: &pb.LabelSelector{
						MatchLabels: map[string]string{
							"key3": "value3",
							"key4": "value4",
						},
					},
				},
			},
			Partition:           "Partition1",
			UserData:            "#cloud-config",
			InstanceGroup:       "InstanceGroup1",
			ComputeNodePools:    []string{"pool1"},
			QuickConnectEnabled: pb.TriState_Undefined,
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
					Subnet:       "Subnet1",
					Gateway:      "Gateway1",
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
	k8sInstance := &privatecloudv1alpha1.Instance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Instance",
			APIVersion: "private.cloud.intel.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "ResourceId1",
			Namespace:         "CloudAccountId1",
			ResourceVersion:   "ResourceVersion1",
			CreationTimestamp: metav1.NewTime(creationTimestamp),
			DeletionTimestamp: &k8sDeletionTimestamp,
			Labels: map[string]string{
				"cloud-account-id":      "CloudAccountId1",
				"instance-category":     string(privatecloudv1alpha1.InstanceCategoryVirtualMachine),
				"supercompute-group-id": "SuperComputeGroupId1",
				"cluster-group-id":      "ClusterGroupId1",
				"cluster-id":            "ClusterId1",
				"region":                "Region1",
				"node-id":               "NodeId1",
				"instance-group":        "InstanceGroup1",
			},
		},
		Spec: privatecloudv1alpha1.InstanceSpec{
			AvailabilityZone:  "AvailabilityZone1",
			InstanceType:      "InstanceType1",
			MachineImage:      "MachineImage1",
			RunStrategy:       "Halted",
			SshPublicKeyNames: []string{"SshPublicKeyName1", "SshPublicKeyName2"},
			Interfaces: []privatecloudv1alpha1.InterfaceSpec{{
				Name:        "Interface1",
				VNet:        "VNet1",
				DnsName:     "DnsName1",
				Nameservers: []string{"nameserver1"},
			}},
			InstanceTypeSpec: privatecloudv1alpha1.InstanceTypeSpec{
				Name:             "InstanceTypeSpecName1",
				DisplayName:      "Tiny VM",
				Description:      "Description1",
				InstanceCategory: privatecloudv1alpha1.InstanceCategoryVirtualMachine,
				Cpu: privatecloudv1alpha1.CpuSpec{
					Cores:     4,
					Id:        "0x806F2",
					ModelName: "ModelName1",
					Sockets:   1,
					Threads:   2,
				},
				Gpu: privatecloudv1alpha1.GpuSpec{
					Count: 0,
				},
				HBMMode: "",
				Memory: privatecloudv1alpha1.MemorySpec{
					Size:      "16Gi",
					DimmSize:  "8Gi",
					DimmCount: 2,
					Speed:     3200,
				},
				Disks: []privatecloudv1alpha1.DiskSpec{
					{
						Size: "100Gi",
					},
				},
			},
			MachineImageSpec: privatecloudv1alpha1.MachineImageSpec{
				Name: ""},
			SshPublicKeySpecs: []privatecloudv1alpha1.SshPublicKeySpec{
				{
					SshPublicKey: "SshPublicKey1",
				},
			},
			SuperComputeGroupId: "SuperComputeGroupId1",
			ClusterGroupId:      "ClusterGroupId1",
			ClusterId:           "ClusterId1",
			Region:              "Region1",
			NodeId:              "NodeId1",
			ServiceType:         pb.InstanceServiceType_ComputeAsAService.String(),
			InstanceName:        "Name1",
			TopologySpreadConstraints: []privatecloudv1alpha1.TopologySpreadConstraints{
				{
					LabelSelector: privatecloudv1alpha1.LabelSelector{
						MatchLabels: map[string]string{
							"key3": "value3",
							"key4": "value4",
						},
					},
				},
			},
			Labels: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			Partition:           "Partition1",
			UserData:            "#cloud-config",
			InstanceGroup:       "InstanceGroup1",
			ComputeNodePools:    []string{"pool1"},
			QuickConnectEnabled: "Undefined",
		},
		Status: privatecloudv1alpha1.InstanceStatus{
			Phase:   privatecloudv1alpha1.PhaseReady,
			Message: "Message1",
			Interfaces: []privatecloudv1alpha1.InstanceInterfaceStatus{
				{
					Name:         "InterfaceName1",
					VNet:         "VNet1",
					DnsName:      "DnsName1",
					PrefixLength: 24,
					Addresses:    []string{"1.2.3.4"},
					Subnet:       "Subnet1",
					Gateway:      "Gateway1",
					VlanId:       1001,
				},
			},
			SshProxy: privatecloudv1alpha1.SshProxyTunnelStatus{
				ProxyUser:    "ProxyUser1",
				ProxyAddress: "ProxyAddress1",
				ProxyPort:    2222,
			}}}
	return pbInstance, k8sInstance
}

var _ = Describe("PbToK8s", func() {
	It("Baseline should succeed", func() {
		converter := NewInstanceConverter()
		pbInstance, k8sInstanceExpected := baseline()
		k8sInstanceActual, err := converter.PbToK8s(pbInstance)
		Expect(err).Should(Succeed())
		diff := cmp.Diff(k8sInstanceActual, k8sInstanceExpected)
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})

	It("Nil DeletionTimestamp should succeed", func() {
		converter := NewInstanceConverter()
		pbInstance, k8sInstanceExpected := baseline()
		pbInstance.Metadata.DeletionTimestamp = nil
		k8sInstanceExpected.ObjectMeta.DeletionTimestamp = nil
		k8sInstanceActual, err := converter.PbToK8s(pbInstance)
		Expect(err).Should(Succeed())
		diff := cmp.Diff(k8sInstanceActual, k8sInstanceExpected)
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})
})

var _ = Describe("K8sToPb", func() {
	It("Baseline should succeed", func() {
		converter := NewInstanceConverter()
		pbInstanceExpected, k8sInstance := baseline()
		// Delete fields from expected value that do not get converted.
		pbInstanceExpected.Metadata.Name = ""
		pbInstanceActual, err := converter.K8sToPb(k8sInstance)
		Expect(err).Should(Succeed())
		diff := cmp.Diff(pbInstanceActual, pbInstanceExpected, protocmp.Transform())
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})

	It("Nil DeletionTimestamp should succeed", func() {
		converter := NewInstanceConverter()
		pbInstanceExpected, k8sInstance := baseline()
		pbInstanceExpected.Metadata.DeletionTimestamp = nil
		k8sInstance.ObjectMeta.DeletionTimestamp = nil
		// Delete fields from expected value that do not get converted.
		pbInstanceExpected.Metadata.Name = ""
		pbIstanceActual, err := converter.K8sToPb(k8sInstance)
		Expect(err).Should(Succeed())
		diff := cmp.Diff(pbIstanceActual, pbInstanceExpected, protocmp.Transform())
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})
})
