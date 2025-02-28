// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func BenchmarkModifyRouter(b *testing.B) {
	const maxConcurrency = 32
	const numRequest = 256

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c = v1.NewOvnnetClient(conn)

	// Create VPCs in advance for all following tests
	vpcIds := make([]string, 0)
	for i := 0; i < maxConcurrency; i++ {
		vpcId := uuid.NewString()
		vpcIds = append(vpcIds, vpcId)
		_, err = c.CreateVPC(context.Background(), &v1.CreateVPCRequest{VpcId: &v1.VPCId{Uuid: vpcId}, RegionId: vpcId, TenantId: vpcId, Name: vpcId})
		if err != nil {
			log.Fatalf("fail to initialize VPC: %v", err)
		}
		defer c.DeleteVPC(context.Background(), &v1.DeleteVPCRequest{VpcId: &v1.VPCId{Uuid: vpcId}})
	}
	// Encode requests
	createRouterReqs := make([]*v1.CreateRouterRequest, 0)
	deleteRouterReqs := make([]*v1.DeleteRouterRequest, 0)
	for i := 0; i < numRequest; i++ {
		routerId := uuid.NewString()
		createRouterReqs = append(createRouterReqs, &v1.CreateRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
			Name:     routerId,
			// Leave VPC for empty now
		})
		deleteRouterReqs = append(deleteRouterReqs, &v1.DeleteRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
		})
	}

	numVpc := []int{1, 2, 4, 8}
	numWorker := []int{1, 2, 4, 8}

	for _, w := range numWorker {
		for _, v := range numVpc {
			b.Run(fmt.Sprintf("%d VPC,%d Worker", v, w), func(benchmark *testing.B) {
				numReqPerWorker := numRequest / w
				for i := 0; i < benchmark.N; i++ {
					var wg sync.WaitGroup
					benchmark.StartTimer()
					wg.Add(w)
					for j := 0; j < w; j++ {
						go func(offset int) {
							defer wg.Done()
							for k := 0; k < numReqPerWorker; k++ {
								createRouterReqs[offset+k].VpcId = &v1.VPCId{Uuid: vpcIds[(offset+k)%v]}
								c.CreateRouter(context.Background(), createRouterReqs[offset+k])
							}
							for k := 0; k < numReqPerWorker; k++ {
								c.DeleteRouter(context.Background(), deleteRouterReqs[offset+k])
							}
						}(j * numReqPerWorker)
					}
					wg.Wait()
					benchmark.StopTimer()
				}
			})
		}
	}
}
