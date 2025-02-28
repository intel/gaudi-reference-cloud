// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	machineimageserver "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/machine_image"
	vnetserver "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/vnet"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

var _ = Describe("ProtoJsonTable Machine Image Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	baseline := func() *pb.MachineImage {
		machineImage := &pb.MachineImage{
			Metadata: &pb.MachineImage_Metadata{
				Name: "machine-image-name1",
			},
			Spec: &pb.MachineImageSpec{
				DisplayName: "Ubuntu 22.04 LTS (Jammy Jellyfish) v20221204",
				UserName:    "ubuntu",
			},
		}
		return machineImage
	}

	It("Put should succeed", func() {
		machineImage := baseline()
		machineImage.Metadata.Name = uuid.NewString()
		protoJsonTable := machineimageserver.NewMachineImageProtoJsonTable(sqlDb)
		Expect(protoJsonTable.Put(ctx, machineImage)).Should(Succeed())
	})

	It("Delete should succeed", func() {
		machineImage := baseline()
		machineImage.Metadata.Name = uuid.NewString()
		protoJsonTable := machineimageserver.NewMachineImageProtoJsonTable(sqlDb)
		By("Put")
		Expect(protoJsonTable.Put(ctx, machineImage)).Should(Succeed())
		By("Delete")
		err := protoJsonTable.Delete(ctx, &pb.MachineImageDeleteRequest{
			Metadata: &pb.MachineImageDeleteRequest_Metadata{
				Name: machineImage.Metadata.Name,
			}})
		Expect(err).Should(Succeed())
		By("Get")
		respGet, err := protoJsonTable.Get(ctx, &pb.MachineImageGetRequest{
			Metadata: &pb.MachineImageGetRequest_Metadata{
				Name: machineImage.Metadata.Name,
			}})
		Expect(err).ShouldNot(Succeed())
		Expect(respGet).Should(BeNil())
	})

	It("Get should return a previously Put message", func() {
		machineImage := baseline()
		machineImage.Metadata.Name = uuid.NewString()
		protoJsonTable := machineimageserver.NewMachineImageProtoJsonTable(sqlDb)
		By("Put")
		Expect(protoJsonTable.Put(ctx, machineImage)).Should(Succeed())
		By("Get")
		respGet, err := protoJsonTable.Get(ctx, &pb.MachineImageGetRequest{
			Metadata: &pb.MachineImageGetRequest_Metadata{
				Name: machineImage.Metadata.Name,
			}})
		Expect(err).Should(Succeed())
		Expect(respGet).ShouldNot(BeNil())
		Expect(cmp.Diff(respGet, machineImage, protocmp.Transform())).Should(Equal(""))
	})

	It("Search should return a previously Put message", func() {
		machineImage := baseline()
		machineImage.Metadata.Name = uuid.NewString()
		protoJsonTable := machineimageserver.NewMachineImageProtoJsonTable(sqlDb)
		By("Put")
		Expect(protoJsonTable.Put(ctx, machineImage)).Should(Succeed())
		By("Search")
		foundMatch := false
		handlerFunc := func(m proto.Message) error {
			log.Info("handler", "m", m)
			if proto.Equal(m, machineImage) {
				foundMatch = true
			}
			return nil
		}
		searchReq := &pb.MachineImageSearchRequest{}
		Expect(protoJsonTable.Search(ctx, searchReq, handlerFunc)).Should(Succeed())
		Expect(foundMatch).Should(BeTrue())
	})
})

var _ = Describe("ProtoJsonTable VNet Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	baseline := func() *pb.VNetPrivate {
		vNetPrivate := &pb.VNetPrivate{
			Metadata: &pb.VNetPrivate_Metadata{
				CloudAccountId: "090631835288",
				ResourceId:     uuid.NewString(),
				Name:           "vnet-name-1",
			},
			Spec: &pb.VNetSpecPrivate{
				AvailabilityZone: "az-1",
				PrefixLength:     16,
			},
		}
		return vNetPrivate
	}

	It("Put should succeed", func() {
		protoJsonTable := vnetserver.NewVNetProtoJsonTable(sqlDb)
		putReq := baseline()
		Expect(protoJsonTable.Put(ctx, putReq)).Should(Succeed())
	})

	It("Delete should succeed", func() {
		protoJsonTable := vnetserver.NewVNetProtoJsonTable(sqlDb)
		putReq := baseline()
		By("Put")
		Expect(protoJsonTable.Put(ctx, putReq)).Should(Succeed())
		By("Delete")
		delReq := &pb.VNetDeleteRequest{
			Metadata: &pb.VNetDeleteRequest_Metadata{
				NameOrId: &pb.VNetDeleteRequest_Metadata_ResourceId{
					ResourceId: putReq.Metadata.ResourceId,
				},
			},
		}
		err := protoJsonTable.Delete(ctx, delReq)
		Expect(err).Should(Succeed())
		By("Get")
		respGet, err := protoJsonTable.Get(ctx, &pb.VNetGetRequest{
			Metadata: &pb.VNetGetRequest_Metadata{
				CloudAccountId: putReq.Metadata.CloudAccountId,
				NameOrId: &pb.VNetGetRequest_Metadata_Name{
					Name: putReq.Metadata.Name,
				},
			}})
		Expect(err).ShouldNot(Succeed())
		Expect(respGet).Should(BeNil())
	})

	It("Get should return a previously Put message", func() {
		protoJsonTable := vnetserver.NewVNetProtoJsonTable(sqlDb)
		putReq := baseline()
		By("Put")
		Expect(protoJsonTable.Put(ctx, putReq)).Should(Succeed())
		By("Get")
		respGet, err := protoJsonTable.Get(ctx, &pb.VNetGetRequest{
			Metadata: &pb.VNetGetRequest_Metadata{
				CloudAccountId: putReq.Metadata.CloudAccountId,
				NameOrId: &pb.VNetGetRequest_Metadata_ResourceId{
					ResourceId: putReq.Metadata.ResourceId,
				},
			}})
		Expect(err).Should(Succeed())
		Expect(respGet).ShouldNot(BeNil())
	})

	It("Search should return a previously Put message", func() {
		protoJsonTable := vnetserver.NewVNetProtoJsonTable(sqlDb)
		putReq := baseline()
		By("Put")
		Expect(protoJsonTable.Put(ctx, putReq)).Should(Succeed())
		By("Search")
		foundMatch := false
		expected := putReq
		handlerFunc := func(m proto.Message) error {
			if proto.Equal(m, expected) {
				foundMatch = true
			}
			return nil
		}
		searchReq := &pb.VNetSearchRequest{
			Metadata: &pb.VNetSearchRequest_Metadata{
				CloudAccountId: putReq.Metadata.CloudAccountId,
			},
		}
		Expect(protoJsonTable.Search(ctx, searchReq, handlerFunc)).Should(Succeed())
		Expect(foundMatch).Should(BeTrue())
	})
})
