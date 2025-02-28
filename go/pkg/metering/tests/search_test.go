// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/db/query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSearch(t *testing.T) {
	id := "123456789102"
	filter := &pb.UsageFilter{
		CloudAccountId: &id,
	}

	res, err := client.Search(context.Background(), filter)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if res == nil {
		t.Fatalf("Search returned nil result")
	}

}

func TestSearchWithNonMatchingFilter(t *testing.T) {
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
				"instanceType":        "tiny",
				"serviceType":         "ComputeAsAService",
				"region":              "us-dev-1",
				"runningSeconds":      "192624.136468704",
			},
			Timestamp: timestamppb.Now(),
		})
	id := "123456789102"
	filter := &pb.UsageFilter{
		CloudAccountId: &id,
	}
	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	expectedRes := id
	stream, _ := client.Search(context.Background(), filter)

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected errort %v", err)
		}

		if !reflect.DeepEqual(resp.CloudAccountId, expectedRes) {
			t.Fatalf("Expected result %v, but got %v", expectedRes, resp)
		}
	}
}

func TestSearchWithNonMatchingRecord(t *testing.T) {
	_, err := client.Create(context.Background(),
		&pb.UsageCreate{
			CloudAccountId: "123456789102",
			ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1999",
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e199",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"serviceType":         "ComputeAsAService",
				"instanceType":        "tiny",
				"region":              "us-dev-1",
				"runningSeconds":      "192624.136468704",
			},
			Timestamp: timestamppb.Now(),
		})
	rid := "3bc52387-da79-4947-a562-ab7a88c38e1999"
	filter := &pb.UsageFilter{
		ResourceId: &rid,
	}
	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	expectedRes := rid

	stream, _ := client.Search(context.Background(), filter)
	resp, err := stream.Recv()
	if err != nil {
		t.Fatalf("unexpected errort %v", err)
	}
	if !reflect.DeepEqual(resp.ResourceId, expectedRes) {
		t.Fatalf("Expected result %v, but got %v", expectedRes, resp)
	}
}

func TestSearchWithNonMatchingTransaction(t *testing.T) {
	_, err := client.Create(context.Background(),
		&pb.UsageCreate{
			CloudAccountId: "123456789102",
			ResourceId:     "3bc52387-da79-4947-a562-ab7a88c38e1997",
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e197",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "tiny",
				"region":              "us-dev-1",
				"runningSeconds":      "192624.136468704",
			},
			Timestamp: timestamppb.Now(),
		})
	tid := "3bc52387-da79-4947-a562-ab7a88c38e197"
	filter := &pb.UsageFilter{
		TransactionId: &tid,
	}
	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	expectedRes := tid

	stream, _ := client.Search(context.Background(), filter)
	resp, err := stream.Recv()
	if !reflect.DeepEqual(resp.TransactionId, expectedRes) {
		t.Fatalf("Expected result %v, but got %v", expectedRes, resp)
	}
	if err != nil {
		t.Fatalf("unexpected errort %v", err)
	}
}

func TestSearchWithMultipleFilters(t *testing.T) {
	id := "123456789102"
	rid := "3bc52387-da79-4947-a562-ab7a88c38e1998"
	filter := &pb.UsageFilter{
		CloudAccountId: &id,
		ResourceId:     &rid,
	}

	_, err := client.Create(context.Background(),
		&pb.UsageCreate{
			CloudAccountId: "123456789102",
			ResourceId:     rid,
			TransactionId:  "3bc52387-da79-4947-a562-ab7a88c38e199",
			Properties: map[string]string{
				"availabilityZone":    "us-dev-1a",
				"clusterId":           "harvester1",
				"deleted":             "false",
				"firstReadyTimestamp": "2023-02-16T14:53:29Z",
				"instanceType":        "tiny",
				"region":              "us-dev-1",
				"runningSeconds":      "192624.136468704",
			},
			Timestamp: timestamppb.Now(),
		})
	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	expectedRes := rid

	stream, _ := client.Search(context.Background(), filter)
	resp, err := stream.Recv()
	if !reflect.DeepEqual(resp.ResourceId, expectedRes) {
		t.Fatalf("Expected result %v, but got %v", expectedRes, resp)
	}
	if err != nil {
		t.Fatalf("unexpected errort %v", err)
	}
}

