// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Copyright © 2023 Intel Corporation
//

package secrets

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/approle"

	"github.com/muonsoft/validation/validate"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

var (
	VaultAddr    = ""
	RoleID       = ""
	SecretIDFile = ""
)

func NewVaultClient(ctx context.Context) (*vault.Client, error) {
	log := log.FromContext(ctx).WithName("Secrets.NewVaultClient")
	log.Info("NewClient Environment")

	// Read the Vault address from the environment
	VaultAddr = getEnv("VAULT_ADDR", VaultAddr)
	if VaultAddr == "" {
		return nil, fmt.Errorf("vault server address not set in env VAULT_ADDR")
	}

	if err := unblockServer(VaultAddr); err != nil {
		return nil, fmt.Errorf("unable to fix no_proxy: %w", err)
	}

	config := vault.DefaultConfig() // modify for more granular configuration
	config.Address = VaultAddr

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	return client, nil
}

func AppRoleLogin(ctx context.Context, client *vault.Client, roleID string, secretIDFile string) (*vault.Secret, error) {
	log := log.FromContext(ctx).WithName("Secrets.AppRoleLogin")
	log.Info("New AppRole Login", "roleID", roleID)

	// A combination of a Role ID and Secret ID is required to log in to Vault
	// with an AppRole.

	// Use the passed roleID
	RoleID = roleID
	if RoleID == "" {
		return nil, fmt.Errorf("no approle role ID file was provided to NewClient")
	}

	// The Secret ID is a value that needs to be protected, so instead of the
	// app having knowledge of the secret ID directly, we have a trusted orchestrator (https://learn.hashicorp.com/tutorials/vault/secure-introduction?in=vault/app-integration#trusted-orchestrator)
	// give the app access to a short-lived response-wrapping token (https://www.vaultproject.io/docs/concepts/response-wrapping).
	// Read more at: https://learn.hashicorp.com/tutorials/vault/approle-best-practices?in=vault/auth-methods#secretid-delivery-best-practices

	// Use the passed secretIDFile if it is set
	SecretIDFile = secretIDFile
	if SecretIDFile == "" {
		return nil, fmt.Errorf("no approle secret ID file was provided to NewClient")
	}
	if _, err := os.Stat(SecretIDFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("approle secret ID file %s not found", SecretIDFile)
		}
	}

	secretID := &auth.SecretID{FromFile: secretIDFile}

	appRoleAuth, err := auth.NewAppRoleAuth(
		roleID,
		secretID,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AppRole auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(ctx, appRoleAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AppRole auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return authInfo, nil
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func unblockServer(serverURL string) error {
	u, err := url.Parse(serverURL)
	if err != nil {
		return err
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}
	if host == "" {
		host = u.Host
	}

	err = cleanNoProxy(host, "no_proxy")
	if err != nil {
		return err
	}
	err = cleanNoProxy(host, "NO_PROXY")
	if err != nil {
		return err
	}

	return nil
}
func cleanNoProxy(host string, env string) error {

	removes := []string{host}

	// If the hostname is an IP address, only remove this single IP
	// it would be very unlikely the IP is in the no_proxy env
	if validate.IP(host) != nil {
		// Find all possible matches in the no_proxy variable
		// hostname.subdomain.domain.com = hostname.subdomain.domain.com, subdomain.domain.com, domain.com, com
		h := host
		for len(h) > 0 {
			parts := strings.SplitN(h, ".", 2)
			if len(parts) > 1 {
				h = parts[1]
				removes = append(removes, h)
			} else {
				h = ""
			}
		}
	}

	// WE NEED A PROXY server Inside of Intel because the Vault severs
	// are running in AWS but have a name which looks like it is inside of Intel
	proxies := strings.Split(os.Getenv(env), ",")

	var found bool
	cleanProxies := []string{}

	for _, proxy := range proxies {
		found = false
		for _, remove := range removes {
			if strings.EqualFold(remove, proxy) || strings.EqualFold("."+remove, proxy) {
				found = true
				break
			}
		}
		if !found {
			cleanProxies = append(cleanProxies, proxy)
		}
	}

	if osErr := os.Setenv(env, strings.Join(cleanProxies, ",")); osErr != nil {
		return fmt.Errorf("Failed to modify %s for vault host %s: %v\n", env, host, osErr)
	}

	return nil
}
