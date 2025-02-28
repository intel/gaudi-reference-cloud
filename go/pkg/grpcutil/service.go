// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package grpcutil

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Config interface {
	GetListenPort() uint16
	SetListenPort(port uint16)
}

type ListenConfig struct {
	ListenPort uint16 `koanf:"listenPort"`
}

func (config *ListenConfig) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *ListenConfig) SetListenPort(port uint16) {
	config.ListenPort = port
}

type Service[C Config] interface {
	Init(ctx context.Context, config C, resolver Resolver, grpcServer *grpc.Server) error
	// returns the dns name of the service
	Name() string
}

// Optional Done method for Service that requires cleanup
type ServiceCleanup interface {
	Done()
}

func ServiceDone[C Config](svc Service[C]) {
	if sc, ok := svc.(ServiceCleanup); ok {
		sc.Done()
	}
}

func LoadConfig[C Config](ctx context.Context, cfg C) error {
	log.BindFlags()
	configFile := ""
	flag.StringVar(&configFile, "config", "", "config file")
	flag.Parse()
	if configFile == "" {
		return fmt.Errorf("config flag can't be an empty string")
	}
	if err := conf.LoadConfigFile(ctx, configFile, cfg); err != nil {
		return fmt.Errorf("error loading config file (%s): %w", configFile, err)
	}
	return nil
}

func RunService[C Config](ctx context.Context, svc Service[C], config C) error {
	log.SetDefaultLogger()
	logger := log.FromContext(ctx)

	port := config.GetListenPort()
	if port == 0 {
		return fmt.Errorf("listenPort must be set in the config")
	}

	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(otelgrpc.UnaryServerInterceptor(), GrpcAuthzServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    120 * time.Second, // The time a connection is kept alive without any activity.
			Timeout: 20 * time.Second,  // Maximum time the server waits for activity before closing the connection.
		}),
	}

	grpcServer, err := NewServer(ctx, serverOptions...)
	if err != nil {
		return err
	}

	logger.Info("service start", "port", port)
	if err := svc.Init(ctx, config, &DnsResolver{}, grpcServer); err != nil {
		return err
	}
	defer ServiceDone(svc)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}

	return grpcServer.Serve(listener)
}

// Loads config file and runs the service
func Run[C Config](ctx context.Context, svc Service[C], config C) error {
	logger := log.FromContext(ctx)
	if err := LoadConfig(ctx, config); err != nil {
		logger.Error(err, "error loading config", "err", err)
		// Keep going. The caller may decide to provide reasonable defaults.
	}
	logger.Info("Run", "cfg", config)
	return RunService(ctx, svc, config)
}
