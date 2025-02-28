// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultInvalidProductUsageRecordQty = 10

func TestProductUsageRecord(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestProductUsageRecord")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageRecordServiceClient := pb.NewUsageRecordServiceClient(clientConn)

	usageRecordData := NewUsageRecordData(usageDb)

	err := usageRecordData.DeleteAllUsageRecords(ctx)
	if err != nil {
		t.Fatalf("could not delete all usage records")
	}

	cloudAccountId := MustNewCloudAcctId()
	productName := uuid.NewString()
	region := "us-west-1"
	transactionId := uuid.NewString()

	productUsageRecordCreate := GetProductUsageRecordCreate(cloudAccountId, &productName, region,
		transactionId, defaultProductUsageRecordQty,
		map[string]string{
			"availabilityZone": region,
			"serviceType":      "ComputeAsAService",
		})

	_, err = usageRecordServiceClient.CreateProductUsageRecord(ctx, productUsageRecordCreate)
	if err != nil {
		t.Fatalf("failed to create product usage record %v", err)
	}

	productUsageRecordsFilter := &pb.ProductUsageRecordsFilter{
		CloudAccountId: &cloudAccountId,
	}

	productUsageRecordsReturned := getProductUsageRecordsBySearch(t, ctx, productUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search product usage records for cloud account filter %v", err)
	}

	if productUsageRecordsReturned == nil || len(productUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of product usage records returned for cloud account filter")
	}

	if checkIfValuesForProductUsageRecordDoNotMatch(productUsageRecordCreate, productUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for cloud account filter")
	}

	productUsageRecordsFilter = &pb.ProductUsageRecordsFilter{
		Region: &region,
	}

	productUsageRecordsReturned = getProductUsageRecordsBySearch(t, ctx, productUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search product usage records for region filter %v", err)
	}

	if productUsageRecordsReturned == nil || len(productUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of product usage records returned for region filter")
	}

	if checkIfValuesForProductUsageRecordDoNotMatch(productUsageRecordCreate, productUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for region filter")
	}

	productUsageRecordsFilter = &pb.ProductUsageRecordsFilter{
		TransactionId: &transactionId,
	}

	productUsageRecordsReturned = getProductUsageRecordsBySearch(t, ctx, productUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search product usage records for transaction id filter %v", err)
	}

	if productUsageRecordsReturned == nil || len(productUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of product usage records returned for transaction id filter")
	}

	if checkIfValuesForProductUsageRecordDoNotMatch(productUsageRecordCreate, productUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for transaction id filter")
	}

	reported := false

	productUsageRecordsFilter = &pb.ProductUsageRecordsFilter{
		Reported: &reported,
	}

	productUsageRecordsReturned = getProductUsageRecordsBySearch(t, ctx, productUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search product usage records for reported filter %v", err)
	}

	if productUsageRecordsReturned == nil || len(productUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of product usage records returned for reported filter")
	}

	if checkIfValuesForProductUsageRecordDoNotMatch(productUsageRecordCreate, productUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for reported filter")
	}

	startTime := time.Now().AddDate(0, -1, 0)
	productUsageRecordsFilter = &pb.ProductUsageRecordsFilter{
		StartTime: timestamppb.New(startTime),
	}

	productUsageRecordsReturned = getProductUsageRecordsBySearch(t, ctx, productUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search product usage records for start time filter %v", err)
	}

	if productUsageRecordsReturned == nil || len(productUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of product usage records returned for start time filter")
	}

	if checkIfValuesForProductUsageRecordDoNotMatch(productUsageRecordCreate, productUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for start time filter")
	}

	endTime := time.Now().AddDate(0, +1, 0)
	productUsageRecordsFilter = &pb.ProductUsageRecordsFilter{
		EndTime: timestamppb.New(endTime),
	}

	productUsageRecordsReturned = getProductUsageRecordsBySearch(t, ctx, productUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search product usage records for end time filter %v", err)
	}

	if productUsageRecordsReturned == nil || len(productUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of product usage records returned for end time filter")
	}

	if checkIfValuesForProductUsageRecordDoNotMatch(productUsageRecordCreate, productUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for start time filter")
	}

	invalidCloudAcctId := MustNewCloudAcctId()
	verifyNoProductUsageRecordsReturned(ctx, t, &pb.ProductUsageRecordsFilter{
		CloudAccountId: &invalidCloudAcctId,
	})

	verifyNoProductUsageRecordsReturned(ctx, t, &pb.ProductUsageRecordsFilter{
		StartTime: timestamppb.New(endTime),
	})

	verifyNoProductUsageRecordsReturned(ctx, t, &pb.ProductUsageRecordsFilter{
		EndTime: timestamppb.New(startTime),
	})
}

