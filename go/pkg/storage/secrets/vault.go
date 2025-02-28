// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package secrets

//go:generate mockgen -destination ../mocks/secret_manager.go -package mocks github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/secrets SecretManager
//go:generate mockgen -destination ./vault_mocks.go -package secrets github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/secrets VaultHelper

import (
	"context"
	"fmt"
	"os"
	"time"

	vault "github.com/hashicorp/vault/api"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/secrets"
)

const (
	// Mount point Env Vars
	vaultEnvApproleRoleIDFile        = "VAULT_APPROLE_ROLE_ID_FILE"
	vaultEnvApproleSecretIDFile      = "VAULT_APPROLE_SECRET_ID_FILE"
	vaultEnvSecretEngineSTORAGE      = "VAULT_SECRET_ENGINE_STORAGE"
	vaultEnvSecretEngineControlplane = "VAULT_SECRET_ENGINE_CONTROLPLANE"

	// Vault Address
	VaultAddressEnvVar = "VAULT_ADDR"

	// Enrollment approle secretID and roleID
	secretsRoleIDFile   = "/vault/secrets/role-id"
	secretsSecretIDFile = "/vault/secrets/secret-id"

	// Mount points
	defaultSecretEngineControlPlane = "controlplane"
	defaultSecretEngineSTORAGE      = "storage"
)

type SecretManager interface {
	STORAGESecretAccessor
	ValidateVaultClientAccessor
	ApiServiceSecretAccessor
}

type STORAGESecretAccessor interface {
	GetStorageCredentials(ctx context.Context, secretPath string, renewToken bool) (string, string, string, string, error)
	PutStorageSecrets(ctx context.Context, secretsPath string, kv map[string]interface{}, renewToken bool) (*vault.KVSecret, error)
	DeleteStorageCredentials(ctx context.Context, secretPath string, renewToken bool) error
}

type ValidateVaultClientAccessor interface {
	ValidateVaultClient(ctx context.Context) error
}

type VaultHelper interface {
	getVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, renewToken bool) (*vault.KVSecret, error)
	putVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, kv map[string]interface{}, renewToken bool) (*vault.KVSecret, error)
	getVaultClient(ctx context.Context) error
	deleteVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, renewToken bool) error
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
	renewalTime  time.Time
	vaultHelper  VaultHelper
}

func NewVaultClient(ctx context.Context) *Vault {
	log := log.FromContext(ctx).WithName("NewVaultClient")
	log.Info("Creating Vault Connection")
	return &Vault{}
}

func fetchRoleID(roleIdPath string) (string, error) {
	b, err := os.ReadFile(roleIdPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s", err)
	}
	return string(b), nil
}

func (v *Vault) getRoleID() error {
	roleID, err := fetchRoleID("/vault/secrets/role-id")
	fmt.Print("default val")
	fmt.Print(secretsRoleIDFile)
	if roleID == "" || err != nil {
		return fmt.Errorf("no role ID was provided by file '%s'", vaultEnvApproleRoleIDFile)
	}
	v.roleID = roleID
	return nil
}

func (v *Vault) getSecretIDFile() error {
	secretIDFile := GetEnv(vaultEnvApproleSecretIDFile, secretsSecretIDFile)
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

func (v *Vault) getStorageVaultSecrets(ctx context.Context, secretsPath string, renewToken bool) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.getStorageVaultSecrets")
	log.Info("get Storage vault secrets")

	log.Info("parameters for review", logkeys.VaultSecretEngineStorage, vaultEnvSecretEngineSTORAGE, logkeys.DefaultSecretEngineStorage, defaultSecretEngineSTORAGE, logkeys.SecretsPath, secretsPath)

	return v.vaultHelper.getVaultSecrets(ctx, vaultEnvSecretEngineSTORAGE, defaultSecretEngineSTORAGE, secretsPath, renewToken)
}

func (v *Vault) deleteStorageVaultSecrets(ctx context.Context, secretsPath string, renewToken bool) error {
	log := log.FromContext(ctx).WithName("Vault.deleteStorageVaultSecrets")
	log.Info("delete Storage vault secrets")

	log.Info("parameters for delete review", logkeys.VaultSecretEngineStorage, vaultEnvSecretEngineSTORAGE, logkeys.DefaultSecretEngineStorage, defaultSecretEngineSTORAGE, logkeys.SecretsPath, secretsPath)

	return v.vaultHelper.deleteVaultSecrets(ctx, vaultEnvSecretEngineSTORAGE, defaultSecretEngineSTORAGE, secretsPath, renewToken)
}

func (v *Vault) GetStorageCredentials(ctx context.Context, secretPath string, renewToken bool) (string, string, string, string, error) {
	log := log.FromContext(ctx).WithName("Vault.GetStorageCredentials")
	log.Info("Requesting STORAGE default credentials")

	secret, err := v.getStorageVaultSecrets(ctx, secretPath, renewToken)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to retrieve storage secrets from vault: %v", err)
	}
	// data map can contain more than one key-value pair,
	// in this case we're just grabbing one of them
	value, ok := secret.Data["username"].(string)
	if !ok {
		return "", "", "", "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["username"], secret.Data["username"])
	}
	log.Info("STORAGE Default credentials", logkeys.UserName, value)
	namespaceUsername := value
	value, ok = secret.Data["password"].(string)
	if !ok {
		return "", "", "", "", fmt.Errorf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"])
	}
	namespacePassword := value
	userId, ok := secret.Data["userId"].(string)
	if !ok {
		userId = "" //  handle the missing userId case as empty
	}
	namespaceId, ok := secret.Data["namespaceId"].(string)
	if !ok {
		namespaceId = "" // or handle the missing namespaceId case as empty
	}
	return namespaceUsername, namespacePassword, userId, namespaceId, nil
}

