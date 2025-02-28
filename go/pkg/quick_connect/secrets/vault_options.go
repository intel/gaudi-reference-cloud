// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
//
// The contents of this file are based upon existing files in go/pkg/baremetal_enrollment/secrets/
// and go/pkg/storage/secrets/. See IDCCOMP-2524 for more details.
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
