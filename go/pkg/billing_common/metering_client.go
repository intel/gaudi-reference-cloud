// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type MeteringClient struct {
	meteringServiceClient pb.MeteringServiceClient
}

type AggrUsage struct {
	ProductUsages []ProductUsage
}

type ProductUsage struct {
	InstanceType string
	InstanceName string
	ServiceName  string
	RegionName   string
	Rate         *pb.Rate
	Usage        float64
	UsageStart   time.Time
	UsageEnd     time.Time
}

func NewTestMeteringClient(meteringServiceClient pb.MeteringServiceClient) *MeteringClient {
	return &MeteringClient{meteringServiceClient: meteringServiceClient}
}

func NewMeteringClient(ctx context.Context, resolver grpcutil.Resolver) (*MeteringClient, error) {
	logger := log.FromContext(ctx).WithName("MeteringClient.InitUsageClient")
	meteringAddr, err := resolver.Resolve(ctx, "metering")
	if err != nil {
		logger.Error(err, "grpc resolver not able to connect", "addr", meteringAddr)
		return nil, err
	}
	meteringConn, err := grpcConnect(ctx, meteringAddr)
	if err != nil {
		return nil, err
	}

	return &MeteringClient{meteringServiceClient: pb.NewMeteringServiceClient(meteringConn)}, nil
}

func NewMeteringClientForTest(meteringServiceClient pb.MeteringServiceClient) *MeteringClient {
	return &MeteringClient{meteringServiceClient: meteringServiceClient}
}

// GetUsagesToReport to get the usages that have not been reported and need to be reported.
func (meteringClient *MeteringClient) GetUsagesToReport(ctx context.Context) ([]*pb.Usage, error) {
	logger := log.FromContext(ctx).WithName("MeteringClient.GetUsagesToReport")
	logger.Info("start get usage to report")
	usageReported := false
	usageFilter := pb.UsageFilter{Reported: &usageReported}
	usagesToReport, err := meteringClient.GetFilteredUsagesToReport(ctx, &usageFilter)
	if err != nil {
		logger.Error(err, "error in getting usages to report")
		return nil, err
	}
	logger.Info("end get usage to report", "usagesToReport", usagesToReport)
	return usagesToReport, nil
}

func (meteringClient *MeteringClient) GetFilteredUsagesToReport(ctx context.Context, usageFilter *pb.UsageFilter) ([]*pb.Usage, error) {
	logger := log.FromContext(ctx).WithName("MeteringClient.GetUsagesToReport").V(1)
	defer logger.Info("get filtered end usage to report")
	logger.Info("get filtered start usage to report ", "usageFilter", usageFilter)
	meteringServiceSearchClient, err := meteringClient.meteringServiceClient.Search(ctx, usageFilter)
	if err != nil {
		logger.Error(err, "error in usage service client response")
		return nil, err
	}
	usagesToReport := []*pb.Usage{}
	for {
		usage, err := meteringServiceSearchClient.Recv()
		logger.Info("metering record", "usage", usage)
		if err == io.EOF {
			logger.Info("usages to be reported", "report", usagesToReport)
			close_err := meteringServiceSearchClient.CloseSend()
			if close_err != nil {
				logger.Error(close_err, "error closing the metering service search client")
			}
			return usagesToReport, nil
		}
		if err != nil {
			logger.Error(err, "failed to read usage not reported")
			return nil, err
		}
		usagesToReport = append(usagesToReport, usage)
	}
}

// UpdateUsageAsReported to update the usage once reported.
// For the caller to handle usages that have not been updated to reported, returns such usages.
func (meteringClient *MeteringClient) UpdateUsageAsReported(ctx context.Context, usageToUpdateAsReported *pb.Usage) error {
	logger := log.FromContext(ctx).WithName("MeteringClient.UpdateUsageAsReported")

	usageUpdate := pb.UsageUpdate{
		Id:       []int64{usageToUpdateAsReported.GetId()},
		Reported: true,
	}
	_, err := meteringClient.meteringServiceClient.Update(ctx, &usageUpdate)
	if err != nil {
		logger.Error(err, "failed to update usage to reported for", "id", usageToUpdateAsReported.GetId())
		return err
	}
	return nil
}

// instead of a two step call to invalidate and mark as reported in the caller, doing it here.
func (meteringClient *MeteringClient) InvalidateMeteringRecords(ctx context.Context, meteringRecords []*pb.MeteringRecord,
	meteringInvalidityReason pb.MeteringRecordInvalidityReason) error {
	logger := log.FromContext(ctx).WithName("MeteringClient.InvalidateMeteringRecords")

	var usageIds []int64
	invalidateMeteringRecordsCreate := &pb.CreateInvalidMeteringRecords{}
	for _, meteringRecord := range meteringRecords {
		usageIds = append(usageIds, meteringRecord.Id)
		invalidateMeteringRecordsCreate.CreateInvalidMeteringRecords = append(invalidateMeteringRecordsCreate.CreateInvalidMeteringRecords,
			&pb.InvalidMeteringRecordCreate{
				RecordId:                       fmt.Sprint(meteringRecord.Id),
				TransactionId:                  meteringRecord.TransactionId,
				ResourceId:                     meteringRecord.ResourceId,
				ResourceName:                   meteringRecord.ResourceName,
				CloudAccountId:                 meteringRecord.CloudAccountId,
				Region:                         meteringRecord.Region,
				Timestamp:                      meteringRecord.Timestamp,
				Properties:                     meteringRecord.Properties,
				MeteringRecordInvalidityReason: meteringInvalidityReason,
			})
	}
	_, err := meteringClient.meteringServiceClient.CreateInvalidRecords(ctx, invalidateMeteringRecordsCreate)
	if err != nil {
		logger.Error(err, "failed to update metering records as invalid")
	}
	err = meteringClient.UpdateUsagesAsReported(ctx, usageIds)
	if err != nil {
		logger.Error(err, "failed to update invalid metering records as reported")
		return err
	}
	return nil
}

