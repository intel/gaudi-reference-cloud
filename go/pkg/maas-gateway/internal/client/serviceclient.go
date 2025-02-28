// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"google.golang.org/grpc"
)

type ServiceClient[T any] struct {
	creator        Creator[T]
	client         T
	grpcConnection *grpc.ClientConn
}

type Creator[T any] func(cc grpc.ClientConnInterface) T

func NewServiceClient[T any](creator Creator[T]) *ServiceClient[T] {

	return &ServiceClient[T]{
		creator: creator,
	}
}

func (sc *ServiceClient[T]) Connect(grpcConnection *grpc.ClientConn) {
	sc.grpcConnection = grpcConnection
	sc.client = sc.creator(grpcConnection)
}

func (sc *ServiceClient[T]) Close() error {
	return sc.grpcConnection.Close()
}

func (sc *ServiceClient[T]) Client() T {
	return sc.client
}

func (sc *ServiceClient[T]) GrpcConnection() *grpc.ClientConn {
	return sc.grpcConnection
}
