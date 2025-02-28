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

func TestBulkUpload(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestBulkUpload")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageServiceClient := pb.NewUsageServiceClient(clientConn)
	usageData := NewUsageData(usageDb)

	err := usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	resourceId := uuid.NewString()
	cloudAccountId := MustNewCloudAcctId()
	productId := uuid.NewString()
	region := "us-west-1"
	usageUnitType := "mins"
	resourceUsages := []*pb.ResourceUsageCreate{{
		CloudAccountId: cloudAccountId,
		ResourceId:     resourceId,
		ResourceName:   uuid.NewString(),
		ProductId:      productId,
		ProductName:    uuid.NewString(),
		Region:         region,
		Quantity:       10,
		Rate:           10,
		UsageUnitType:  usageUnitType,
	}}

	bulkUploadResourceUsages := &pb.BulkUploadResourceUsages{ResourceUsages: resourceUsages}

	bulkUploadResourceUsageFailedRecords, err := usageServiceClient.PostBulkUploadResourceUsages(ctx, bulkUploadResourceUsages)
	if err != nil {
		t.Fatalf("failed to do bulk upload of resource usages %v", err)
	}

	if len(bulkUploadResourceUsageFailedRecords.ResourceUsages) != 0 {
		t.Fatalf("should not have received any failed usages for upload")
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestCreate")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageServiceClient := pb.NewUsageServiceClient(clientConn)
	usageData := NewUsageData(usageDb)

	err := usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	resourceId := uuid.NewString()
	cloudAccountId := MustNewCloudAcctId()
	productId := uuid.NewString()
	region := "us-west-1"
	usageUnitType := "mins"
	transactionId := uuid.NewString()

	resourceUsageCreate := GetResourceUsageCreate(cloudAccountId, resourceId, productId, region,
		transactionId, usageUnitType)

	_, err = usageServiceClient.CreateResourceUsage(ctx, resourceUsageCreate)
	if err != nil {
		t.Fatalf("failed to do creation of resource usage %v", err)
	}

}

func TestMarkAsReported(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestMarkAsReported")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageServiceClient := pb.NewUsageServiceClient(clientConn)
	usageData := NewUsageData(usageDb)

	err := usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	resourceId := uuid.NewString()
	cloudAccountId := MustNewCloudAcctId()
	productId := uuid.NewString()
	region := "us-west-1"
	usageUnitType := "mins"

	transactionId := uuid.NewString()

	resourceUsageCreate := GetResourceUsageCreate(cloudAccountId, resourceId, productId, region,
		transactionId, usageUnitType)

	resourceUsageCreated, err := usageServiceClient.CreateResourceUsage(ctx, resourceUsageCreate)
	if err != nil {
		t.Fatalf("failed to do creation of resource usage %v", err)
	}

	resourceUsageReturned, err := usageServiceClient.GetResourceUsageById(ctx, &pb.ResourceUsageId{Id: resourceUsageCreated.Id})
	if err != nil {
		t.Fatalf("failed to do get resource usage by id after marking as reported %v", err)
	}

	if resourceUsageReturned.Reported != false {
		t.Fatalf("should not have been set to reported after create")
	}

	_, err = usageServiceClient.MarkResourceUsageAsReported(ctx, &pb.ResourceUsageId{Id: resourceUsageCreated.Id})
	if err != nil {
		t.Fatalf("failed to mark resource usage as reported %v", err)
	}

	resourceUsageReturnedAfterSetReported, err := usageServiceClient.GetResourceUsageById(ctx, &pb.ResourceUsageId{Id: resourceUsageCreated.Id})
	if err != nil {
		t.Fatalf("failed to do get resource usage by id after marking as reported %v", err)
	}

	if checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate, resourceUsageReturnedAfterSetReported) {
		t.Fatalf("values apart from reported updated")
	}

	if resourceUsageReturnedAfterSetReported.Reported != true {
		t.Fatalf("should have been set to reported")
	}
}

