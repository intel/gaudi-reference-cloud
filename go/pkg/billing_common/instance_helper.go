// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	instanceDeletedPropertyKey = "deleted"
)

type Instance struct {
	instanceDeleted string
	instanceType    string
}

func GetPaidInstanceTypes(ctx context.Context, meteringServiceClient *MeteringClient, productServiceClient *ProductClient, instFilter *pb.UsageFilter, accountType pb.AccountType) ([]string, error) {
	log := log.FromContext(ctx).WithName("InstanceHelper.GetPaidInstanceTypes")
	log.Info("InstanceHelper.GetPaidInstanceTypes: enter")
	defer log.Info("InstanceHelper.GetPaidInstanceTypes: return")

	meteringServiceSearchClient, err := meteringServiceClient.meteringServiceClient.Search(ctx, instFilter)
	if err != nil {
		log.Error(err, "error in metering service client response")
		return nil, status.Errorf(codes.Internal, "error encountered in metering service client response")
	}

	allInstances := map[string]*Instance{}
	for {
		record, err := meteringServiceSearchClient.Recv()
		if err == io.EOF {
			if err := meteringServiceSearchClient.CloseSend(); err != nil {
				log.Error(err, "error closing the send stream")
				return nil, status.Errorf(codes.Internal, "error closing the send stream")
			}
			break
		}
		if err != nil {
			log.Error(err, "failed to read records")
			return nil, status.Errorf(codes.Internal, "error reading records")
		}
		log.Info("Received", "record", record)

		instanceTypeStr, found := record.GetProperties()[instanceTypePropertyKey]
		if !found {
			log.Info("instance type not found")
		}

		instanceDeletedStr, found := record.GetProperties()[instanceDeletedPropertyKey]
		if !found {
			log.Info("instance delete value not found")
		}

		instance, found := allInstances[record.ResourceId]
		if !found {
			ins := &Instance{}
			ins.instanceType = instanceTypeStr
			ins.instanceDeleted = instanceDeletedStr
			allInstances[record.ResourceId] = ins
			log.Info("Add instance to map", "instance", ins)
		} else {
			// continue if instance is deleted
			if instance.instanceDeleted == "true" || instance.instanceDeleted == instanceDeletedStr {
				continue
			}
			instance.instanceDeleted = instanceDeletedStr
			allInstances[record.ResourceId] = instance
		}
	}

	var paidInstancetypes []string
	if len(allInstances) == 0 {
		log.Info("No running instances or records found")
		return paidInstancetypes, nil
	}

	rateList, err := DetermineInstanceTypesRates(ctx, productServiceClient, accountType)
	if err != nil {
		log.Error(err, "failed to determine instance types rates")
		return nil, status.Errorf(codes.Internal, "error determining instance types rates")
	}

	log.Info("DetermineInstanceTypesRates", "rates", rateList)

	// Filter free or deleted instances
	addedInstypesHash := make(map[string]bool)
	for resourceId, instance := range allInstances {
		log.Info("Instance", resourceId, instance)
		rate, found := rateList[instance.instanceType]
		if !found {
			log.Info("Instance type not found", "instanceType", instance.instanceType)
			continue
		}

		if !(rate > 0) || instance.instanceDeleted == "true" {
			log.Info("Free or deleted Instance found", resourceId, instance.instanceType)
			continue
		}

		_, found = addedInstypesHash[instance.instanceType]
		if !found {
			paidInstancetypes = append(paidInstancetypes, instance.instanceType)
			addedInstypesHash[instance.instanceType] = true
		}
	}

	// Return the list of paid instancetypes.
	return paidInstancetypes, nil
}

func DetermineInstanceTypesRates(ctx context.Context, productServiceClient *ProductClient, accountType pb.AccountType) (map[string]float32, error) {
	log := log.FromContext(ctx).WithName("InstanceHelper.DetermineInstanceTypesRates")
	log.Info("InstanceHelper.DetermineInstanceTypesRates: enter")
	defer log.Info("InstanceHelper.DetermineInstanceTypesRates: return")

	rateList := make(map[string]float32)
	filter := pb.ProductFilter{AccountType: &accountType}
	regProducts, err := productServiceClient.GetProductCatalogProductsWithFilter(ctx, &filter)
	if err != nil {
		log.Error(err, "error reading product catalog entries")
		return nil, fmt.Errorf("error reading product catalog entries")
	}

	for _, regproduct := range regProducts {

		rates := regproduct.GetRates()
		instanceType := regproduct.Metadata["instanceType"]

		for _, rate := range rates {
			floatValue, err := strconv.ParseFloat(rate.GetRate(), 32)
			if err != nil {
				return nil, fmt.Errorf("error parsing rate")
			}
			rateList[instanceType] = float32(floatValue)
		}

	}

	return rateList, nil
}

func DetermineInstanceSearchWindow(ctx context.Context, timeWindow time.Duration) (*timestamppb.Timestamp, *timestamppb.Timestamp, error) {
	// func DetermineInstanceSearchWindow(ctx context.Context) (*timestamppb.Timestamp, *timestamppb.Timestamp, error) {
	log := log.FromContext(ctx).WithName("InstanceHelper.DetermineInstanceSearchWindow")
	log.Info("InstanceHelper.DetermineInstanceSearchWindow: enter")
	defer log.Info("InstanceHelper.DetermineInstanceSearchWindow: return")
	if !(timeWindow > 0) {
		log.Info("invalid search duration")
		return nil, nil, fmt.Errorf("invalid search duration")
	}

	endTimeProto := timestamppb.Now()
	duration := -1 * (timeWindow * time.Hour)
	startTimeUTC := endTimeProto.AsTime().Add(duration)
	startTimeProto := timestamppb.New(startTimeUTC)

	formattedStartTime := startTimeProto.AsTime().Format("2006-01-02 15:04:05")
	formattedEndTime := endTimeProto.AsTime().Format("2006-01-02 15:04:05")
	log.Info("InstanceHelper.DetermineInstanceSearchWindows", "startTime", formattedStartTime, "endTime", formattedEndTime)

	return startTimeProto, endTimeProto, nil
}
