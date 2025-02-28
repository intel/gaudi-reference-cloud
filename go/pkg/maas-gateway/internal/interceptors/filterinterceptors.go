// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package interceptors

import (
	"context"
	"google.golang.org/grpc"
)

// GrpcMethodMatcher is a function that takes a gRPC full method name as input
// and returns a boolean indicating whether the method matches a specific criteria.
// It is used to selectively apply interceptors to specific gRPC methods.
type GrpcMethodMatcher func(fullMethod string) bool

func FilterStreamServerInterceptor(interceptor grpc.StreamServerInterceptor, matcher GrpcMethodMatcher) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if matcher(info.FullMethod) {
			return handler(srv, ss)
		}
		return interceptor(srv, ss, info, handler)
	}
}

func FilterUnaryServerInterceptor(interceptor grpc.UnaryServerInterceptor, matcher GrpcMethodMatcher) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if matcher(info.FullMethod) {
			return handler(ctx, req)
		}
		return interceptor(ctx, req, info, handler)
	}
}
