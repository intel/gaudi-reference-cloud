// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AriaPaymentService struct {
	ariaController *AriaController
	pb.UnimplementedPaymentServiceServer
}

func (ariaPaymentService AriaPaymentService) AddPaymentPreProcessing(ctx context.Context, in *pb.PrePaymentRequest) (*pb.PrePaymentResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaPaymentService.AddPaymentPreProcessing").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	resp, err := ariaPaymentService.ariaController.AddPaymentPreProcessing(ctx, in.CloudAccountId)
	if err != nil {
		logger.Error(err, "aria payment service add payment pre processing api error", "in", in)
		return nil, status.Errorf(codes.Internal, "aria payment service add payment pre processing api error %v", err)
	}

	prePaymentResponse := &pb.PrePaymentResponse{
		SessionId:     resp,
		DirectPostUrl: config.Cfg.GetAriaSystemDirectPostUrl(),
		FunctionMode:  config.Cfg.GetAriaSystemFunctionMode(),
	}

	return prePaymentResponse, nil
}

func (ariaPaymentService AriaPaymentService) AddPaymentPostProcessing(ctx context.Context, in *pb.PostPaymentRequest) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaPaymentService.AddPaymentPostProcessing").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	err := ariaPaymentService.ariaController.AddPaymentPostProcessing(ctx, in.CloudAccountId, in.PrimaryPaymentMethodNo)
	if err != nil {
		logger.Error(err, "aria payment service add payment post processing api error")
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "aria payment service add payment post processing api error %v", err)
	}

	return &emptypb.Empty{}, nil
}
