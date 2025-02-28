// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ = Describe("Subnet API Integration Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)

	})

	Context("Create subnet API", func() {
		It("Create Subnet should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())

			Expect(gotSubnet.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(gotSubnet.Spec.AvailabilityZone).Should(Equal(availabilityZone))
			Expect(gotSubnet.Spec.VpcId).Should(Equal(gotVPC.Metadata.ResourceId))

		})

		It("Create Subnet without name should succeed and name should be equal to resource_id", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create a subnet without a name
			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())

			// Verify the name of the created subnet is equal to the resource_id
			Expect(gotSubnet.Metadata.Name).Should(Equal(gotSubnet.Metadata.ResourceId))
		})

		It("Create subnet with the same name for different VPCs should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create the first VPC
			cidr := "10.0.0.0/16"
			name := "vpc1"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create the first subnet - name it 'sbbName'
			subnetName := "SbbName"
			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           subnetName,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(gotSubnet.Metadata.Name).Should(Equal(subnetName))

			// create another vpc.
			name = "vpc2"

			createVPCReq = NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC2, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create another subnet with a name of 'SbbName' in the new vpc
			createReq = &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           subnetName,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC2.Metadata.ResourceId,
				},
			}
			gotSubnet, err = subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(gotSubnet.Metadata.Name).Should(Equal(subnetName))

		})
		It("Create Subnet with invalid name should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			invalidNames := []string{
				" ", // empty name is not allowed.
				" leading_space not allowed",
				"trailing_space not allowed ",

				"invalid@name",  // @ is not allowed.
				"invalid#name",  // # is not allowed.
				"invalid$name",  // $ is not allowed.
				"invalid%name",  // % is not allowed.
				"invalid^name",  // ^ is not allowed.
				"invalid&name",  // & is not allowed.
				"invalid*name",  // * is not allowed.
				"invalid(name",  // ( is not allowed.
				"invalid)name",  // ) is not allowed.
				"invalid+name",  // + is not allowed.
				"invalid=name",  // = is not allowed.
				"invalid{name",  // { is not allowed.
				"invalid}name",  // } is not allowed.
				"invalid[name",  // [ is not allowed.
				"invalid]name",  // ] is not allowed.
				"invalid|name",  // | is not allowed.
				"invalid\\name", // \ is not allowed.
				"invalid/name",  // / is not allowed.
				"invalid~name",  // ~ is not allowed.
				"invalid!name",  // ! is not allowed.

				strings.Repeat("a", 66), // name is too long

			}

			// Create a VPC
			cidr := "192.168.1.14/24"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, "vpc1", cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			for _, invalidName := range invalidNames {
				createReq := &pb.SubnetCreateRequest{
					Metadata: &pb.SubnetMetadataCreate{
						CloudAccountId: cloudAccountId,
						Name:           invalidName,
					},
					Spec: &pb.SubnetSpec{
						CidrBlock:        cidr,
						AvailabilityZone: availabilityZone,
						VpcId:            gotVPC.Metadata.ResourceId,
					},
				}

				res, err := subnetServiceClient.Create(ctx, createReq)
				Expect(err).ShouldNot(Succeed())
				Expect(res).Should(BeNil())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			}

		})
		It("Create Subnet with existing name within the same vpc, should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create the first subnet
			createReq1 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet1",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			firstSubnet, err := subnetServiceClient.Create(ctx, createReq1)
			Expect(err).Should(Succeed())

			// Attempt to create a second subnet with the same name
			createReq2 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           firstSubnet.Metadata.Name,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.1.0/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq2)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.AlreadyExists))
		})

		It("Create subnet with invalid CIDR should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			invalidCidrs := []string{
				" ",
				"invalid",
				"10.0.1.0/15",
				"10.0.1.0/30",
				"192.168.5.1/24", // address not in CIDR - 10.0.1.1/16
			}

			// Create a VPC
			name := "default"
			cidr := "10.0.1.1/16"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			for _, invalidCidr := range invalidCidrs {
				createReq := &pb.SubnetCreateRequest{
					Metadata: &pb.SubnetMetadataCreate{
						CloudAccountId: cloudAccountId,
					},
					Spec: &pb.SubnetSpec{
						CidrBlock:        invalidCidr,
						AvailabilityZone: availabilityZone,
						VpcId:            gotVPC.Metadata.ResourceId,
					},
				}
				_, err = subnetServiceClient.Create(ctx, createReq)
				Expect(err).ShouldNot(Succeed())
			}
		})

		It("Invalid AZ should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "invalid"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq)
			Expect(err).ShouldNot(Succeed())
		})
		It("Invalid vpcId should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			_, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create an invalid vpcid
			invalidResourceId, err := uuid.NewRandom()
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            invalidResourceId.String(),
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq)
			Expect(err).ShouldNot(Succeed())
		})

		It("Malformed vpcId should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			_, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            "invalid",
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq)
			Expect(err).ShouldNot(Succeed())
		})

		It("Create Subnet with same CIDR in the same VPC should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create the first subnet
			createReq1 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet1",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.64/26",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq1)
			Expect(err).Should(Succeed())

			// Attempt to create a second subnet with a partially overlapping CIDR
			createReq2 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet2",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.64/26",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq2)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})

		It("Create Subnet with partially overlapping CIDR in the same VPC should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create the first subnet
			createReq1 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet1",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.64/26",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq1)
			Expect(err).Should(Succeed())

			// Attempt to create a second subnet with a partially overlapping CIDR
			createReq2 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet2",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.96/26",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq2)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})

		It("Create Subnet fully contained in the existing CIDR in the same VPC should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create the first subnet
			createReq1 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet1",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.0/22",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq1)
			Expect(err).Should(Succeed())

			// Attempt to create a second subnet fully contained in the existing CIDR
			createReq2 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet2",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.1.0/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq2)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})

		It("Create Subnet fully containing the existing CIDR in the same VPC should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create the first subnet
			createReq1 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet1",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.0/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq1)
			Expect(err).Should(Succeed())

			// Attempt to create a second subnet fully containing the existing CIDR
			createReq2 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet4",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.0/23",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq2)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})

		It("Create Subnet with overlapping CIDR in a different VPC should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create the first VPC
			cidr1 := "10.0.0.0/16"
			name1 := "vpc1"
			availabilityZone := "us-dev-1a"

			createVPCReq1 := NewCreateVPCRequest(cloudAccountId, name1, cidr1)
			gotVPC1, err := vpcServiceClient.Create(ctx, createVPCReq1)
			Expect(err).Should(Succeed())

			// Create the second VPC
			cidr2 := "10.0.0.0/16"
			name2 := "vpc2"

			createVPCReq2 := NewCreateVPCRequest(cloudAccountId, name2, cidr2)
			gotVPC2, err := vpcServiceClient.Create(ctx, createVPCReq2)
			Expect(err).Should(Succeed())

			// Create a subnet in the first VPC
			createReq1 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet1",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.0/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC1.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq1)
			Expect(err).Should(Succeed())

			// Create a subnet with an overlapping CIDR in the second VPC
			createReq2 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet2",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.0/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC2.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq2)
			Expect(err).Should(Succeed())
		})

		It("Create Subnet with non-overlapping CIDR in the same VPC should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create the first subnet
			createReq1 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet1",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.0/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq1)
			Expect(err).Should(Succeed())

			// Create a second subnet with a non-overlapping CIDR
			createReq2 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet2",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.1.0/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq2)
			Expect(err).Should(Succeed())
		})
	})

	Context("UpdateStatus subnet API", func() {
		It("Update status should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(gotSubnet.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(gotSubnet.Status.Phase).Should(Equal(pb.SubnetPhase_SubnetPhase_Provisioning))
			Expect(gotSubnet.Status.Message).Should(Equal("Subnet is provisioning"))

			// Update status
			_, err = subnetPrivateServiceClient.UpdateStatus(ctx, &pb.SubnetUpdateStatusRequest{
				Metadata: &pb.SubnetIdReference{
					CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
					ResourceId:      gotSubnet.Metadata.ResourceId,
					ResourceVersion: gotSubnet.Metadata.ResourceVersion,
				},
				Status: &pb.SubnetStatusPrivate{
					Phase:   pb.SubnetPhase_SubnetPhase_Ready,
					Message: "Subnet ready",
				},
			})

			// Fetch the subnet
			newSubnet, err := subnetServiceClient.Get(ctx, &pb.SubnetGetRequest{
				Metadata: &pb.SubnetMetadataReference{
					CloudAccountId: gotSubnet.Metadata.CloudAccountId,
					NameOrId:       &pb.SubnetMetadataReference_ResourceId{ResourceId: gotSubnet.Metadata.ResourceId},
				},
			})

			Expect(err).Should(Succeed())
			Expect(newSubnet.Status.Phase).Should(Equal(pb.SubnetPhase_SubnetPhase_Ready))
			Expect(newSubnet.Status.Message).Should(Equal("Subnet ready"))
		})

		It("Update status with delete should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(gotSubnet.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(gotSubnet.Status.Phase).Should(Equal(pb.SubnetPhase_SubnetPhase_Provisioning))
			Expect(gotSubnet.Status.Message).Should(Equal("Subnet is provisioning"))

			// Update status
			_, err = subnetPrivateServiceClient.UpdateStatus(ctx, &pb.SubnetUpdateStatusRequest{
				Metadata: &pb.SubnetIdReference{
					CloudAccountId:   gotSubnet.Metadata.CloudAccountId,
					ResourceId:       gotSubnet.Metadata.ResourceId,
					ResourceVersion:  gotSubnet.Metadata.ResourceVersion,
					DeletedTimestamp: timestamppb.New(time.Unix(1737501600, 0)),
				},
				Status: &pb.SubnetStatusPrivate{
					Phase:   pb.SubnetPhase_SubnetPhase_Deleted,
					Message: "Deleted",
				},
			})

			// Fetch the subnet
			_, err = subnetServiceClient.Get(ctx, &pb.SubnetGetRequest{
				Metadata: &pb.SubnetMetadataReference{
					CloudAccountId: gotSubnet.Metadata.CloudAccountId,
					NameOrId:       &pb.SubnetMetadataReference_ResourceId{ResourceId: gotSubnet.Metadata.ResourceId},
				},
			})

			Expect(err).ShouldNot(Succeed())
			// resource deleted, therefor -> not found
			Expect(status.Code(err)).Should(Equal(codes.NotFound)) //

		})

	})
	Context("Update subnet Private API", func() {
		It("Update subnet should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(gotSubnet.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(gotSubnet.Status.Phase).Should(Equal(pb.SubnetPhase_SubnetPhase_Provisioning))
			Expect(gotSubnet.Status.Message).Should(Equal("Subnet is provisioning"))

			// Update subnet using private API client
			newName := "newName3331"
			newLabels := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			_, err = subnetPrivateServiceClient.UpdatePrivate(ctx, &pb.SubnetUpdatePrivateRequest{
				Metadata: &pb.SubnetMetadata{
					CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
					ResourceId:      gotSubnet.Metadata.ResourceId,
					ResourceVersion: gotSubnet.Metadata.ResourceVersion,
					Name:            newName,
					Labels:          newLabels,
				},
			})
			newSubnet, err := subnetServiceClient.Get(ctx, &pb.SubnetGetRequest{
				Metadata: &pb.SubnetMetadataReference{
					CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
					NameOrId:        &pb.SubnetMetadataReference_ResourceId{ResourceId: gotSubnet.Metadata.ResourceId},
					ResourceVersion: gotSubnet.Metadata.ResourceVersion,
				},
			})
			Expect(err).Should(Succeed())
			Expect(newSubnet.Metadata.Name).Should(Equal(newName))
			Expect(newSubnet.Metadata.Labels).Should(Equal(newLabels))
			Expect(newSubnet.Spec.VpcId).Should(Equal(gotVPC.Metadata.ResourceId))
		})
	})

	Context("Update subnet API", func() {
		It("Update subnet should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(gotSubnet.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(gotSubnet.Status.Phase).Should(Equal(pb.SubnetPhase_SubnetPhase_Provisioning))
			Expect(gotSubnet.Status.Message).Should(Equal("Subnet is provisioning"))

			// Update
			newName := "newName3331"
			newLabels := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			_, err = subnetServiceClient.Update(ctx, &pb.SubnetUpdateRequest{
				Metadata: &pb.SubnetMetadata{
					CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
					ResourceId:      gotSubnet.Metadata.ResourceId,
					ResourceVersion: gotSubnet.Metadata.ResourceVersion,
					Name:            newName,
					Labels:          newLabels,
				},
			})
			newSubnet, err := subnetServiceClient.Get(ctx, &pb.SubnetGetRequest{
				Metadata: &pb.SubnetMetadataReference{
					CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
					NameOrId:        &pb.SubnetMetadataReference_ResourceId{ResourceId: gotSubnet.Metadata.ResourceId},
					ResourceVersion: gotSubnet.Metadata.ResourceVersion,
				},
			})
			Expect(err).Should(Succeed())
			Expect(newSubnet.Metadata.Name).Should(Equal(newName))
			Expect(newSubnet.Metadata.Labels).Should(Equal(newLabels))
			Expect(newSubnet.Spec.VpcId).Should(Equal(gotVPC.Metadata.ResourceId))

		})
		It("Update subnet with same name should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(gotSubnet.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(gotSubnet.Status.Phase).Should(Equal(pb.SubnetPhase_SubnetPhase_Provisioning))
			Expect(gotSubnet.Status.Message).Should(Equal("Subnet is provisioning"))

			// Update
			newName := name
			newLabels := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			_, err = subnetServiceClient.Update(ctx, &pb.SubnetUpdateRequest{
				Metadata: &pb.SubnetMetadata{
					CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
					ResourceId:      gotSubnet.Metadata.ResourceId,
					ResourceVersion: gotSubnet.Metadata.ResourceVersion,
					Name:            newName,
					Labels:          newLabels,
				},
			})
			newSubnet, err := subnetServiceClient.Get(ctx, &pb.SubnetGetRequest{
				Metadata: &pb.SubnetMetadataReference{
					CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
					NameOrId:        &pb.SubnetMetadataReference_ResourceId{ResourceId: gotSubnet.Metadata.ResourceId},
					ResourceVersion: gotSubnet.Metadata.ResourceVersion,
				},
			})
			Expect(err).Should(Succeed())
			Expect(newSubnet.Metadata.Name).Should(Equal(newName))
			Expect(newSubnet.Metadata.Labels).Should(Equal(newLabels))
			Expect(newSubnet.Spec.VpcId).Should(Equal(gotVPC.Metadata.ResourceId))
		})

		It("Update subnet should fail with invalid labels", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(gotSubnet.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(gotSubnet.Status.Phase).Should(Equal(pb.SubnetPhase_SubnetPhase_Provisioning))
			Expect(gotSubnet.Status.Message).Should(Equal("Subnet is provisioning"))

			// Update
			newName := "newName3331"
			newLabels := make(map[string]string)

			// generate too many labels
			for i := 1; i <= 30; i++ {
				key := fmt.Sprintf("key%d", i)
				value := fmt.Sprintf("value%d", i)
				newLabels[key] = value
			}

			_, err = subnetServiceClient.Update(ctx, &pb.SubnetUpdateRequest{
				Metadata: &pb.SubnetMetadata{
					CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
					ResourceId:      gotSubnet.Metadata.ResourceId,
					ResourceVersion: gotSubnet.Metadata.ResourceVersion,
					Name:            newName,
					Labels:          newLabels,
				},
			})

			Expect(err).ShouldNot(Succeed())
			Expect(err.Error()).To(ContainSubstring("the number of labels must not exceed"))
		})

		It("Update subnet name with invalid name should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// create subnet.
			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet1",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			_, err = subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())

			invalidNames := []string{
				" ", // empty name is not allowed.
				" leading_space not allowed",
				"trailing_space not allowed ",

				"invalid@name", // @ is not allowed.
				"invalid#name", // # is not allowed.

				strings.Repeat("a", 66), // name is too long
			}

			for _, invalidName := range invalidNames {
				_, err := subnetServiceClient.Update(ctx, &pb.SubnetUpdateRequest{
					Metadata: &pb.SubnetMetadata{
						CloudAccountId: cloudAccountId,
						Name:           invalidName,
					},
				})
				Expect(err).ShouldNot(Succeed())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			}

		})
		It("Update subnet name with existing name should fail", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create the first subnet
			createReq1 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet1",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			firstSubnet, err := subnetServiceClient.Create(ctx, createReq1)
			Expect(err).Should(Succeed())
			Expect(firstSubnet.Metadata.Name).Should(Equal("subnet1"))

			// Create the second subnet
			createReq2 := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
					Name:           "subnet2",
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.1.0/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			secondSubnet, err := subnetServiceClient.Create(ctx, createReq2)
			Expect(err).Should(Succeed())

			// Attempt to update the second subnet with the name of the first subnet
			_, err = subnetServiceClient.Update(ctx, &pb.SubnetUpdateRequest{
				Metadata: &pb.SubnetMetadata{
					CloudAccountId: secondSubnet.Metadata.CloudAccountId,
					ResourceId:     secondSubnet.Metadata.ResourceId,
					Name:           firstSubnet.Metadata.Name,
				},
			})
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.AlreadyExists))
		})
	})
	Context("Search subnet API", func() {
		It("Search subnet should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "192.168.1.1/16"
			name := "myvpc"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create a subnet
			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "192.168.1.1/20",
					VpcId:            gotVPC.Metadata.ResourceId,
					AvailabilityZone: availabilityZone,
				},
			}
			subnetServiceClient.Create(ctx, createReq)

			res, _ := subnetServiceClient.Search(ctx, &pb.SubnetSearchRequest{
				Metadata: &pb.SubnetMetadataSearch{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					VpcId: gotVPC.Metadata.ResourceId,
				},
			})
			Expect(err).Should(Succeed())
			Expect(len(res.Items)).Should(Equal(1))
			Expect(res.Items[0].Spec.CidrBlock).Should(Equal("192.168.1.1/20"))

		})
		It("Search Subnet should fail with invalid VPC ID", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.1.1.2/20"
			name := "myvpc"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create a subnet
			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "192.168.1.1/20",
					VpcId:            gotVPC.Metadata.ResourceId,
					AvailabilityZone: availabilityZone,
				},
			}
			subnetServiceClient.Create(ctx, createReq)

			// search for subnet with invalid VPC ID
			_, err = subnetServiceClient.Search(ctx, &pb.SubnetSearchRequest{
				Metadata: &pb.SubnetMetadataSearch{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					VpcId: "invalid",
				},
			})
			Expect(err).ShouldNot(Succeed())
		})

	})

	Context("Get subnet API", func() {
		It("Get subnet should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.1/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())

			Expect(gotSubnet.Metadata.ResourceId).Should(Not(BeEmpty()))

			// Get the subnet
			newSubnet, err := subnetServiceClient.Get(ctx, &pb.SubnetGetRequest{
				Metadata: &pb.SubnetMetadataReference{
					CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
					NameOrId:        &pb.SubnetMetadataReference_ResourceId{ResourceId: gotSubnet.Metadata.ResourceId},
					ResourceVersion: gotSubnet.Metadata.ResourceVersion,
				},
			})
			Expect(err).Should(Succeed())
			Expect(newSubnet).Should(Equal(gotSubnet))
			Expect(newSubnet.Spec.AvailabilityZone).Should(Equal(availabilityZone))
			Expect(newSubnet.Spec.VpcId).Should(Equal(gotVPC.Metadata.ResourceId))
		})
	})
	Context("Delete subnet API", func() {
		It("Delete subnet should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create a subnet
			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.6/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(gotSubnet.Metadata.ResourceId).Should(Not(BeEmpty()))

			// Delete the subnet
			deleteSubnetReq := &pb.SubnetDeleteRequest{
				Metadata: &pb.SubnetMetadataReference{
					CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
					NameOrId:        &pb.SubnetMetadataReference_ResourceId{ResourceId: gotSubnet.Metadata.ResourceId},
					ResourceVersion: gotSubnet.Metadata.ResourceVersion,
				},
				Spec: &pb.SubnetSpec{
					VpcId: gotVPC.Metadata.ResourceId,
				},
			}

			_, err = subnetServiceClient.Delete(ctx, deleteSubnetReq)
			Expect(err).Should(Succeed())

		})

		It("Delete subnet should fail with invalid VPC ID", func() {

			// Create a VPC
			cloudAccountId := cloudaccount.MustNewId()

			// Create a VPC
			cidr := "10.0.0.0/16"
			name := "default"
			availabilityZone := "us-dev-1a"

			createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
			Expect(err).Should(Succeed())

			// Create a subnet
			createReq := &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.6/24",
					AvailabilityZone: availabilityZone,
					VpcId:            gotVPC.Metadata.ResourceId,
				},
			}
			gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(gotSubnet.Metadata.ResourceId).Should(Not(BeEmpty()))

			invalidVpcIds := []string{
				"",                           // empry
				"invalid",                    // invalid uuid
				uuid.New().String() + "aaaa", // invalid format
			}

			for _, invalidVpcId := range invalidVpcIds {
				// Delete the subnet
				deleteSubnetReq := &pb.SubnetDeleteRequest{
					Metadata: &pb.SubnetMetadataReference{
						CloudAccountId:  gotSubnet.Metadata.CloudAccountId,
						NameOrId:        &pb.SubnetMetadataReference_ResourceId{ResourceId: gotSubnet.Metadata.ResourceId},
						ResourceVersion: gotSubnet.Metadata.ResourceVersion,
					},
					Spec: &pb.SubnetSpec{
						VpcId: invalidVpcId,
					},
				}

				_, err = subnetServiceClient.Delete(ctx, deleteSubnetReq)
				Expect(err).ShouldNot(Succeed())
				Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
			}
		})
	})
})