func TestGetResourceUsageById(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestGetResourceUsageById")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageServiceClient := pb.NewUsageServiceClient(clientConn)
	usageData := NewUsageData(usageDb)

	err := usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	resourceId := uuid.NewString()
	cloudAccountId := MustNewCloudAcctId()
	productId := uuid.NewString()
	region := "us-west-1"
	usageUnitType := "mins"

	transactionId := uuid.NewString()

	resourceUsageCreate := GetResourceUsageCreate(cloudAccountId, resourceId, productId, region,
		transactionId, usageUnitType)

	resourceUsageCreated, err := usageServiceClient.CreateResourceUsage(ctx, resourceUsageCreate)
	if err != nil {
		t.Fatalf("failed to do creation of resource usage %v", err)
	}

	resourceUsageReturned, err := usageServiceClient.GetResourceUsageById(ctx, &pb.ResourceUsageId{Id: resourceUsageCreated.Id})
	if err != nil {
		t.Fatalf("failed to do get resource usage by id %v", err)
	}

	if checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate, resourceUsageReturned) {
		t.Fatalf("values do not match after create and get")
	}
}

func TestGetProductUsage(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestGetProductUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageServiceClient := pb.NewUsageServiceClient(clientConn)

	usageData := NewUsageData(usageDb)

	err := usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	productId := uuid.NewString()
	region := "us-west-2"
	usageUnitType := "mins"
	cloudAcctId := MustNewCloudAcctId()

	productUsageRecordCreate := &ProductUsageCreate{
		Id:             uuid.NewString(),
		CloudAccountId: cloudAcctId,
		ProductId:      productId,
		ProductName:    "someName",
		Region:         region,
		Quantity:       10,
		Rate:           10,
		UsageUnitType:  usageUnitType,
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	err = usageData.StoreProductUsage(ctx, productUsageRecordCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing product usage %v", err)
	}

	productUsagesFilter := &pb.ProductUsagesFilter{
		CloudAccountId: &cloudAcctId,
	}
	productUsagesReturned, err := usageServiceClient.SearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search product usages %v", err)
	}

	if len(productUsagesReturned.ProductUsages) != 1 {
		t.Fatalf("invalid length of product usages returned")
	}

	productUsageReturned, err := usageServiceClient.GetProductUsageById(ctx, &pb.ProductUsageId{Id: productUsagesReturned.ProductUsages[0].Id})
	if err != nil {
		t.Fatalf("failed to do get resource usage by id %v", err)
	}

	if checkIfValuesForProductUsageDoNotMatch(productUsageRecordCreate, productUsageReturned) {
		t.Fatalf("values not same after create and get")
	}
}

