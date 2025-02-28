// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package secrets

//go:generate mockgen -destination ../mocks/secret_manager.go -package mocks github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets SecretManager
//go:generate mockgen -destination ./vault_mocks.go -package secrets github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets VaultHelper

import (
	"context"
	"fmt"
	"path"
	"time"

	vault "github.com/hashicorp/vault/api"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/secrets"
)

const (
	// Mount point Env Vars
	vaultEnvApproleRoleIDFile        = "VAULT_APPROLE_ROLE_ID_FILE"
	vaultEnvApproleSecretIDFile      = "VAULT_APPROLE_SECRET_ID_FILE"
	vaultEnvSecretEngineEnrollment   = "VAULT_SECRET_ENGINE_ENROLLMENT"
	vaultEnvSecretEngineBMC          = "VAULT_SECRET_ENGINE_BMC"
	vaultEnvSecretEngineControlplane = "VAULT_SECRET_ENGINE_CONTROLPLANE"

	// Vault Address
	VaultAddressEnvVar = "VAULT_ADDR"

	// Enrollment approle secretID and roleID
	secretsRoleIDFile   = "/vault/secrets/role-id"
	secretsSecretIDFile = "/vault/secrets/secret-id"

	// Mount points
	defaultSecretEngineEnrollment   = "enrollment"
	defaultSecretEngineControlPlane = "controlplane"
	defaultSecretEngineBMC          = "bmc"
)

var (
	// Secret paths
	DefaultSecretPathNetBox = path.Join(defaultSecretEngineControlPlane, defaultSecretEngineEnrollment, "netbox")
)

type SecretManager interface {
	BMCSecretAccessor
	ControlPlaneSecretAccessor
	NetBoxSecretAccessor
	DDISecretAccessor
	IpaImageSecretAccessor
	ValidateVaultClientAccessor
	ApiServiceSecretAccessor
}

type BMCSecretAccessor interface {
	GetBMCCredentials(ctx context.Context, secretPath string) (string, string, error)
	GetBMCBIOSPassword(ctx context.Context, secretPath string) (string, error)
	PutBMCSecrets(ctx context.Context, secretsPath string, kv map[string]interface{}) (*vault.KVSecret, error)
	DeleteBMCSecrets(ctx context.Context, secretsPath string) error
}

type DDISecretAccessor interface {
	GetDDICredentials(ctx context.Context, secretPath string) (string, string, error)
}

type ControlPlaneSecretAccessor interface {
	GetControlPlaneSecrets(ctx context.Context, secretsPath string) (*vault.KVSecret, error)
}

type NetBoxSecretAccessor interface {
	GetNetBoxAPIToken(ctx context.Context, secretPath string) (string, error)
}

type IpaImageSecretAccessor interface {
	GetIPAImageSSHPrivateKey(ctx context.Context, secretPath string) (string, error)
}

type ValidateVaultClientAccessor interface {
	ValidateVaultClient(ctx context.Context) error
}

type VaultHelper interface {
	getVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, renewToken bool) (*vault.KVSecret, error)
	putVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, kv map[string]interface{}, renewToken bool) (*vault.KVSecret, error)
	deleteVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, renewToken bool) error
	getVaultClient(ctx context.Context) error
	getVaultAuthInfo(ctx context.Context) error
}

type ApiServiceSecretAccessor interface {
	GetEnrollBasicAuth(ctx context.Context, secretPath string, renewToken bool) (string, string, error)
}

var _ SecretManager = (*Vault)(nil)

type Vault struct {
	client       *vault.Client
	secret       *vault.Secret
	roleID       string
	secretIDFile string
	validate     bool
	renewToken   bool
	renewalTime  time.Time
	vaultHelper  VaultHelper
}

func NewVaultClient(ctx context.Context, opts ...VaultOption) (*Vault, error) {
	log := log.FromContext(ctx).WithName("Vault")
	log.Info("Creating Vault Connection")

	client := &Vault{}
	for _, opt := range opts {
		opt(client)
	}

	if client.validate {
		if err := client.ValidateVaultClient(ctx); err != nil {
			return nil, fmt.Errorf("failed to validate vault client: %v", err)
		}
	}

	return client, nil
}