// retain this function along with the option to update one at a time.
// update one at a time when there is only usage to be updated to be able to identify the mapping between usage failed to be updated
// and usage - currently not returned by metering.
func (meteringClient *MeteringClient) UpdateUsagesAsReported(ctx context.Context, usageIds []int64) error {
	logger := log.FromContext(ctx).WithName("MeteringClient.UpdateUsageAsReported")

	if len(usageIds) == 0 {
		return nil
	}

	usageUpdate := pb.UsageUpdate{
		Id:       usageIds,
		Reported: true,
	}
	_, err := meteringClient.meteringServiceClient.Update(ctx, &usageUpdate)
	if err != nil {
		logger.Error(err, "failed to update usages to be updated as reported")
		return err
	}
	return nil
}

// FindPreviousUsage to get the previous usage for a chosen usage.
func (meteringClient *MeteringClient) FindPreviousUsage(ctx context.Context, usageForFindingPrevious *pb.Usage) (*pb.Usage, error) {
	logger := log.FromContext(ctx).WithName("MeteringClient.FindPreviousUsage").V(1)
	usagePrevious := pb.UsagePrevious{
		Id:         usageForFindingPrevious.GetId(),
		ResourceId: usageForFindingPrevious.GetResourceId(),
	}
	usage, err := meteringClient.meteringServiceClient.FindPrevious(ctx, &usagePrevious)
	if err != nil {
		logger.Info("failed to find the previous usage for", "id", usageForFindingPrevious.GetId(), "error", err)
		return nil, err
	}
	return usage, nil
}

// CreateUsage to create usage for integration testing
func (meteringClient *MeteringClient) CreateUsage(ctx context.Context, usageToCreate *pb.UsageCreate) error {
	logger := log.FromContext(ctx).WithName("MeteringClient.CreateUsage")
	_, err := meteringClient.meteringServiceClient.Create(ctx, usageToCreate)
	if err != nil {
		logger.Error(err, "failed to create usage")
		return err
	}
	return nil
}

func (meteringClient *MeteringClient) SearchUnreportedResourceMeteringRecords(ctx context.Context) (*pb.ResourceMeteringRecordsList, error) {
	logger := log.FromContext(ctx).WithName("MeteringClient.SearchUnreportedMeteringRecordsForCloudAccount")
	logger.Info("getting unreported resource metering records")
	reported := false
	return meteringClient.SearchResourceMeteringRecordsUsingStream(ctx, &reported)
}

func (meteringClient *MeteringClient) SearchReportedResourceMeteringRecords(ctx context.Context) (*pb.ResourceMeteringRecordsList, error) {
	logger := log.FromContext(ctx).WithName("MeteringClient.SearchUnreportedMeteringRecordsForCloudAccount")
	logger.Info("getting reported resource metering records")
	reported := true
	return meteringClient.SearchResourceMeteringRecordsUsingStream(ctx, &reported)
}

func (meteringClient *MeteringClient) SearchResourceMeteringRecordsUsingStream(ctx context.Context, reported *bool) (*pb.ResourceMeteringRecordsList, error) {
	logger := log.FromContext(ctx).WithName("MeteringClient.SearchResourceMeteringRecordsUsingStream")

	meteringFilter := &pb.MeteringFilter{
		Reported: reported,
	}
	resourceMeteringRecords := &pb.ResourceMeteringRecordsList{}
	meteringSearchClient, err := meteringClient.meteringServiceClient.SearchResourceMeteringRecordsAsStream(context.Background(), meteringFilter)
	if err != nil {
		logger.Error(err, "failed to get resource metering records client")
		return nil, err
	}

	for {
		resourceMeteringRecordsR, err := meteringSearchClient.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		resourceMeteringRecords.ResourceMeteringRecordsList = append(resourceMeteringRecords.ResourceMeteringRecordsList,
			resourceMeteringRecordsR.ResourceMeteringRecordsList...)
	}

	return resourceMeteringRecords, nil
}

func (meteringClient *MeteringClient) IsMeteringRecordAvailable(ctx context.Context, filter *pb.MeteringAvailableFilter) (bool, error) {
	logger := log.FromContext(ctx).WithName("MeteringClient.IsMeteringRecordAvailable")
	resp, err := meteringClient.meteringServiceClient.IsMeteringRecordAvailable(ctx, filter)
	if err != nil {
		logger.Info("failed to find the previous usage for", "CloudAccountId", filter.CloudAccountId, "error", err)
		return false, err
	}
	return resp.MeteringDataAvailable, nil
}
