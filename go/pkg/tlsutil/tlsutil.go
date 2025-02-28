// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tlsutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type TlsProvider struct {
	// If true, TLS will be enabled on the server (HTTPS).
	serverTlsEnabled bool `koanf:"serverTlsEnabled"`
	// If true, TLS will be enabled for clients (HTTPS).
	clientTlsEnabled bool `koanf:"clientTlsEnabled"`
	// If true, GRPC Authz will check allowed_callers ACL
	grpcTlsAuthzEnabled        bool   `koanf:"grpcTlsAuthzEnabled"`
	grpcTlsAuthzAllowedCallers string `koanf:"grpcTlsAuthzAllowedCallers"`
	// If RequireAndVerifyClientCert, clients connecting to this server must provide a valid certificate.
	clientAuth         tls.ClientAuthType `koanf:"clientAuth"`
	insecureSkipVerify bool               `koanf:"insecureSkipVerify"`
	certFile           string             `koanf:"certFile"`
	keyFile            string             `koanf:"keyFile"`
	caCertFile         string             `koanf:"caCertFile"`
}

// Return a default TlsProvider.
// Parameters will come from the environment.
func NewTlsProvider() *TlsProvider {
	serverTlsEnabled := true
	if b, err := strconv.ParseBool(os.Getenv("IDC_SERVER_TLS_ENABLED")); err == nil {
		serverTlsEnabled = b
	}
	clientTlsEnabled := true
	if b, err := strconv.ParseBool(os.Getenv("IDC_CLIENT_TLS_ENABLED")); err == nil {
		clientTlsEnabled = b
	}
	clientAuth := tls.RequireAndVerifyClientCert
	if b, err := strconv.ParseBool(os.Getenv("IDC_REQUIRE_CLIENT_CERTIFICATE")); err == nil && !b {
		clientAuth = tls.NoClientCert
	}
	insecureSkipVerify := false
	if b, err := strconv.ParseBool(os.Getenv("IDC_INSECURE_SKIP_VERIFY")); err == nil {
		insecureSkipVerify = b
	}

	// grpc tls authz
	grpcTlsAuthzEnabled := false
	if b, err := strconv.ParseBool(os.Getenv("IDC_GRPC_TLS_AUTHZ_ENABLED")); err == nil {
		grpcTlsAuthzEnabled = b
	}

	// Initialize a slice to hold the parsed values
	var grpcTlsAuthzAllowedCallers string
	allowedCallersEnv := os.Getenv("IDC_GRPC_TLS_AUTHZ_ALLOWED_CALLERS")
	if len(allowedCallersEnv) != 0 {
		grpcTlsAuthzAllowedCallers = allowedCallersEnv
	}

	return &TlsProvider{
		serverTlsEnabled:           serverTlsEnabled,
		clientTlsEnabled:           clientTlsEnabled,
		clientAuth:                 clientAuth,
		insecureSkipVerify:         insecureSkipVerify,
		grpcTlsAuthzEnabled:        grpcTlsAuthzEnabled,
		grpcTlsAuthzAllowedCallers: grpcTlsAuthzAllowedCallers,
		certFile:                   GetenvOrDefault("IDC_TLS_CERT", "/vault/secrets/cert.pem"),
		keyFile:                    GetenvOrDefault("IDC_TLS_KEY", "/vault/secrets/cert.key"),
		caCertFile:                 GetenvOrDefault("IDC_TLS_CA", "/vault/secrets/ca.pem"),
	}
}

// Return a TlsProvider for testing.
func NewTestTlsProvider(certFile string, keyFile string, caCertFile string) *TlsProvider {
	return &TlsProvider{
		serverTlsEnabled:   true,
		clientTlsEnabled:   true,
		clientAuth:         tls.RequireAndVerifyClientCert,
		insecureSkipVerify: false,
		certFile:           certFile,
		keyFile:            keyFile,
		caCertFile:         caCertFile,
	}
}

func (p *TlsProvider) ServerTlsEnabled() bool {
	return p.serverTlsEnabled
}

func (p *TlsProvider) ClientTlsEnabled() bool {
	return p.clientTlsEnabled
}

func (p *TlsProvider) InsecureSkipVerify() bool {
	return p.insecureSkipVerify
}

func (p *TlsProvider) GrpcTlsAuthzEnabled() bool {
	return p.grpcTlsAuthzEnabled
}

func (p *TlsProvider) GrpcTlsAuthzAllowedCallers() string {
	return p.grpcTlsAuthzAllowedCallers
}

// Loads the server certificate from cert and key files.
// This can be used for tls.Config.GetCertificate functions.
func (p *TlsProvider) GetCertificate(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(p.certFile, p.keyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// Loads the client certificate from cert and key files.
// This can be used for tls.Config.GetClientCertificate functions.
func (p *TlsProvider) GetClientCertificate(requestInfo *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(p.certFile, p.keyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// Get the TLS config for a TLS server.
func (p *TlsProvider) ServerTlsConfig(ctx context.Context) (*tls.Config, error) {
	log := log.FromContext(ctx).WithName("tlsutil.TlsProvider.ServerTlsConfig")
	log.Info("Creating TLS server", "clientAuth", p.clientAuth, "certFile", p.certFile, "keyFile", p.keyFile, "caCertFile", p.caCertFile)
	certPool := x509.NewCertPool()
	if err := appendCertsFromPEMFile(certPool, p.caCertFile); err != nil {
		return nil, err
	}
	config := &tls.Config{
		ClientAuth: p.clientAuth,
		ClientCAs:  certPool,
		// GetCertificate will be called by the server whenever a client connects to the server.
		GetCertificate: p.GetCertificate,
	}
	return config, nil
}

// Get the TLS config for a TLS client.
func (p *TlsProvider) ClientTlsConfig(ctx context.Context) (*tls.Config, error) {
	log := log.FromContext(ctx).WithName("tlsutil.TlsProvider.ClientTlsConfig")
	log.Info("Creating TLS client", "insecureSkipVerify", p.insecureSkipVerify, "certFile", p.certFile, "keyFile", p.keyFile, "caCertFile", p.caCertFile)
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	if err := appendCertsFromPEMFile(certPool, p.caCertFile); err != nil {
		return nil, err
	}
	config := &tls.Config{
		RootCAs: certPool,
		// GetClientCertificate will be called when the server requests a client certificate.
		GetClientCertificate: p.GetClientCertificate,
	}
	return config, nil
}

func appendCertsFromPEMFile(certPool *x509.CertPool, caCertFile string) error {
	pemCA, err := os.ReadFile(caCertFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file %s: %w", caCertFile, err)
	}
	if !certPool.AppendCertsFromPEM(pemCA) {
		return fmt.Errorf("failed to add certificate from file %s", caCertFile)
	}
	return nil
}

// Get string value from an environment variable.
// If environment variable does not exist or is an empty string, return a default value.
func GetenvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	} else {
		return value
	}
}
