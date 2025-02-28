// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"crypto/sha256"
	"net/http"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	"github.com/gorilla/mux"
	"golang.org/x/text/language"

	"github.com/zitadel/oidc/pkg/op"
)

const (
	pathLoggedOut = "/logged-out"
)

type Storage interface {
	op.Storage
}

// SetupServer creates an OIDC server with Issuer=http://localhost:<port>
// Use one of the pre-made clients in storage/clients.go or register a new one.
func SetupServer(ctx context.Context, issuer string, storage Storage) *mux.Router {
	logger := log.FromContext(ctx)
	// this will allow us to use an issuer with http:// instead of https://
	if err := os.Setenv(op.OidcDevMode, "true"); err != nil {
		logger.Error(err, "Failed to set zitadel dev mode enviornment variable")
	}

	// the OpenID Provider requires a 32-byte key for (token) encryption
	// be sure to create a proper crypto random key and manage it securely!
	key := sha256.Sum256([]byte("test"))

	router := mux.NewRouter()

	// for simplicity, we provide a very small default page for users who have signed out
	router.HandleFunc(pathLoggedOut, func(w http.ResponseWriter, req *http.Request) {
		logger.Info("Log out function invoked.")
		_, err := w.Write([]byte("signed out successfully"))
		if err != nil {
			logger.Error(err, "error serving logged out page")
		}
	})

	// creation of the OpenIDProvider with the just created in-memory Storage
	provider, err := newOP(ctx, storage, issuer, key)
	if err != nil {
		logger.Error(err, "Error while creating OpenIDProvider with in-memory storage")
		os.Exit(1)
	}

	router.Handle("/token", GetTokenHandler(provider))

	// we register the http handler of the OP on the root, so that the discovery endpoint (/.well-known/openid-configuration)
	// is served on the correct path
	router.PathPrefix("/").Handler(provider.HttpHandler())

	return router
}

// newOP will create an OpenID Provider for localhost on a specified port with a given encryption key
// and a predefined default logout uri
// it will enable all options (see descriptions)
func newOP(ctx context.Context, storage op.Storage, issuer string, key [32]byte) (op.OpenIDProvider, error) {
	config := &op.Config{
		Issuer:    issuer,
		CryptoKey: key,

		// will be used if the end_session endpoint is called without a post_logout_redirect_uri
		DefaultLogoutRedirectURI: pathLoggedOut,

		// enables code_challenge_method S256 for PKCE (and therefore PKCE in general)
		CodeMethodS256: true,

		// enables additional client_id/client_secret authentication by form post (not only HTTP Basic Auth)
		AuthMethodPost: true,

		// enables additional authentication by using private_key_jwt
		AuthMethodPrivateKeyJWT: true,

		// enables refresh_token grant use
		GrantTypeRefreshToken: true,

		// enables use of the `request` Object parameter
		RequestObjectSupported: true,

		// this example has only static texts (in English), so we'll set the here accordingly
		SupportedUILocales: []language.Tag{language.English},
	}
	handler, err := op.NewOpenIDProvider(ctx, config, storage,
		// as an example on how to customize an endpoint this will change the authorization_endpoint from /authorize to /auth
		op.WithCustomAuthEndpoint(op.NewEndpoint("auth")),
	)
	if err != nil {
		return nil, err
	}
	return handler, nil
}
