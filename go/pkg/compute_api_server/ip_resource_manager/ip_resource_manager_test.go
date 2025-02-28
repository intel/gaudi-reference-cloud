// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ip_resource_manager

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/praserx/ipconv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func NewPutSubnetRequest(index int, numAddresses int) *pb.CreateSubnetRequest {
	region := "us-dev-1"
	availabilityZone := "us-dev-1a"
	prefixLength := int32(16)
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

func NewPutSubnetWithPrefixLengthRequest(index int, prefixLength int) *pb.CreateSubnetRequest {
	region := "us-dev-1"
	availabilityZone := "us-dev-1a"
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
		Subnet:           fmt.Sprintf("%s/%d", subnetIp.String(), prefixLength),
		PrefixLength:     int32(prefixLength),
		Gateway:          gatewayIp.String(),
		VlanId:           vlanId,
	}
	return req
}

func NewDeleteSubnetRequest(createSubnetRequest *pb.CreateSubnetRequest) *pb.DeleteSubnetRequest {
	req := &pb.DeleteSubnetRequest{
		Region:           createSubnetRequest.Region,
		AvailabilityZone: createSubnetRequest.AvailabilityZone,
		AddressSpace:     createSubnetRequest.AddressSpace,
		Subnet:           createSubnetRequest.Subnet,
		PrefixLength:     createSubnetRequest.PrefixLength,
	}
	return req
}

func NewSearchSubnetRequest() *pb.SearchSubnetRequest {
	req := &pb.SearchSubnetRequest{}
	return req
}

func NewReserveSubnetRequest(createSubnetRequest *pb.CreateSubnetRequest) *pb.ReserveSubnetRequest {
	req := &pb.ReserveSubnetRequest{
		SubnetReference: &pb.SubnetReference{
			SubnetConsumerId: fmt.Sprintf("%s.0123456789012.vnet", uuid.NewString()),
		},
		Spec: &pb.ReserveSubnetRequest_Spec{
			Region:           createSubnetRequest.Region,
			AvailabilityZone: createSubnetRequest.AvailabilityZone,
			PrefixLengthHint: createSubnetRequest.PrefixLength,
		},
	}
	return req
}

func NewReleaseSubnetRequest(reserveSubnetRequest *pb.ReserveSubnetRequest) *pb.ReleaseSubnetRequest {
	req := &pb.ReleaseSubnetRequest{
		SubnetReference: reserveSubnetRequest.SubnetReference,
	}
	return req
}

func NewReserveAddressRequest(reserveSubnetRequest *pb.ReserveSubnetRequest) *pb.ReserveAddressRequest {
	req := &pb.ReserveAddressRequest{
		SubnetReference: reserveSubnetRequest.SubnetReference,
		AddressReference: &pb.AddressReference{
			AddressConsumerId: fmt.Sprintf("%s.0123456789012.cloud.intel.com", uuid.NewString()),
		},
	}
	return req
}

func NewReleaseAddressRequest(reserveAddressRequest *pb.ReserveAddressRequest) *pb.ReleaseAddressRequest {
	req := &pb.ReleaseAddressRequest{
		SubnetReference:  reserveAddressRequest.SubnetReference,
		AddressReference: reserveAddressRequest.AddressReference,
	}
	return req
}

func getMapSize(subnetMap *sync.Map) int {
	var size int
	subnetMap.Range(func(key, value interface{}) bool {
		size++
		return true
	})
	return size
}