func (v *Vault) getRoleID() error {
	roleID, err := util.GetRoleID(util.GetEnv(vaultEnvApproleRoleIDFile, secretsRoleIDFile))
	if roleID == "" || err != nil {
		return fmt.Errorf("no role ID was provided by file '%s'", vaultEnvApproleRoleIDFile)
	}
	v.roleID = roleID
	return nil
}

func (v *Vault) getSecretIDFile() error {
	secretIDFile := util.GetEnv(vaultEnvApproleSecretIDFile, secretsSecretIDFile)
	if secretIDFile == "" {
		return fmt.Errorf("no secret ID file was provided by env '%s'", vaultEnvApproleSecretIDFile)
	}

	v.secretIDFile = secretIDFile
	return nil
}

func (v *Vault) getVaultClient(ctx context.Context) error {
	vaultClient, err := secrets.NewVaultClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize Vault client: %v roleID %s", err, v.roleID)
	}
	v.client = vaultClient
	return nil
}

func (v *Vault) getVaultAuthInfo(ctx context.Context) error {
	authIno, err := secrets.AppRoleLogin(ctx, v.client, v.roleID, v.secretIDFile)
	if err != nil {
		return fmt.Errorf("failed to get the Vault auth info: %v roleID %s", err, v.roleID)
	}
	v.secret = authIno
	if !v.renewalTime.IsZero() {
		v.renewalTime = time.Time{}
	}
	return nil
}

func (v *Vault) UpdateVaultHelper() {
	v.vaultHelper = v
}

func (v *Vault) getAuthInfo(ctx context.Context) error {
	return v.vaultHelper.getVaultAuthInfo(ctx)
}

func (v *Vault) getClient(ctx context.Context) error {
	return v.vaultHelper.getVaultClient(ctx)
}

func (v *Vault) ValidateVaultClient(ctx context.Context) error {
	//update vaultHelperInterface
	v.UpdateVaultHelper()

	err := v.getRoleID()
	if err != nil {
		return err
	}
	err = v.getSecretIDFile()
	if err != nil {
		return err
	}
	ctx, cancelContextFunc := context.WithCancel(ctx)
	defer cancelContextFunc()

	err = v.getClient(ctx)
	if err != nil {
		return err
	}
	err = v.getAuthInfo(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (v *Vault) getBMCVaultSecrets(ctx context.Context, secretsPath string) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.getBMCVaultSecrets")
	log.Info("get BMC vault secrets")

	log.Info("parameters for review", "vaultEnvSecretEngineBMC", vaultEnvSecretEngineBMC, "defaultSecretEngineBMC", defaultSecretEngineBMC, "secretsPath", secretsPath)

	return v.vaultHelper.getVaultSecrets(ctx, vaultEnvSecretEngineBMC, defaultSecretEngineBMC, secretsPath, v.renewToken)
}

func (v *Vault) GetBMCCredentials(ctx context.Context, secretPath string) (string, string, error) {
	log := log.FromContext(ctx).WithName("Vault.GetBMCCredentials")
	log.Info("Requesting BMC default credentials")

	secret, err := v.getBMCVaultSecrets(ctx, secretPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve secrets from vault: %v", err)
	}
	// data map can contain more than one key-value pair,
	// in this case we're just grabbing one of them
	value, ok := secret.Data["username"].(string)
	if !ok {
		return "", "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["username"], secret.Data["username"])
	}
	log.Info("BMC Default credentials", "username", value)
	bmcUsername := value
	value, ok = secret.Data["password"].(string)
	if !ok {
		return "", "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"])
	}
	bmcPassword := value
	return bmcUsername, bmcPassword, nil
}

func (v *Vault) GetBMCBIOSPassword(ctx context.Context, secretPath string) (string, error) {
	log := log.FromContext(ctx).WithName("Vault.GetBMCBIOSPassword")
	log.Info("Requesting BMC BIOS Password")

	secret, err := v.getBMCVaultSecrets(ctx, secretPath)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secrets from path '%s' in vault: %v", secretPath, err)
	}

	// data map can contain more than one key-value pair,
	// in this case we're just grabbing one of them
	value, ok := secret.Data["password"].(string)
	if !ok {
		return "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"])
	}

	biosPassword := value
	return biosPassword, nil
}