func TestSearchResourceUsages(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestSearchResourceUsages")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageServiceClient := pb.NewUsageServiceClient(clientConn)

	usageData := NewUsageData(usageDb)

	err := usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	resourceId := uuid.NewString()
	cloudAccountId := MustNewCloudAcctId()
	productId := uuid.NewString()
	region := "us-west-1"
	usageUnitType := "mins"

	transactionId := uuid.NewString()

	resourceUsageCreate := GetResourceUsageCreate(cloudAccountId, resourceId, productId, region,
		transactionId, usageUnitType)

	_, err = usageServiceClient.CreateResourceUsage(ctx, resourceUsageCreate)
	if err != nil {
		t.Fatalf("failed to do creation of resource usage %v", err)
	}

	notMatchingResourceUsageRecordCreate := GetResourceUsageCreate(MustNewCloudAcctId(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString())

	notMatchingResourceUsageRecord, err := usageServiceClient.CreateResourceUsage(ctx, notMatchingResourceUsageRecordCreate)
	if err != nil {
		t.Fatalf("failed to do creation of resource usage %v", err)
	}

	_, err = usageServiceClient.MarkResourceUsageAsReported(ctx, &pb.ResourceUsageId{Id: notMatchingResourceUsageRecord.Id})
	if err != nil {
		t.Fatalf("failed to mark resource usage as reported %v", err)
	}

	resourceUsagesFilter := &pb.ResourceUsagesFilter{
		CloudAccountId: &cloudAccountId,
	}
	resourceUsagesReturned, err := usageServiceClient.SearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for cloud account filter %v", err)
	}

	if len(resourceUsagesReturned.ResourceUsages) != 1 {
		t.Fatalf("invalid length of resource usages returned for cloud account filter ")
	}

	if checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate, resourceUsagesReturned.ResourceUsages[0]) {
		t.Fatalf("values do not match after create and get for cloud account filter ")
	}

	resourceUsagesFilter = &pb.ResourceUsagesFilter{
		ResourceId: &resourceId,
	}

	resourceUsagesReturned, err = usageServiceClient.SearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for resource filter %v", err)
	}

	if len(resourceUsagesReturned.ResourceUsages) != 1 {
		t.Fatalf("invalid length of resource usages returned for resource filter")
	}

	if checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate, resourceUsagesReturned.ResourceUsages[0]) {
		t.Fatalf("values do not match after create and get for resource filter")
	}

	resourceUsagesFilter = &pb.ResourceUsagesFilter{
		Region: &region,
	}

	resourceUsagesReturned, err = usageServiceClient.SearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for region filter %v", err)
	}

	if len(resourceUsagesReturned.ResourceUsages) != 1 {
		t.Fatalf("invalid length of resource usages returned for region filter")
	}

	if checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate, resourceUsagesReturned.ResourceUsages[0]) {
		t.Fatalf("values do not match after create and get for region filter")
	}

	startTime := time.Now().AddDate(0, -1, 0)
	resourceUsagesFilter = &pb.ResourceUsagesFilter{
		StartTime: timestamppb.New(startTime),
	}

	resourceUsagesReturned, err = usageServiceClient.SearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for start time filter %v", err)
	}

	if len(resourceUsagesReturned.ResourceUsages) != 2 {
		t.Fatalf("invalid length of resource usages returned for start time filter")
	}

	endTime := time.Now().AddDate(0, +1, 0)
	resourceUsagesFilter = &pb.ResourceUsagesFilter{
		EndTime: timestamppb.New(endTime),
	}

	resourceUsagesReturned, err = usageServiceClient.SearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for end time filter %v", err)
	}

	if len(resourceUsagesReturned.ResourceUsages) != 2 {
		t.Fatalf("invalid length of resource usages returned for end time filter")
	}

	reported := false
	resourceUsagesFilter = &pb.ResourceUsagesFilter{
		Reported: &reported,
	}

	resourceUsagesReturned, err = usageServiceClient.SearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for reported filter %v", err)
	}

	if len(resourceUsagesReturned.ResourceUsages) != 1 {
		t.Fatalf("invalid length of resource usages reported for region filter")
	}

	if checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate, resourceUsagesReturned.ResourceUsages[0]) {
		t.Fatalf("values do not match after create and get for reported filter")
	}

	invalidCloudAcctId := MustNewCloudAcctId()
	verifyNoResourceUsagesReturned(ctx, t, &pb.ResourceUsagesFilter{
		CloudAccountId: &invalidCloudAcctId,
	})

	invalidResourceId := uuid.NewString()
	verifyNoResourceUsagesReturned(ctx, t, &pb.ResourceUsagesFilter{
		ResourceId: &invalidResourceId,
	})

	verifyNoResourceUsagesReturned(ctx, t, &pb.ResourceUsagesFilter{
		StartTime: timestamppb.New(endTime),
	})

	verifyNoResourceUsagesReturned(ctx, t, &pb.ResourceUsagesFilter{
		EndTime: timestamppb.New(startTime),
	})
}

