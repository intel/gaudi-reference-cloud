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
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type RegionAccessService struct {
	pb.UnimplementedRegionAccessServiceServer
	repo               *RegionAccessRepository
	cloudAccountClient pb.CloudAccountServiceClient
}

func NewRegionAccessService(repo *RegionAccessRepository, cloudAccountClient pb.CloudAccountServiceClient) (*RegionAccessService, error) {
	if cloudAccountClient == nil {
		return nil, fmt.Errorf("cloudaccount client is required")
	}
	return &RegionAccessService{
		repo:               repo,
		cloudAccountClient: cloudAccountClient,
	}, nil
}

func (srv *RegionAccessService) ReadAccess(ctx context.Context, request *pb.GetRegionAccessRequest) (*pb.GetRegionAccessResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionAccessService.ReadAccess").Start()
	defer span.End()
	logger.V(9).Info("begin")
	defer logger.V(9).Info("end")

	accessList, err := srv.repo.ReadAccess(ctx, request.CloudaccountId, request.RegionAccessType)
	if err != nil {
		logger.Error(err, "failed to read cloud account region access")
		return nil, status.Errorf(codes.Internal, "failed to read cloud account region access")
	}

	return &pb.GetRegionAccessResponse{Acl: accessList}, nil
}

func (srv *RegionAccessService) AddAccess(ctx context.Context, request *pb.RegionAccessRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionAccessService.AddAccess").Start()
	defer span.End()
	logger.V(9).Info("begin")
	defer logger.V(9).Info("end")

	cloudAccountId := request.GetCloudaccountId()
	_, err := srv.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil {
		logger.Error(err, "invalid cloud account")
		return nil, fmt.Errorf("invalid cloud account")
	}

	err = srv.repo.AddAccess(ctx, request.CloudaccountId, request.RegionName)
	if err != nil {
		logger.Error(err, "error inserting account into db")
		return nil, status.Errorf(codes.Internal, "access record insertion failed")
	}

	logger.Info("database transaction completed")
	return &emptypb.Empty{}, nil
}

func (srv *RegionAccessService) RemoveAccess(ctx context.Context, request *pb.DeleteRegionAccessRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionAccessService.RemoveAccess").WithValues("cloudAccountId", request.CloudaccountId, "regionName", request.RegionName).Start()
	defer span.End()
	logger.V(9).Info("begin")
	defer logger.V(9).Info("end")

	logger.Info("delete request for:", "cloudaccountid", request.CloudaccountId, "regionName", request.RegionName)
	err := srv.repo.RemoveAccess(ctx, request.CloudaccountId, request.RegionName)
	if err != nil {
		logger.Error(err, "failed to delete cloud account with regionName invoked", "cloudAccountId", request.CloudaccountId, "regionName", request.RegionName)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (srv *RegionAccessService) CheckRegionAccess(ctx context.Context, request *pb.RegionAccessCheckRequest) (*wrapperspb.BoolValue, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionAccessService.CheckRegionAccess").Start()
	defer span.End()
	logger.V(9).Info("begin")
	defer logger.V(9).Info("end")

	logger.Info("checking region access", "cloudAccountId", request.CloudaccountId, "regionName", request.RegionName)

	exists, err := srv.repo.CheckRegionAccess(ctx, request.CloudaccountId, request.RegionName)
	if err != nil {
		logger.Error(err, "failed to check cloud account region access", "cloudAccountId", request.CloudaccountId, "regionName", request.RegionName)
		return nil, err
	}

	return &wrapperspb.BoolValue{Value: exists}, nil
}
