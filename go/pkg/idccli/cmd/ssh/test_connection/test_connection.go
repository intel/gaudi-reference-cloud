// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package testconnection

import (
	"fmt"
	"net"
	"os/exec"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type testConnectionOpts struct {
	proxyServer string
}

func NewCmdTestConnection() *cobra.Command {
	opts := testConnectionOpts{}
	var cmd = &cobra.Command{
		Use:   "test-proxy",
		Short: "Test given ssh proxy sever connectivity.",
		Example: heredoc.Doc(`
			$ idccli ssh test-proxy -p <proxysever>
			$ idccli ssh test-proxy --proxy-server <proxysever>
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTestConnection(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.proxyServer, "proxy-server", "p", "", "Test connectivty of the given proxy sever")
	return cmd
}

func runTestConnection(opts testConnectionOpts) error {
	proxyServer := opts.proxyServer
	if proxyServer == "" {
		return fmt.Errorf("provided proxy server is empty, please provide proper proxy server")
	}
	conn, err := net.DialTimeout("tcp", proxyServer+":22", 5*time.Second)
	if err != nil {
		fmt.Println("Failed to connect to SSH server:", err)
		return err
	}

	defer conn.Close()

	fmt.Println("Connection to given SSH proxy server is successful!")

	cmd := exec.Command("ssh-keyscan", "-H", proxyServer)

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Failed to get SSH keys:", err)
		return err
	}

	fmt.Println("SSH keys for", proxyServer+":")
	fmt.Println(string(output))
	return nil
}