func TestSearchProductUsages(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestSearchProductUsages")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageServiceClient := pb.NewUsageServiceClient(clientConn)
	usageData := NewUsageData(usageDb)

	err := usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	cloudAccountId := MustNewCloudAcctId()
	productId := uuid.NewString()
	region := "us-west-1"
	usageUnitType := "mins"

	productUsageRecordCreate := &ProductUsageCreate{
		Id:             uuid.NewString(),
		CloudAccountId: cloudAccountId,
		ProductId:      productId,
		ProductName:    "someName",
		Region:         region,
		Quantity:       10,
		Rate:           10,
		UsageUnitType:  usageUnitType,
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	err = usageData.StoreProductUsage(ctx, &ProductUsageCreate{
		Id:             uuid.NewString(),
		CloudAccountId: MustNewCloudAcctId(),
		ProductId:      uuid.NewString(),
		ProductName:    uuid.NewString(),
		Region:         uuid.NewString(),
		Quantity:       10,
		Rate:           10,
		UsageUnitType:  uuid.NewString(),
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing product usage %v", err)
	}

	err = usageData.StoreProductUsage(ctx, productUsageRecordCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing product usage %v", err)
	}

	productUsagesFilter := &pb.ProductUsagesFilter{
		ProductId: &productId,
	}

	productUsagesReturned, err := usageServiceClient.SearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search product usages for product id filter %v", err)
	}

	if len(productUsagesReturned.ProductUsages) != 1 {
		t.Fatalf("invalid length of product usages returned for product id filter ")
	}

	if checkIfValuesForProductUsageDoNotMatch(productUsageRecordCreate, productUsagesReturned.ProductUsages[0]) {
		t.Fatalf("values do not match after create and get for product id filter ")
	}

	productUsagesFilter = &pb.ProductUsagesFilter{
		Region: &region,
	}

	productUsagesReturned, err = usageServiceClient.SearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search product usages for region filter %v", err)
	}

	if len(productUsagesReturned.ProductUsages) != 1 {
		t.Fatalf("invalid length of product usages returned for region filter ")
	}

	if checkIfValuesForProductUsageDoNotMatch(productUsageRecordCreate, productUsagesReturned.ProductUsages[0]) {
		t.Fatalf("values do not match after create and get for region filter ")
	}

	startTime := time.Now().AddDate(0, -1, 0)
	productUsagesFilter = &pb.ProductUsagesFilter{
		StartTime: timestamppb.New(startTime),
	}

	productUsagesReturned, err = usageServiceClient.SearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search product usages for start time filter %v", err)
	}

	if len(productUsagesReturned.ProductUsages) != 2 {
		t.Fatalf("invalid length of product usages returned for start time filter ")
	}

	endTime := time.Now().AddDate(0, +1, 0)
	productUsagesFilter = &pb.ProductUsagesFilter{
		EndTime: timestamppb.New(endTime),
	}

	productUsagesReturned, err = usageServiceClient.SearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search product usages for end time filter %v", err)
	}

	if len(productUsagesReturned.ProductUsages) != 2 {
		t.Fatalf("invalid length of product usages returned for end time filter ")
	}

	invalidCloudAcctId := MustNewCloudAcctId()
	verifyNoProductUsagesReturned(ctx, t, &pb.ProductUsagesFilter{
		CloudAccountId: &invalidCloudAcctId,
	})

	invalidProductId := uuid.NewString()
	verifyNoProductUsagesReturned(ctx, t, &pb.ProductUsagesFilter{
		ProductId: &invalidProductId,
	})

	verifyNoProductUsagesReturned(ctx, t, &pb.ProductUsagesFilter{
		StartTime: timestamppb.New(endTime),
	})

	verifyNoProductUsagesReturned(ctx, t, &pb.ProductUsagesFilter{
		EndTime: timestamppb.New(startTime),
	})
}

