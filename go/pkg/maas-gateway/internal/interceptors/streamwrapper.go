// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package interceptors

import (
	"context"
	"google.golang.org/grpc"
)

type StreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *StreamWrapper) Context() context.Context {
	return s.ctx
}
