// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package secrets

import (
	context "context"
	"fmt"
	"sync"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

var (
	// Token renewal configuration
	RenewalTimeout = 90 * time.Second
	RenewPeriod    = 5 * time.Minute
)

// - renew/recreate vault token.
// - It returns immediately if token is renewed
// - wait up to renewalTimeout to recreate the token
func (v *Vault) renewVaultToken(ctx context.Context) bool {
	ctx, cancelContextFunc := context.WithTimeout(ctx, RenewalTimeout)
	log := log.FromContext(ctx).WithName("Vault.renewVaultToken")
	defer cancelContextFunc()
	var wg sync.WaitGroup
	wg.Add(1)
	renewalStatus := make(chan bool, 1)
	go func(chan bool) {
		// renew token
		err, status := v.tokenRenewer(ctx)
		if err != nil {
			log.Error(err, "failed to renew the token", "roleID", v.roleID)
		} else {
			renewalStatus <- status
			log.Info("renewal status", "status", status)
			wg.Done()
		}
	}(renewalStatus)
	defer func() {
		wg.Wait()
	}()

	return <-renewalStatus
}

func (v *Vault) tokenRenewer(ctx context.Context) (error, bool) {
	log := log.FromContext(ctx).WithName("Vault.vaultTokenRenewer")
	log.Info("vault token renewal: start")
	defer log.Info("vault token renewal: end")

	// get tokenTTL
	tokenTTL, err := v.secret.TokenTTL()
	if err != nil {
		// continue with the token renewal if failed to get the tokenTTL
		log.Info("failed to get the token TTL from secret")
	}
	// skip renewal if time.Since(h.renewalTime) < (tokenTTL-RenewPeriod)
	if tokenTTL != 0 && !v.renewalTime.IsZero() && time.Since(v.renewalTime) < (tokenTTL-RenewPeriod) {
		log.Info("valid token. skipping renewal", "roleID", v.roleID, "renewal time",
			v.renewalTime, "tokenTTL", tokenTTL)
		return nil, true
	}
	// token renewal/recreation
	for {
		// retry token renewal/recreate when err != nil
		err, status := v.manageTokenLifecycle(ctx)
		if status {
			log.Info("renewed vault token", "roleID", v.roleID)
		} else {
			log.Info("failed to renew vault token", "roleID", v.roleID)
		}
		if err == nil {
			return nil, status
		}
	}
}

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
		log.Info("token is not configured to be renewable. re-attempting login.", "roleID", v.roleID)
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
			log.Info("timeout while waiting to renew the token. re-attempting login", "roleID", v.roleID)
			return nil, false

		// DoneCh will return if renewal fails, or if the remaining lease
		// duration is under a built-in threshold and either renewing is not
		// extending it or renewing is disabled.  In both cases, the caller
		// should attempt a re-read of the secret. Clients should check the
		// return value of the channel to see if renewal was successful.
		case err := <-authTokenWatcher.DoneCh():
			// Leases created by a token get revoked when the token is revoked.
			if err != nil {
				log.Info("failed to renew token. re-attempting login.", "roleID", v.roleID)
			}
			log.Info("token can no longer be renewed. re-attempting login.", "roleID", v.roleID)
			return nil, false

		// RenewCh is a channel that receives a message when a successful
		// renewal takes place and includes metadata about the renewal.
		case renewal := <-authTokenWatcher.RenewCh():
			log.Info("renewed auth token", "roleID", v.roleID)
			fmt.Printf("renewed auth token: %v", renewal)
			// update renewal time
			if !renewal.RenewedAt.IsZero() {
				log.Info("renewed token", "roleID", v.roleID, "renewedAt", renewal.RenewedAt)
				v.renewalTime = renewal.RenewedAt
			}
			return nil, true
		}
	}
}
