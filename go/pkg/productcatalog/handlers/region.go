// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package handlers

import (
	"context"
	"fmt"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type RegionService struct {
	pb.UnimplementedRegionServiceServer
	cloudAccountClient pb.CloudAccountServiceClient
	repo               *RegionRepository
}

func NewRegionService(repo *RegionRepository, cloudAccountClient pb.CloudAccountServiceClient) (*RegionService, error) {
	if cloudAccountClient == nil {
		return nil, fmt.Errorf("cloudaccount client is required")
	}

	return &RegionService{
		cloudAccountClient: cloudAccountClient,
		repo:               repo,
	}, nil
}

func (srv *RegionService) Add(ctx context.Context, request *pb.AddRegionRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionService.Add").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := request.Validate(); err != nil {
		logger.Error(err, "validation failed for region request")
		return nil, status.Error(codes.InvalidArgument, "invalid region request provided")
	}

	if request.IsDefault {
		if request.Type != "open" {
			logger.Error(fmt.Errorf("region type is not 'open'"), "invalid region type for default")
			return nil, status.Error(codes.InvalidArgument, "only regions with type 'open' can be set as default")
		}

		if err := srv.repo.ResetDefaultRegion(ctx, request.Name); err != nil {
			logger.Error(err, "error resetting default region flag in database")
			return nil, status.Errorf(codes.Internal, "resetting default region flag failed")
		}
	}

	if err := srv.repo.AddRegion(ctx, request); err != nil {
		logger.Error(err, "error creating region record in database")
		return nil, status.Errorf(codes.Internal, "region insertion failed")
	}

	return &emptypb.Empty{}, nil
}

func (srv *RegionService) AdminRead(ctx context.Context, filter *pb.RegionFilter) (*pb.RegionResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionService.AdminRead").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := filter.Validate(); err != nil {
		logger.Error(err, "validation failed for region filter")
		return nil, status.Error(codes.InvalidArgument, "invalid region filter provided")
	}

	regions, err := srv.repo.GetRegions(ctx, filter)
	if err != nil {
		logger.Error(err, "error reading region records from database")
		return nil, status.Errorf(codes.Internal, "region read failed")
	}

	return &pb.RegionResponse{Regions: regions}, nil
}

func (srv *RegionService) UserRead(ctx context.Context, filter *pb.RegionUserFilter) (*pb.RegionUserResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionService.UserRead").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := filter.Validate(); err != nil {
		logger.Error(err, "validation failed for region user filter")
		return nil, status.Error(codes.InvalidArgument, "invalid region user filter provided")
	}

	if filter.CloudaccountId == "" {
		logger.Error(nil, "cloudaccount_id must be provided")
		return nil, status.Error(codes.InvalidArgument, "cloudaccount_id must be provided")
	}

	cloudAccountId := filter.GetCloudaccountId()
	_, err := srv.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil {
		logger.Error(err, "invalid cloud account")
		return &pb.RegionUserResponse{}, fmt.Errorf("invalid cloud account: %v", err)
	}

	regions, err := srv.repo.GetUserRegions(ctx, filter)
	if err != nil {
		logger.Error(err, "error reading region records from database")
		return nil, status.Errorf(codes.Internal, "region read failed")
	}

	return &pb.RegionUserResponse{Regions: regions}, nil
}

func (srv *RegionService) Update(ctx context.Context, request *pb.UpdateRegionRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionService.Update").WithValues("regionName", request.Name).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	if err := request.Validate(); err != nil {
		logger.Error(err, "validation failed for region request")
		return nil, status.Error(codes.InvalidArgument, "invalid region request provided")
	}

	// Fetch the current region details
	region, err := srv.repo.GetRegionByName(ctx, request.Name)
	if err != nil {
		logger.Error(err, "error fetching region type from database")
		return nil, status.Errorf(codes.Internal, "fetching region type failed")
	}

	// Check if the region is being set to default and validate its type
	if request.IsDefault != nil && *request.IsDefault {
		// Allow the change if the region's current type is "controlled" and the request is changing it to "open"
		if region.Type == "controlled" && request.Type != nil && *request.Type == "open" {
			if err := srv.repo.ResetDefaultRegion(ctx, request.Name); err != nil {
				logger.Error(err, "error resetting default region flag in database")
				return nil, status.Errorf(codes.Internal, "resetting default region flag failed")
			}
		} else if (region.Type == "controlled") || (request.Type != nil && *request.Type == "controlled") {
			logger.Error(fmt.Errorf("region type is 'controlled'"), "invalid region type for default")
			return nil, status.Error(codes.InvalidArgument, "regions with type 'controlled' cannot be set as default")
		}
		if err := srv.repo.ResetDefaultRegion(ctx, request.Name); err != nil {
			logger.Error(err, "error resetting default region flag in database")
			return nil, status.Errorf(codes.Internal, "resetting default region flag failed")
		}
	}

	// Check if the region is currently default and the update request is attempting to change its type to "controlled"
	if (region.IsDefault || (request.IsDefault != nil && *request.IsDefault)) && request.Type != nil && *request.Type == "controlled" {
		logger.Error(fmt.Errorf("cannot change default region to 'controlled'"), "invalid type change for default region")
		return nil, status.Error(codes.InvalidArgument, "default regions cannot be changed to type 'controlled'")
	}

	if err := srv.repo.UpdateRegion(ctx, request); err != nil {
		logger.Error(err, "error updating region record in database", "key", request.Name)
		return nil, status.Errorf(codes.Internal, "region update failed")
	}

	//TODO if region change from controlled to open, clean whitelisting? or add a filter to list be type in region access endpoint?

	return &emptypb.Empty{}, nil
}
func (srv *RegionService) Delete(ctx context.Context, request *pb.DeleteRegionRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionService.Delete").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if err := request.Validate(); err != nil {
		logger.Error(err, "validation failed for region request")
		return nil, status.Error(codes.InvalidArgument, "invalid region request provided")
	}

	if err := srv.repo.DeleteRegion(ctx, request.Name); err != nil {
		logger.Error(err, "error deleting region record in database", "key", request.Name)
		return nil, status.Errorf(codes.Internal, "region deletion failed")
	}

	return &emptypb.Empty{}, nil
}