func TestSearchUsages(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestSearchUsages")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageServiceClient := pb.NewUsageServiceClient(clientConn)
	usageData := NewUsageData(usageDb)

	err := usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	cloudAccountId := MustNewCloudAcctId()
	productId := uuid.NewString()
	region := "us-west-1"
	usageUnitType := "mins"

	productUsageRecordCreate := &ProductUsageCreate{
		Id:             uuid.NewString(),
		CloudAccountId: cloudAccountId,
		ProductId:      productId,
		ProductName:    "someName",
		Region:         region,
		Quantity:       10,
		Rate:           10,
		UsageUnitType:  usageUnitType,
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	err = usageData.StoreProductUsage(ctx, productUsageRecordCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing product usage %v", err)
	}

	usagesFilter := &pb.UsagesFilter{
		CloudAccountId: &cloudAccountId,
	}

	usagesReturned, err := usageServiceClient.SearchUsages(ctx, usagesFilter)
	if err != nil {
		t.Fatalf("failed to do search product usages for product id filter %v", err)
	}

	if len(usagesReturned.Usages) != 1 {
		t.Fatalf("invalid length of usages returned")
	}

}

func TestStreamSearchResourceUsages(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestStreamSearchResourceUsages")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageServiceClient := pb.NewUsageServiceClient(clientConn)

	usageData := NewUsageData(usageDb)

	err := usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	resourceId := uuid.NewString()
	cloudAccountId := MustNewCloudAcctId()
	productId := uuid.NewString()
	region := "us-west-1"
	usageUnitType := "mins"

	transactionId := uuid.NewString()

	resourceUsageCreate := GetResourceUsageCreate(cloudAccountId, resourceId, productId, region,
		transactionId, usageUnitType)

	_, err = usageServiceClient.CreateResourceUsage(ctx, resourceUsageCreate)
	if err != nil {
		t.Fatalf("failed to do creation of resource usage %v", err)
	}

	notMatchingResourceUsageRecordCreate := GetResourceUsageCreate(MustNewCloudAcctId(), uuid.NewString(), uuid.NewString(), uuid.NewString(),
		uuid.NewString(), uuid.NewString())

	notMatchingResourceUsageRecord, err := usageServiceClient.CreateResourceUsage(ctx, notMatchingResourceUsageRecordCreate)
	if err != nil {
		t.Fatalf("failed to do creation of resource usage %v", err)
	}

	_, err = usageServiceClient.MarkResourceUsageAsReported(ctx, &pb.ResourceUsageId{Id: notMatchingResourceUsageRecord.Id})
	if err != nil {
		t.Fatalf("failed to mark resource usage as reported %v", err)
	}

	resourceUsagesFilter := &pb.ResourceUsagesFilter{
		CloudAccountId: &cloudAccountId,
	}
	resourceUsagesStream, err := usageServiceClient.StreamSearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for cloud account id filter %v", err)
	}
	resourceUsagesReturned := getResourceUsageOverStream(t, resourceUsagesStream)

	if len(resourceUsagesReturned.ResourceUsages) != 1 {
		t.Fatalf("invalid length of resource usages returned for cloud account filter ")
	}

	if checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate, resourceUsagesReturned.ResourceUsages[0]) {
		t.Fatalf("values do not match after create and get for cloud account filter ")
	}

	resourceUsagesFilter = &pb.ResourceUsagesFilter{
		ResourceId: &resourceId,
	}

	resourceUsagesStream, err = usageServiceClient.StreamSearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for resource id filter %v", err)
	}
	resourceUsagesReturned = getResourceUsageOverStream(t, resourceUsagesStream)

	if len(resourceUsagesReturned.ResourceUsages) != 1 {
		t.Fatalf("invalid length of resource usages returned for resource filter")
	}

	if checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate, resourceUsagesReturned.ResourceUsages[0]) {
		t.Fatalf("values do not match after create and get for resource filter")
	}

	resourceUsagesFilter = &pb.ResourceUsagesFilter{
		Region: &region,
	}
	resourceUsagesStream, err = usageServiceClient.StreamSearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for region filter %v", err)
	}
	resourceUsagesReturned = getResourceUsageOverStream(t, resourceUsagesStream)

	if len(resourceUsagesReturned.ResourceUsages) != 1 {
		t.Fatalf("invalid length of resource usages returned for region filter")
	}

	if checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate, resourceUsagesReturned.ResourceUsages[0]) {
		t.Fatalf("values do not match after create and get for region filter")
	}

	startTime := time.Now().AddDate(0, -1, 0)
	resourceUsagesFilter = &pb.ResourceUsagesFilter{
		StartTime: timestamppb.New(startTime),
	}
	resourceUsagesStream, err = usageServiceClient.StreamSearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for startTime filter %v", err)
	}
	resourceUsagesReturned = getResourceUsageOverStream(t, resourceUsagesStream)

	if len(resourceUsagesReturned.ResourceUsages) != 2 {
		t.Fatalf("invalid length of resource usages returned for start time filter")
	}

	endTime := time.Now().AddDate(0, +1, 0)
	resourceUsagesFilter = &pb.ResourceUsagesFilter{
		EndTime: timestamppb.New(endTime),
	}

	resourceUsagesStream, err = usageServiceClient.StreamSearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for endTime filter %v", err)
	}
	resourceUsagesReturned = getResourceUsageOverStream(t, resourceUsagesStream)

	if len(resourceUsagesReturned.ResourceUsages) != 2 {
		t.Fatalf("invalid length of resource usages returned for end time filter")
	}

	reported := false
	resourceUsagesFilter = &pb.ResourceUsagesFilter{
		Reported: &reported,
	}

	resourceUsagesStream, err = usageServiceClient.StreamSearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search resource usages for reported filter %v", err)
	}
	resourceUsagesReturned = getResourceUsageOverStream(t, resourceUsagesStream)

	if len(resourceUsagesReturned.ResourceUsages) != 1 {
		t.Fatalf("invalid length of resource usages reported for region filter")
	}

	if checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate, resourceUsagesReturned.ResourceUsages[0]) {
		t.Fatalf("values do not match after create and get for reported filter")
	}

	invalidCloudAcctId := MustNewCloudAcctId()
	verifyNoResourceUsagesReturned(ctx, t, &pb.ResourceUsagesFilter{
		CloudAccountId: &invalidCloudAcctId,
	})

	invalidResourceId := uuid.NewString()
	verifyNoResourceUsagesReturned(ctx, t, &pb.ResourceUsagesFilter{
		ResourceId: &invalidResourceId,
	})

	verifyNoResourceUsagesReturned(ctx, t, &pb.ResourceUsagesFilter{
		StartTime: timestamppb.New(endTime),
	})

	verifyNoResourceUsagesReturned(ctx, t, &pb.ResourceUsagesFilter{
		EndTime: timestamppb.New(startTime),
	})
}

