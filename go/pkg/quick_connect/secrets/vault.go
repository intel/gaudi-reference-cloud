// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
//
// The contents of this file are based upon existing files in go/pkg/baremetal_enrollment/secrets/
// and go/pkg/storage/secrets/. See IDCCOMP-2524 for more details.
package secrets

//go:generate mockgen -destination ../mocks/secret_manager.go -package mocks github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quick_connect/secrets SecretManager
//go:generate mockgen -destination ./vault_mocks.go -package secrets github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quick_connect/secrets VaultHelper

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	vault "github.com/hashicorp/vault/api"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/secrets"
)

const (
	// Env Vars
	vaultEnvApproleRoleIDFile   = "VAULT_APPROLE_ROLE_ID_FILE"
	vaultEnvApproleSecretIDFile = "VAULT_APPROLE_SECRET_ID_FILE"
	vaultEnvPKIEngine           = "VAULT_PKI_ENGINE"
	vaultEnvPKIRole             = "VAULT_PKI_ROLE"

	// AppRole secretID and roleID
	secretsRoleIDFile   = "/vault/secrets/role-id"
	secretsSecretIDFile = "/vault/secrets/secret-id"

	// vault.Secret.Data keys
	CertificateSecret = "certificate"
	PrivateKeySecret  = "private_key"
	CAChainSecret     = "ca_chain"
)

type SecretManager interface {
	QuickConnectCertificateIssuer
	ValidateVaultClientAccessor
}

type QuickConnectCertificateIssuer interface {
	IssueQuickConnectCertificate(ctx context.Context, commonName string, ttl time.Duration) (*tls.Certificate, error)
}

type ValidateVaultClientAccessor interface {
	ValidateVaultClient(ctx context.Context) error
}

type VaultHelper interface {
	issueVaultCertificate(ctx context.Context, data map[string]interface{}, renewToken bool) (*vault.Secret, error)
	getVaultClient(ctx context.Context) error
	getVaultAuthInfo(ctx context.Context) error
}

var _ SecretManager = (*Vault)(nil)

type Vault struct {
	client       *vault.Client
	secret       *vault.Secret
	pkiRole      string
	roleID       string
	secretIDFile string
	validate     bool
	renewToken   bool
	pkiEngine    string
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

func (v *Vault) getPKIEngine() error {
	if value, ok := os.LookupEnv(vaultEnvPKIEngine); ok {
		v.pkiEngine = value
		return nil
	}
	return fmt.Errorf("no PKI engine was provided by env '%s'", vaultEnvPKIEngine)
}

func (v *Vault) getPKIRole() error {
	if value, ok := os.LookupEnv(vaultEnvPKIRole); ok {
		v.pkiRole = value
		return nil
	}
	return fmt.Errorf("no PKI role was provided by env '%s'", vaultEnvPKIRole)
}

func fetchRoleID(roleIdPath string) (string, error) {
	b, err := os.ReadFile(roleIdPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s", err)
	}
	return string(b), nil
}

func (v *Vault) getRoleID() error {
	roleIDFile := getEnv(vaultEnvApproleRoleIDFile, secretsRoleIDFile)
	if roleIDFile == "" {
		return fmt.Errorf("no secret ID file was provided by env '%s'", vaultEnvApproleSecretIDFile)
	}
	roleID, err := fetchRoleID(roleIDFile)
	if roleID == "" || err != nil {
		return fmt.Errorf("no role ID was provided by file '%s'", vaultEnvApproleRoleIDFile)
	}
	v.roleID = roleID
	return nil
}

func (v *Vault) getSecretIDFile() error {
	secretIDFile := getEnv(vaultEnvApproleSecretIDFile, secretsSecretIDFile)
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

	err := v.getPKIRole()
	if err != nil {
		return err
	}
	err = v.getRoleID()
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

	err = v.getPKIEngine()
	if err != nil {
		return err
	}
	return nil
}

func (v *Vault) Login(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Vault.Login")
	log.Info("login", logkeys.RoleId, v.roleID)
	return v.getVaultAuthInfo(ctx)
}

func (v *Vault) RenewToken(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Vault.RenewToken")
	for {
		err := v.Login(ctx)
		if err != nil {
			return err
		}
		err, status := v.manageTokenLifecycle(ctx)
		if err != nil {
			return err
		}
		if status {
			log.Info("token renewed", logkeys.RoleId, v.roleID)
		}
	}
}

func (v *Vault) IssueQuickConnectCertificate(ctx context.Context, commonName string, ttl time.Duration) (*tls.Certificate, error) {
	log := log.FromContext(ctx).WithName("Vault.IssueQuickConnectCertificate")
	log.Info("Issuing quick connect certificate", logkeys.CommonName, commonName, logkeys.TTL, ttl)

	data := map[string]interface{}{
		"common_name": commonName,
		"ttl":         ttl.String(),
	}
	secret, err := v.vaultHelper.issueVaultCertificate(ctx, data, v.renewToken)
	if err != nil {
		return nil, fmt.Errorf("failed to issue certificate in vault: %v", err)
	}

	var certs []string
	certificate, ok := secret.Data[CertificateSecret].(string)
	if !ok {
		return nil, fmt.Errorf("value type assertion failed: %T", secret.Data[CertificateSecret])
	}
	certs = append(certs, certificate)
	// Add all CAs not including the final root CA to the certificate chain
	caChain, ok := secret.Data[CAChainSecret].([]interface{})
	if !ok {
		return nil, fmt.Errorf("value type assertion failed: %T", secret.Data[CAChainSecret])
	}
	for i := 0; i < len(caChain)-1; i++ {
		ca, ok := caChain[i].(string)
		if !ok {
			return nil, fmt.Errorf("value type assertion failed: %T", ca)
		}
		certs = append(certs, ca)
	}
	log.V(9).Info("Certificate chain", "certs", certs)
	privateKey, ok := secret.Data[PrivateKeySecret].(string)
	if !ok {
		return nil, fmt.Errorf("value type assertion failed: %T", secret.Data[PrivateKeySecret])
	}

	tlsCert, err := tls.X509KeyPair([]byte(strings.Join(certs, "\n")), []byte(privateKey))
	if err != nil {
		return nil, err
	}
	return &tlsCert, nil
}

func (v *Vault) issueVaultCertificate(ctx context.Context, data map[string]interface{}, renewToken bool) (*vault.Secret, error) {
	path := fmt.Sprintf("%s/issue/%s", v.pkiEngine, v.pkiRole)
	secret, err := v.client.Logical().Write(path, data)
	if err != nil {
		return nil, fmt.Errorf("unable to issue certificate: %w", err)
	}
	return secret, nil
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
