// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"context"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type UsageRecordServer struct {
	pb.UnimplementedUsageRecordServiceServer
	err     error
	timeout time.Duration
}

func NewUsageRecordServer() *UsageRecordServer {
	return &UsageRecordServer{}
}

func (m *UsageRecordServer) SetUsageRecordError(err error, timeout time.Duration) {
	m.err = err
	m.timeout = timeout
}

func (m *UsageRecordServer) CreateProductUsageRecord(_ context.Context, _ *pb.ProductUsageRecordCreate) (*emptypb.Empty, error) {
	// If timeout is set, simulate delay
	if m.timeout > 0 {
		time.Sleep(m.timeout)
	}

	// If error is set, return it
	if m.err != nil {
		return nil, m.err
	}

	// Default success case
	return &emptypb.Empty{}, nil
}

func (m *UsageRecordServer) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
