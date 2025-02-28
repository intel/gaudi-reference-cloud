// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	"context"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"google.golang.org/grpc/health/grpc_health_v1"
	"time"

	"google.golang.org/grpc"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewHealthCommand() *cobra.Command {
	var configPath, endpoint string

	cmd := &cobra.Command{
		Use:   "health",
		Short: "checks the health state of the service",
		Long:  `checks the health state of the service on the specified endpoint`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetDefaultLogger()
			log := log.FromContext(cmd.Context()).WithValues("endpoint", endpoint)

			cfg := config.DispatcherConfig{}
			if err := conf.LoadConfigFile(cmd.Context(), configPath, &cfg); err != nil {
				return errors.Wrap(err, "failed to load config file")
			}
			log.Info("loaded configuration", "cfg", cfg)

			clientCtx, cancel := context.WithTimeout(cmd.Context(), time.Minute)
			defer cancel()

			dialOptions := []grpc.DialOption{}
			conn, err := grpcutil.NewClient(clientCtx, fmt.Sprintf("localhost:%d", cfg.ListenPort), dialOptions...)
			if err != nil {
				return errors.Wrap(err, "failed to connect to the service dispatcher")
			}

			client := grpc_health_v1.NewHealthClient(conn)
			response, err := client.Check(cmd.Context(), &grpc_health_v1.HealthCheckRequest{Service: endpoint})
			if err != nil {
				return errors.Wrap(err, "failed to call health service")
			}

			log.Info("got health response", "response", response)
			if response.Status != grpc_health_v1.HealthCheckResponse_SERVING {
				return errors.Errorf("Status=%s", response.Status)
			}
			return nil
		},
	}

	viper.AutomaticEnv()
	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "/config.yaml", "config file path")
	cmd.PersistentFlags().StringVar(&endpoint, "endpoint", "alive", "health endpoint [ready|alive]")

	return cmd
}
