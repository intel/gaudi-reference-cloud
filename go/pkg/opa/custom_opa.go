// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package opa

import (
	"context"
	"flag"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	builtin "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/opa/builtin"
	tradecheck "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tradecheck/tradecheckintel"
	"github.com/open-policy-agent/opa-envoy-plugin/plugin"
	"github.com/open-policy-agent/opa/cmd"
	"github.com/open-policy-agent/opa/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	// AWS Cognito client for auth token generation, and required by global grpc-proxy
	cognitoEnabled, _ = strconv.ParseBool(os.Getenv("IDC_COGNITO_ENABLED"))
	cognitoURL, _     = url.Parse(os.Getenv("IDC_COGNITO_ENDPOINT"))
)

func resolve(ctx context.Context, name string) string {
	resolver := grpcutil.DnsResolver{}
	addr, err := resolver.Resolve(ctx, name)
	if err != nil {
		_, logger, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OPA.resolve").Start()
		logger.Error(err, "unable to resolve %s service", name)
		os.Exit(1)
	}
	return addr
}

func newClient(ctx context.Context, addr string) *grpc.ClientConn {

	var conn *grpc.ClientConn
	var err error

	_, logger, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("OPA.newClient").Start()
	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}

	if cognitoEnabled {
		// create the cognitoClient to access AWS Cognito
		cognitoClient, err := authutil.NewCognitoClient(&authutil.CognitoConfig{
			URL:     cognitoURL,
			Timeout: 1 * time.Minute,
		})
		if err != nil {
			logger.Error(err, "unable to read AWS Cognito credentials", "addr", addr)
			os.Exit(1)
		}

		// prefetch the access token to access global: cloudaccount svc
		token, err := cognitoClient.GetGlobalAuthToken(ctx)
		if err != nil {
			logger.Error(err, "unable to get AWS Cognito token", "addr", addr)
			os.Exit(1)
		}
		logger.V(9).Info("Prefetched Cognito Token: ", "cognitoToken", token)
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(authutil.NewCognitoAuth(ctx, cognitoClient)))
		conn, err = grpcutil.NewClient(ctx, addr, dialOptions...)
		if err != nil {
			logger.Error(err, "not able to connect to gRPC service using grpcutil.NewClient", "addr", addr)
			os.Exit(1)
		}
	} else {
		conn, err = grpcutil.NewClient(ctx, addr, dialOptions...)
		if err != nil {
			logger.Error(err, "not able to connect to gRPC service using grpcutil.NewClient", "addr", addr)
			os.Exit(1)
		}
	}

	return conn
}

func RunService(ctx context.Context) {
	log.BindFlags()
	flag.Parse()
	log.SetDefaultLogger()
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("OPA.RunService").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("Starting Custom OPA")

	logger.Info("Initializing Custom OPA")
	cloudAccountAddr := os.Getenv("CLOUDACCOUNT_ADDR")
	if cloudAccountAddr == "" {
		cloudAccountAddr = resolve(ctx, "cloudaccount")
	}
	cloudAccountConn := newClient(ctx, cloudAccountAddr)
	defer cloudAccountConn.Close()
	cloudAccountBuiltIn := builtin.NewCloudAccountBuiltIn(cloudAccountConn)

	productCatalogAddr := os.Getenv("PRODUCTCATALOG_ADDR")
	if productCatalogAddr == "" {
		productCatalogAddr = resolve(ctx, "productcatalog")
	}
	productCatalogConn := newClient(ctx, productCatalogAddr)
	defer productCatalogConn.Close()
	productCatalogBuiltIn := builtin.NewProductCatalogBuiltIn(productCatalogConn)

	authzEnabled, err := strconv.ParseBool(os.Getenv("AUTHZ_ENABLED"))
	if err != nil {
		logger.Error(err, "There was an error parsing AUTHZ_ENABLED env var")
	}
	var authzBuiltIn *builtin.AuthzBuiltIn
	var authzAddr string
	if authzEnabled {
		authzAddr = os.Getenv("AUTHZ_ADDR")
		if authzAddr == "" {
			authzAddr = resolve(ctx, "authz")
		}
		authzConn := newClient(ctx, authzAddr)
		defer authzConn.Close()
		authzBuiltIn = builtin.NewAuthzBuiltIn(authzConn)
	}

	gtsUsername := os.Getenv("usernameFile")
	gtsPassword := os.Getenv("passwordFile")
	gtsTokenUrl := os.Getenv("gts_get_token_url")
	createOrderUrl := os.Getenv("gts_create_order_url")
	logger.Info("GTS Configuration", "userName", gtsUsername, "tokenUrl", gtsTokenUrl, "orderUrl", createOrderUrl)
	cfg, err := tradecheck.CreateConfig(gtsUsername, gtsPassword, gtsTokenUrl, "", createOrderUrl, "")
	if err != nil {
		logger.Error(err, "Failed to create a GTS client config")
		os.Exit(1)
	}
	gtsClient, err := tradecheck.NewClient(cfg)
	if err != nil {
		logger.Error(err, "Failed to create a GTS client")
		os.Exit(1)
	}
	gtsCheckBuiltIn := builtin.NewGTSCheckBuiltIn(gtsClient)

	if authzEnabled {
		logger.Info("Configuration", "cloudAccountAddr", cloudAccountAddr, "productCatalogAddr", productCatalogAddr, "authz", authzAddr)
	} else {
		logger.Info("Configuration", "cloudAccountAddr", cloudAccountAddr, "productCatalogAddr", productCatalogAddr)
	}

	// Ensure that we can ping the Cloud Account service before starting.
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := cloudAccountBuiltIn.GetCloudAccountClient().Ping(pingCtx, &emptypb.Empty{}); err != nil {
		logger.Error(err, "unable to ping Cloud Account service")
		os.Exit(1)
	}
	logger.Info("Ping to Cloud Account service was successful")

	// Ensure that we can ping the Product Catalog service before starting.
	pingCtx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := productCatalogBuiltIn.GetProductCatalogClient().Ping(pingCtx, &emptypb.Empty{}); err != nil {
		logger.Error(err, "unable to ping Product Catalog service")
		os.Exit(1)
	}
	logger.Info("ping to Product Catalog service was successful")

	if authzEnabled {
		// Ensure that we can ping the Authz service before starting.
		pingCtx, cancel = context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if _, err := authzBuiltIn.GetAuthzClient().Ping(pingCtx, &emptypb.Empty{}); err != nil {
			logger.Error(err, "unable to ping Authz service")
			os.Exit(1)
		}
		logger.Info("Ping to Authz service was successful")
	}

	runtime.RegisterPlugin("envoy.ext_authz.grpc", plugin.Factory{}) // for backwards compatibility
	runtime.RegisterPlugin(plugin.PluginName, plugin.Factory{})

	cloudAccountBuiltIn.Register()
	productCatalogBuiltIn.Register()
	authzBuiltIn.Register()
	gtsCheckBuiltIn.Register()

	logger.Info("service running")

	// allow passing zap-log configuration OPA args with default values
	cmd.RootCommand.PersistentFlags().String("zap-encoder", "console", "")
	cmd.RootCommand.PersistentFlags().String("zap-log-level", "debug", "")
	cmd.RootCommand.PersistentFlags().String("zap-stacktrace-level", "error", "")
	if err := cmd.RootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
