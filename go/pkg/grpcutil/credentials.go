// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package grpcutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tlsutil"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// _test.go code sets UseTLS = false to disable TLS on localhost servers
var (
	UseTLS = tlsutil.NewTlsProvider().ServerTlsEnabled()
)

func GetClientCredentials(ctx context.Context) (credentials.TransportCredentials, error) {
	tlsProvider := tlsutil.NewTlsProvider()
	if UseTLS && tlsProvider.ClientTlsEnabled() {
		config, err := tlsProvider.ClientTlsConfig(ctx)
		if err != nil {
			return nil, err
		}
		return credentials.NewTLS(config), nil
	} else {
		return insecure.NewCredentials(), nil
	}
}

func GetServerCredentials(ctx context.Context) (credentials.TransportCredentials, error) {
	tlsProvider := tlsutil.NewTlsProvider()
	if UseTLS && tlsProvider.ServerTlsEnabled() {
		tlsConfig, err := tlsProvider.ServerTlsConfig(ctx)
		if err != nil {
			return nil, err
		}
		return credentials.NewTLS(tlsConfig), nil
	} else {
		return insecure.NewCredentials(), nil
	}
}

func addrIsLocalhost(addr string) bool {
	strs := strings.Split(addr, ":")
	if len(strs) < 1 {
		return false
	}
	return strs[0] == "127.0.0.1" || strs[0] == "localhost"
}

func createClient(ctx context.Context, addr string, clientOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	logger := log.FromContext(ctx).WithName("grpcutil.createClient")
	creds, err := GetClientCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't load TLS credentials for client: %w", err)
	}
	logger.Info("Appending client options", "options", clientOptions)
	clientOptions = append(clientOptions, grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))

	logger.Info("Creating client", "address", addr)
	return grpc.DialContext(ctx, addr, clientOptions...)
}

func NewClient(ctx context.Context, addr string, clientOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	if addrIsLocalhost(addr) {
		// Testing with embedded servers; we don't use TLS for this
		return grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	return createClient(ctx, addr, clientOptions...)
}

func NewServer(ctx context.Context, serverOptions ...grpc.ServerOption) (*grpc.Server, error) {
	creds, err := GetServerCredentials(ctx)
	if err != nil {
		return nil, err
	}
	serverOptions = append(serverOptions, grpc.Creds(creds))
	return grpc.NewServer(serverOptions...), nil
}