func TestStreamSearchProductUsages(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestStreamSearchProductUsages")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageServiceClient := pb.NewUsageServiceClient(clientConn)
	usageData := NewUsageData(usageDb)

	err := usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	cloudAccountId := MustNewCloudAcctId()
	productId := uuid.NewString()
	region := "us-west-1"
	usageUnitType := "mins"

	productUsageRecordCreate := &ProductUsageCreate{
		Id:             uuid.NewString(),
		CloudAccountId: cloudAccountId,
		ProductId:      productId,
		ProductName:    "someName",
		Region:         region,
		Quantity:       10,
		Rate:           10,
		UsageUnitType:  usageUnitType,
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	err = usageData.StoreProductUsage(ctx, &ProductUsageCreate{
		Id:             uuid.NewString(),
		CloudAccountId: MustNewCloudAcctId(),
		ProductId:      uuid.NewString(),
		ProductName:    uuid.NewString(),
		Region:         uuid.NewString(),
		Quantity:       10,
		Rate:           10,
		UsageUnitType:  uuid.NewString(),
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing product usage %v", err)
	}

	err = usageData.StoreProductUsage(ctx, productUsageRecordCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing product usage %v", err)
	}

	productUsagesFilter := &pb.ProductUsagesFilter{
		ProductId: &productId,
	}

	productUsagesStream, err := usageServiceClient.StreamSearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search product usages for product id filter %v", err)
	}
	productUsagesReturned := getProductUsageOverStream(t, productUsagesStream)

	if len(productUsagesReturned.ProductUsages) != 1 {
		t.Fatalf("invalid length of product usages returned for product id filter ")
	}

	if checkIfValuesForProductUsageDoNotMatch(productUsageRecordCreate, productUsagesReturned.ProductUsages[0]) {
		t.Fatalf("values do not match after create and get for product id filter ")
	}

	productUsagesFilter = &pb.ProductUsagesFilter{
		Region: &region,
	}
	productUsagesStream, err = usageServiceClient.StreamSearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search product usages for region filter %v", err)
	}
	productUsagesReturned = getProductUsageOverStream(t, productUsagesStream)

	if len(productUsagesReturned.ProductUsages) != 1 {
		t.Fatalf("invalid length of product usages returned for region filter ")
	}

	if checkIfValuesForProductUsageDoNotMatch(productUsageRecordCreate, productUsagesReturned.ProductUsages[0]) {
		t.Fatalf("values do not match after create and get for region filter ")
	}

	startTime := time.Now().AddDate(0, -1, 0)
	productUsagesFilter = &pb.ProductUsagesFilter{
		StartTime: timestamppb.New(startTime),
	}
	productUsagesStream, err = usageServiceClient.StreamSearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search product usages for start time filter %v", err)
	}
	productUsagesReturned = getProductUsageOverStream(t, productUsagesStream)

	if len(productUsagesReturned.ProductUsages) != 2 {
		t.Fatalf("invalid length of product usages returned for start time filter ")
	}

	endTime := time.Now().AddDate(0, +1, 0)
	productUsagesFilter = &pb.ProductUsagesFilter{
		EndTime: timestamppb.New(endTime),
	}

	productUsagesStream, err = usageServiceClient.StreamSearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to do search product usages for end time filter %v", err)
	}
	productUsagesReturned = getProductUsageOverStream(t, productUsagesStream)

	if len(productUsagesReturned.ProductUsages) != 2 {
		t.Fatalf("invalid length of product usages returned for end time filter ")
	}

	invalidCloudAcctId := MustNewCloudAcctId()
	verifyNoProductUsagesReturned(ctx, t, &pb.ProductUsagesFilter{
		CloudAccountId: &invalidCloudAcctId,
	})

	invalidProductId := uuid.NewString()
	verifyNoProductUsagesReturned(ctx, t, &pb.ProductUsagesFilter{
		ProductId: &invalidProductId,
	})

	verifyNoProductUsagesReturned(ctx, t, &pb.ProductUsagesFilter{
		StartTime: timestamppb.New(endTime),
	})

	verifyNoProductUsagesReturned(ctx, t, &pb.ProductUsagesFilter{
		EndTime: timestamppb.New(startTime),
	})
}

