// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
//
// The contents of this file are based upon existing files in go/pkg/baremetal_enrollment/secrets/
// and go/pkg/storage/secrets/. See IDCCOMP-2524 for more details.
package secrets

import (
	context "context"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
)

// manageTokenLifecycle is a blocking helper function that uses LifetimeWatcher
// instances to periodically renew the given secrets when they are close to
// their 'token_ttl' expiration times until one of the secrets is close to its
// 'token_max_ttl' lease expiration time.
func (v *Vault) manageTokenLifecycle(ctx context.Context) (error, bool) {
	log := log.FromContext(ctx).WithName("Vault.manageTokenLifecycle")
	log.Info("manage token lifecycle: start")
	defer log.Info("manage token lifecycle: end")

	renew := v.secret.Auth.Renewable
	if !renew {
		log.Info("token is not configured to be renewable. re-attempting login.", logkeys.RoleId, v.roleID)
		return nil, false
	}

	// auth token
	authTokenWatcher, err := v.client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret:    v.secret,
		Increment: v.secret.Auth.LeaseDuration,
	})

	if err != nil {
		return (fmt.Errorf("unable to initialize auth token lifetime watcher. roleID: %v, error: %v",
			v.roleID, err)), false
	}

	go authTokenWatcher.Start()
	defer authTokenWatcher.Stop()

	// monitor events from watchers
	for {
		select {
		case <-ctx.Done():
			log.Info("timeout while waiting to renew the token. re-attempting login", logkeys.RoleId, v.roleID)
			return nil, false

		// DoneCh will return if renewal fails, or if the remaining lease
		// duration is under a built-in threshold and either renewing is not
		// extending it or renewing is disabled.  In both cases, the caller
		// should attempt a re-read of the secret. Clients should check the
		// return value of the channel to see if renewal was successful.
		case err := <-authTokenWatcher.DoneCh():
			// Leases created by a token get revoked when the token is revoked.
			if err != nil {
				log.Info("failed to renew token", logkeys.RoleId, v.roleID)
			}
			log.Info("token can no longer be renewed", logkeys.RoleId, v.roleID)
			return nil, false

		// RenewCh is a channel that receives a message when a successful
		// renewal takes place and includes metadata about the renewal.
		case <-authTokenWatcher.RenewCh():
			log.Info("renewed auth token", logkeys.RoleId, v.roleID)
		}
	}
}
