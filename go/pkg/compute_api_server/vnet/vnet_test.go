// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vnet

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/praserx/ipconv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewPutSubnetRequest(index int, numAddresses int, prefixLen int) *pb.CreateSubnetRequest {
	region := "us-dev-1"
	availabilityZone := "us-dev-1a"
	prefixLength := int32(prefixLen)
	baseIp := net.ParseIP("172.16.0.0")
	baseIpInt, err := ipconv.IPv4ToInt(baseIp)
	Expect(err).Should(Succeed())
	subnetIpInt := baseIpInt + (uint32(index) * 256 * 256)
	subnetIp := ipconv.IntToIPv4(subnetIpInt)
	gatewayIp := ipconv.IntToIPv4(subnetIpInt + 1)
	vlanId := int32(1000 + index)
	req := &pb.CreateSubnetRequest{
		Region:           region,
		AvailabilityZone: availabilityZone,
		Subnet:           subnetIp.String(),
		PrefixLength:     prefixLength,
		Gateway:          gatewayIp.String(),
		VlanId:           vlanId,
	}
	for i := 0; i < numAddresses; i++ {
		req.Address = append(req.Address, ipconv.IntToIPv4(subnetIpInt+2+uint32(i)).String())
	}
	return req
}

func NewVNetPutRequest(cloudAccountId string, createSubnetRequest *pb.CreateSubnetRequest) *pb.VNetPutRequest {
	name := fmt.Sprintf("%s-%s", createSubnetRequest.AvailabilityZone, uuid.NewString())
	return NewVNetPutRequestWithName(cloudAccountId, createSubnetRequest, name)
}

func NewVNetPutRequestWithName(cloudAccountId string, createSubnetRequest *pb.CreateSubnetRequest, name string) *pb.VNetPutRequest {
	req := &pb.VNetPutRequest{
		Metadata: &pb.VNetPutRequest_Metadata{
			CloudAccountId: cloudAccountId,
			Name:           name,
		},
		Spec: &pb.VNetSpec{
			Region:           createSubnetRequest.Region,
			AvailabilityZone: createSubnetRequest.AvailabilityZone,
			PrefixLength:     createSubnetRequest.PrefixLength,
		},
	}
	return req
}

func NewVNetDeleteRequest(vNetPutRequest *pb.VNetPutRequest) *pb.VNetDeleteRequest {
	req := &pb.VNetDeleteRequest{
		Metadata: &pb.VNetDeleteRequest_Metadata{
			CloudAccountId: vNetPutRequest.Metadata.CloudAccountId,
			NameOrId: &pb.VNetDeleteRequest_Metadata_Name{
				Name: vNetPutRequest.Metadata.Name,
			},
		},
	}
	return req
}

func NewVNetGetRequest(vNetPutRequest *pb.VNetPutRequest) *pb.VNetGetRequest {
	req := &pb.VNetGetRequest{
		Metadata: &pb.VNetGetRequest_Metadata{
			CloudAccountId: vNetPutRequest.Metadata.CloudAccountId,
			NameOrId: &pb.VNetGetRequest_Metadata_Name{
				Name: vNetPutRequest.Metadata.Name,
			},
		},
	}
	return req
}

func NewVNetSearchRequest(vNetPutRequest *pb.VNetPutRequest) *pb.VNetSearchRequest {
	req := &pb.VNetSearchRequest{
		Metadata: &pb.VNetSearchRequest_Metadata{
			CloudAccountId: vNetPutRequest.Metadata.CloudAccountId,
		},
	}
	return req
}

func NewVNetReserveSubnetRequest(vNetPutRequest *pb.VNetPutRequest) *pb.VNetReserveSubnetRequest {
	req := &pb.VNetReserveSubnetRequest{
		VNetReference: &pb.VNetReference{
			CloudAccountId: vNetPutRequest.Metadata.CloudAccountId,
			Name:           vNetPutRequest.Metadata.Name,
		},
	}
	return req
}

