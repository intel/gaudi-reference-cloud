// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const (
	bufSize = 1024 * 1024
)

type Server struct {
	grpcServer *grpc.Server
	listener   *bufconn.Listener
}

func NewGrpcServer() *Server {
	return &Server{
		listener:   bufconn.Listen(bufSize),
		grpcServer: grpc.NewServer(),
	}
}

func (m *Server) GetGrpcServer() *grpc.Server {
	return m.grpcServer
}

func (m *Server) Run() (*grpc.ClientConn, error) {
	go func() {
		if err := m.grpcServer.Serve(m.listener); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	return grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(m.bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func (m *Server) bufDialer(context.Context, string) (net.Conn, error) {
	return m.listener.Dial()
}
