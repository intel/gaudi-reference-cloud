// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	"github.com/friendsofgo/errors"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-agent/agent"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/spf13/cobra"
)

const (
	FlagEnvironment = "environment"
)

func NewServerCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts the infaas agent",
		Long:  `Starts the infaas agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetDefaultLogger()
			log := log.FromContext(cmd.Context())

			log.Info("initializing agent...")
			cfg := agent.AgentConfig{}
			if err := conf.LoadConfigFile(cmd.Context(), configPath, &cfg); err != nil {
				return errors.Wrap(err, "failed to load config file")
			}
			log.Info("loaded configuration", "cfg", cfg)

			a, err := agent.New(cmd.Context(), cfg, log)
			if err != nil {
				return errors.Wrap(err, "failed to create agent")
			}

			log.Info("starting agent...")
			err = a.Start(cmd.Context())
			if err != nil {
				return errors.Wrap(err, "failed to start agent")
			}

			return a.Close()
		},
	}

	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "pkg/infaas-agent/config/dev.yaml", "config file path")
	// cmd.PersistentFlags().StringVarP(&env, FlagEnvironment, "e", viper.GetString(FlagEnvironment), "runtime environment type")
	// _ = viper.BindPFlag(FlagEnvironment, cmd.Flag(FlagEnvironment)) //nolint:errcheck

	return cmd
}