func NewVNetReleaseSubnetRequest(vNetReserveSubnetRequest *pb.VNetReserveSubnetRequest) *pb.VNetReleaseSubnetRequest {
	req := &pb.VNetReleaseSubnetRequest{
		VNetReference: vNetReserveSubnetRequest.VNetReference,
	}
	return req
}

func NewVNetReserveAddressRequest(vNetReserveSubnetRequest *pb.VNetReserveSubnetRequest) *pb.VNetReserveAddressRequest {
	req := &pb.VNetReserveAddressRequest{
		VNetReference: &pb.VNetReference{
			CloudAccountId: vNetReserveSubnetRequest.VNetReference.CloudAccountId,
			Name:           vNetReserveSubnetRequest.VNetReference.Name,
		},
		AddressReference: &pb.VNetAddressReference{
			AddressConsumerId: fmt.Sprintf("%s.%s.cloud.intel.com", uuid.NewString(), vNetReserveSubnetRequest.VNetReference.CloudAccountId)},
	}
	return req
}

func NewVNetReleaseAddressRequest(vNetReserveAddressRequest *pb.VNetReserveAddressRequest) *pb.VNetReleaseAddressRequest {
	req := &pb.VNetReleaseAddressRequest{
		VNetReference:    vNetReserveAddressRequest.VNetReference,
		AddressReference: vNetReserveAddressRequest.AddressReference,
	}
	return req
}

var _ = Describe("VNet Unit Tests", func() {
	getSubnetConsumerId := func(subnetConsumerIdPattern string) string {
		return subnetConsumerIdForVNet(subnetConsumerIdPattern, "123456789012", "us-dev-1a-default", "d70ce0b0-cef5-4cfb-bb05-934a4389f7c4", "us-dev-1", "us-dev-1a")
	}

	It("subnetConsumerIdForVNet for production should succeed", func() {
		Expect(getSubnetConsumerId("{ResourceId}.{CloudAccountId}.vnet")).Should(Equal("d70ce0b0-cef5-4cfb-bb05-934a4389f7c4.123456789012.vnet"))
	})

	It("subnetConsumerIdForVNet for JF development cluster should succeed", func() {
		Expect(getSubnetConsumerId("{AvailabilityZone}.vnet")).Should(Equal("us-dev-1a.vnet"))
	})
})

