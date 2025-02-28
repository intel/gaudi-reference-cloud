// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/adminserver"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	FlagEnvironment = "environment"
)

func NewServerCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts the kafka dispatcher server",
		Long:  `Starts the kafka dispatcher server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			g, ctx := errgroup.WithContext(ctx)

			log.SetDefaultLogger()
			logger := log.FromContext(ctx)

			// Load configuration
			cfg := config.DispatcherConfig{}
			if err := conf.LoadConfigFile(ctx, configPath, &cfg); err != nil {
				return errors.Wrap(err, "failed to load config file")
			}
			logger.Info("loaded configuration", "cfg", cfg)

			// Create HTTP server
			adminServer := adminserver.New(fmt.Sprintf("%s:%d", os.Getenv("POD_IP"), cfg.MetricsPort))

			//create metrics sdk and register http handler
			metricsSDK := metrics.NewPromMetrics(logger, cfg, server.ServiceName)
			logger.Info("registering prometheus metrics")
			adminServer.RegisterHandler("/metrics", promhttp.Handler())

			// Create dispatcher
			dispatcher, err := server.New(ctx, cfg, metricsSDK, logger)
			if err != nil {
				return errors.Wrap(err, "failed to create server")
			}

			// Start HTTP server
			g.Go(func() error {
				return startAdminServer(logger, adminServer)
			})

			// Start the main server
			g.Go(func() error {
				return startDispatcher(ctx, cfg, dispatcher)
			})

			// Setup graceful shutdown
			g.Go(func() error {
				return setupShutdown(ctx, logger, adminServer, dispatcher)
			})

			// Wait for all goroutines to complete or for an error to occur
			if err := g.Wait(); err != nil {
				logger.Error(err, "server stopped with error")
				return err
			}

			return nil
		},
	}

	viper.AutomaticEnv()
	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "pkg/infaas-dispatcher/config/dev.yaml", "config file path")
	// cmd.PersistentFlags().StringVarP(&env, FlagEnvironment, "e", viper.GetString(FlagEnvironment), "runtime environment type")
	// _ = viper.BindPFlag(FlagEnvironment, cmd.Flag(FlagEnvironment)) //nolint:errcheck

	//zapopts := &zap.Options{
	//	Development: true,
	//	TimeEncoder: zapcore.RFC3339TimeEncoder,
	//}
	//zapopts.BindFlags(flag.CommandLine)
	//flag.Parse()

	return cmd
}

func startDispatcher(ctx context.Context, cfg config.DispatcherConfig, dispatcher *server.Dispatcher) error {
	bindAddr := fmt.Sprintf("0.0.0.0:%d", cfg.ListenPort)
	ln, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on server address %q", bindAddr)
	}

	return dispatcher.Run(ctx, ln)
}

func startAdminServer(logger logr.Logger, adminServer *adminserver.HTTPServer) error {
	logger.Info("starting admin server...")
	if err := adminServer.Start(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("http server closed")
			return nil
		}
		return errors.Wrap(err, "unable to start http server")
	}
	return nil
}

func setupShutdown(ctx context.Context, logger logr.Logger, adminServer *adminserver.HTTPServer, dispatcher *server.Dispatcher) error {
	<-ctx.Done()
	logger.Info("initiating graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := adminServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(err, "failed to shutdown http server")
		return errors.Wrap(err, "failed to shutdown http server")
	}

	dispatcher.Stop(false)
	return nil
}
