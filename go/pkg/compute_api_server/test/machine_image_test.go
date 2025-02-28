// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func checkMachineImageExists(MachineImages []*pb.MachineImage, machineImageName string) bool {
	machineImageFound := false
	for _, item := range MachineImages {
		if item.Metadata.Name == machineImageName {
			machineImageFound = true
			break
		}
	}
	return machineImageFound
}

var _ = Describe("Search and SearchStream Machine Image", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	It("Search machineimage should not list machineimage when hidden is set to true", func() {
		machineImageSearchReq := &pb.MachineImageSearchRequest{}
		machineImageName := "Test1"
		hidden := true
		machineImageClient, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{"vm-spr-sml"}, machineImageName, hidden)
		Expect(err).Should(Succeed())
		machineImageSearchResponse, err := machineImageClient.Search(ctx, machineImageSearchReq)
		Expect(err).Should(Succeed())
		MachineImages := machineImageSearchResponse.Items
		machineImageFound := checkMachineImageExists(MachineImages, machineImageName)
		Expect(machineImageFound).To(BeFalse())
	})

	It("SearchStream machineimage should not list machineimage when hidden is set to true", func() {
		machineImageName := "Test2"
		hidden := true
		machineImageSearchReq := &pb.MachineImageSearchRequest{}
		machineImageClient, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{"vm-spr-sml"}, machineImageName, hidden)
		Expect(err).Should(Succeed())
		stream, err := machineImageClient.SearchStream(ctx, machineImageSearchReq)
		Expect(err).Should(Succeed())

		var images []*pb.MachineImage
		for {
			image, err := stream.Recv()
			if err != nil {
				break
			}
			images = append(images, image)
		}

		machineImageFound := checkMachineImageExists(images, machineImageName)
		Expect(machineImageFound).To(BeFalse())
	})

	It("Search machineimage should list machineimage when hidden is set to false", func() {
		machineImageName := "Test3"
		hidden := false
		machineImageClient, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{"vm-spr-sml"}, machineImageName, hidden)
		Expect(err).Should(Succeed())
		machineImageSearchReq := &pb.MachineImageSearchRequest{}
		machineImageSearchResponse, err := machineImageClient.Search(ctx, machineImageSearchReq)
		Expect(err).Should(Succeed())
		MachineImages := machineImageSearchResponse.Items
		machineImageFound := checkMachineImageExists(MachineImages, machineImageName)
		Expect(machineImageFound).To(BeTrue())
	})

	It("Search machineimage should list machineimage when hidden is set to false", func() {
		machineImageName := "Test4"
		hidden := false
		machineImageClient, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{"vm-spr-sml"}, machineImageName, hidden)
		Expect(err).Should(Succeed())
		machineImageSearchReq := &pb.MachineImageSearchRequest{}
		stream, err := machineImageClient.SearchStream(ctx, machineImageSearchReq)
		Expect(err).Should(Succeed())

		var images []*pb.MachineImage
		for {
			image, err := stream.Recv()
			if err != nil {
				break
			}
			images = append(images, image)
		}

		machineImageFound := checkMachineImageExists(images, machineImageName)
		Expect(machineImageFound).To(BeTrue())
	})

	It("MachineImage search with filter", func() {

		By("Initialize with one image")
		machineImageName := "ubuntu-20.04-gaudi-metal-cloudimg-amd64-latest"
		hidden := false
		CreateInstanceType(ctx, "bm-virtual")
		_, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_BareMetalHost}, []string{"bm-virtual", "bm-icp-gaudi2"}, machineImageName, hidden)
		Expect(err).Should(Succeed())
		computeApiServerAddress := fmt.Sprintf("localhost:%d", grpcListenPort)
		clientConn, err := grpc.Dial(computeApiServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).Should(Succeed())

		By("1. Check for valid filter")
		machineImageClient := pb.NewMachineImageServiceClient(clientConn)
		searchResult, err := machineImageClient.Search(ctx, &pb.MachineImageSearchRequest{
			Metadata: &pb.MachineImageSearchRequest_Metadata{
				InstanceType: "bm-virtual",
			},
		})
		Expect(err).Should(Succeed())
		Expect(len(searchResult.Items)).Should(Equal(1))
		Expect(searchResult.Items[0].Metadata.Name).Should(Equal("ubuntu-20.04-gaudi-metal-cloudimg-amd64-latest"))

		By("1. Check for invalid filter")
		searchResult, err = machineImageClient.Search(ctx, &pb.MachineImageSearchRequest{
			Metadata: &pb.MachineImageSearchRequest_Metadata{
				InstanceType: "XX",
			},
		})
		Expect(err).Should(Succeed())
		Expect(len(searchResult.Items)).Should(Equal(0))

		By("1. Check for no filter")
		searchResult, err = machineImageClient.Search(ctx, &pb.MachineImageSearchRequest{})
		Expect(err).Should(Succeed())
		Expect(len(searchResult.Items)).Should(Equal(1))

		machineImageName = "ubuntu-20.04-gaudi-metal-cloudimg-amd64-v20231013"
		hidden = false
		_, err = CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_BareMetalHost}, []string{"bm-virtual", "bm-icp-gaudi2"}, machineImageName, hidden)
		Expect(err).Should(Succeed())

		By("2. Check for valid filter with two entries")
		searchResult, err = machineImageClient.Search(ctx, &pb.MachineImageSearchRequest{
			Metadata: &pb.MachineImageSearchRequest_Metadata{
				InstanceType: "bm-icp-gaudi2",
			},
		})
		Expect(err).Should(Succeed())
		Expect(len(searchResult.Items)).Should(Equal(2))
		// Verify that the result order is deterministic
		Expect(searchResult.Items[0].Metadata.Name).Should(Equal("ubuntu-20.04-gaudi-metal-cloudimg-amd64-v20231013"))

		By("2. Check for invalid filter")
		searchResult, err = machineImageClient.Search(ctx, &pb.MachineImageSearchRequest{
			Metadata: &pb.MachineImageSearchRequest_Metadata{
				InstanceType: "XX",
			},
		})
		Expect(err).Should(Succeed())
		Expect(len(searchResult.Items)).Should(Equal(0))

		By("2. Check for no filter")
		searchResult, err = machineImageClient.Search(ctx, &pb.MachineImageSearchRequest{})
		Expect(err).Should(Succeed())
		Expect(len(searchResult.Items)).Should(Equal(2))

	})

})