var _ = Describe("VNet Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	It("VNet happy path should succeed", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put (simulate call by public API user)")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNet, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNet", vNet)

		By("VNetPrivate.ReserveSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		vNetReserveSubnetReq := NewVNetReserveSubnetRequest(vNetPutReq1)
		vNetPrivate, err := vNetPrivateClient.ReserveSubnet(ctx, vNetReserveSubnetReq)
		Expect(err).Should(Succeed())
		log.Info("ReserveSubnet", "vNetPrivate", vNetPrivate)

		By("VNetPrivate.ReserveAddress (simulate call by Instance Operator)")
		vNetReserveAddressReq := NewVNetReserveAddressRequest(vNetReserveSubnetReq)
		vNetReserveAddressResp, err := vNetPrivateClient.ReserveAddress(ctx, vNetReserveAddressReq)
		Expect(err).Should(Succeed())
		log.Info("ReserveAddress", "vNetReserveAddressResp", vNetReserveAddressResp)

		By("VNetPrivate.ReleaseAddress (simulate call by Instance Operator)")
		vNetReleaseAddressReq := NewVNetReleaseAddressRequest(vNetReserveAddressReq)
		_, err = vNetPrivateClient.ReleaseAddress(ctx, vNetReleaseAddressReq)
		Expect(err).Should(Succeed())

		By("VNetPrivate.ReleaseSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().RemoveBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		vNetReleaseSubnetReq := NewVNetReleaseSubnetRequest(vNetReserveSubnetReq)
		_, err = vNetPrivateClient.ReleaseSubnet(ctx, vNetReleaseSubnetReq)
		Expect(err).Should(Succeed())

		By("VNet.Delete (simulate call by public API user)")
		vNetDeleteReq1 := NewVNetDeleteRequest(vNetPutReq1)
		_, err = vNetClient.Delete(ctx, vNetDeleteReq1)
		Expect(err).Should(Succeed())
	})

	It("Repeated VNet put with same request should succeed", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put 1st time (simulate call by public API user)")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNetResponse1, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNet", vNetResponse1)

		vNetGetReq1 := NewVNetGetRequest(vNetPutReq1)
		vNetGetRes, err := vNetClient.Get(ctx, vNetGetReq1)
		Expect(err).Should(Succeed())
		Expect(vNetGetRes.Metadata.CloudAccountId).Should(Equal(vNetResponse1.Metadata.CloudAccountId))
		Expect(vNetGetRes.Metadata.ResourceId).Should(Equal(vNetResponse1.Metadata.ResourceId))
		Expect(vNetGetRes.Metadata.Name).Should(Equal(vNetResponse1.Metadata.Name))

		By("VNet.Put 2nd time (simulate call by public API user)")
		vNetResponse2, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())
		Expect(vNetGetRes.Metadata.CloudAccountId).Should(Equal(vNetResponse2.Metadata.CloudAccountId))
		Expect(vNetGetRes.Metadata.ResourceId).Should(Equal(vNetResponse2.Metadata.ResourceId))
		Expect(vNetGetRes.Metadata.Name).Should(Equal(vNetResponse2.Metadata.Name))

		log.Info("Put", "vNet", vNetResponse2)
	})

	It("VNet Put that updates the PrefixLength should succeed", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put 1st time (simulate call by public API user)")
		cloudAccountId := cloudaccount.MustNewId()
		vNetName := "vnetname1"
		vNetPutReq1 := NewVNetPutRequestWithName(cloudAccountId, putSubnetReq1, vNetName)
		vNetResponse1, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNetResponse1", vNetResponse1)

		putSubnetReq2 := NewPutSubnetRequest(2, 4, 22)
		_, err = ipResourceManagerClient.PutSubnet(ctx, putSubnetReq2)
		Expect(err).Should(Succeed())

		vNetPutReq2 := NewVNetPutRequestWithName(cloudAccountId, putSubnetReq2, vNetName)
		By("VNet.Put 2nd time (simulate call by public API user)")
		vNetResponse2, err := vNetClient.Put(ctx, vNetPutReq2)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNetResponse2", vNetResponse2)

		vNetGetReq1 := NewVNetGetRequest(vNetPutReq2)
		vNetGetRes, err := vNetClient.Get(ctx, vNetGetReq1)
		Expect(err).Should(Succeed())
		Expect(vNetGetRes.Metadata.CloudAccountId).Should(Equal(vNetPutReq1.Metadata.CloudAccountId))
		Expect(vNetGetRes.Metadata.ResourceId).Should(Equal(vNetResponse1.Metadata.ResourceId))
		Expect(vNetGetRes.Metadata.Name).Should(Equal(vNetResponse1.Metadata.Name))
		Expect(vNetGetRes.Spec.Region).Should(Equal(vNetPutReq1.Spec.Region))
		Expect(vNetGetRes.Spec.AvailabilityZone).Should(Equal(vNetPutReq1.Spec.AvailabilityZone))
		Expect(vNetGetRes.Spec.PrefixLength).Should(Equal(vNetPutReq2.Spec.PrefixLength))
		Expect(vNetGetRes.Spec.PrefixLength).ShouldNot(Equal(vNetPutReq1.Spec.PrefixLength))
	})

	It("VNet delete with reserved addresses should return FailedPrecondition", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put (simulate call by public API user)")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNet, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNet", vNet)

		By("VNetPrivate.ReserveSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		vNetReserveSubnetReq := NewVNetReserveSubnetRequest(vNetPutReq1)
		vNetPrivate, err := vNetPrivateClient.ReserveSubnet(ctx, vNetReserveSubnetReq)
		Expect(err).Should(Succeed())
		log.Info("ReserveSubnet", "vNetPrivate", vNetPrivate)

		By("VNetPrivate.ReserveAddress (simulate call by Instance Operator)")
		vNetReserveAddressReq := NewVNetReserveAddressRequest(vNetReserveSubnetReq)
		vNetReserveAddressResp, err := vNetPrivateClient.ReserveAddress(ctx, vNetReserveAddressReq)
		Expect(err).Should(Succeed())
		log.Info("ReserveAddress", "vNetReserveAddressResp", vNetReserveAddressResp)

		By("VNet.Delete should return FailedPrecondition (simulate call by public API user)")
		vNetDeleteReq1 := NewVNetDeleteRequest(vNetPutReq1)
		_, err = vNetClient.Delete(ctx, vNetDeleteReq1)
		Expect(status.Code(err)).Should(Equal(codes.FailedPrecondition))
	})

	It("VNet Get should return a previously Put message", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNetPutResponse, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())

		By("VNet.Get")
		vNetGetReq1 := NewVNetGetRequest(vNetPutReq1)
		vNet, err := vNetClient.Get(ctx, vNetGetReq1)
		Expect(err).Should(Succeed())
		Expect(vNet.Metadata.CloudAccountId).Should(Equal(vNetPutReq1.Metadata.CloudAccountId))
		Expect(vNet.Metadata.Name).Should(Equal(vNetPutReq1.Metadata.Name))
		Expect(vNet.Metadata.ResourceId).Should(Equal(vNetPutResponse.Metadata.ResourceId))
		Expect(vNet.Spec.Region).Should(Equal(vNetPutReq1.Spec.Region))
		Expect(vNet.Spec.AvailabilityZone).Should(Equal(vNetPutReq1.Spec.AvailabilityZone))
		Expect(vNet.Spec.PrefixLength).Should(Equal(vNetPutReq1.Spec.PrefixLength))
	})

	It("VNet Search should return a previously Put message", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNetPutResponse, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())

		By("VNet.Search")
		vNetSearchReq1 := NewVNetSearchRequest(vNetPutReq1)
		vNetSearchResponse, err := vNetClient.Search(ctx, vNetSearchReq1)
		Expect(err).Should(Succeed())
		Expect(len(vNetSearchResponse.Items)).Should(Equal(1))
		vNet := vNetSearchResponse.Items[0]
		Expect(vNet.Metadata.CloudAccountId).Should(Equal(vNetPutReq1.Metadata.CloudAccountId))
		Expect(vNet.Metadata.Name).Should(Equal(vNetPutReq1.Metadata.Name))
		Expect(vNet.Metadata.ResourceId).Should(Equal(vNetPutResponse.Metadata.ResourceId))
		Expect(vNet.Spec.Region).Should(Equal(vNetPutReq1.Spec.Region))
		Expect(vNet.Spec.AvailabilityZone).Should(Equal(vNetPutReq1.Spec.AvailabilityZone))
		Expect(vNet.Spec.PrefixLength).Should(Equal(vNetPutReq1.Spec.PrefixLength))
	})

	It("VNet SearchStream should return a previously Put message", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNetPutResponse, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())

		By("VNet.SearchStream")
		vNetSearchReq1 := NewVNetSearchRequest(vNetPutReq1)
		stream, err := vNetClient.SearchStream(ctx, vNetSearchReq1)
		Expect(err).Should(Succeed())
		vNet, err := stream.Recv()
		Expect(err).Should(Succeed())
		Expect(vNet.Metadata.CloudAccountId).Should(Equal(vNetPutReq1.Metadata.CloudAccountId))
		Expect(vNet.Metadata.Name).Should(Equal(vNetPutReq1.Metadata.Name))
		Expect(vNet.Metadata.ResourceId).Should(Equal(vNetPutResponse.Metadata.ResourceId))
		Expect(vNet.Spec.Region).Should(Equal(vNetPutReq1.Spec.Region))
		Expect(vNet.Spec.AvailabilityZone).Should(Equal(vNetPutReq1.Spec.AvailabilityZone))
		err = stream.CloseSend()
		Expect(err).Should(Succeed())
	})

	It("VNet Search should not return VNet for other CloudAccount", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put CloudAccount 1")
		cloudAccountId1 := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId1, putSubnetReq1)
		vNetPutResponse1, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put CloudAccount 2")
		cloudAccountId2 := cloudaccount.MustNewId()
		vNetPutReq2 := NewVNetPutRequest(cloudAccountId2, putSubnetReq1)
		_, err = vNetClient.Put(ctx, vNetPutReq2)
		Expect(err).Should(Succeed())

		By("VNet.Search CloudAccount 1")
		vNetSearchReq1 := NewVNetSearchRequest(vNetPutReq1)
		vNetSearchResponse, err := vNetClient.Search(ctx, vNetSearchReq1)
		Expect(err).Should(Succeed())
		Expect(len(vNetSearchResponse.Items)).Should(Equal(1))
		vNet := vNetSearchResponse.Items[0]
		Expect(vNet.Metadata.CloudAccountId).Should(Equal(vNetPutReq1.Metadata.CloudAccountId))
		Expect(vNet.Metadata.Name).Should(Equal(vNetPutReq1.Metadata.Name))
		Expect(vNet.Metadata.ResourceId).Should(Equal(vNetPutResponse1.Metadata.ResourceId))
		Expect(vNet.Spec.Region).Should(Equal(vNetPutReq1.Spec.Region))
		Expect(vNet.Spec.AvailabilityZone).Should(Equal(vNetPutReq1.Spec.AvailabilityZone))
		Expect(vNet.Spec.PrefixLength).Should(Equal(vNetPutReq1.Spec.PrefixLength))
	})

	It("VNet ReserveAddress with address should succeed", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put (simulate call by public API user)")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNet, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNet", vNet)

		By("VNetPrivate.ReserveSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		vNetReserveSubnetReq := NewVNetReserveSubnetRequest(vNetPutReq1)
		vNetPrivate, err := vNetPrivateClient.ReserveSubnet(ctx, vNetReserveSubnetReq)
		Expect(err).Should(Succeed())
		log.Info("ReserveSubnet", "vNetPrivate", vNetPrivate)

		By("VNetPrivate.ReserveAddress (simulate call by Instance Operator)")
		vNetReserveAddressReq := NewVNetReserveAddressRequest(vNetReserveSubnetReq)
		vNetReserveAddressReq.AddressReference.Address = putSubnetReq1.Address[0]
		vNetReserveAddressResp, err := vNetPrivateClient.ReserveAddress(ctx, vNetReserveAddressReq)
		Expect(err).Should(Succeed())
		Expect(vNetReserveAddressResp.Address).Should(Equal(vNetReserveAddressReq.AddressReference.Address))
		log.Info("ReserveAddress", "vNetReserveAddressResp", vNetReserveAddressResp)

		By("VNetPrivate.ReserveAddress again (simulate second call by Instance Operator)")
		vNetReserveAddressResp2, err := vNetPrivateClient.ReserveAddress(ctx, vNetReserveAddressReq)
		Expect(err).Should(Succeed())
		Expect(vNetReserveAddressResp2.Address).Should(Equal(vNetReserveAddressReq.AddressReference.Address))
	})

	It("VNet ReserveAddress with address reserved by another addressConsumerId should fail", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put (simulate call by public API user)")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNet, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNet", vNet)

		By("VNetPrivate.ReserveSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		vNetReserveSubnetReq := NewVNetReserveSubnetRequest(vNetPutReq1)
		vNetPrivate, err := vNetPrivateClient.ReserveSubnet(ctx, vNetReserveSubnetReq)
		Expect(err).Should(Succeed())
		log.Info("ReserveSubnet", "vNetPrivate", vNetPrivate)

		By("VNetPrivate.ReserveAddress (simulate call by Instance Operator)")
		vNetReserveAddressReq := NewVNetReserveAddressRequest(vNetReserveSubnetReq)
		vNetReserveAddressReq.AddressReference.Address = putSubnetReq1.Address[0]
		vNetReserveAddressResp, err := vNetPrivateClient.ReserveAddress(ctx, vNetReserveAddressReq)
		Expect(err).Should(Succeed())
		Expect(vNetReserveAddressResp.Address).Should(Equal(vNetReserveAddressReq.AddressReference.Address))
		log.Info("ReserveAddress", "vNetReserveAddressResp", vNetReserveAddressResp)

		By("VNetPrivate.ReserveAddress with another addressConsumerId (simulate second call by Instance Operator)")
		vNetReserveAddressReq2 := NewVNetReserveAddressRequest(vNetReserveSubnetReq)
		vNetReserveAddressReq2.AddressReference.Address = vNetReserveAddressReq.AddressReference.Address
		_, err = vNetPrivateClient.ReserveAddress(ctx, vNetReserveAddressReq2)
		Expect(err).ShouldNot(Succeed())
	})

	It("VNet ReserveAddress should fail when AddBucketSubnet fails", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put (simulate call by public API user)")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNet, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNet", vNet)

		By("VNetPrivate.ReserveSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
		vNetReserveSubnetReq := NewVNetReserveSubnetRequest(vNetPutReq1)
		_, err = vNetPrivateClient.ReserveSubnet(ctx, vNetReserveSubnetReq)
		Expect(err).ShouldNot(Succeed())
	})

	It("VNet ReleaseAddress should fail when RemoveBucketSubnet fails", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put (simulate call by public API user)")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNet, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNet", vNet)

		By("VNetPrivate.ReserveSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		vNetReserveSubnetReq := NewVNetReserveSubnetRequest(vNetPutReq1)
		vNetPrivate, err := vNetPrivateClient.ReserveSubnet(ctx, vNetReserveSubnetReq)
		Expect(err).Should(Succeed())
		log.Info("ReserveSubnet", "vNetPrivate", vNetPrivate)

		By("VNetPrivate.ReserveAddress (simulate call by Instance Operator)")
		vNetReserveAddressReq := NewVNetReserveAddressRequest(vNetReserveSubnetReq)
		vNetReserveAddressResp, err := vNetPrivateClient.ReserveAddress(ctx, vNetReserveAddressReq)
		Expect(err).Should(Succeed())
		log.Info("ReserveAddress", "vNetReserveAddressResp", vNetReserveAddressResp)

		By("VNetPrivate.ReleaseAddress (simulate call by Instance Operator)")
		vNetReleaseAddressReq := NewVNetReleaseAddressRequest(vNetReserveAddressReq)
		_, err = vNetPrivateClient.ReleaseAddress(ctx, vNetReleaseAddressReq)
		Expect(err).Should(Succeed())

		By("VNetPrivate.ReleaseSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().RemoveBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
		vNetReleaseSubnetReq := NewVNetReleaseSubnetRequest(vNetReserveSubnetReq)
		_, err = vNetPrivateClient.ReleaseSubnet(ctx, vNetReleaseSubnetReq)
		Expect(err).ShouldNot(Succeed())
	})

	It("VNet ReleaseSubnet should not call RemoveBucketSubnet if subnet has consumers", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("VNet.Put (simulate call by public API user)")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq1 := NewVNetPutRequest(cloudAccountId, putSubnetReq1)
		vNet, err := vNetClient.Put(ctx, vNetPutReq1)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNet", vNet)

		By("VNetPrivate.ReserveSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		vNetReserveSubnetReq := NewVNetReserveSubnetRequest(vNetPutReq1)
		vNetPrivate, err := vNetPrivateClient.ReserveSubnet(ctx, vNetReserveSubnetReq)
		Expect(err).Should(Succeed())
		log.Info("ReserveSubnet", "vNetPrivate", vNetPrivate)

		By("VNetPrivate.ReserveAddress (simulate call by Instance Operator)")
		vNetReserveAddressReq := NewVNetReserveAddressRequest(vNetReserveSubnetReq)
		vNetReserveAddressResp, err := vNetPrivateClient.ReserveAddress(ctx, vNetReserveAddressReq)
		Expect(err).Should(Succeed())
		log.Info("ReserveAddress", "vNetReserveAddressResp", vNetReserveAddressResp)

		By("VNetPrivate.ReleaseSubnet (simulate call by Instance Operator)")
		// Expected calls to RemoveBucketSubnet is 0 since the address has been released yet
		objectStorageServicePrivateClient.EXPECT().RemoveBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(0)
		vNetReleaseSubnetReq := NewVNetReleaseSubnetRequest(vNetReserveSubnetReq)
		_, err = vNetPrivateClient.ReleaseSubnet(ctx, vNetReleaseSubnetReq)
		Expect(err).ShouldNot(Succeed())
	})

	It("ReserveSubnet should pick the subnet whose PrefixLength is the largest value not exceeding MaximumPrefixLength", func() {
		prefixLength1 := 16
		vlanIndex1 := 1
		prefixLength2 := 22
		vlanIndex2 := 2
		prefixLength3 := 27
		vlanIndex3 := 3
		numAddresses := 4
		var maximumPrefixLength int32 = 23

		By("PutSubnet subnet1 with prefixLength1")
		putSubnetReq1 := NewPutSubnetRequest(vlanIndex1, numAddresses, prefixLength1)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("PutSubnet subnet2 with prefixLength2")
		putSubnetReq2 := NewPutSubnetRequest(vlanIndex2, numAddresses, prefixLength2)
		_, err = ipResourceManagerClient.PutSubnet(ctx, putSubnetReq2)
		Expect(err).Should(Succeed())

		By("PutSubnet subnet3 with prefixLength3")
		putSubnetReq3 := NewPutSubnetRequest(vlanIndex3, numAddresses, prefixLength3)
		_, err = ipResourceManagerClient.PutSubnet(ctx, putSubnetReq3)
		Expect(err).Should(Succeed())

		By("VNet.Put (simulate call by public API user)")
		cloudAccountId := cloudaccount.MustNewId()
		vNetPutReq := NewVNetPutRequest(cloudAccountId, putSubnetReq2)
		vNet, err := vNetClient.Put(ctx, vNetPutReq)
		Expect(err).Should(Succeed())
		log.Info("Put", "vNet", vNet)

		By("VNetPrivate.ReserveSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		vNetReserveSubnetReq := NewVNetReserveSubnetRequest(vNetPutReq)
		vNetReserveSubnetReq.MaximumPrefixLength = maximumPrefixLength
		log.Info("vNetReserveSubnetReq", "vNetReserveSubnetReq", vNetReserveSubnetReq)

		vNetPrivate, err := vNetPrivateClient.ReserveSubnet(ctx, vNetReserveSubnetReq)
		Expect(err).Should(Succeed())
		Expect(vNetPrivate.Spec.PrefixLength).Should(Equal(int32(prefixLength2)))
		Expect(vNetPrivate.Spec.VlanId).Should(Equal(int32(1000 + vlanIndex2)))

		By("VNetPrivate.ReserveAddress (simulate call by Instance Operator)")
		vNetReserveAddressReq := NewVNetReserveAddressRequest(vNetReserveSubnetReq)
		_, err = vNetPrivateClient.ReserveAddress(ctx, vNetReserveAddressReq)
		Expect(err).Should(Succeed())

		By("VNetPrivate.ReleaseAddress (simulate call by Instance Operator)")
		vNetReleaseAddressReq := NewVNetReleaseAddressRequest(vNetReserveAddressReq)
		_, err = vNetPrivateClient.ReleaseAddress(ctx, vNetReleaseAddressReq)
		Expect(err).Should(Succeed())

		By("VNetPrivate.ReleaseSubnet (simulate call by Instance Operator)")
		objectStorageServicePrivateClient.EXPECT().RemoveBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		vNetReleaseSubnetReq := NewVNetReleaseSubnetRequest(vNetReserveSubnetReq)
		_, err = vNetPrivateClient.ReleaseSubnet(ctx, vNetReleaseSubnetReq)
		Expect(err).Should(Succeed())

		By("VNet.Delete (simulate call by public API user)")
		vNetDeleteReq := NewVNetDeleteRequest(vNetPutReq)
		_, err = vNetClient.Delete(ctx, vNetDeleteReq)
		Expect(err).Should(Succeed())
	})
})