func (v *Vault) DeleteStorageCredentials(ctx context.Context, secretPath string, renewToken bool) error {
	log := log.FromContext(ctx).WithName("Vault.DeleteStorageCredentials")
	log.Info("delete storage path default credentials")

	err := v.deleteStorageVaultSecrets(ctx, secretPath, renewToken)
	if err != nil {
		return fmt.Errorf("failed to delete storage secrets from vault: %v", err)
	}

	return nil
}

func (v *Vault) GetControlPlaneSecrets(ctx context.Context, secretsPath string, renewToken bool) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.GetControlPlaneVaultSecrets")
	log.Info("get Controlplane vault secrets")

	return v.vaultHelper.getVaultSecrets(ctx, vaultEnvSecretEngineControlplane, defaultSecretEngineControlPlane, secretsPath, renewToken)
}

func (v *Vault) getVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, renewToken bool) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.getVaultSecret")
	// renew/recreate vault token
	if renewToken {
		renewalStatus := v.renewVaultToken(ctx)
		if renewalStatus {
			log.Info("renewed vault token as part of getting vault secret", logkeys.RoleId, v.roleID)
		} else {
			log.Info("failed to renew vault token as part of getting vault secret", logkeys.RoleId, v.roleID)
			err := v.getVaultAuthInfo(ctx)
			if err != nil {
				return nil, fmt.Errorf("unable to get secret as token renewal failed: %w", err)
			}
		}
	}
	secretEngineMount := GetEnv(vaultEnvSecretEngineMount, defaultSecretEngineMount)
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
		log.Info("Secret has more than one version, find most current version", logkeys.SecretsPath, secretsPath, logkeys.Versions, versions)
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
		log.Info("Selecting Secret version", logkeys.SecretsPath, secretsPath, logkeys.Version, version)
		secret, err = v.client.KVv2(secretEngineMount).GetVersion(context.Background(), secretsPath, version)
		if err != nil {
			return nil, fmt.Errorf("unable to read secret version '%d': %w", version, err)
		}
	} else {
		log.Info("Selecting default Secret version", logkeys.SecretsPath, secretsPath)
		secret, err = v.client.KVv2(secretEngineMount).Get(context.Background(), secretsPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read secret: %w", err)
		}
	}
	return secret, nil
}

func (v *Vault) deleteVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, renewToken bool) error {
	log := log.FromContext(ctx).WithName("Vault.deleteVaultSecrets")

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
	secretEngineMount := GetEnv(vaultEnvSecretEngineMount, defaultSecretEngineMount)
	if secretEngineMount == "" {
		return fmt.Errorf("no secret engine mount point was provided by env '%s'", vaultEnvSecretEngineMount)
	}

	_, err := v.client.KVv2(secretEngineMount).Get(context.Background(), secretsPath)
	if err != nil {

		return fmt.Errorf("error checking if secret exists: %w", err)
	}
	// delete the secret from mount path for KV v2
	err = v.client.KVv2(secretEngineMount).Delete(context.Background(), secretsPath)
	if err != nil {
		return fmt.Errorf("unable to delete secret: %w", err)
	}

	return nil
}

func (v *Vault) PutStorageSecrets(ctx context.Context, secretsPath string, kv map[string]interface{}, renewToken bool) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.PutStorageSecrets")
	log.Info("write STORAGE vault secrets")

	return v.vaultHelper.putVaultSecrets(ctx, vaultEnvSecretEngineSTORAGE, defaultSecretEngineSTORAGE, secretsPath, kv, renewToken)
}

func (v *Vault) PutControlplaneVaultSecrets(ctx context.Context, secretsPath string, kv map[string]interface{}, renewToken bool) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.PutControlplaneVaultSecrets")
	log.Info("write storage controlplane vault secrets")

	return v.vaultHelper.putVaultSecrets(ctx, vaultEnvSecretEngineControlplane, defaultSecretEngineControlPlane, secretsPath, kv, renewToken)
}

func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func (v *Vault) putVaultSecrets(ctx context.Context, vaultEnvSecretEngineMount, defaultSecretEngineMount, secretsPath string, kv map[string]interface{}, renewToken bool) (*vault.KVSecret, error) {
	log := log.FromContext(ctx).WithName("Vault.putVaultSecrets")
	if renewToken {
		renewalStatus := v.renewVaultToken(ctx)
		if renewalStatus {
			log.Info("renewed vault token as part of creating vault secret", logkeys.RoleId, v.roleID)
		} else {
			log.Info("failed to renew vault token vault token as part of creating vault secret", logkeys.RoleId, v.roleID)
			log.Info("re-attempting to get vault auth token")
			err := v.getVaultAuthInfo(ctx)
			if err != nil {
				return nil, fmt.Errorf("unable to write secret as token renewal failed: %w", err)
			}
		}
	}

	secretEngineMount := GetEnv(vaultEnvSecretEngineMount, defaultSecretEngineMount)
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

func (v *Vault) GetEnrollBasicAuth(ctx context.Context, secretPath string, renewToken bool) (string, string, error) {
	log := log.FromContext(ctx).WithName("Vault.GetEnrollBasicAuth")
	log.Info("Requesting Enroll Basic Auth credentials")

	secret, err := v.GetControlPlaneSecrets(ctx, secretPath, renewToken)
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
