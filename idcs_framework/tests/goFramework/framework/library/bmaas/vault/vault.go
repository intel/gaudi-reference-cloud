package vault

import (
	"context"
	"fmt"
	"log"
	"os"

	vault "github.com/hashicorp/vault/api"
)

// This is the accompanying code for the Developer Quick Start.
// WARNING: Using root tokens is insecure and should never be done in production!

func GetVaultClient() (*vault.Client, error) {
	config := vault.DefaultConfig()
	config.Address = os.Getenv("VAULT_ADDR")

	// this should be perviouly set using the local/secret
	token := os.Getenv("VAULT_TOKEN")

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	// Authenticate
	client.SetToken(token)

	return client, nil
}

func GetNetBoxSecret(client *vault.Client, secretPath string) (string, error) {
	// Read a secret from the default mount path for KV v2 in dev mode, "secret"
	enrollSecret, err := client.KVv2("controlplane").Get(context.Background(), secretPath)
	if err != nil {
		return "", fmt.Errorf("unable to read enrollment secret: %v", err)
	}
	// fmt.Println(secret)
	token, ok := enrollSecret.Data["token"].(string)
	if !ok {
		return "", fmt.Errorf("value type assertion failed: %T %#v", enrollSecret.Data["token"], enrollSecret.Data["role_id"])
	}

	return token, nil
}

func GetApproleSecret(client *vault.Client) {
	// netbox region
	// From IDC_ENV=kind-jenkins
	region := "us-dev-1"
	secretPath := fmt.Sprintf("%s/baremetal/enrollment/approle", region)
	// Read a secret from the default mount path for KV v2 in dev mode, "secret"
	enrollSecret, err := client.KVv2("controlplane").Get(context.Background(), secretPath)
	if err != nil {
		log.Fatalf("unable to read enrollment secret: %v", err)
	}
	// fmt.Println(secret)
	role_id, ok := enrollSecret.Data["role_id"].(string)
	if !ok {
		log.Fatalf("value type assertion failed: %T %#v", enrollSecret.Data["role_id"], enrollSecret.Data["role_id"])
	}
	secret_id, ok := enrollSecret.Data["secret_id"].(string)
	if !ok {
		log.Fatalf("value type assertion failed: %T %#v", enrollSecret.Data["secret_id"], enrollSecret.Data["secret_id"])
	}

	fmt.Printf("Role ID: %v\t", role_id)
	fmt.Printf("Secret ID: %v\n", secret_id)
}
