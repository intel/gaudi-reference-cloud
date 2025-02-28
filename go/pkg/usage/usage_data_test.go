// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStoreResourceMetering(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestStoreResourceMetering")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageData := NewUsageData(usageDb)
	resourceId := uuid.NewString()
	transactionId := uuid.NewString()
	region := "us-west-2"

	resourceMetering := &ResourceMeteringCreate{
		ResourceId:     resourceId,
		CloudAccountId: MustNewCloudAcctId(),
		TransactionId:  transactionId,
		Region:         region,
		LastRecorded:   time.Now(),
	}

	_, err := usageData.StoreResourceMetering(ctx, resourceMetering)

	if err != nil {
		t.Fatalf("unexpected error for storing resource metering %v", err)
	}
}

func TestGetResourceMeteringLastTimestamp(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestGetResourceMeteringLastTimestamp")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageData := NewUsageData(usageDb)
	resourceId := uuid.NewString()
	transactionId := uuid.NewString()
	region := "us-west-2"
	lastRecorded := time.Now()

	resourceMetering := &ResourceMeteringCreate{
		ResourceId:     resourceId,
		CloudAccountId: MustNewCloudAcctId(),
		TransactionId:  transactionId,
		Region:         region,
		LastRecorded:   lastRecorded,
	}

	_, err := usageData.StoreResourceMetering(ctx, resourceMetering)
	if err != nil {
		t.Fatalf("unexpected error for storing resource metering %v", err)
	}

	_, err = usageData.GetLastMeteringTimestampForResource(ctx, resourceId)
	if err != nil {
		t.Fatalf("unexpected error for getting last metering timestamp %v", err)
	}
}

func TestUpdateResourceMetering(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestUpdateResourceMetering")
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAccountId := MustNewCloudAcctId()
	usageData := NewUsageData(usageDb)
	resourceId := uuid.NewString()
	transactionId := uuid.NewString()
	region := "us-west-2"
	lastRecorded := time.Now()

	transactionId1 := uuid.NewString()

	resourceMeteringCreate := &ResourceMeteringCreate{
		ResourceId:     resourceId,
		CloudAccountId: cloudAccountId,
		TransactionId:  transactionId,
		Region:         region,
		LastRecorded:   lastRecorded,
	}

	updatedResourceMetering := &ResourceMeteringUpdate{
		TransactionId: transactionId1,
		LastRecorded:  lastRecorded,
	}

	_, err := usageData.StoreResourceMetering(ctx, resourceMeteringCreate)
	if err != nil {
		t.Fatalf("unexpected error for storing resource metering %v", err)
	}

	resourceMetering, err := usageData.GetMeteringForResource(ctx, resourceId)
	if err != nil {
		t.Fatalf("failed to get resource metering %v", err)
	}

	if (resourceMeteringCreate.ResourceId != resourceMetering.ResourceId) ||
		(resourceMeteringCreate.CloudAccountId != resourceMetering.CloudAccountId) ||
		(resourceMeteringCreate.TransactionId != resourceMetering.TransactionId) ||
		(resourceMeteringCreate.Region != resourceMetering.Region) {
		t.Fatalf("values not same after create and get")
	}

	err = usageData.UpdateResourceMetering(ctx, resourceId, updatedResourceMetering)
	if err != nil {
		t.Fatalf("unexpected error for updating resource metering %v", err)
	}

	resourceMeteringAfterUpdate, err := usageData.GetMeteringForResource(ctx, resourceId)
	if err != nil {
		t.Fatalf("failed to get resource metering after update %v", err)
	}

	if resourceMeteringAfterUpdate.TransactionId != transactionId1 {
		t.Fatalf("values not updated")
	}

	if (resourceMeteringCreate.ResourceId != resourceMeteringAfterUpdate.ResourceId) ||
		(resourceMeteringCreate.CloudAccountId != resourceMeteringAfterUpdate.CloudAccountId) ||
		(resourceMeteringCreate.Region != resourceMeteringAfterUpdate.Region) {
		t.Fatalf("incorrect values updated")
	}
}

func TestStoreResourceUsage(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestStoreResourceUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageData := NewUsageData(usageDb)
	resourceId := uuid.NewString()
	productId := uuid.NewString()
	transactionId := uuid.NewString()
	region := "us-west-2"
	usageUnitType := "mins"

	resourceUsageCreate := GetResourceUsageCreate(MustNewCloudAcctId(), resourceId, productId, region,
		transactionId, usageUnitType)

	_, err := usageData.StoreResourceUsage(ctx, resourceUsageCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing resource usage %v", err)
	}
}

