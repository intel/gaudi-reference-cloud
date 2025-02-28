// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	instanceserver "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/instance"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ = Describe("InstanceSqlTransformer Tests", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	baseline := func() *pb.InstancePrivate {
		return &pb.InstancePrivate{
			Metadata: &pb.InstanceMetadataPrivate{
				CloudAccountId:    cloudaccount.MustNewId(),
				Name:              uuid.NewString(),
				ResourceId:        uuid.NewString(),
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
	}

	insertInstance := func(instance *pb.InstancePrivate) {
		sqlTransformer := instanceserver.NewInstanceSqlTransformer()
		flattened, err := sqlTransformer.Flatten(ctx, instance)
		Expect(err).Should(Succeed())
		query := fmt.Sprintf(`insert into instance (resource_id, cloud_account_id, name, %s) values ($1, $2, $3, %s)`,
			flattened.GetColumnsString(), flattened.GetInsertValuesString(4))
		args := append([]any{instance.Metadata.ResourceId, instance.Metadata.CloudAccountId, instance.Metadata.Name}, flattened.Values...)
		_, err = sqlDb.ExecContext(ctx, query, args...)
		Expect(err).Should(Succeed())
	}

	selectInstance := func(instance *pb.InstancePrivate) *pb.InstancePrivate {
		sqlTransformer := instanceserver.NewInstanceSqlTransformer()
		query := fmt.Sprintf(`select %s from instance where resource_id = $1 and cloud_account_id = $2`,
			sqlTransformer.ColumnsForFromRow())
		rows, err := sqlDb.QueryContext(ctx, query, instance.Metadata.ResourceId, instance.Metadata.CloudAccountId)
		Expect(err).Should(Succeed())
		defer rows.Close()
		Expect(rows.Next()).Should(BeTrue())
		selectedInstance, err := sqlTransformer.FromRow(ctx, rows)
		Expect(err).Should(Succeed())
		return selectedInstance
	}

	It("Insert of instance to database should succeed", func() {
		instance := baseline()
		insertInstance(instance)
	})

	It("Select of instance from database should succeed and return what was inserted", func() {
		insertedInstance := baseline()
		insertInstance(insertedInstance)
		selectedInstance := selectInstance(insertedInstance)
		// Do not compare ResourceVersion since it is not in insertedInstance.
		selectedInstance.Metadata.ResourceVersion = insertedInstance.Metadata.ResourceVersion
		diff := cmp.Diff(selectedInstance, insertedInstance, protocmp.Transform())
		Expect(diff).Should(Equal(""))
	})
})