func getProductUsageRecordsBySearch(t *testing.T, ctx context.Context,
	productUsageRecordsFilter *pb.ProductUsageRecordsFilter) []*pb.ProductUsageRecord {

	productUsageRecords := []*pb.ProductUsageRecord{}

	usageRecordServiceClient := pb.NewUsageRecordServiceClient(clientConn)
	stream, _ := usageRecordServiceClient.SearchProductUsageRecords(ctx, productUsageRecordsFilter)

	for {

		productUsageRecord, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("unexpected error for searching product usage records %v", err)
		}

		productUsageRecords = append(productUsageRecords, productUsageRecord)
	}

	return productUsageRecords
}

func checkIfValuesForProductUsageRecordDoNotMatch(productUsageRecordCreate *pb.ProductUsageRecordCreate,
	productUsageRecordReturned *pb.ProductUsageRecord) bool {
	if (productUsageRecordCreate.CloudAccountId != productUsageRecordReturned.CloudAccountId) ||
		(*productUsageRecordCreate.ProductName != *productUsageRecordReturned.ProductName) ||
		(productUsageRecordCreate.TransactionId != productUsageRecordReturned.TransactionId) ||
		(productUsageRecordCreate.Region != productUsageRecordReturned.Region) ||
		(productUsageRecordCreate.Quantity != productUsageRecordReturned.Quantity) {
		return true
	}
	return false
}

func verifyNoProductUsageRecordsReturned(ctx context.Context, t *testing.T, productUsageRecordsWithNoResultsFilter *pb.ProductUsageRecordsFilter) {
	productUsageRecordsReturned := getProductUsageRecordsBySearch(t, ctx, productUsageRecordsWithNoResultsFilter)
	if len(productUsageRecordsReturned) != 0 {
		t.Fatalf("should not have got product usage records")
	}
}

func TestInvalidProductUsageRecord(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestInvalidProductUsageRecord")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageRecordData := NewUsageRecordData(usageDb)

	err := usageRecordData.DeleteAllUsageRecords(ctx)
	if err != nil {
		t.Fatalf("could not delete all usage records")
	}

	cloudAccountId := MustNewCloudAcctId()
	productName := uuid.NewString()
	region := "us-west-1"
	transactionId := uuid.NewString()
	recordId := uuid.NewString()

	invalidProductUsageRecordCreate := GetInvalidProductUsageRecordCreate(&recordId, cloudAccountId, &productName, region,
		transactionId, defaultInvalidProductUsageRecordQty,
		map[string]string{
			"availabilityZone": region,
			"serviceType":      "ComputeAsAService",
		}, pb.ProductUsageRecordInvalidityReason_DEFAULT_PRODUCT_USAGE_RECORD_INVALIDITY_REASON)

	err = usageRecordData.StoreInvalidProductUsageRecord(ctx, invalidProductUsageRecordCreate)

	if err != nil {
		t.Fatalf("unexpected error for storing invalid product usage record %v", err)
	}

	invalidProductUsageRecordsFilter := &pb.InvalidProductUsageRecordsFilter{
		CloudAccountId: &cloudAccountId,
	}

	invalidProductUsageRecordsReturned := getInvalidProductUsageRecordsBySearch(t, ctx, invalidProductUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search invalid invalid product usage records for cloud account filter %v", err)
	}

	if invalidProductUsageRecordsReturned == nil || len(invalidProductUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of invalid invalid product usage records returned for cloud account filter")
	}

	if checkIfValuesForInvalidProductUsageRecordDoNotMatch(invalidProductUsageRecordCreate, invalidProductUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for cloud account filter")
	}

	invalidProductUsageRecordsFilter = &pb.InvalidProductUsageRecordsFilter{
		Region: &region,
	}

	invalidProductUsageRecordsReturned = getInvalidProductUsageRecordsBySearch(t, ctx, invalidProductUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search invalid invalid product usage records for region filter %v", err)
	}

	if invalidProductUsageRecordsReturned == nil || len(invalidProductUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of invalid invalid product usage records returned for region filter")
	}

	if checkIfValuesForInvalidProductUsageRecordDoNotMatch(invalidProductUsageRecordCreate, invalidProductUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for region filter")
	}

	invalidProductUsageRecordsFilter = &pb.InvalidProductUsageRecordsFilter{
		TransactionId: &transactionId,
	}

	invalidProductUsageRecordsReturned = getInvalidProductUsageRecordsBySearch(t, ctx, invalidProductUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search invalid invalid product usage records for transaction id filter %v", err)
	}

	if invalidProductUsageRecordsReturned == nil || len(invalidProductUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of invalid invalid product usage records returned for transaction id filter")
	}

	if checkIfValuesForInvalidProductUsageRecordDoNotMatch(invalidProductUsageRecordCreate, invalidProductUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for transaction id filter")
	}

	startTime := time.Now().AddDate(0, -1, 0)
	invalidProductUsageRecordsFilter = &pb.InvalidProductUsageRecordsFilter{
		StartTime: timestamppb.New(startTime),
	}

	invalidProductUsageRecordsReturned = getInvalidProductUsageRecordsBySearch(t, ctx, invalidProductUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search invalid invalid product usage records for start time filter %v", err)
	}

	if invalidProductUsageRecordsReturned == nil || len(invalidProductUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of invalid invalid product usage records returned for start time filter")
	}

	if checkIfValuesForInvalidProductUsageRecordDoNotMatch(invalidProductUsageRecordCreate, invalidProductUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for start time filter")
	}

	endTime := time.Now().AddDate(0, +1, 0)
	invalidProductUsageRecordsFilter = &pb.InvalidProductUsageRecordsFilter{
		EndTime: timestamppb.New(endTime),
	}

	invalidProductUsageRecordsReturned = getInvalidProductUsageRecordsBySearch(t, ctx, invalidProductUsageRecordsFilter)
	if err != nil {
		t.Fatalf("failed to do search invalid invalid product usage records for end time filter %v", err)
	}

	if invalidProductUsageRecordsReturned == nil || len(invalidProductUsageRecordsReturned) != 1 {
		t.Fatalf("invalid length of invalid invalid product usage records returned for end time filter")
	}

	if checkIfValuesForInvalidProductUsageRecordDoNotMatch(invalidProductUsageRecordCreate, invalidProductUsageRecordsReturned[0]) {
		t.Fatalf("values do not match after create and get for start time filter")
	}

	invalidCloudAcctId := MustNewCloudAcctId()
	verifyNoInvalidProductUsageRecordsReturned(ctx, t, &pb.InvalidProductUsageRecordsFilter{
		CloudAccountId: &invalidCloudAcctId,
	})

	verifyNoInvalidProductUsageRecordsReturned(ctx, t, &pb.InvalidProductUsageRecordsFilter{
		StartTime: timestamppb.New(endTime),
	})

	verifyNoInvalidProductUsageRecordsReturned(ctx, t, &pb.InvalidProductUsageRecordsFilter{
		EndTime: timestamppb.New(startTime),
	})
}

