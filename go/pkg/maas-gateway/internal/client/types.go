// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"google.golang.org/grpc"
)

type Connector interface {
	GetIdcConnection(ctx context.Context, addr string) (*grpc.ClientConn, error)
	GetIksConnection(ctx context.Context, addr string) (*grpc.ClientConn, error)
}