func (v *Vault) GetNetBoxAPIToken(ctx context.Context, secretsPath string) (string, error) {
	log := log.FromContext(ctx).WithName("Vault.GetNetBoxToken")
	log.Info("get NetBox token")

	// get vault secrets
	secret, err := v.GetControlPlaneSecrets(ctx, secretsPath)
	if err != nil {
		return "", fmt.Errorf("failed to get vault secrets with error:  %v", err)
	}

	// get netbox token
	netboxToken, ok := secret.Data["token"].(string)
	if !ok {
		return "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["token"], secret.Data["token"])
	}

	return netboxToken, nil
}

func (v *Vault) GetControlPlaneSecrets(ctx context.Context, secretsPath string) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.GetControlplaneVaultSecrets")
	log.Info("get Controlplane vault secrets")

	return v.vaultHelper.getVaultSecrets(ctx, vaultEnvSecretEngineControlplane, defaultSecretEngineControlPlane, secretsPath, v.renewToken)
}

func (v *Vault) getVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, renewToken bool) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.getVaultSecret")
	// renew/recreate vault token
	if renewToken {
		renewalStatus := v.renewVaultToken(ctx)
		if renewalStatus {
			log.Info("renewed vault token as part of getting vault secret", "roleID", v.roleID)
		} else {
			log.Info("failed to renew vault token as part of getting vault secret", "roleID", v.roleID)
			log.Info("re-attempting to get vault auth token")
			err := v.getVaultAuthInfo(ctx)
			if err != nil {
				return nil, fmt.Errorf("unable to get secret as token renewal failed: %w", err)
			}
		}
	}
	secretEngineMount := util.GetEnv(vaultEnvSecretEngineMount, defaultSecretEngineMount)
	if secretEngineMount == "" {
		return nil, fmt.Errorf("no secret engine mount point was provided by env '%s'", vaultEnvSecretEngineMount)
	}

	// First get the metadata so we know the current version of the secret
	versions, err := v.client.KVv2(secretEngineMount).GetVersionsAsList(context.Background(), secretsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get versions of secret: %w", err)
	}
	// Iterate through the list of versions and find the most recent one that has not been deleted
	if len(versions) > 1 {
		log.Info("Secret has more than one version, find most current version", "secretsPath", secretsPath, "versions", versions)
	}
	var version int = 0
	for i := len(versions) - 1; i >= 0; i-- {
		if versions[i].DeletionTime.IsZero() {
			version = versions[i].Version
			break
		}
	}

	// get secret from the Enrollment mount path for KV v2
	var secret *vault.KVSecret
	if version > 1 {
		log.Info("Selecting Secret version", "secretsPath", secretsPath, "version", version)
		secret, err = v.client.KVv2(secretEngineMount).GetVersion(context.Background(), secretsPath, version)
		if err != nil {
			return nil, fmt.Errorf("unable to read secret version '%d': %w", version, err)
		}
	} else {
		log.Info("Selecting default Secret version", "secretsPath", secretsPath)
		secret, err = v.client.KVv2(secretEngineMount).Get(context.Background(), secretsPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read secret: %w", err)
		}
	}
	return secret, nil
}

func (v *Vault) PutBMCSecrets(ctx context.Context, secretsPath string, kv map[string]interface{}) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.PutBMCVaultSecrets")
	log.Info("write BMC vault secrets")

	return v.vaultHelper.putVaultSecrets(ctx, vaultEnvSecretEngineBMC, defaultSecretEngineBMC, secretsPath, kv, v.renewToken)
}

func (v *Vault) PutControlplaneVaultSecrets(ctx context.Context, secretsPath string, kv map[string]interface{}, renewToken bool) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.PutControlplaneVaultSecrets")
	log.Info("write controlplane vault secrets")

	return v.vaultHelper.putVaultSecrets(ctx, vaultEnvSecretEngineControlplane, defaultSecretEngineControlPlane, secretsPath, kv, renewToken)
}

func (v *Vault) putVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, kv map[string]interface{}, renewToken bool) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.putVaultSecrets")
	if renewToken {
		renewalStatus := v.renewVaultToken(ctx)
		if renewalStatus {
			log.Info("renewed vault token as part of creating vault secret", "roleID", v.roleID)
		} else {
			log.Info("failed to renew vault token vault token as part of creating vault secret", "roleID", v.roleID)
			log.Info("re-attempting to get vault auth token")
			err := v.getVaultAuthInfo(ctx)
			if err != nil {
				return nil, fmt.Errorf("unable to write secret as token renewal failed: %w", err)
			}
		}
	}

	secretEngineMount := util.GetEnv(vaultEnvSecretEngineMount, defaultSecretEngineMount)
	if secretEngineMount == "" {
		return nil, fmt.Errorf("no secret engine mount point was provided by env '%s'", vaultEnvSecretEngineMount)
	}

	// put the secret into the mount path for KV v2
	secret, err := v.client.KVv2(secretEngineMount).Put(context.Background(), secretsPath, kv)
	if err != nil {
		return nil, fmt.Errorf("unable to write secret: %w", err)
	}

	return secret, nil
}

