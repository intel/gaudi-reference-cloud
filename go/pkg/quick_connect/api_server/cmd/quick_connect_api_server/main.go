// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quick_connect/api_server/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quick_connect/secrets"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tlsutil"
	"golang.org/x/sync/errgroup"
)

// Entry point for quick connect service
func main() {
	ctx := context.Background()

	// Parse command line.
	var configFile string
	flag.StringVar(&configFile, "config", "",
		"The application will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	log.BindFlags()
	flag.Parse()

	// Initialize logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	err := func() error {
		log.Info("main", "configFile", configFile)

		var cfg api.Config
		if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
			return err
		}
		log.Info("main", "config", cfg)

		tlsClientConfig, err := tlsutil.NewTlsProvider().ClientTlsConfig(ctx)
		if err != nil {
			return err
		}

		vaultClient, err := secrets.NewVaultClient(ctx,
			secrets.VaultOptionValidateClient(true),
			secrets.VaultOptionRenewToken(true),
		)
		if err != nil {
			return err
		}
		// Do an initial login to prevent a race between RenewToken and router.Run needing
		// to use the vault token.
		err = vaultClient.Login(ctx)
		if err != nil {
			return err
		}

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		gin.SetMode(gin.ReleaseMode)
		// Note: when run in DebugMode, gin logs the following:
		//   [GIN-debug] [WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.
		//   Please check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.
		// However we do trust the downstream Envoy and Nginx proxies so this warning
		// can be safely ignored. Additionally, we don't make any decisions based on the
		// client IP (i.e. the X-Forwarded-For header that is being warned as unsafe).

		group, ctx := errgroup.WithContext(ctx)
		group.Go(func() error {
			// Assume any errors from RenewToken are transient. To avoid shutting down the
			// router on a transient Vault error just sleep and retry.
			for {
				err := vaultClient.RenewToken(ctx)
				if err != nil {
					log.Error(err, "RenewToken failed")
				}
				log.Info("RenewToken exited, sleeping and retrying")
				time.Sleep(5 * time.Second)
			}
		})
		group.Go(func() error {
			router, err := api.NewRouter(ctx, &cfg, tlsClientConfig, vaultClient)
			if err != nil {
				return err
			}
			// This service runs behind Envoy and therefore listens only on localhost.
			// Envoy will listen on INADDR_ANY and proxy the connection here.
			addr := fmt.Sprintf("127.0.0.1:%d", cfg.ListenPort)
			return router.Run(addr)
		})
		return group.Wait()
	}()
	if err != nil {
		log.Error(err, "fatal error")
		os.Exit(1)
	}
}
