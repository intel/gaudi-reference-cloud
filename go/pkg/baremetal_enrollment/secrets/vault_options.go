// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package secrets

type VaultOption func(vault *Vault)

func VaultOptionRenewToken(renewToken bool) VaultOption {
	return func(vault *Vault) {
		vault.renewToken = renewToken
	}
}

func VaultOptionValidateClient(validate bool) VaultOption {
	return func(vault *Vault) {
		vault.validate = validate
	}
}