func (v *Vault) DeleteBMCSecrets(ctx context.Context, secretsPath string) error {
	log := log.FromContext(ctx).WithName("Vault.DeleteBMCVaultSecrets")
	log.Info("delete BMC vault secrets", "secretsPath", secretsPath)

	return v.vaultHelper.deleteVaultSecrets(ctx, vaultEnvSecretEngineBMC, defaultSecretEngineBMC, secretsPath, v.renewToken)
}

func (v *Vault) deleteVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, renewToken bool) error {
	log := log.FromContext(ctx).WithName("Vault.deleteVaultSecrets")
	// renew/recreate vault token
	if renewToken {
		renewalStatus := v.renewVaultToken(ctx)
		if renewalStatus {
			log.Info("renewed vault token as part of delete vault secret", "roleID", v.roleID)
		} else {
			log.Info("failed to renew vault token as part of delete vault secret", "roleID", v.roleID)
			log.Info("re-attempting to get vault auth token")
			err := v.getVaultAuthInfo(ctx)
			if err != nil {
				return fmt.Errorf("unable to delete secret as token renewal failed: %w", err)
			}
		}
	}
	secretEngineMount := util.GetEnv(vaultEnvSecretEngineMount, defaultSecretEngineMount)
	if secretEngineMount == "" {
		return fmt.Errorf("no secret engine mount point was provided by env '%s'", vaultEnvSecretEngineMount)
	}

	// delete the secret from mount path for KV v2
	err := v.client.KVv2(secretEngineMount).Delete(context.Background(), secretsPath)
	if err != nil {
		return fmt.Errorf("unable to delete secret: %w", err)
	}

	return nil
}

func (v *Vault) GetDDICredentials(ctx context.Context, secretPath string) (string, string, error) {
	log := log.FromContext(ctx).WithName("Vault.GetDDICredentials")
	log.Info("Requesting DDI credentials")

	secret, err := v.GetControlPlaneSecrets(ctx, secretPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve secrets from vault: %v", err)
	}

	value, ok := secret.Data["username"].(string)
	if !ok {
		return "", "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["username"], secret.Data["username"])
	}

	ddiUsername := value

	value, ok = secret.Data["password"].(string)
	if !ok {
		return "", "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"])
	}

	ddiPassword := value
	return ddiUsername, ddiPassword, nil
}

func (v *Vault) GetEnrollBasicAuth(ctx context.Context, secretPath string, renewToken bool) (string, string, error) {
	log := log.FromContext(ctx).WithName("Vault.GetEnrollBasicAuth")
	log.Info("Requesting Enroll Basic Auth credentials")

	secret, err := v.GetControlPlaneSecrets(ctx, secretPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve secrets from vault: %v", err)
	}

	value, ok := secret.Data["username"].(string)
	if !ok {
		return "", "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["username"], secret.Data["username"])
	}

	enrollUsername := value

	value, ok = secret.Data["password"].(string)
	if !ok {
		return "", "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"])
	}

	enrollPassword := value
	return enrollUsername, enrollPassword, nil
}

func (v *Vault) GetIPAImageSSHPrivateKey(ctx context.Context, secretPath string) (string, error) {
	log := log.FromContext(ctx).WithName("Vault.GetIPAImageSSHPrivateKey")
	log.Info("Requesting IPA Image credentials")

	// get vault secrets
	secret, err := v.GetControlPlaneSecrets(ctx, secretPath)
	if err != nil {
		return "", fmt.Errorf("failed to get vault secrets with error:  %v", err)
	}
	// get the private key
	privateKey, ok := secret.Data["privateKey"].(string)
	if !ok {
		// Do NOT log the value of the private key
		return "", fmt.Errorf("value type assertion failed for privateKey: %T", secret.Data["privateKey"])
	}
	if len(privateKey) == 0 {
		return "", fmt.Errorf("privateKey is empty")
	}

	return privateKey, nil
}