var _ = Describe("IP Resource Manager Unit Tests", func() {
	It("generateStandardAddresses with /29 should succeed", func() {
		gateway, addresses, err := generateStandardAddresses("172.16.11.0/29")
		Expect(err).Should(Succeed())
		Expect(gateway).Should(Equal("172.16.11.1"))
		Expect(addresses).Should(Equal([]string{"172.16.11.4", "172.16.11.5", "172.16.11.6"}))
	})

	It("generateStandardAddresses should generate enough usable IPs for a given netmask", func() {
		for usableIPs := 27; usableIPs <= 30; usableIPs++ {
			prefix := utils.GetMaximumPrefixLength(int32(usableIPs))
			_, addresses, err := generateStandardAddresses(fmt.Sprintf("172.16.11.0/%d", prefix))
			Expect(err).Should(Succeed())
			Expect(len(addresses)).Should(BeNumerically(">=", usableIPs))
		}
	})

	It("validateAndNormalizeCreateSubnetRequest with /29 CIDR should succeed", func() {
		req := &pb.CreateSubnetRequest{
			Subnet: "172.16.11.0/29",
		}
		Expect(validateAndNormalizeCreateSubnetRequest(req)).Should(Succeed())
		Expect(req.Subnet).Should(Equal("172.16.11.0/29"))
		Expect(req.PrefixLength).Should(Equal(int32(29)))
		Expect(req.Gateway).Should(Equal("172.16.11.1"))
		Expect(req.Address).Should(Equal([]string{"172.16.11.4", "172.16.11.5", "172.16.11.6"}))
	})

	It("validateAndNormalizeCreateSubnetRequest with /29 non-CIDR should succeed", func() {
		req := &pb.CreateSubnetRequest{
			Subnet:       "172.16.11.0",
			PrefixLength: 29,
		}
		Expect(validateAndNormalizeCreateSubnetRequest(req)).Should(Succeed())
		Expect(req.Subnet).Should(Equal("172.16.11.0/29"))
		Expect(req.PrefixLength).Should(Equal(int32(29)))
		Expect(req.Gateway).Should(Equal("172.16.11.1"))
		Expect(req.Address).Should(Equal([]string{"172.16.11.4", "172.16.11.5", "172.16.11.6"}))
	})
})

