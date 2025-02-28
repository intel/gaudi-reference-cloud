// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	meteringQuery "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/db/query"
	meteringServer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CreateAndGetCloudAccount(t *testing.T, ctx context.Context, user string, acctType pb.AccountType) *pb.CloudAccount {
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	acctCreate := pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  acctType,
	}
	id, err := cloudAcctClient.Create(ctx, &acctCreate)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	acctOut, err := cloudAcctClient.GetById(context.Background(), &pb.CloudAccountId{Id: id.GetId()})
	if err != nil {
		t.Fatalf("failed to read account: %v", err)
	}
	return acctOut
}

func GetProductUsageRecordCreate(cloudAccountId string, productName *string, region string,
	transactionId string, quantity float64, properties map[string]string) *pb.ProductUsageRecordCreate {

	return &pb.ProductUsageRecordCreate{
		TransactionId:  transactionId,
		CloudAccountId: cloudAccountId,
		Region:         region,
		Quantity:       quantity,
		ProductName:    productName,
		Timestamp:      timestamppb.New(time.Now()),
		StartTime:      timestamppb.New(time.Now()),
		EndTime:        timestamppb.New(time.Now()),
		Properties:     properties,
	}
}

func GetInvalidProductUsageRecordCreate(recordId *string, cloudAccountId string, productName *string, region string,
	transactionId string, quantity float64, properties map[string]string, invalidityReason pb.ProductUsageRecordInvalidityReason) *pb.InvalidProductUsageRecordCreate {

	return &pb.InvalidProductUsageRecordCreate{
		RecordId:                           recordId,
		TransactionId:                      transactionId,
		CloudAccountId:                     cloudAccountId,
		Region:                             region,
		Quantity:                           quantity,
		ProductName:                        productName,
		Timestamp:                          timestamppb.New(time.Now()),
		StartTime:                          timestamppb.New(time.Now()),
		EndTime:                            timestamppb.New(time.Now()),
		Properties:                         properties,
		ProductUsageRecordInvalidityReason: invalidityReason,
	}
}

func GetResourceUsageCreate(cloudAccountId string, resourceId string, productId string, region string,
	transactionId string, usageUnitType string) *pb.ResourceUsageCreate {
	return &pb.ResourceUsageCreate{
		CloudAccountId: cloudAccountId,
		ResourceId:     resourceId,
		ResourceName:   uuid.NewString(),
		ProductId:      productId,
		ProductName:    "someName",
		TransactionId:  transactionId,
		Region:         region,
		Quantity:       10,
		Rate:           10,
		UsageUnitType:  usageUnitType,
		StartTime:      timestamppb.New(time.Now()),
		EndTime:        timestamppb.New(time.Now()),
	}
}

func InitializeTests(ctx context.Context, t *testing.T) {
	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	err := meteringQuery.DeleteAllRecords(ctx, meteringServer.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	err = usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all usages")
	}

	err = usageRecordData.DeleteAllUsageRecords(ctx)
	if err != nil {
		t.Fatalf("could not delete all usage records")
	}
}

func CreateComputeMeteringRecord(ctx context.Context, t *testing.T, cloudAcctId string, resourceId string, instanceType string,
	transactionId string, meteringRecordTimeMetric float32) {
	meteringServiceClient := pb.NewMeteringServiceClient(meteringConn)
	_, err := meteringServiceClient.Create(ctx,
		GetComputeUsageRecordCreate(cloudAcctId, resourceId, transactionId,
			DefaultServiceRegion, idcComputeServiceName, instanceType, meteringRecordTimeMetric))

	if err != nil {
		t.Fatalf("failed to create metering record: %v", err)
	}
}

func CreateStorageMeteringRecord(ctx context.Context, t *testing.T, cloudAcctId string, serviceType string, resourceId string,
	transactionId string, meteringRecordTimeMetric float32, meteringRecordStorageTimeMetric float32) {
	meteringServiceClient := pb.NewMeteringServiceClient(meteringConn)
	_, err := meteringServiceClient.Create(ctx,
		GetStorageUsageRecordCreate(cloudAcctId, serviceType, resourceId, transactionId,
			DefaultServiceRegion, serviceType, meteringRecordTimeMetric, meteringRecordStorageTimeMetric))

	if err != nil {
		t.Fatalf("failed to create metering record: %v", err)
	}
}

func MarkAllMeteringAsReported(ctx context.Context, t *testing.T) {
	meteringServiceClient := pb.NewMeteringServiceClient(meteringConn)
	reported := false
	meteringFilter := &pb.UsageFilter{
		Reported: &reported,
	}

	meteringSearchClient, err := meteringServiceClient.Search(ctx, meteringFilter)
	if err != nil {
		t.Fatalf("failed to get metering search client: %v", err)
	}

	var usageIds []int64
	for {
		resourceMeteringRecordR, err := meteringSearchClient.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				t.Fatalf("invalid error received for searching records: %v", err)
			}
		}
		usageIds = append(usageIds, resourceMeteringRecordR.Id)
	}

	usageUpdate := pb.UsageUpdate{
		Id:       usageIds,
		Reported: true,
	}
	_, err = meteringServiceClient.Update(ctx, &usageUpdate)
	if err != nil {
		t.Fatalf("failed to update metering record: %v", err)
	}
}
