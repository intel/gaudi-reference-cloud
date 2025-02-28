// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	"k8s.io/client-go/util/retry"
)

func (e *Deployer) DeleteVault(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("DeleteVault")
	log.Info("BEGIN")
	defer log.Info("END")

	if err := e.RunHelmfile(ctx,
		"destroy",
		"--file", "helmfile-vault.yaml",
		"--selector", "chart=vault",
	); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, e.KubectlBinary, "delete", "-n", "kube-system", "pvc", "data-vault-0")
	if err := util.RunCmd(ctx, cmd); err != nil {
		log.Error(err, "unable to delete data-vault-0")
	}

	return nil
}

func (e *Deployer) DeployVault(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("DeployVault")
	logger.Info("BEGIN")
	defer logger.Info("END")

	verboseLoggerCtx := log.IntoContext(ctx, logger.V(9))

	// Deploy Vault Helm release.
	if err := e.RunHelmfile(
		ctx,
		"apply",
		"--file", "helmfile-vault.yaml",
		"--selector", "chart=vault",
		"--skip-diff-on-install",
		"--wait",
	); err != nil {
		return err
	}

	if err := e.StartVaultPortForward(ctx); err != nil {
		return err
	}

	// Initialize (unseal) Vault Server if needed.
	// This also writes the Vault token file.
	cmd := exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "deployment/common/vault/vault-operator-init-with-cli.sh"))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "JQ="+e.JqBinary)
	cmd.Env = append(cmd.Env, "JWK_SOURCE_DIR="+e.JwkSourceDir)
	cmd.Env = append(cmd.Env, "KUBECONFIG="+e.KubeConfig)
	cmd.Env = append(cmd.Env, "KUBECTL="+e.KubectlBinary)
	cmd.Env = append(cmd.Env, "SECRETS_DIR="+e.SecretsDir)
	cmd.Env = append(cmd.Env, "VAULT="+e.VaultBinary)
	cmd.Env = append(cmd.Env, "VAULT_ADDR="+e.VaultAddr)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	// Read Vault token file.
	if err := e.ReadVaultToken(ctx); err != nil {
		return err
	}

	// Configure Vault (policies, roles, PKI).
	if e.Options.IncludeVaultConfigure {
		cmd = exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "deployment/common/vault/configure.sh"))
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "HELMFILE_DUMP="+e.HelmfileDumpYamlFile)
		cmd.Env = append(cmd.Env, "IDC_ENV="+e.IdcEnv)
		cmd.Env = append(cmd.Env, "JQ="+e.JqBinary)
		cmd.Env = append(cmd.Env, "SECRETS_DIR="+e.SecretsDir)
		cmd.Env = append(cmd.Env, "VAULT="+e.VaultBinary)
		cmd.Env = append(cmd.Env, "VAULT_ADDR="+e.VaultAddr)
		cmd.Env = append(cmd.Env, "VAULT_DRY_RUN=false")
		cmd.Env = append(cmd.Env, "VAULT_TOKEN="+e.VaultToken)
		cmd.Env = append(cmd.Env, "YQ="+e.YqBinary)
		if err := util.RunCmd(verboseLoggerCtx, cmd); err != nil {
			return err
		}
	}

	// Load secrets into Vault.
	if e.Options.IncludeVaultLoadSecrets {
		cmd = exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "deployment/common/vault/load-secrets.sh"))
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "HELMFILE_DUMP="+e.HelmfileDumpYamlFile)
		cmd.Env = append(cmd.Env, "IDC_ENV="+e.IdcEnv)
		cmd.Env = append(cmd.Env, "JQ="+e.JqBinary)
		cmd.Env = append(cmd.Env, "SECRETS_DIR="+e.SecretsDir)
		cmd.Env = append(cmd.Env, "VAULT="+e.VaultBinary)
		cmd.Env = append(cmd.Env, "VAULT_ADDR="+e.VaultAddr)
		cmd.Env = append(cmd.Env, "VAULT_TOKEN="+e.VaultToken)
		cmd.Env = append(cmd.Env, "YQ="+e.YqBinary)
		if err := util.RunCmd(verboseLoggerCtx, cmd); err != nil {
			return err
		}
	}

	// TODO: Wait for Vault Agent Injector to be running.

	return nil
}

// Read Vault token created by unseal.
func (e *Deployer) ReadVaultToken(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("ReadVaultToken")
	log.Info("BEGIN")
	defer log.Info("END")

	vaultTokenBytes, err := os.ReadFile(filepath.Join(e.SecretsDir, "VAULT_TOKEN"))
	if err != nil {
		return err
	}
	e.VaultToken = strings.TrimRight(string(vaultTokenBytes), "\n")
	log.V(0).Info("Vault initialized", "VaultToken", e.VaultToken)
	return nil
}

func (e *Deployer) StartVaultPortForward(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("StartVaultPortForward")
	localPort, err := e.StartPortForward(ctx, "kube-system", "service/vault", "http")
	if err != nil {
		return err
	}

	backoff := LinearBackoff(3*time.Minute, 1*time.Second)
	if err := retry.OnError(backoff, func(error) bool { return true }, func() error {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d", localPort))
		return err
	}); err != nil {
		return fmt.Errorf("checking status of Vault through port forward: %w", err)
	}

	e.VaultAddr = fmt.Sprintf("http://localhost:%d", localPort)
	log.Info("VaultAddr", "VaultAddr", e.VaultAddr)
	return nil
}

func (e *Deployer) CreateVaultPkiCert(ctx context.Context, commonName string) error {
	log := log.FromContext(ctx).WithName("CreatePkiCert")
	log.Info("BEGIN")
	defer log.Info("END")

	cmd := exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "hack/generate-vault-pki-cert.sh"))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "IDC_ENV="+e.IdcEnv)
	cmd.Env = append(cmd.Env, "JQ="+e.JqBinary)
	cmd.Env = append(cmd.Env, "SECRETS_DIR="+e.SecretsDir)
	cmd.Env = append(cmd.Env, "VAULT="+e.VaultBinary)
	cmd.Env = append(cmd.Env, "VAULT_ADDR="+e.VaultAddr)
	cmd.Env = append(cmd.Env, "VAULT_TOKEN="+e.VaultToken)
	cmd.Env = append(cmd.Env, "COMMON_NAME="+commonName)
	cmd.Env = append(cmd.Env, "CREATE_ROLE=1")
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}
	return nil
}
