// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

import (
	"context"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StandardBillingOptionService struct {
	pb.UnimplementedBillingOptionServiceServer
}

func (standardBillingOptionService *StandardBillingOptionService) Read(ctx context.Context, filter *pb.BillingOptionFilter) (*pb.BillingOption, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("StandardBillingOptionService.Read").WithValues("Req", filter).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}
