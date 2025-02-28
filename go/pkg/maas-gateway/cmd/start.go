// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	"context"
	stderrors "errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/adminserver"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"time"
)

const (
	FlagConfig = "config"
)

func NewStartCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the Model as a Service Gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.BindFlags()
			log.SetDefaultLogger()
			group, ctx := errgroup.WithContext(cmd.Context())

			logger := log.FromContext(ctx).WithName(gateway.ServiceName)

			//create config
			cfg := config.NewConfig(logger)
			if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
				return errors.Wrapf(err, "couldn't load config file, %s", configFile)
			}
			logger.Info("loaded configuration", "config", cfg)

			connector := client.NewGrpcServiceConnector(logger)
			// Create Admin server
			adminServer := adminserver.New(fmt.Sprintf(":%d", cfg.MetricsConfig.MetricsPort))

			// Create metrics
			metricSDK := metrics.NewPromMetrics(logger, gateway.ServiceName, gateway.MetricsPrefix)

			// Register prometheus handler
			if cfg.MetricsConfig.Enabled {
				username, password, err := cfg.MetricsConfig.GetCredentials()
				if err != nil {
					return errors.Wrap(err, "could not get credentials for Prometheus")
				}
				adminServer.RegisterHandlerWithBasicAuth("/metrics", promhttp.Handler(), username, password)
			}

			//Run the admin server
			group.Go(func() error {
				return runAdminServer(ctx, adminServer, logger)
			})

			//create port listener for the gateway
			listener, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.ListenPort))
			if err != nil {
				return errors.Wrapf(err, "couldn't create listener for port %v", cfg.ListenPort)
			}

			// Create & Run Gateway
			maasGateway := gateway.NewGateway(cfg, logger, connector, metricSDK)
			group.Go(func() error {
				return runMaasGateway(ctx, maasGateway, listener, logger)
			})

			group.Go(func() error {
				return shutdown(ctx, logger, adminServer, maasGateway)
			})

			// Wait for all goroutines to complete.
			if err := group.Wait(); err != nil {
				logger.Error(err, "service stopped with error")
				return err
			}
			logger.Info("service stopped")
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&configFile, FlagConfig, "c", viper.GetString(FlagConfig), "The application will load its configuration from this file.")

	return cmd
}

func runMaasGateway(ctx context.Context, maasGateway *gateway.Gateway, listener net.Listener, logger logr.Logger) error {
	logger.Info("starting gateway...")
	err := maasGateway.Run(ctx, listener)
	if err != nil {
		logger.Error(err, "couldn't start maasGateway")
		return errors.Wrap(err, "couldn't start maasGateway")
	}
	return nil
}

func runAdminServer(ctx context.Context, adminServer *adminserver.HTTPServer, logger logr.Logger) error {
	logger.Info("starting admin server...")
	if err := adminServer.Start(ctx); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("admin server closed")
			return nil
		}
		logger.Error(err, "unable to start admin server")
		return errors.Wrap(err, "unable to start admin server")
	}
	return nil
}

func shutdown(ctx context.Context, logger logr.Logger, adminServer *adminserver.HTTPServer, maasGateway *gateway.Gateway) error {
	<-ctx.Done()
	logger.Info("initiating graceful shutdown")

	const shutdownTimeout = 30 * time.Second
	var errs []error

	// Admin Server shutdown
	adminCtx, adminCtxCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer adminCtxCancel()
	if err := adminServer.Shutdown(adminCtx); err != nil {
		logger.Error(err, "failed to shutdown admin server")
		errs = append(errs, errors.Wrap(err, "failed to shutdown admin server"))
	}

	maasGateway.Shutdown()

	if len(errs) > 0 {
		return errors.Wrap(stderrors.Join(errs...), "failed to shutdown services")
	}

	return nil
}
