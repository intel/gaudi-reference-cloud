// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	"context"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"io"
	"time"

	"google.golang.org/grpc"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewGenCommand() *cobra.Command {
	var prompt string
	var configPath string

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "runs generate prompt",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetDefaultLogger()
			log := log.FromContext(cmd.Context())

			cfg := config.DispatcherConfig{}
			if err := conf.LoadConfigFile(cmd.Context(), configPath, &cfg); err != nil {
				return errors.Wrap(err, "failed to load config file")
			}
			log.Info("loaded configuration", "cfg", cfg)
			log.Info("generating prompt ...", "prompt", prompt)

			clientCtx, cancel := context.WithTimeout(cmd.Context(), time.Minute)
			defer cancel()

			dialOptions := []grpc.DialOption{}
			conn, err := grpcutil.NewClient(clientCtx, fmt.Sprintf("localhost:%d", cfg.ListenPort), dialOptions...)
			if err != nil {
				return errors.Wrap(err, "failed to connect to the service dispatcher")
			}

			client := pb.NewDispatcherClient(conn)
			req := &pb.DispatcherRequest{
				Model:   cfg.SupportedModels[0],
				Request: &pb.GenerateStreamRequest{Prompt: prompt},
			}

			respStream, err := client.GenerateStream(clientCtx, req)
			if err != nil {
				return errors.Wrap(err, "failed to call Generate()")
			}
			log.Info("got resp stream...")
			for {
				response, err := respStream.Recv()
				if err != nil {
					if err == io.EOF {
						fmt.Println("end of streams")
						return nil
					}
					return errors.Wrap(err, "failed to receive response chunk")
				}
				fmt.Printf("%+v", response)
			}
		},
	}

	viper.AutomaticEnv()
	cmd.PersistentFlags().StringVarP(&prompt, "prompt", "p", "what is the meaning of life?", "generate prompt")
	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "go/pkg/infaas-dispatcher/config/dev.yaml", "config file path")

	return cmd
}
