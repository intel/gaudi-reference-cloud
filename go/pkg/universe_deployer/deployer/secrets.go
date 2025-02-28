// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

func (e *Deployer) MakeSecrets(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "deployment/common/vault/make-secrets.sh"))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HELMFILE_DUMP="+e.HelmfileDumpYamlFile)
	cmd.Env = append(cmd.Env, "IDC_ENV="+e.IdcEnv)
	cmd.Env = append(cmd.Env, "SECRETS_DIR="+e.SecretsDir)
	cmd.Env = append(cmd.Env, "YQ="+e.YqBinary)
	return util.RunCmd(ctx, cmd)
}

func (e *Deployer) ReadSecret(ctx context.Context, secretFile string) (string, error) {
	secretPath := filepath.Join(e.SecretsDir, secretFile)
	fileBytes, err := os.ReadFile(secretPath)
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

func (e *Deployer) ReadSecrets(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("ReadSecrets")

	gitPassword, err := e.ReadSecret(ctx, "gitea_admin_password")
	if err != nil {
		return err
	}
	log.V(0).Info("gitPassword", "gitPassword", gitPassword)
	e.GitPassword = gitPassword
	return nil
}