func getInvalidProductUsageRecordsBySearch(t *testing.T, ctx context.Context,
	invalidProductUsageRecordsFilter *pb.InvalidProductUsageRecordsFilter) []*pb.InvalidProductUsageRecord {

	invalidProductUsageRecords := []*pb.InvalidProductUsageRecord{}

	usageRecordServiceClient := pb.NewUsageRecordServiceClient(clientConn)
	stream, _ := usageRecordServiceClient.SearchInvalidProductUsageRecords(ctx, invalidProductUsageRecordsFilter)

	for {

		invalidProductUsageRecord, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("unexpected error for searching invalid invalid product usage records %v", err)
		}

		invalidProductUsageRecords = append(invalidProductUsageRecords, invalidProductUsageRecord)
	}

	return invalidProductUsageRecords
}

func checkIfValuesForInvalidProductUsageRecordDoNotMatch(invalidProductUsageRecordCreate *pb.InvalidProductUsageRecordCreate,
	invalidProductUsageRecordReturned *pb.InvalidProductUsageRecord) bool {
	if (invalidProductUsageRecordCreate.CloudAccountId != invalidProductUsageRecordReturned.CloudAccountId) ||
		(*invalidProductUsageRecordCreate.ProductName != *invalidProductUsageRecordReturned.ProductName) ||
		(invalidProductUsageRecordCreate.TransactionId != invalidProductUsageRecordReturned.TransactionId) ||
		(invalidProductUsageRecordCreate.Region != invalidProductUsageRecordReturned.Region) ||
		(invalidProductUsageRecordCreate.Quantity != invalidProductUsageRecordReturned.Quantity) {
		return true
	}
	return false
}

func verifyNoInvalidProductUsageRecordsReturned(ctx context.Context, t *testing.T, invalidProductUsageRecordsWithNoResultsFilter *pb.InvalidProductUsageRecordsFilter) {
	invalidProductUsageRecordsReturned := getInvalidProductUsageRecordsBySearch(t, ctx, invalidProductUsageRecordsWithNoResultsFilter)
	if len(invalidProductUsageRecordsReturned) != 0 {
		t.Fatalf("should not have got invalid invalid product usage records")
	}
}
