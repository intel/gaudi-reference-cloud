// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("IPRM API Integration Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	Context("Reserve Port", func() {
		It("Reserve Port should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			_, subnet, err := NewCreateVpcAndSubnet(ctx, cloudAccountId, vpcServiceClient, subnetServiceClient)

			ipuSerialNumber := "ipuSerialNumber"
			chassisId := "30"
			ipAddress := "ipAddress"
			macAddress := "macAddress"
			sshEnabled := true
			internetAccess := true

			reservePortRequest := &pb.ReservePortRequest{
				Metadata: &pb.PortMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.PortSpecPrivate{
					SubnetId:        subnet.Metadata.ResourceId,
					IpuSerialNumber: ipuSerialNumber,
					ChassisId:       chassisId,
					IpAddress:       ipAddress,
					MacAddress:      macAddress,
					SshEnabled:      sshEnabled,
					InternetAccess:  internetAccess,
				},
			}

			reservePortResp, err := iprmPrivateServiceClient.ReservePort(ctx, reservePortRequest)

			Expect(err).Should(Succeed())
			Expect(reservePortResp).ShouldNot(BeNil())
			Expect(reservePortResp.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(reservePortResp.Spec.SubnetId).Should(Equal(subnet.Metadata.ResourceId))
			Expect(reservePortResp.Spec.IpuSerialNumber).Should(Equal(ipuSerialNumber))
			Expect(reservePortResp.Spec.ChassisId).Should(Equal(chassisId))
			Expect(reservePortResp.Spec.IpAddress).Should(Equal(ipAddress))
			Expect(reservePortResp.Spec.MacAddress).Should(Equal(macAddress))
			Expect(reservePortResp.Spec.SshEnabled).Should(BeTrue())
			Expect(reservePortResp.Spec.InternetAccess).Should(BeTrue())

		})

		It("Reserve Port should succeed and return the same port on duplicate call", func() {
			cloudAccountId := cloudaccount.MustNewId()

			_, subnet, err := NewCreateVpcAndSubnet(ctx, cloudAccountId, vpcServiceClient, subnetServiceClient)

			ipuSerialNumber := "ipuSerialNumber"
			chassisId := "34"
			ipAddress := "ipAddress"
			macAddress := "macAddress"
			sshEnabled := true
			internetAccess := true

			reservePortRequest := &pb.ReservePortRequest{
				Metadata: &pb.PortMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.PortSpecPrivate{
					SubnetId:        subnet.Metadata.ResourceId,
					IpuSerialNumber: ipuSerialNumber,
					ChassisId:       chassisId,
					IpAddress:       ipAddress,
					MacAddress:      macAddress,
					SshEnabled:      sshEnabled,
					InternetAccess:  internetAccess,
				},
			}

			reservePortResp, err := iprmPrivateServiceClient.ReservePort(ctx, reservePortRequest)

			Expect(err).Should(Succeed())
			Expect(reservePortResp).ShouldNot(BeNil())
			Expect(reservePortResp.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(reservePortResp.Spec.SubnetId).Should(Equal(subnet.Metadata.ResourceId))

			// make anther call
			originalPortId := reservePortResp.Metadata.ResourceId
			reservePortResp, err = iprmPrivateServiceClient.ReservePort(ctx, reservePortRequest)
			Expect(err).Should(Succeed())

			// make sure the same port is returned
			Expect(reservePortResp.Metadata.ResourceId).Should(Equal(originalPortId))
			Expect(reservePortResp.Spec.ChassisId).Should(Equal(chassisId))
			Expect(reservePortResp.Spec.IpAddress).Should(Equal(ipAddress))
			Expect(reservePortResp.Spec.MacAddress).Should(Equal(macAddress))
		})

		It("Reserve Port should fail with invalid subnet", func() {
			cloudAccountId := cloudaccount.MustNewId()

			randUuid, _ := uuid.NewRandom()
			randUUIDString := randUuid.String()
			ipuSerialNumber := "ipuSerialNumber"
			chassisId := "38"
			ipAddress := "ipAddress"
			macAddress := "macAddress"

			invalidSubnetIds := []string{
				"",                  // empty subnet id
				"invalid-subnet-id", // invalud uuid
				randUUIDString,      // valid uuid but subnet not exists in db.
			}

			for _, invalidSubnetId := range invalidSubnetIds {
				reservePortRequest := &pb.ReservePortRequest{
					Metadata: &pb.PortMetadataCreatePrivate{
						CloudAccountId: cloudAccountId,
					},
					Spec: &pb.PortSpecPrivate{
						SubnetId:        invalidSubnetId,
						IpuSerialNumber: ipuSerialNumber,
						ChassisId:       chassisId,
						IpAddress:       ipAddress,
						MacAddress:      macAddress,
					},
				}

				reservePortResp, err := iprmPrivateServiceClient.ReservePort(ctx, reservePortRequest)

				Expect(err).ShouldNot(Succeed())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
				Expect(reservePortResp).Should(BeNil())
			}

		})
	})
	Context("Update Status api", func() {
		It("Update status should succeed", func() {
			// Create a port to change its status.
			cloudAccountId := cloudaccount.MustNewId()

			_, subnet, err := NewCreateVpcAndSubnet(ctx, cloudAccountId, vpcServiceClient, subnetServiceClient)

			ipuSerialNumber := "ipuSerialNumber"
			chassisId := "31"
			ipAddress := "ipAddress"
			macAddress := "macAddress"

			reservePortRequest := &pb.ReservePortRequest{
				Metadata: &pb.PortMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.PortSpecPrivate{
					SubnetId:        subnet.Metadata.ResourceId,
					IpuSerialNumber: ipuSerialNumber,
					ChassisId:       chassisId,
					IpAddress:       ipAddress,
					MacAddress:      macAddress,
				},
			}

			gotPort, err := iprmPrivateServiceClient.ReservePort(ctx, reservePortRequest)

			Expect(err).Should(Succeed())
			Expect(gotPort).ShouldNot(BeNil())
			Expect(gotPort.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(gotPort.Status.Phase).Should(Equal(pb.PortPhase_PortPhase_Provisioning))
			Expect(gotPort.Status.Message).Should(Equal("Port is provisioning"))

			// Update port status
			_, err = iprmPrivateServiceClient.UpdateStatus(ctx, &pb.PortUpdateStatusRequest{
				Metadata: &pb.PortMetadataReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     gotPort.Metadata.ResourceId,
				},
				Status: &pb.PortStatusPrivate{
					Phase:   pb.PortPhase_PortPhase_Ready,
					Message: "Port ready",
				},
			})
			Expect(err).Should(Succeed())
		})
	})
	Context("Get Port Private API", func() {
		It("Get port should succeed", func() {
			// Create a port to be fetched.
			cloudAccountId := cloudaccount.MustNewId()

			_, subnet, err := NewCreateVpcAndSubnet(ctx, cloudAccountId, vpcServiceClient, subnetServiceClient)

			ipuSerialNumber := "ipuSerialNumber"
			chassisId := "40"
			ipAddress := "ipAddress"
			macAddress := "macAddress"

			reservePortRequest := &pb.ReservePortRequest{
				Metadata: &pb.PortMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.PortSpecPrivate{
					SubnetId:        subnet.Metadata.ResourceId,
					IpuSerialNumber: ipuSerialNumber,
					ChassisId:       chassisId,
					IpAddress:       ipAddress,
					MacAddress:      macAddress,
				},
			}

			gotPort, err := iprmPrivateServiceClient.ReservePort(ctx, reservePortRequest)

			Expect(err).Should(Succeed())
			Expect(gotPort).ShouldNot(BeNil())
			Expect(gotPort.Metadata.ResourceId).ShouldNot(BeEmpty())

			// Get port
			getPortResp, err := iprmPrivateServiceClient.GetPortPrivate(ctx, &pb.GetPortPrivateRequest{
				Metadata: &pb.PortMetadataReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     gotPort.Metadata.ResourceId,
				},
			})
			Expect(err).Should(Succeed())
			Expect(getPortResp).ShouldNot(BeNil())
			Expect(getPortResp.Metadata.ResourceId).Should(Equal(gotPort.Metadata.ResourceId))
		})
	})
	Context("Release Port API", func() {
		It("Release Port should succeed", func() {
			// Create a port to be released.
			cloudAccountId := cloudaccount.MustNewId()

			_, subnet, err := NewCreateVpcAndSubnet(ctx, cloudAccountId, vpcServiceClient, subnetServiceClient)

			ipuSerialNumber := "ipuSerialNumber"
			chassisId := "41"
			ipAddress := "ipAddress"
			macAddress := "macAddress"

			reservePortRequest := &pb.ReservePortRequest{
				Metadata: &pb.PortMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.PortSpecPrivate{
					SubnetId:        subnet.Metadata.ResourceId,
					IpuSerialNumber: ipuSerialNumber,
					ChassisId:       chassisId,
					IpAddress:       ipAddress,
					MacAddress:      macAddress,
				},
			}

			gotPort, err := iprmPrivateServiceClient.ReservePort(ctx, reservePortRequest)

			Expect(err).Should(Succeed())
			Expect(gotPort).ShouldNot(BeNil())
			Expect(gotPort.Metadata.ResourceId).ShouldNot(BeEmpty())

			// // Release port
			_, err = iprmPrivateServiceClient.ReleasePort(ctx, &pb.ReleasePortRequest{
				Metadata: &pb.PortMetadataReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     gotPort.Metadata.ResourceId,
				},
			})
			Expect(err).Should(Succeed())
		})

	})
})
