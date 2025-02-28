// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package root

import (
	sshcmd "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/idccli/cmd/ssh"
	"github.com/spf13/cobra"
)

func NewCmdRoot() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "idccli <command> <subcommand> [flags]",
		Short:   "IDC CLI",
		Long:    `Idccli is a tool for managing idc tasks via command line .`,
		Version: "1.0",
	}

	cmd.AddGroup(&cobra.Group{
		ID:    "core",
		Title: "Core commands",
	})

	// Child commands
	cmd.AddCommand(sshcmd.NewCmdSsh())

	return cmd, nil
}
