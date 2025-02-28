// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const defaultProductUsageRecordQty = 10

func TestProductUsageRecordData(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestProductUsageRecordData")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageRecordData := NewUsageRecordData(usageDb)
	productName := uuid.NewString()
	transactionId := uuid.NewString()
	region := "us-west-2"

	productUsageRecordCreate := GetProductUsageRecordCreate(MustNewCloudAcctId(), &productName, region,
		transactionId, defaultProductUsageRecordQty,
		map[string]string{
			"availabilityZone": region,
			"serviceType":      "ComputeAsAService",
		})

	err := usageRecordData.StoreProductUsageRecord(ctx, productUsageRecordCreate)

	if err != nil {
		t.Fatalf("unexpected error for storing product usage record %v", err)
	}

	unReportedProductUsageRecords, err := usageRecordData.GetUnreportedProductUsageRecords(ctx)

	if err != nil {
		t.Fatalf("unexpected error for getting unreported product usage records %v", err)
	}

	if len(unReportedProductUsageRecords) != 1 {
		t.Fatalf("should have received one unreported product usage record")
	}

	unReportedProductUsageRecord := unReportedProductUsageRecords[0]

	if *unReportedProductUsageRecord.ProductName != productName ||
		unReportedProductUsageRecord.Quantity != defaultProductUsageRecordQty ||
		unReportedProductUsageRecord.Region != region {
		t.Fatalf("values do not match for product usage record")
	}

	err = usageRecordData.MarkProductUsageRecordAsReported(ctx, unReportedProductUsageRecord.Id)
	if err != nil {
		t.Fatalf("unexpected error for marking product usage record as reported %v", err)
	}

	unReportedProductUsageRecords, err = usageRecordData.GetUnreportedProductUsageRecords(ctx)

	if err != nil {
		t.Fatalf("unexpected error for getting unreported product usage records %v", err)
	}

	if len(unReportedProductUsageRecords) != 0 {
		t.Fatalf("should not have received any unreported product usage record")
	}
}

func TestInvalidProductUsageRecordData(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestInvalidProductUsageRecordData")
	logger.Info("BEGIN")
	defer logger.Info("End")

	usageRecordData := NewUsageRecordData(usageDb)
	cloudAccountId := MustNewCloudAcctId()
	recordId := uuid.NewString()
	productName := uuid.NewString()
	transactionId := uuid.NewString()
	var invalidProductUsageRecordQty float64 = 10
	region := "us-west-2"

	invalidProductUsageRecordCreate := GetInvalidProductUsageRecordCreate(&recordId, cloudAccountId, &productName, region,
		transactionId, invalidProductUsageRecordQty,
		map[string]string{
			"availabilityZone": region,
			"serviceType":      "ComputeAsAService",
		}, pb.ProductUsageRecordInvalidityReason_DEFAULT_PRODUCT_USAGE_RECORD_INVALIDITY_REASON)

	err := usageRecordData.StoreInvalidProductUsageRecord(ctx, invalidProductUsageRecordCreate)

	if err != nil {
		t.Fatalf("unexpected error for storing invalid product usage record %v", err)
	}

}