var _ = Describe("Create Machine Image", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	It("Success case", func() {
		machineImageSearchReq := &pb.MachineImageSearchRequest{}
		machineImageName := "Test1"
		hidden := true
		machineImageClient, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{"vm-spr-sml"}, machineImageName, hidden)
		Expect(err).Should(Succeed())
		machineImageSearchResponse, err := machineImageClient.Search(ctx, machineImageSearchReq)
		Expect(err).Should(Succeed())
		MachineImages := machineImageSearchResponse.Items
		machineImageFound := checkMachineImageExists(MachineImages, machineImageName)
		Expect(machineImageFound).To(BeFalse())
	})

	It("fails if virtualMachine image name exceeds 32 characters", func() {
		machineImageName := "TestMachineabcdefghijklmnopqrstuvwxyz" // More than 32 characters
		hidden := false
		_, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{"vm-spr-sml"}, machineImageName, hidden)
		// Assert error and check its details
		Expect(err).To(HaveOccurred())

		st, ok := status.FromError(err)
		Expect(ok).To(BeTrue(), "error should be a gRPC status")
		Expect(st.Code()).To(Equal(codes.InvalidArgument), "error code should be InvalidArgument")
		Expect(st.Message()).To(ContainSubstring("virtualMachine image name must be less than 32 characters"), "error message mismatch")
	})

	It("success if the virtual machine image name exceeds 32 characters but is included in the allowedList", func() {
		machineImageSearchReq := &pb.MachineImageSearchRequest{}
		machineImageName := "iks-vm-u22-cd-cp-1-27-11-v20240227" // More than 32 characters but is included in the allowedList
		hidden := false
		machineImageClient, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{"vm-spr-sml"}, machineImageName, hidden)
		Expect(err).Should(Succeed())
		machineImageSearchResponse, err := machineImageClient.Search(ctx, machineImageSearchReq)
		Expect(err).Should(Succeed())
		MachineImages := machineImageSearchResponse.Items
		machineImageFound := checkMachineImageExists(MachineImages, machineImageName)
		Expect(machineImageFound).To(BeTrue())
	})
})
