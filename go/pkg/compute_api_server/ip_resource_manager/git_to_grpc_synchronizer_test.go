// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ip_resource_manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func searchSubnets(ctx context.Context) []*pb.Subnet {
	var subnets []*pb.Subnet
	searchReq1 := NewSearchSubnetRequest()
	stream, err := ipResourceManagerClient.SearchSubnetStream(ctx, searchReq1)
	Expect(err).Should(Succeed())

	for {
		subnet, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return subnets
		}
		Expect(err).Should(Succeed())
		subnets = append(subnets, subnet)
	}
}

var _ = Describe("GitToGrpcSynchronizer Subnet Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	runSync := func(dir string, expectedCount int) {
		By("Connecting to GRPC server")
		clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).Should(Succeed())
		By("Creating synchronizer (" + dir + ")")
		synchronizer, err := NewSubnetSynchronizer(
			os.DirFS("../testdata/"+dir),
			clientConn,
		)
		Expect(err).Should(Succeed())
		By("Synchronize should succeed (" + dir + ")")
		_, err = synchronizer.Synchronize(ctx)
		Expect(err).Should(Succeed())

		By("SearchSubnetStream should return expected count")
		subnets := searchSubnets(ctx)
		Expect(len(subnets)).Should(Equal(expectedCount))
	}

	It("Synchronize", Serial, func() {
		runSync("Subnet1", 2)
		runSync("Subnet2-delete", 1)
		runSync("Subnet3-update", 1)
	})
})
