// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package grpcserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/configmgr"
	handlers "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/handlers"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"google.golang.org/grpc/credentials"

	"github.com/go-logr/logr"
	libovsdbclient "github.com/ovn-org/libovsdb/client"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	Logger logr.Logger
)

func loadTLSCredentials(verify bool, certFile, privKeyFile, caCertFile string) (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(certFile, privKeyFile)
	if err != nil {
		return nil, err
	}
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAnyClientCert,
		RootCAs:      caCertPool,
	}
	if verify {
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return credentials.NewTLS(config), nil
}

func Server(ctx context.Context, srvport *string, nbClient libovsdbclient.Client, dbClient *sql.DB, sslConfig *configmgr.SslConfig) {
	logger := Logger.WithName("Server")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *srvport))

	// Check for errors
	if err != nil {
		logger.Error(err, "Server: failed to listen")
		return
	}
	v := handlers.OvnBasedHandler{
		OvnClient: nbClient,
		DbClient:  dbClient,
	}

	// Instantiate the server. Apply SSL options if enabled
	var s *grpc.Server
	var sslEnabled bool
	sslEnabled, err = strconv.ParseBool(sslConfig.Enabled)
	if err != nil {
		logger.Error(err, "Cannot parse SSL enabled cfg for gRPC server.\n")
		return
	}
	if sslEnabled {
		var sslVerify bool
		sslVerify, err := strconv.ParseBool(sslConfig.Verify)
		if err != nil {
			logger.Error(err, "Cannot parse SSL verify cfg for gRPC server.\n")
			return
		}
		tlsConfig, err := loadTLSCredentials(sslVerify,
			sslConfig.Cert, sslConfig.Key, sslConfig.Ca)
		if err != nil {
			logger.Error(err, "Cannot create TLS config for gRPC server.\n")
			return
		}
		s = grpc.NewServer(grpc.Creds(tlsConfig))
	} else {
		s = grpc.NewServer()
		// TODO: for testing
		reflection.Register(s)
	}

	// Register server method (actions the server will do)
	v1.RegisterOvnnetServer(s, &v)

	logger.Info(fmt.Sprintf("server listening at %v", lis.Addr()))
	if err := s.Serve(lis); err != nil {
		logger.Error(err, "Server: failed to Serve")
		return
	}
	logger.Info("Server task exit")
}