func TestSearchInvalidMeteringRecordWithMultipleFilters(t *testing.T) {
	cloudAcctId := "123456789102"
	rid := "3bc52387-da79-4947-a562-ab7a88c38e1998"
	filter := &pb.InvalidMeteringRecordFilter{
		CloudAccountId: &cloudAcctId,
		ResourceId:     &rid,
	}

	invalidMeteringRecord := &pb.InvalidMeteringRecordCreate{
		CloudAccountId:                 cloudAcctId,
		ResourceId:                     rid,
		RecordId:                       "123456789103",
		Region:                         "us-west-1",
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

	stream, _ := client.SearchInvalid(context.Background(), filter)
	resp, err := stream.Recv()

	if resp.ResourceId != rid {
		t.Fatalf("resource id does not match")
	}
	if err != nil {
		t.Fatalf("unexpected errort %v", err)
	}
}

func getDefaultUsageCreate(cloudAcctId string, resourceId string, transactionId string, timestamp *timestamppb.Timestamp) *pb.UsageCreate {
	return &pb.UsageCreate{
		CloudAccountId: cloudAcctId,
		ResourceId:     resourceId,
		TransactionId:  transactionId,
		Properties: map[string]string{
			"availabilityZone":    "us-dev-1a",
			"clusterId":           "harvester1",
			"deleted":             "false",
			"firstReadyTimestamp": "2023-02-16T14:53:29Z",
			"instanceType":        "tiny",
			"region":              "us-dev-1",
			"runningSeconds":      "192624.136468704",
		},
		Timestamp: timestamp,
	}
}

func createMultipleUsage(cloudAcctId string, resourceId string, numberOfUsage int, t *testing.T) {
	for i := 0; i < numberOfUsage; i++ {
		transactionId := uuid.NewString()
		_, err := client.Create(context.Background(),
			getDefaultUsageCreate(cloudAcctId, resourceId, transactionId, timestamppb.Now()))

		if err != nil {
			t.Fatalf("create metering record: %v", err)
		}
	}
}

func TestSearchResourceMeteringRecordsNoMRecords(t *testing.T) {

	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestSearchResourceMeteringRecordsNoMRecords")
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAcctId := "123456789102"
	filter := &pb.MeteringFilter{
		CloudAccountId: &cloudAcctId,
	}

	err := query.DeleteAllRecords(ctx, server.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	resp, err := client.SearchResourceMeteringRecords(ctx, filter)

	if err != nil {
		t.Fatalf("unexpected error for searching %v", err)
	}

	if resp.ResourceMeteringRecordsList != nil {
		t.Fatalf("should not have expected resource metering records list")
	}
}

func TestSearchResourceMeteringRecordsAsStreamNoMRecords(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestSearchResourceMeteringRecordsAsStreamNoMRecords")
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAcctId := "123456789102"

	filter := &pb.MeteringFilter{
		CloudAccountId: &cloudAcctId,
	}

	err := query.DeleteAllRecords(ctx, server.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	stream, _ := client.SearchResourceMeteringRecordsAsStream(ctx, filter)
	_, err = stream.Recv()

	if err == nil {
		t.Fatalf("should have received a error")
	}
}

func TestSearchResourceMeteringRecords(t *testing.T) {

	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestSearchResourceMeteringRecords")
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAcctId := "123456789102"
	resourceId := uuid.NewString()
	resourceId1 := uuid.NewString()

	filter := &pb.MeteringFilter{
		CloudAccountId: &cloudAcctId,
	}

	transactionId := "oldTransactionId"
	transactionId1 := "newTransactionId"

	err := query.DeleteAllRecords(ctx, server.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	_, err = client.Create(ctx, getDefaultUsageCreate(cloudAcctId, resourceId, transactionId, timestamppb.Now()))

	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	_, err = client.Create(ctx, getDefaultUsageCreate(cloudAcctId, resourceId,
		transactionId1, timestamppb.New(time.Now().Add(300))))
	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	_, err = client.Create(ctx, getDefaultUsageCreate(cloudAcctId, resourceId1,
		uuid.NewString(), timestamppb.Now()))

	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	resp, err := client.SearchResourceMeteringRecords(ctx, filter)

	if err != nil {
		t.Fatalf("unexpected error for searching %v", err)
	}

	if len(resp.ResourceMeteringRecordsList) != 2 {
		t.Fatalf("expected 2 instances for the cloud acct, but got %d", len(resp.ResourceMeteringRecordsList))
	}

	for _, resourceMeteringRecords := range resp.ResourceMeteringRecordsList {
		if resourceMeteringRecords.ResourceId == resourceId {
			if len(resourceMeteringRecords.MeteringRecords) != 2 {
				t.Fatalf("expected 2 metering records resource, but got %d", len(resp.ResourceMeteringRecordsList[0].MeteringRecords))
			}

			if resourceMeteringRecords.MeteringRecords[0].TransactionId != transactionId1 {
				t.Fatalf("the ordering of the metering records is wrong for the resource: " + resourceMeteringRecords.ResourceId)
			}
		}

		if resourceMeteringRecords.ResourceId == resourceId1 {
			if len(resourceMeteringRecords.MeteringRecords) != 1 {
				t.Fatalf("expected 1 metering records resource, but got %d", len(resp.ResourceMeteringRecordsList[0].MeteringRecords))
			}
		}
	}
}

func TestSearchResourceMeteringRecordsAsStream(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestSearchResourceMeteringRecordsAsStream")
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAcctId := "123456789102"
	resourceId := uuid.NewString()
	resourceId1 := uuid.NewString()

	filter := &pb.MeteringFilter{
		CloudAccountId: &cloudAcctId,
	}

	transactionId := "oldTransactionId"
	transactionId1 := "newTransactionId"

	err := query.DeleteAllRecords(ctx, server.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	_, err = client.Create(ctx, getDefaultUsageCreate(cloudAcctId, resourceId, transactionId, timestamppb.Now()))

	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	_, err = client.Create(ctx, getDefaultUsageCreate(cloudAcctId, resourceId,
		transactionId1, timestamppb.New(time.Now().Add(300))))

	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	_, err = client.Create(ctx, getDefaultUsageCreate(cloudAcctId, resourceId1,
		uuid.NewString(), timestamppb.Now()))

	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	stream, _ := client.SearchResourceMeteringRecordsAsStream(ctx, filter)
	resp, err := stream.Recv()

	if err != nil {
		t.Fatalf("failed to get the resource metering record: %v", err)
	}

	lengthOfResourceMeteringRecords := len(resp.ResourceMeteringRecordsList)

	if lengthOfResourceMeteringRecords != 2 {
		t.Fatalf("expected 2 instances for the cloud acct, but got %d", len(resp.ResourceMeteringRecordsList))
	}

	for _, resourceMeteringRecords := range resp.ResourceMeteringRecordsList {
		if resourceMeteringRecords.ResourceId == resourceId {
			if len(resourceMeteringRecords.MeteringRecords) != 2 {
				t.Fatalf("expected 2 metering records resource, but got %d", len(resp.ResourceMeteringRecordsList[0].MeteringRecords))
			}

			if resourceMeteringRecords.MeteringRecords[0].TransactionId != transactionId1 {
				t.Fatalf("the ordering of the metering records is wrong for the resource: " + resourceMeteringRecords.ResourceId)
			}
		}
		if resourceMeteringRecords.ResourceId == resourceId1 {
			if len(resourceMeteringRecords.MeteringRecords) != 1 {
				t.Fatalf("expected 1 metering records resource, but got %d", len(resp.ResourceMeteringRecordsList[0].MeteringRecords))
			}
		}
	}
}

func TestSearchResourceMeteringRecordsAsStreamWithPagination(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestSearchResourceMeteringRecordsAsStreamWithPagination")
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAcctId := "123456789102"

	filter := &pb.MeteringFilter{
		CloudAccountId: &cloudAcctId,
	}

	err := query.DeleteAllRecords(ctx, server.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	numOfResources := 2*query.FixedPaginationSize + 1

	for i := 0; i < numOfResources; i++ {
		_, err := client.Create(ctx,
			getDefaultUsageCreate(cloudAcctId, uuid.NewString(), uuid.NewString(), timestamppb.Now()))

		if err != nil {
			t.Fatalf("create metering record: %v", err)
		}
	}
	stream, _ := client.SearchResourceMeteringRecordsAsStream(ctx, filter)

	// the check for lengths can move to a method..
	resp, err := stream.Recv()

	if err != nil {
		t.Fatalf("failed to get the resource metering record: %v", err)
	}

	lengthOfResourceMeteringRecords := len(resp.ResourceMeteringRecordsList)

	// the pagination size is hard coded and should be configurable.
	if lengthOfResourceMeteringRecords != query.FixedPaginationSize {
		t.Fatalf("there should be %d records", query.FixedPaginationSize)
	}

	resp, err = stream.Recv()

	if err != nil {
		t.Fatalf("failed to get the resource metering record: %v", err)
	}

	lengthOfResourceMeteringRecords = len(resp.ResourceMeteringRecordsList)

	// the pagination size is hard coded and should be configurable.
	if lengthOfResourceMeteringRecords != query.FixedPaginationSize {
		t.Fatalf("there should be %d records", query.FixedPaginationSize)
	}

	resp, err = stream.Recv()

	if err != nil {
		t.Fatalf("failed to get the resource metering record: %v", err)
	}

	lengthOfResourceMeteringRecords = len(resp.ResourceMeteringRecordsList)

	if lengthOfResourceMeteringRecords != 1 {
		t.Fatalf("there should be only five record")
	}

	_, err = stream.Recv()

	if err == nil {
		t.Fatalf("should have received a error")
	}

	if !errors.Is(err, io.EOF) {
		t.Fatalf("should have received a end of file")
	}
}

func TestSearchResourceMeteringRecordsPerf(t *testing.T) {
	t.Skip()
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestSearchResourceMeteringRecordsPerf")
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAcctId := "123456789102"

	filter := &pb.MeteringFilter{
		CloudAccountId: &cloudAcctId,
	}

	err := query.DeleteAllRecords(ctx, server.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	for i := 0; i < 20; i++ {
		createMultipleUsage(cloudAcctId, uuid.NewString(), 1000, t)
	}

	_, err = client.SearchResourceMeteringRecords(ctx, filter)

	if err != nil {
		t.Fatalf("unexpected error for searching %v", err)
	}
}

func TestSearchResourceMeteringRecordFilters(t *testing.T) {

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestSearchResourceMeteringRecordFilters")
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAcctId := "123456789102"
	resourceId := uuid.NewString()
	resourceId1 := uuid.NewString()

	transactionId := "oldTransactionId"
	transactionId1 := "newTransactionId"

	err := query.DeleteAllRecords(ctx, server.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	_, err = client.Create(ctx, getDefaultUsageCreate(cloudAcctId, resourceId, transactionId, timestamppb.Now()))

	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	_, err = client.Create(ctx, getDefaultUsageCreate(cloudAcctId, resourceId1, transactionId1, timestamppb.Now()))

	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	logger.Info("testing for cloud account id filter")
	getAndVerifyForSearchFilter(ctx, t, &pb.MeteringFilter{CloudAccountId: &cloudAcctId}, 2)

	logger.Info("testing for not reported")
	reported := false
	getAndVerifyForSearchFilter(ctx, t, &pb.MeteringFilter{Reported: &reported}, 2)
}

func getAndVerifyForSearchFilter(ctx context.Context, t *testing.T, filter *pb.MeteringFilter, sizeOfResourceRecordsExpected int) {
	resp, err := client.SearchResourceMeteringRecords(ctx, filter)

	if err != nil {
		t.Fatalf("unexpected error for searching %v", err)
	}

	if len(resp.ResourceMeteringRecordsList) != sizeOfResourceRecordsExpected {
		t.Fatalf("expected %d instances for the cloud acct, but got %d", sizeOfResourceRecordsExpected, len(resp.ResourceMeteringRecordsList))
	}
}

func TestSearchResourceMeteringRecordInvalidFilters(t *testing.T) {

	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestSearchResourceMeteringRecordInvalidFilters")
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAcctId := "123456789102"
	resourceId := uuid.NewString()
	resourceId1 := uuid.NewString()

	transactionId := "oldTransactionId"
	transactionId1 := "newTransactionId"

	err := query.DeleteAllRecords(ctx, server.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	_, err = client.Create(ctx, getDefaultUsageCreate(cloudAcctId, resourceId, transactionId, timestamppb.Now()))

	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	_, err = client.Create(ctx, getDefaultUsageCreate(cloudAcctId, resourceId1, transactionId1, timestamppb.Now()))

	if err != nil {
		t.Fatalf("create metering record: %v", err)
	}

	logger.Info("testing for reported true")
	reported := true
	getAndVerifyNoResourceMeteringRecords(ctx, t, &pb.MeteringFilter{Reported: &reported})
}

func getAndVerifyNoResourceMeteringRecords(ctx context.Context, t *testing.T, filter *pb.MeteringFilter) {
	resp, err := client.SearchResourceMeteringRecords(ctx, filter)

	if err != nil {
		t.Fatalf("unexpected error for searching %v", err)
	}

	if resp.ResourceMeteringRecordsList != nil {
		t.Fatalf("should not have received resource metering records")
	}
}