func TestDeleteResourceUsage(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestDeleteResourceUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageData := NewUsageData(usageDb)
	resourceId := uuid.NewString()
	productId := uuid.NewString()
	transactionId := uuid.NewString()
	region := "us-west-2"
	usageUnitType := "mins"

	resourceUsageCreate := GetResourceUsageCreate(MustNewCloudAcctId(), resourceId, productId, region,
		transactionId, usageUnitType)

	resourceUsageCreated, err := usageData.StoreResourceUsage(ctx, resourceUsageCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing resource usage %v", err)
	}

	err = usageData.DeleteResourceUsage(ctx, resourceUsageCreated.Id)

	if err != nil {
		t.Fatalf("unexpected error for deleting resource usage %v", err)
	}
}

func TestSumResourceUsageQty(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestSumResourceUsageQty")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageData := NewUsageData(usageDb)
	resourceId := uuid.NewString()
	productId := uuid.NewString()
	transactionId := uuid.NewString()
	region := "us-west-2"
	usageUnitType := "mins"

	quantity := 10
	resourceUsageCreate := &pb.ResourceUsageCreate{
		CloudAccountId: MustNewCloudAcctId(),
		ResourceId:     resourceId,
		ResourceName:   uuid.NewString(),
		ProductId:      productId,
		ProductName:    "someName",
		TransactionId:  transactionId,
		Region:         region,
		Quantity:       float64(quantity),
		Rate:           10,
		UsageUnitType:  usageUnitType,
		StartTime:      timestamppb.New(time.Now()),
		EndTime:        timestamppb.New(time.Now()),
	}

	_, err := usageData.StoreResourceUsage(ctx, resourceUsageCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing resource usage %v", err)
	}

	_, err = usageData.StoreResourceUsage(ctx, resourceUsageCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing resource usage %v", err)
	}

	sumOfResourceUsageQty, err := usageData.GetTotalResourceUsageQty(ctx, resourceId)

	if err != nil {
		t.Fatalf("unexpected error for deleting resource usage %v", err)
	}

	if sumOfResourceUsageQty != (2 * float64(quantity)) {
		t.Fatalf("incorrect sum of resource usage qty")
	}
}

func TestUpdateAndGetResourceUsage(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestUpdateAndGetResourceUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageData := NewUsageData(usageDb)
	resourceId := uuid.NewString()
	productId := uuid.NewString()
	transactionId := uuid.NewString()
	region := "us-west-2"
	usageUnitType := "mins"

	cloudAcctId := MustNewCloudAcctId()
	resourceUsageCreate := GetResourceUsageCreate(cloudAcctId, resourceId, productId, region,
		transactionId, usageUnitType)

	resourceUsageCreated, err := usageData.StoreResourceUsage(ctx, resourceUsageCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing resource usage %v", err)
	}

	resourceUsage, err := usageData.GetResourceUsageById(ctx, resourceUsageCreated.Id)
	if err != nil {
		t.Fatalf("failed to get resource usage by id %v", err)
	}

	if (resourceUsageCreate.CloudAccountId != resourceUsage.CloudAccountId) ||
		(resourceUsageCreate.ResourceId != resourceUsage.ResourceId) ||
		(resourceUsageCreate.ResourceName != resourceUsage.ResourceName) ||
		(resourceUsageCreate.ProductId != resourceUsage.ProductId) ||
		(resourceUsageCreate.ProductName != resourceUsage.ProductName) ||
		(resourceUsageCreate.TransactionId != resourceUsage.TransactionId) ||
		(resourceUsageCreate.Region != resourceUsage.Region) ||
		(resourceUsageCreate.Quantity != resourceUsage.Quantity) ||
		(resourceUsageCreate.Rate != resourceUsage.Rate) ||
		(resourceUsageCreate.UsageUnitType != resourceUsage.UsageUnitType) {
		t.Fatalf("values not same after create and get")
	}

	resourceUsageUpdate := &ResourceUsageUpdate{
		UnReportedQuantity: 5,
	}

	err = usageData.UpdateResourceUsage(ctx, resourceUsageCreated.Id, resourceUsageUpdate)

	if err != nil {
		t.Fatalf("unexpected error for updating resource usage %v", err)
	}

	resourceUsagesFilter := &pb.ResourceUsagesFilter{
		ResourceId: &resourceId,
	}

	resourceUsages, err := usageData.SearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		t.Fatalf("failed to search resource usages %v", err)
	}

	if resourceUsages.ResourceUsages[0].UnReportedQuantity != 5 {
		t.Fatalf("update failed")
	}

	resourceUsageAfterUpdate, err := usageData.GetResourceUsageById(ctx, resourceUsages.ResourceUsages[0].Id)
	if err != nil {
		t.Fatalf("failed to get resource usage by id after update %v", err)
	}

	if resourceUsages.ResourceUsages[0].Id != resourceUsageAfterUpdate.Id {
		t.Fatalf("ids do not match after getting")
	}

	if (resourceUsageCreate.CloudAccountId != resourceUsageAfterUpdate.CloudAccountId) ||
		(resourceUsageCreate.ResourceId != resourceUsageAfterUpdate.ResourceId) ||
		(resourceUsageCreate.ResourceName != resourceUsageAfterUpdate.ResourceName) ||
		(resourceUsageCreate.ProductId != resourceUsageAfterUpdate.ProductId) ||
		(resourceUsageCreate.ProductName != resourceUsageAfterUpdate.ProductName) ||
		(resourceUsageCreate.TransactionId != resourceUsage.TransactionId) ||
		(resourceUsageCreate.Region != resourceUsageAfterUpdate.Region) ||
		(resourceUsageCreate.Quantity != resourceUsageAfterUpdate.Quantity) ||
		(resourceUsageCreate.Rate != resourceUsageAfterUpdate.Rate) ||
		(resourceUsageCreate.UsageUnitType != resourceUsageAfterUpdate.UsageUnitType) {
		t.Fatalf("values changed after update")
	}

}

