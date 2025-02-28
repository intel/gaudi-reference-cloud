// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMembers(t *testing.T) {

	_, err := client.Create(context.Background(),
		&pb.UsageCreate{
			// hardcoding the values for now, use uuid,variables to get it
			CloudAccountId: "123456789102",
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e1999",
			ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1998",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "vm-spr-sml",
				"region":              "us-dev-1",
				"runningSeconds":      "192624.136468704",
				"serviceType":         "ComputeAsAService",
			},
			Timestamp: timestamppb.Now(),
		})
	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}
	// pre-define it
	id := "123456789102"
	t.Run("addMembers", func(t *testing.T) { testCreateRecords(t, client, id) })
}

func testCreateRecords(t *testing.T, metClient pb.MeteringServiceClient,
	id string) {

	_, err := metClient.Search(context.Background(), &pb.UsageFilter{
		CloudAccountId: &id,
	})
	if err != nil {
		t.Fatalf(" %v", err)
	}

}

func TestCreateRecordMissingCloudAccountId(t *testing.T) {
	t.Skip()
	// create a new record without providing the CloudAccountId
	_, err := client.Create(context.Background(),
		&pb.UsageCreate{
			ResourceId:    "3bc52387-da79-4947-a562-ab7a88c38e1998",
			TransactionId: "3bc52387-da79-4947-a562-ab7a88c38e199",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "vm-spr-sml",
				"region":              "us-dev-1",
				"runningSeconds":      "192624.136468704",
				"serviceType":         "ComputeAsAService",
			},
			Timestamp: timestamppb.Now(),
		})
	if err == nil {
		t.Fatal("Expected an error, but none occurred")
	}

	expectedErrorMessage := "rpc error: code = InvalidArgument desc = invalid input arguments, ignoring record creation."
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Fatalf("Expected error message to contain %q, but got: %v", expectedErrorMessage, err)
	}
}

func TestCreateInvalidMeteringRecord(t *testing.T) {
	t.Skip()
	invalidMeteringRecord := &pb.InvalidMeteringRecordCreate{
		CloudAccountId:                 "123456789102",
		RecordId:                       "123456789103",
		Region:                         "us-west-1",
		ResourceId:                     uuid.NewString(),
		TransactionId:                  uuid.NewString(),
		Timestamp:                      timestamppb.Now(),
		MeteringRecordInvalidityReason: pb.MeteringRecordInvalidityReason_DUPLICATE_TRANSACTION_ID,
		Properties: map[string]string{
			"availabilityZone":    "us-dev-1a",
			"clusterId":           "harvester1",
			"deleted":             "false",
			"firstReadyTimestamp": "2023-02-16T14:53:29Z",
			"instanceType":        "vm-spr-sml",
			"region":              "us-dev-1",
			"runningSeconds":      "192624.136468704",
			"serviceType":         "ComputeAsAService",
		},
	}
	_, err := client.CreateInvalidRecords(context.Background(),
		&pb.CreateInvalidMeteringRecords{CreateInvalidMeteringRecords: []*pb.InvalidMeteringRecordCreate{invalidMeteringRecord}})
	if err != nil {
		t.Fatal("failed to create a invalid metering record")
	}
}

func TestCreateInvalidMeteringRecordMissingCloudAcctId(t *testing.T) {
	t.Skip()
	invalidMeteringRecord := &pb.InvalidMeteringRecordCreate{
		RecordId:                       "123456789103",
		Region:                         "us-west-1",
		ResourceId:                     uuid.NewString(),
		TransactionId:                  uuid.NewString(),
		Timestamp:                      timestamppb.Now(),
		MeteringRecordInvalidityReason: pb.MeteringRecordInvalidityReason_DUPLICATE_TRANSACTION_ID,
		Properties: map[string]string{
			"availabilityZone":    "us-dev-1a",
			"clusterId":           "harvester1",
			"deleted":             "false",
			"firstReadyTimestamp": "2023-02-16T14:53:29Z",
			"instanceType":        "vm-spr-sml",
			"region":              "us-dev-1",
			"runningSeconds":      "192624.136468704",
			"serviceType":         "ComputeAsAService",
		},
	}
	_, err := client.CreateInvalidRecords(context.Background(),
		&pb.CreateInvalidMeteringRecords{CreateInvalidMeteringRecords: []*pb.InvalidMeteringRecordCreate{invalidMeteringRecord}})
	if err == nil {
		t.Fatal("Expected an error, but none occurred")
	}

	expectedErrorMessage := "rpc error: code = InvalidArgument desc = invalid input arguments, ignoring invalid records creation"
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Fatalf("Expected error message to contain %q, but got: %v", expectedErrorMessage, err)
	}
}

