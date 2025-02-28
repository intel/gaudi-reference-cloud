// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package adminserver

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tlsutil"
	"github.com/pkg/errors"
	"net/http"
)

type HTTPServer struct {
	server *http.Server
	mux    *http.ServeMux
}

// New initializes a new HttpServer with a specified configuration
func New(addr string) *HTTPServer {
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    addr,
	}
	return &HTTPServer{
		server: server,
		mux:    mux,
	}
}

func createAuthMiddleware(username string, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RegisterHandler adds a new route and its associated handler to the server
func (hs *HTTPServer) RegisterHandler(route string, handler http.Handler) {
	hs.mux.Handle(route, handler)
}

func (hs *HTTPServer) RegisterHandlerWithBasicAuth(route string, handler http.Handler, username, password string) {
	hs.mux.Handle(route, createAuthMiddleware(username, password)(handler))
}

// Start begins running the HTTP server
func (hs *HTTPServer) Start(ctx context.Context) error {
	tlsConfig, err := tlsutil.NewTlsProvider().ServerTlsConfig(ctx)

	// Disallow TLS 1.0 and 1.1
	tlsConfig.MinVersion = tls.VersionTLS12

	// As per CT-35: https://readthedocs.intel.com/cryptoteam/crypto_bkms/tls.html#id3
	tlsConfig.CipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	}
	tlsConfig.ClientAuth = tls.NoClientCert

	if err != nil {
		return errors.Wrap(err, "failed to get server tls config")
	}

	hs.server.TLSConfig = tlsConfig
	return hs.server.ListenAndServeTLS("", "")
}

func (hs *HTTPServer) Shutdown(ctx context.Context) error {
	return hs.server.Shutdown(ctx)
}
