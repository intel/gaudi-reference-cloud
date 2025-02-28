// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package keygen

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/idccli/cmd/ssh/keygen"
	testconnection "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/idccli/cmd/ssh/test_connection"
	"github.com/spf13/cobra"
)

func NewCmdSsh() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ssh <command>",
		Short: "ssh commands.",
		Example: heredoc.Doc(`
			$ idccli ssh keygen --outputdir <output_dir>
			$ idccli ssh test-proxy --proxy-server <proxysever>
		`),
		GroupID: "core",
	}

	cmd.AddCommand(keygen.NewCmdKeyGen())
	cmd.AddCommand(testconnection.NewCmdTestConnection())
	return cmd
}