var _ = Describe("IP Resource Manager Integration Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	It("Create, reserve, release subnet should succeed", MustPassRepeatedly(10), func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("PutSubnet subnet2 in different AZ")
		createSubnetReq2 := NewPutSubnetRequest(2, 4)
		createSubnetReq2.AvailabilityZone = "az2"
		_, err = ipResourceManagerClient.PutSubnet(ctx, createSubnetReq2)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should reserve and return subnet1")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		reserveSubnetResp1, err := ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())
		Expect(reserveSubnetResp1.Region).Should(Equal(putSubnetReq1.Region))
		Expect(reserveSubnetResp1.AvailabilityZone).Should(Equal(putSubnetReq1.AvailabilityZone))
		Expect(reserveSubnetResp1.Subnet).Should(Equal(putSubnetReq1.Subnet))
		Expect(reserveSubnetResp1.PrefixLength).Should(Equal(putSubnetReq1.PrefixLength))
		Expect(reserveSubnetResp1.Gateway).Should(Equal(putSubnetReq1.Gateway))
		Expect(reserveSubnetResp1.VlanId).Should(Equal(putSubnetReq1.VlanId))

		By("ReserveSubnet again should return subnet1")
		reserveSubnetResp2, err := ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())
		Expect(reserveSubnetResp2.Region).Should(Equal(putSubnetReq1.Region))
		Expect(reserveSubnetResp2.AvailabilityZone).Should(Equal(putSubnetReq1.AvailabilityZone))
		Expect(reserveSubnetResp2.Subnet).Should(Equal(putSubnetReq1.Subnet))
		Expect(reserveSubnetResp2.PrefixLength).Should(Equal(putSubnetReq1.PrefixLength))
		Expect(reserveSubnetResp2.Gateway).Should(Equal(putSubnetReq1.Gateway))
		Expect(reserveSubnetResp2.VlanId).Should(Equal(putSubnetReq1.VlanId))

		By("ReleaseSubnet should succeed")
		releaseSubnetReq := NewReleaseSubnetRequest(reserveSubnetReq)
		_, err = ipResourceManagerClient.ReleaseSubnet(ctx, releaseSubnetReq)
		Expect(err).Should(Succeed())

		By("ReleaseSubnet again should return NotFound")
		_, err = ipResourceManagerClient.ReleaseSubnet(ctx, releaseSubnetReq)
		Expect(status.Code(err)).Should(Equal(codes.NotFound))
	})

	It("Delete subnet should succeed", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetRequest(1, 4)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("DeleteSubnet should succeed")
		deleteSubnetReq1 := NewDeleteSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.DeleteSubnet(ctx, deleteSubnetReq1)
		Expect(err).Should(Succeed())
	})

	It("Search subnet should return previously put subnet", func() {
		By("PutSubnet subnet1")
		putSubnetReq1 := NewPutSubnetWithPrefixLengthRequest(1, 29)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("SearchSubnetStream")
		searchReq1 := NewSearchSubnetRequest()
		stream, err := ipResourceManagerClient.SearchSubnetStream(ctx, searchReq1)
		Expect(err).Should(Succeed())
		subnet, err := stream.Recv()
		Expect(err).Should(Succeed())
		Expect(subnet.Region).Should(Equal(putSubnetReq1.Region))
		Expect(subnet.AvailabilityZone).Should(Equal(putSubnetReq1.AvailabilityZone))
		Expect(subnet.AddressSpace).Should(Equal(putSubnetReq1.AddressSpace))
		Expect(subnet.Subnet).Should(Equal(putSubnetReq1.Subnet))
		Expect(subnet.PrefixLength).Should(Equal(putSubnetReq1.PrefixLength))
		Expect(subnet.Gateway).Should(Equal(putSubnetReq1.Gateway))
		Expect(subnet.VlanId).Should(Equal(putSubnetReq1.VlanId))
		Expect(subnet.SubnetConsumerId).Should(Equal(reserveSubnetReq.SubnetReference.SubnetConsumerId))
		err = stream.CloseSend()
		Expect(err).Should(Succeed())
	})

	It("Reserve subnet with 1 of 1 subnets reserved should return ResourceExhausted", func() {
		By("PutSubnet should succeed")
		putSubnetReq1 := NewPutSubnetRequest(1, 1)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("ReserveSubnet with a different subnetConsumerId should return ResourceExhausted")
		reserveSubnetReq2 := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq2)
		Expect(status.Code(err)).Should(Equal(codes.ResourceExhausted))
	})

	It("Release subnet with reserved address should return FailedPrecondition", func() {
		By("PutSubnet should succeed")
		putSubnetReq1 := NewPutSubnetRequest(1, 4)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress should succeed")
		reserveAddressReq := NewReserveAddressRequest(reserveSubnetReq)
		_, err = ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())

		By("ReleaseSubnet again should return FailedPrecondition")
		releaseSubnetReq := NewReleaseSubnetRequest(reserveSubnetReq)
		_, err = ipResourceManagerClient.ReleaseSubnet(ctx, releaseSubnetReq)
		Expect(status.Code(err)).Should(Equal(codes.FailedPrecondition))
	})

	It("Reserve and release address should succeed", func() {
		By("PutSubnet should succeed")
		putSubnetReq1 := NewPutSubnetRequest(1, 1)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress should succeed and return first address")
		reserveAddressReq := NewReserveAddressRequest(reserveSubnetReq)
		reserveAddressResp1, err := ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())
		Expect(reserveAddressResp1.Address).Should(Equal(putSubnetReq1.Address[0]))

		By("ReserveAddress again should succeed and return first address")
		reserveAddressResp2, err := ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())
		Expect(reserveAddressResp2.Address).Should(Equal(putSubnetReq1.Address[0]))

		By("ReleaseAddress should succeed")
		releaseAddressReq := NewReleaseAddressRequest(reserveAddressReq)
		_, err = ipResourceManagerClient.ReleaseAddress(ctx, releaseAddressReq)
		Expect(err).Should(Succeed())

		By("ReleaseAddress again should return NotFound")
		_, err = ipResourceManagerClient.ReleaseAddress(ctx, releaseAddressReq)
		Expect(status.Code(err)).Should(Equal(codes.NotFound))
	})

	It("Reserve address with 2 subnets should succeed (TWC4728-403)", MustPassRepeatedly(10), func() {
		By("PutSubnet 1 should succeed")
		putSubnetReq1 := NewPutSubnetRequest(1, 10)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("PutSubnet 2 should succeed")
		putSubnetReq2 := NewPutSubnetRequest(2, 10)
		_, err = ipResourceManagerClient.PutSubnet(ctx, putSubnetReq2)
		Expect(err).Should(Succeed())

		By("ReserveSubnet 1 should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress 1 should succeed")
		reserveAddressReq := NewReserveAddressRequest(reserveSubnetReq)
		_, err = ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())
	})

	It("Reserve address with 1 of 1 addresses reserved should return ResourceExhausted", func() {
		By("PutSubnet should succeed")
		putSubnetReq1 := NewPutSubnetRequest(1, 1)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress should succeed and return first address")
		reserveAddressReq := NewReserveAddressRequest(reserveSubnetReq)
		reserveAddressResp1, err := ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())
		Expect(reserveAddressResp1.Address).Should(Equal(putSubnetReq1.Address[0]))

		By("ReserveAddress with a different addressConsumerId should return ResourceExhausted")
		reserveAddressReq2 := NewReserveAddressRequest(reserveSubnetReq)
		_, err = ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq2)
		Expect(status.Code(err)).Should(Equal(codes.ResourceExhausted))
	})

	It("Reserve address after release should succeed", func() {
		By("PutSubnet should succeed")
		putSubnetReq1 := NewPutSubnetRequest(1, 1)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress should succeed and return first address")
		reserveAddressReq := NewReserveAddressRequest(reserveSubnetReq)
		reserveAddressResp1, err := ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())
		Expect(reserveAddressResp1.Address).Should(Equal(putSubnetReq1.Address[0]))

		By("ReleaseAddress should succeed")
		releaseAddressReq := NewReleaseAddressRequest(reserveAddressReq)
		_, err = ipResourceManagerClient.ReleaseAddress(ctx, releaseAddressReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress should succeed and return first address")
		reserveAddressResp2, err := ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())
		Expect(reserveAddressResp2.Address).Should(Equal(putSubnetReq1.Address[0]))
	})

	It("Reserve address with same address and addressConsumerId should succeed", func() {
		By("PutSubnet should succeed")
		putSubnetReq1 := NewPutSubnetRequest(1, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress should succeed and return first address")
		reserveAddressReq := NewReserveAddressRequest(reserveSubnetReq)
		reserveAddressReq.AddressReference.Address = putSubnetReq1.Address[0]
		reserveAddressResp1, err := ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())
		Expect(reserveAddressResp1.Address).Should(Equal(reserveAddressReq.AddressReference.Address))

		By("ReserveAddress again should succeed and return same address")
		reserveAddressResp2, err := ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())
		Expect(reserveAddressResp2.Address).Should(Equal(reserveAddressReq.AddressReference.Address))
	})

	It("Reserve address with same address and different addressConsumerId should fail", func() {
		By("PutSubnet should succeed")
		putSubnetReq1 := NewPutSubnetRequest(1, 16)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress should succeed and return address")
		reserveAddressReq := NewReserveAddressRequest(reserveSubnetReq)
		reserveAddressResp1, err := ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress again with different addressConsumerId should fail")
		reserveAddressReq2 := NewReserveAddressRequest(reserveSubnetReq)
		reserveAddressReq2.AddressReference.Address = reserveAddressResp1.Address
		_, err = ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq2)
		Expect(err).ShouldNot(Succeed())
	})

	It("Reserve address with different address and same addressConsumerId should fail", func() {
		By("PutSubnet should succeed")
		putSubnetReq1 := NewPutSubnetRequest(1, 4)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress should succeed and return first address")
		reserveAddressReq := NewReserveAddressRequest(reserveSubnetReq)
		reserveAddressReq.AddressReference.Address = putSubnetReq1.Address[0]
		_, err = ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress again with different address should fail")
		reserveAddressReq.AddressReference.Address = putSubnetReq1.Address[1]
		_, err = ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).ShouldNot(Succeed())
	})

	It("Reserve address with unavailable address should fail", func() {
		By("PutSubnet should succeed")
		putSubnetReq1 := NewPutSubnetRequest(1, 1)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should succeed")
		reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
		Expect(err).Should(Succeed())

		By("ReserveAddress with unavailable address should fail")
		reserveAddressReq := NewReserveAddressRequest(reserveSubnetReq)
		ipInt, err := ipconv.IPv4ToInt(net.ParseIP(putSubnetReq1.Address[0]))
		Expect(err).Should(Succeed())
		reserveAddressReq.AddressReference.Address = ipconv.IntToIPv4(ipInt + 1).String()
		_, err = ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
		Expect(err).ShouldNot(Succeed())
	})

	It("Create large subnet should succeed", func() {
		putSubnetReq1 := NewPutSubnetRequest(1, 16*1024)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())
	})

	It("Put twice of large subnet should succeed", func() {
		putSubnetReq1 := NewPutSubnetRequest(1, 16*1024)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())
		_, err = ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())
	})

	It("Create subnet with generated addresses should succeed", func() {
		putSubnetReq1 := NewPutSubnetWithPrefixLengthRequest(1, 29)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())
	})

	It("Put modified subnet should succeed", func() {
		putSubnetReq1 := NewPutSubnetRequest(1, 2)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())
		putSubnetReq2 := NewPutSubnetRequest(1, 8)
		_, err = ipResourceManagerClient.PutSubnet(ctx, putSubnetReq2)
		Expect(err).Should(Succeed())
		putSubnetReq3 := NewPutSubnetRequest(1, 3)
		_, err = ipResourceManagerClient.PutSubnet(ctx, putSubnetReq3)
		Expect(err).Should(Succeed())
	})

	It("GetSubnetStatistics should return total number of subnet and total subnets consumed grouped by region, availability_zone, address_space, vlan_domain, prefix_length", func() {
		By("PutSubnet subnet1 with Prefixlength 22")
		putSubnetReq1 := NewPutSubnetWithPrefixLengthRequest(1, 22)
		_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq1)
		Expect(err).Should(Succeed())

		By("PutSubnet subnet2 with Prefixlength 22")
		putSubnetReq2 := NewPutSubnetWithPrefixLengthRequest(2, 22)
		_, err = ipResourceManagerClient.PutSubnet(ctx, putSubnetReq2)
		Expect(err).Should(Succeed())

		By("PutSubnet subnet1 with Prefixlength 24")
		putSubnetReq3 := NewPutSubnetWithPrefixLengthRequest(4, 24)
		_, err = ipResourceManagerClient.PutSubnet(ctx, putSubnetReq3)
		Expect(err).Should(Succeed())

		By("PutSubnet subnet2 with Prefixlength 24")
		putSubnetReq4 := NewPutSubnetWithPrefixLengthRequest(5, 24)
		_, err = ipResourceManagerClient.PutSubnet(ctx, putSubnetReq4)
		Expect(err).Should(Succeed())

		By("GetSubnetStatics should return SubnetStaticsRecord grouped by region, availability_zone, address_space, vlan_domain, prefix_length")
		getSubnetStatisticsResponse, err := ipResourceManagerClient.GetSubnetStatistics(ctx, &emptypb.Empty{})
		Expect(err).Should(Succeed())
		Expect(len(getSubnetStatisticsResponse.GetSubnetStatistics())).Should(Equal(2))

		By("ReserveSubnet should reserve and return subnet1 with Prefixlength 22")
		reserveSubnetReq1 := NewReserveSubnetRequest(putSubnetReq1)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq1)
		Expect(err).Should(Succeed())

		By("ReserveSubnet should reserve and return subnet2 with Prefixlength 22")
		reserveSubnetReq2 := NewReserveSubnetRequest(putSubnetReq2)
		_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq2)
		Expect(err).Should(Succeed())

		By("GetSubnetStatics should return SubnetStaticsRecords with the total number subnets and total number of consumed subnets in each group")
		getSubnetStatisticsResponse, err = ipResourceManagerClient.GetSubnetStatistics(ctx, &emptypb.Empty{})
		Expect(err).Should(Succeed())
		expectedSubnetStatistics := []*pb.SubnetStatisticsRecord{
			{
				Region:               putSubnetReq1.Region,
				AvailabilityZone:     putSubnetReq1.AvailabilityZone,
				AddressSpace:         putSubnetReq1.AddressSpace,
				PrefixLength:         putSubnetReq1.PrefixLength,
				VlanDomain:           putSubnetReq1.VlanDomain,
				TotalSubnets:         int32(2),
				TotalConsumedSubnets: int32(2),
			},
			{
				Region:               putSubnetReq3.Region,
				AvailabilityZone:     putSubnetReq3.AvailabilityZone,
				AddressSpace:         putSubnetReq3.AddressSpace,
				PrefixLength:         putSubnetReq3.PrefixLength,
				VlanDomain:           putSubnetReq3.VlanDomain,
				TotalSubnets:         int32(2),
				TotalConsumedSubnets: int32(0),
			},
		}
		Expect(getSubnetStatisticsResponse.SubnetStatistics).Should(ConsistOf(expectedSubnetStatistics))
	})
})