func TestCreateMeteringRecordInvalidRegion(t *testing.T) {
	t.Skip()
	_, err := client.Create(context.Background(),
		&pb.UsageCreate{
			CloudAccountId: "123456789102",
			ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1998",
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e199",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "vm-spr-sml",
				"runningSeconds":      "192624.136468704",
				"serviceType":         "ComputeAsAService",
			},
			Timestamp: timestamppb.Now(),
		})
	if err == nil {
		t.Fatal("Expected an error, but none occurred")
	}

	expectedErrorMessage := "rpc error: code = InvalidArgument desc = missing region, ignoring record creation."
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Fatalf("Expected error message to contain %q, but got: %v", expectedErrorMessage, err)
	}

	_, err = client.Create(context.Background(),
		&pb.UsageCreate{
			CloudAccountId: "123456789102",
			ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1998",
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e199",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "vm-spr-sml",
				"region":              "",
				"runningSeconds":      "192624.136468704",
				"serviceType":         "ComputeAsAService",
			},
			Timestamp: timestamppb.Now(),
		})
	if err == nil {
		t.Fatal("Expected an error, but none occurred")
	}

	expectedErrorMessage = "rpc error: code = InvalidArgument desc = invalid region value, ignoring record creation."
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Fatalf("Expected error message to contain %q, but got: %v", expectedErrorMessage, err)
	}

	_, err = client.Create(context.Background(),
		&pb.UsageCreate{
			CloudAccountId: "123456789102",
			ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1998",
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e199",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "vm-spr-sml",
				"region":              "$$$",
				"runningSeconds":      "192624.136468704",
				"serviceType":         "ComputeAsAService",
			},
			Timestamp: timestamppb.Now(),
		})
	if err == nil {
		t.Fatal("Expected an error, but none occurred")
	}

	expectedErrorMessage = "rpc error: code = InvalidArgument desc = invalid region value, ignoring record creation."
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Fatalf("Expected error message to contain %q, but got: %v", expectedErrorMessage, err)
	}

	_, err = client.Create(context.Background(),
		&pb.UsageCreate{
			CloudAccountId: "123456789102",
			ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1998",
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e199",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "vm-spr-sml",
				"region":              "valid-region",
				"runningSeconds":      "192624.136468704",
				"serviceType":         "ComputeAsAService",
			},
			Timestamp: timestamppb.Now(),
		})
	if err != nil {
		t.Fatal("should not have had a error")
	}

}

func TestCreateMeteringRecordInvalidRunningSeconds(t *testing.T) {
	t.Skip()
	_, err := client.Create(context.Background(),
		&pb.UsageCreate{
			CloudAccountId: "123456789102",
			ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1998",
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e199",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "vm-spr-sml",
				"region":              "valid-region",
				"serviceType":         "ComputeAsAService",
			},
			Timestamp: timestamppb.Now(),
		})
	if err == nil {
		t.Fatal("Expected an error, but none occurred")
	}

	expectedErrorMessage := "rpc error: code = InvalidArgument desc = missing running seconds, ignoring record creation."
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Fatalf("Expected error message to contain %q, but got: %v", expectedErrorMessage, err)
	}

	_, err = client.Create(context.Background(),
		&pb.UsageCreate{
			CloudAccountId: "123456789102",
			ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1998",
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e199",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "vm-spr-sml",
				"region":              "valid-region",
				"runningSeconds":      "",
				"serviceType":         "ComputeAsAService",
			},
			Timestamp: timestamppb.Now(),
		})
	if err == nil {
		t.Fatal("Expected an error, but none occurred")
	}

	expectedErrorMessage = "rpc error: code = InvalidArgument desc = invalid running seconds value, ignoring record creation."
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Fatalf("Expected error message to contain %q, but got: %v", expectedErrorMessage, err)
	}

	_, err = client.Create(context.Background(),
		&pb.UsageCreate{
			CloudAccountId: "123456789102",
			ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1998",
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e199",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "vm-spr-sml",
				"region":              "valid-region",
				"runningSeconds":      "192624.136468704",
				"serviceType":         "ComputeAsAService",
			},
			Timestamp: timestamppb.Now(),
		})
	if err != nil {
		t.Fatal("should not have had a error")
	}

}
