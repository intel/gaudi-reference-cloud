// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package common

import (
	"context"
	"os"
	"os/exec"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func RunCmd(ctx context.Context, cmd *exec.Cmd) error {
	log := log.FromContext(ctx).WithName("RunCmd")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Info("Running", "Args", cmd.Args, "Dir", cmd.Dir, "Env", cmd.Env)
	err := cmd.Run()
	log.Info("Completed", "Args", cmd.Args, "err", err)
	return err
}