func getResourceUsageOverStream(t *testing.T,
	resourceUsagesStream pb.UsageService_StreamSearchResourceUsagesClient) *pb.ResourceUsages {

	resourceUsages := &pb.ResourceUsages{}
	for {

		resourceUsage, err := resourceUsagesStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("unexpected error for searching resource usage %v", err)
		}

		resourceUsages.ResourceUsages = append(resourceUsages.ResourceUsages, resourceUsage)
	}

	return resourceUsages
}

func getProductUsageOverStream(t *testing.T,
	productUsagesStream pb.UsageService_StreamSearchProductUsagesClient) *pb.ProductUsages {

	productUsages := &pb.ProductUsages{}
	for {

		productUsage, err := productUsagesStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("unexpected error for searching product usage %v", err)
		}

		productUsages.ProductUsages = append(productUsages.ProductUsages, productUsage)
	}

	return productUsages
}

func verifyNoResourceUsagesReturned(ctx context.Context, t *testing.T, filterWithNoResults *pb.ResourceUsagesFilter) {
	usageServiceClient := pb.NewUsageServiceClient(clientConn)

	resourceUsagesReturned, err := usageServiceClient.SearchResourceUsages(ctx, filterWithNoResults)
	if err != nil {
		t.Fatalf("failed to do search resource usages for filter %v", err)
	}

	if resourceUsagesReturned.ResourceUsages != nil {
		t.Fatalf("should not have got resource usages")
	}
}

func verifyNoProductUsagesReturned(ctx context.Context, t *testing.T, filterWithNoResults *pb.ProductUsagesFilter) {
	usageServiceClient := pb.NewUsageServiceClient(clientConn)

	productUsagesReturned, err := usageServiceClient.SearchProductUsages(ctx, filterWithNoResults)
	if err != nil {
		t.Fatalf("failed to do search product usages for filter %v", err)
	}

	if productUsagesReturned.ProductUsages != nil {
		t.Fatalf("should not have got product usages")
	}
}

func checkIfValuesForResourceUsageDoNotMatch(resourceUsageCreate *pb.ResourceUsageCreate,
	resourceUsageReturned *pb.ResourceUsage) bool {
	if (resourceUsageCreate.CloudAccountId != resourceUsageReturned.CloudAccountId) ||
		(resourceUsageCreate.ResourceId != resourceUsageReturned.ResourceId) ||
		(resourceUsageCreate.ResourceName != resourceUsageReturned.ResourceName) ||
		(resourceUsageCreate.ProductId != resourceUsageReturned.ProductId) ||
		(resourceUsageCreate.ProductName != resourceUsageReturned.ProductName) ||
		(resourceUsageCreate.TransactionId != resourceUsageReturned.TransactionId) ||
		(resourceUsageCreate.Region != resourceUsageReturned.Region) ||
		(resourceUsageCreate.Quantity != resourceUsageReturned.Quantity) ||
		(resourceUsageCreate.Rate != resourceUsageReturned.Rate) ||
		(resourceUsageCreate.UsageUnitType != resourceUsageReturned.UsageUnitType) {
		return true
	}
	return false
}

func checkIfValuesForProductUsageDoNotMatch(productUsageCreate *ProductUsageCreate,
	productUsageReturned *pb.ProductUsage) bool {
	if (productUsageCreate.CloudAccountId != productUsageReturned.CloudAccountId) ||
		(productUsageCreate.ProductId != productUsageReturned.ProductId) ||
		(productUsageCreate.ProductName != productUsageReturned.ProductName) ||
		(productUsageCreate.Region != productUsageReturned.Region) ||
		(productUsageCreate.Quantity != productUsageReturned.Quantity) ||
		(productUsageCreate.Rate != productUsageReturned.Rate) ||
		(productUsageCreate.UsageUnitType != productUsageReturned.UsageUnitType) {
		return true
	}
	return false
}