func TestProductUsage(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestProductUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageData := NewUsageData(usageDb)
	productId := uuid.NewString()
	region := "us-west-2"
	usageUnitType := "mins"

	productUsageCreate := &ProductUsageCreate{
		Id:             uuid.NewString(),
		CloudAccountId: MustNewCloudAcctId(),
		ProductId:      productId,
		ProductName:    "someName",
		Region:         region,
		Quantity:       10,
		Rate:           10,
		UsageUnitType:  usageUnitType,
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	err := usageData.StoreProductUsage(ctx, productUsageCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing product usage %v", err)
	}

	productUsagesFilter := &pb.ProductUsagesFilter{
		ProductId: &productId,
	}

	productUsages, err := usageData.SearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to search product usages %v", err)
	}

	productUsage, err := usageData.GetProductUsageById(ctx, productUsages.ProductUsages[0].Id)
	if err != nil {
		t.Fatalf("failed to get product usage by id %v", err)
	}

	if (productUsageCreate.CloudAccountId != productUsage.CloudAccountId) ||
		(productUsageCreate.ProductId != productUsage.ProductId) ||
		(productUsageCreate.ProductName != productUsage.ProductName) ||
		(productUsageCreate.Region != productUsage.Region) ||
		(productUsageCreate.Quantity != productUsage.Quantity) ||
		(productUsageCreate.Rate != productUsage.Rate) ||
		(productUsageCreate.UsageUnitType != productUsage.UsageUnitType) {
		t.Fatalf("values not same after create and get")
	}

	productUsageUpdate := &ProductUsageUpdate{
		CreationTime: time.Now(),
		Quantity:     100,
		StartTime:    time.Now(),
		EndTime:      time.Now(),
	}

	err = usageData.UpdateProductUsage(ctx, productUsages.ProductUsages[0].Id, productUsageUpdate)

	if err != nil {
		t.Fatalf("unexpected error for updating product usage %v", err)
	}

	productUsagesAfterUpdate, err := usageData.SearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to search product usages %v", err)
	}

	if productUsagesAfterUpdate.ProductUsages[0].Quantity != 100 {
		t.Fatalf("update failed")
	}

	productUsageAfterUpdate, err := usageData.GetProductUsageById(ctx, productUsagesAfterUpdate.ProductUsages[0].Id)
	if err != nil {
		t.Fatalf("failed to get product usage by id after update %v", err)
	}

	if productUsages.ProductUsages[0].Id != productUsageAfterUpdate.Id {
		t.Fatalf("ids do not match after getting")
	}

	if (productUsageCreate.CloudAccountId != productUsage.CloudAccountId) ||
		(productUsageCreate.ProductId != productUsage.ProductId) ||
		(productUsageCreate.ProductName != productUsage.ProductName) ||
		(productUsageCreate.Region != productUsage.Region) ||
		(productUsageCreate.Quantity != productUsage.Quantity) ||
		(productUsageCreate.Rate != productUsage.Rate) ||
		(productUsageCreate.UsageUnitType != productUsage.UsageUnitType) {
		t.Fatalf("values changed after update")
	}

}

func TestDeleteProductUsage(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestDeleteProductUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageData := NewUsageData(usageDb)
	productId := uuid.NewString()
	region := "us-west-2"
	usageUnitType := "mins"

	productUsageCreate := &ProductUsageCreate{
		Id:             uuid.NewString(),
		CloudAccountId: MustNewCloudAcctId(),
		ProductId:      productId,
		ProductName:    "someName",
		Region:         region,
		Quantity:       10,
		Rate:           10,
		UsageUnitType:  usageUnitType,
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	err := usageData.StoreProductUsage(ctx, productUsageCreate, time.Now())

	if err != nil {
		t.Fatalf("unexpected error for storing product usage %v", err)
	}

	productUsagesFilter := &pb.ProductUsagesFilter{
		ProductId: &productId,
	}

	productUsages, err := usageData.SearchProductUsages(ctx, productUsagesFilter)
	if err != nil {
		t.Fatalf("failed to search product usages %v", err)
	}

	err = usageData.DeleteProductUsage(ctx, productUsages.ProductUsages[0].ProductId)

	if err != nil {
		t.Fatalf("failed to delete product usage %v", err)
	}
}
