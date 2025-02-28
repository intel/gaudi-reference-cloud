// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"google.golang.org/grpc"
)

type MockGrpcServiceConnector struct {
	connection *grpc.ClientConn
}

func NewMockGrpcServiceConnector(connection *grpc.ClientConn) *MockGrpcServiceConnector {
	return &MockGrpcServiceConnector{
		connection: connection,
	}
}

func (c *MockGrpcServiceConnector) GetIdcConnection(_ context.Context, _ string) (*grpc.ClientConn, error) {
	return c.connection, nil
}

func (c *MockGrpcServiceConnector) GetIksConnection(_ context.Context, _ string) (*grpc.ClientConn, error) {
	return c.connection, nil
}