func TestProductUsageReportApis(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestProductUsageReportApis")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)
	productName := GetXeon3SmallInstanceType()
	region := DefaultServiceRegion

	transactionId := "transaction-id"

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	productUsageRecordCreate := GetProductUsageRecordCreate(premiumCloudAcct.Id, &productName, region,
		transactionId, defaultProductUsageRecordQty,
		map[string]string{
			"availabilityZone": region,
			"instanceType":     GetXeon3SmallInstanceType(),
			"service":          GetIdcComputeServiceName(),
		})

	err := usageRecordData.StoreProductUsageRecord(ctx, productUsageRecordCreate)

	if err != nil {
		t.Fatalf("unexpected error for storing product usage record %v", err)
	}

	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateProductUsages(ctx)

	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	if len(productUsages.ProductUsages) != 1 {
		t.Fatalf("invalid length of product usages")
	}

	if productUsages.ProductUsages[0].Quantity != defaultProductUsageRecordQty ||
		productUsages.ProductUsages[0].CloudAccountId != premiumCloudAcct.Id {
		t.Fatalf("expected values do not match")
	}

	reportProductUsages := searchProductUsagesUnreported(t, ctx)

	if len(reportProductUsages) != 1 {
		t.Fatalf("invalid length of report product usages")
	}

	if (reportProductUsages[0].CloudAccountId != productUsages.ProductUsages[0].CloudAccountId) ||
		(reportProductUsages[0].ProductId != productUsages.ProductUsages[0].ProductId) ||
		(reportProductUsages[0].ProductUsageId != productUsages.ProductUsages[0].Id) ||
		(reportProductUsages[0].Quantity != productUsages.ProductUsages[0].Quantity) ||
		(reportProductUsages[0].UnReportedQuantity != productUsages.ProductUsages[0].Quantity) {
		t.Fatalf("expected values do not match for product usage report")

	}

	unReportedQty := reportProductUsages[0].UnReportedQuantity - 1

	usageServiceClient := pb.NewUsageServiceClient(clientConn)

	_, err = usageServiceClient.UpdateProductUsageReport(ctx, &pb.ReportProductUsageUpdate{
		ProductUsageReportId: reportProductUsages[0].Id,
		UnReportedQuantity:   unReportedQty,
	})

	if err != nil {
		t.Fatalf("failed to update product usage report: %v", err)
	}

	reportProductUsagesAfterUpdate := searchProductUsagesUnreported(t, ctx)

	if len(reportProductUsagesAfterUpdate) != 1 {
		t.Fatalf("invalid length of report product usages")
	}

	if (reportProductUsagesAfterUpdate[0].CloudAccountId != productUsages.ProductUsages[0].CloudAccountId) ||
		(reportProductUsagesAfterUpdate[0].ProductId != productUsages.ProductUsages[0].ProductId) ||
		(reportProductUsagesAfterUpdate[0].ProductUsageId != productUsages.ProductUsages[0].Id) ||
		(reportProductUsagesAfterUpdate[0].Quantity != productUsages.ProductUsages[0].Quantity) ||
		(reportProductUsagesAfterUpdate[0].UnReportedQuantity != unReportedQty) {
		t.Fatalf("expected values do not match for product usage report after update")
	}

	_, err = usageServiceClient.MarkProductUsageAsReported(ctx, &pb.ReportProductUsageId{
		Id: reportProductUsages[0].Id,
	})

	if err != nil {
		t.Fatalf("failed to mark product usage report as reported: %v", err)
	}

	reportProductUsagesAfterMarkingReported := searchProductUsagesUnreported(t, ctx)

	if len(reportProductUsagesAfterMarkingReported) != 0 {
		t.Fatalf("invalid length of report product usages")
	}
}

func searchProductUsagesUnreported(t *testing.T, ctx context.Context) []*pb.ReportProductUsage {

	reportProductUsages := []*pb.ReportProductUsage{}

	usageServiceClient := pb.NewUsageServiceClient(clientConn)
	reported := false
	stream, _ := usageServiceClient.SearchProductUsagesReport(ctx, &pb.ProductUsagesReportFilter{Reported: &reported})

	for {

		reportProductUsage, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("unexpected error for searching report product usages %v", err)
		}

		reportProductUsages = append(reportProductUsages, reportProductUsage)
	}

	return reportProductUsages
}
