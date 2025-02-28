// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/test/mock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
)

func NewDevCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Starts the Model as a Service Gateway in a Dev Mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.BindFlags()
			log.SetDefaultLogger()
			ctx := cmd.Context()

			logger := log.FromContext(cmd.Context()).WithName(gateway.ServiceName)

			//create config
			cfg := config.Config{}
			if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
				return errors.Wrapf(err, "couldn't load config file, %s", configFile)
			}
			logger.Info("loaded configuration", "config", cfg)

			mockServer := mock.NewGrpcServer()
			dispatcherServer := mock.NewDispatcherServer()
			productCatalogServer := mock.NewProductCatalogServer()
			usageRecordServer := mock.NewUsageRecordServer()
			healthServer := mock.NewHealthServer()

			grpcMockServer := mockServer.GetGrpcServer()
			pb.RegisterDispatcherServer(grpcMockServer, dispatcherServer)
			pb.RegisterUsageRecordServiceServer(grpcMockServer, usageRecordServer)
			pb.RegisterProductCatalogServiceServer(grpcMockServer, productCatalogServer)
			grpc_health_v1.RegisterHealthServer(grpcMockServer, healthServer)
			connection, err := mockServer.Run()
			if err != nil {
				return errors.Wrap(err, "couldn't create mock server")
			}

			mockConnector := client.NewMockGrpcServiceConnector(connection)

			metricSDK := metrics.NewPromMetrics(logger, gateway.ServiceName, gateway.MetricsPrefix)

			maasGateway := gateway.NewGateway(&cfg, logger, mockConnector, metricSDK)

			//create port listener for the gateway
			listener, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.ListenPort))
			if err != nil {
				return errors.Wrapf(err, "couldn't create listener for port %v", cfg.ListenPort)
			}

			err = maasGateway.Run(ctx, listener)
			if err != nil {
				return errors.Wrap(err, "couldn't start mock gateway")
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&configFile, FlagConfig, "c", viper.GetString(FlagConfig), "The application will load its configuration from this file.")

	return cmd
}
