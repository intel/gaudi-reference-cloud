// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	mrand "math/rand"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/zitadel/oidc/pkg/crypto"
	"google.golang.org/grpc/metadata"
	"gopkg.in/square/go-jose.v2"
)

func NewUpdateComputeNodePoolsForCloudAccountRequest(cloudAccountId string, computeNodePoolIds []string, createAdmin string) *pb.UpdateComputeNodePoolsForCloudAccountRequest {
	req := &pb.UpdateComputeNodePoolsForCloudAccountRequest{
		CloudAccountId: cloudAccountId,
		CreateAdmin:    createAdmin,
	}
	for _, poolId := range computeNodePoolIds {
		req.ComputeNodePools = append(req.ComputeNodePools, &pb.ComputeNodePoolForInstanceScheduling{
			PoolId: poolId,
		})
	}
	return req
}

func NewSearchComputeNodePoolsForInstanceSchedulingRequest(cloudAccountId string) *pb.SearchComputeNodePoolsForInstanceSchedulingRequest {
	req := &pb.SearchComputeNodePoolsForInstanceSchedulingRequest{
		CloudAccountId: cloudAccountId,
	}
	return req
}

func NewAddCloudAccountToComputeNodePool(cloudAccountId string, poolId string, createAdmin string) *pb.AddCloudAccountToComputeNodePoolRequest {
	req := &pb.AddCloudAccountToComputeNodePoolRequest{
		CloudAccountId: cloudAccountId,
		CreateAdmin:    createAdmin,
		PoolId:         poolId,
	}
	return req
}

func NewDeleteCloudAccountFromComputeNodePool(cloudAccountId string, poolId string) *pb.DeleteCloudAccountFromComputeNodePoolRequest {
	req := &pb.DeleteCloudAccountFromComputeNodePoolRequest{
		CloudAccountId: cloudAccountId,
		PoolId:         poolId,
	}
	return req
}

func NewPutComputeNodePool(poolId string, poolName string, poolAccountManagerAgsRole string) *pb.PutComputeNodePoolRequest {
	req := &pb.PutComputeNodePoolRequest{
		PoolId:                    poolId,
		PoolName:                  poolName,
		PoolAccountManagerAgsRole: poolAccountManagerAgsRole,
	}
	return req
}

func NewSearchCloudAccountsForComputeNodePool(poolId string) *pb.SearchCloudAccountsForComputeNodePoolRequest {
	req := &pb.SearchCloudAccountsForComputeNodePoolRequest{
		PoolId: poolId,
	}
	return req
}

func NewSearchComputeNodePoolsForPoolAccountManager(ctx context.Context, agsRoles []string) (context.Context, error) {
	jwtMap := map[string]interface{}{
		"tid":          "yyy",
		"enterpriseId": "xxx",
		"email":        "test@user1.com",
		"roles":        agsRoles,
		"idp":          "intelcorpintb2c.onmicrosoft.com",
		"countryCode":  "IN",
	}

	jwtToken, err := generateJWT(jwtMap)
	if err != nil {
		return nil, err
	}
	jwtToken = "Bearer " + jwtToken
	header := metadata.New(map[string]string{"authorization": jwtToken})
	ctxWithJwtToken := metadata.NewOutgoingContext(ctx, header)

	return ctxWithJwtToken, nil
}

func generateJWT(input map[string]interface{}) (string, error) {
	claims := map[string]any{}
	claims["iss"] = "http://issuer:port"
	now := time.Now().UTC()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	claims["exp"] = now.Add(5 * time.Minute).Unix()
	for key, val := range input {
		claims[key] = val
	}

	rsaTestKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: rsaTestKey}, nil)
	if err != nil {
		return "", err
	}

	tok, err := crypto.Sign(claims, signer)
	if err != nil {
		return "", err
	}

	return tok, nil

}

// randomInt generates a random integer between min and max using math/rand with a fixed seed
func randomInt(min, max int, rSrcInput *mrand.Rand) int {
	return rSrcInput.Intn(max-min+1) + min
}

// randomString selects a random string from the given options using math/rand with a fixed seed
func randomString(options []string, rSrcInput *mrand.Rand) string {
	return options[randomInt(0, len(options)-1, rSrcInput)]
}

// NewReportNodeStatisticsRequest generates a report for a given number of nodes
func NewReportNodeStatisticsRequest(numberOfNodes int, rSrcInput *mrand.Rand) *pb.ReportNodeStatisticsRequest {
	// Base request
	req := &pb.ReportNodeStatisticsRequest{
		SchedulerNodeStatistics: []*pb.SchedulerNodeStatistics{},
	}

	// Define the available regions and instance types
	regions := []string{"us-dev-1", "us-dev-2"}
	instanceTypes := []string{"vm-spr-sml", "vm-icp-gaudi2-1"}

	// Loop to create multiple SchedulerNodeStatistics based on numberOfNodes
	for i := 1; i <= numberOfNodes; i++ {
		// Randomly select a region
		region := randomString(regions, rSrcInput)

		// Set availability zones based on the selected region
		availabilityZones := []string{
			fmt.Sprintf("%s-1a", region),
			fmt.Sprintf("%s-1b", region),
		}

		// Create a new node with random data
		node := &pb.SchedulerNode{
			Region:           region,
			AvailabilityZone: randomString(availabilityZones, rSrcInput),
			NodeName:         fmt.Sprintf("harvester%d/node-%d", randomInt(1, 5, rSrcInput), i),
			ClusterId:        fmt.Sprintf("harvester%d", randomInt(1, 5, rSrcInput)),
			Partition:        fmt.Sprintf("pdx05-c01-bspr00%d-hvst", randomInt(1, 9, rSrcInput)),
			SourceGvr: &pb.GroupVersionResource{
				Version:  "v1",
				Resource: "nodes",
			},
		}

		// Define random InstanceTypeStatistics for the node
		instanceTypeStats := []*pb.InstanceTypeStatistics{
			{
				InstanceType:     instanceTypes[0],
				RunningInstances: int32(randomInt(1, 10, rSrcInput)),
				MaxNewInstances:  int32(randomInt(1, 10, rSrcInput)),
				InstanceCategory: "VirtualMachine",
			},
			{
				InstanceType:     instanceTypes[1],
				RunningInstances: int32(randomInt(1, 10, rSrcInput)),
				MaxNewInstances:  int32(randomInt(1, 10, rSrcInput)),
				InstanceCategory: "VirtualMachine",
			},
		}

		// Define random NodeResources for the node
		nodeResources := &pb.NodeResources{
			FreeMilliCPU:    int64(randomInt(50000, 120000, rSrcInput)),
			UsedMilliCPU:    int64(randomInt(100000, 150000, rSrcInput)),
			FreeMemoryBytes: int64(randomInt(50000000000, 60000000000, rSrcInput)),
			UsedMemoryBytes: int64(randomInt(400000000000, 500000000000, rSrcInput)),
			FreeGPU:         int32(randomInt(0, 8, rSrcInput)),
			UsedGPU:         int32(randomInt(0, 8, rSrcInput)),
		}

		// Create SchedulerNodeStatistics for the current node
		schedulerNodeStat := &pb.SchedulerNodeStatistics{
			SchedulerNode:          node,
			InstanceTypeStatistics: instanceTypeStats,
			NodeResources:          nodeResources,
		}

		// Add the node statistics to the request
		req.SchedulerNodeStatistics = append(req.SchedulerNodeStatistics, schedulerNodeStat)
	}

	return req
}
