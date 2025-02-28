// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package api

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/alron/ginlogr"
	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quick_connect/secrets"
)

func NewRouter(ctx context.Context, cfg *Config, tlsClientConfig *tls.Config, vaultClient *secrets.Vault) (*gin.Engine, error) {
	router := gin.New()

	// Configure logging.
	log := log.FromContext(ctx)
	utc := true
	router.Use(ginlogr.Ginlogr(log, time.RFC3339, utc))
	printStackTrace := true
	router.Use(ginlogr.RecoveryWithLogr(log, time.RFC3339, utc, printStackTrace))

	// Configure tracing.
	router.Use(otelgin.Middleware("quick_connect"))

	err := AddHealthzRoutes(router)
	if err != nil {
		return nil, err
	}

	proxy, err := NewQuickConnectProxy(ctx, cfg, tlsClientConfig, vaultClient)
	if err != nil {
		return nil, err
	}
	if err = proxy.AddRoutes(router); err != nil {
		return nil, err
	}

	return router, nil
}