var _ = Describe("IP Resource Manager Unit Tests for Concurrency", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	It("Create, reserve, release subnet concurrently should succeed", func() {
		var wg sync.WaitGroup
		numberOfSubnets := 10
		var reservedSubnetsMap sync.Map
		var reserveSubnetReqs []*pb.ReserveSubnetRequest
		var releaseSubnetReqs []*pb.ReleaseSubnetRequest

		for i := 0; i < numberOfSubnets; i++ {
			By("Create PutSubnet, ReserveSubnet, ReleaseSubnet requests")
			putSubnetReq := NewPutSubnetRequest(i+1, 40)
			_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq)
			Expect(err).Should(Succeed())
			reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq)
			reserveSubnetReqs = append(reserveSubnetReqs, reserveSubnetReq)
			releaseSubnetReq := NewReleaseSubnetRequest(reserveSubnetReq)
			releaseSubnetReqs = append(releaseSubnetReqs, releaseSubnetReq)
		}

		for i := 0; i < numberOfSubnets; i++ {
			By("ReserveSubnet subnet concurrently")
			reserveSubnetReq := reserveSubnetReqs[i]
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				By("ReserveSubnet should succeed")
				reserveSubnetResp, err := ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
				Expect(err).Should(Succeed())
				reservedSubnetsMap.Store(reserveSubnetResp.Subnet, true)

				By("ReserveSubnet again should succeed")
				_, err = ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
				Expect(err).Should(Succeed())
			}()
		}
		wg.Wait()
		Expect(getMapSize(&reservedSubnetsMap)).Should(Equal(numberOfSubnets))

		for i := 0; i < numberOfSubnets; i++ {
			By("ReleaseSubnet subnet concurrently")
			releaseSubnetReq := releaseSubnetReqs[i]
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				_, err := ipResourceManagerClient.ReleaseSubnet(ctx, releaseSubnetReq)
				Expect(err).Should(Succeed())

				By("ReleaseSubnet again should return NotFound")
				_, err = ipResourceManagerClient.ReleaseSubnet(ctx, releaseSubnetReq)
				Expect(status.Code(err)).Should(Equal(codes.NotFound))
			}()
		}
		wg.Wait()
	})

	It("Reserve and release address should succeed concurrently", func() {
		var wg sync.WaitGroup
		numberOfSubnets := 10
		var reservedSubnetsMap sync.Map
		var reservedAddressMap sync.Map
		var reserveSubnetReqs []*pb.ReserveSubnetRequest
		var reserveAddressReqs []*pb.ReserveAddressRequest

		for i := 0; i < numberOfSubnets; i++ {
			By("Create PutSubnet, ReserveSubnet, ReserveAddress requests")
			putSubnetReq := NewPutSubnetRequest(i+1, 40)
			_, err := ipResourceManagerClient.PutSubnet(ctx, putSubnetReq)
			Expect(err).Should(Succeed())
			reserveSubnetReq := NewReserveSubnetRequest(putSubnetReq)
			reserveSubnetReqs = append(reserveSubnetReqs, reserveSubnetReq)
			reserveAddressReq := NewReserveAddressRequest(reserveSubnetReq)
			reserveAddressReqs = append(reserveAddressReqs, reserveAddressReq)
		}

		for i := 0; i < numberOfSubnets; i++ {
			By("ReserveSubnet should succeed")
			reserveSubnetReq := reserveSubnetReqs[i]
			reserveSubnetResp, err := ipResourceManagerClient.ReserveSubnet(ctx, reserveSubnetReq)
			Expect(err).Should(Succeed())
			reservedSubnetsMap.Store(reserveSubnetResp.Subnet, true)
		}
		Expect(getMapSize(&reservedSubnetsMap)).Should(Equal(numberOfSubnets))

		for i := 0; i < numberOfSubnets; i++ {
			By("ReserveAddress concurrently should succeed")
			reserveAddressReq := reserveAddressReqs[i]
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				By("ReserveAddress concurrently should succeed")
				reserveAddressResp, err := ipResourceManagerClient.ReserveAddress(ctx, reserveAddressReq)
				Expect(err).Should(Succeed())
				reservedAddressMap.Store(reserveAddressResp.Address, true)
			}()
		}
		wg.Wait()
		Expect(getMapSize(&reservedAddressMap)).Should(Equal(numberOfSubnets))

		for i := 0; i < numberOfSubnets; i++ {
			By("ReleaseAddress address concurrently")
			reserveAddressReq := reserveAddressReqs[i]
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				By("ReleaseAddress concurrently should succeed")
				releaseAddressReq := NewReleaseAddressRequest(reserveAddressReq)
				_, err := ipResourceManagerClient.ReleaseAddress(ctx, releaseAddressReq)
				Expect(err).Should(Succeed())
			}()
		}
		wg.Wait()
	})
})
