// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authutil

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"gopkg.in/square/go-jose.v2/jwt"
)

// vault-agent injects the cert/key and ca files into the svc container
const (
	clientIdFile     = "/vault/secrets/client_id"
	clientSecretFile = "/vault/secrets/client_secret"
)

type CognitoAuth struct {
	client *CognitoClient
}

// PerRPCCredentials implementation
func (cta CognitoAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {

	log := log.FromContext(ctx).WithName("GetRequestMetadata")
	// Get the access token
	token, err := cta.client.GetGlobalAuthToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get AWS Cognito token: %w", err)
	}
	log.V(9).Info("Cognito Token: ", "cognitoToken", token)

	return map[string]string{
		"authorization": "Bearer " + token,
	}, nil
}

// PerRPCCredentials implementation
func (CognitoAuth) RequireTransportSecurity() bool {
	return true
}

func NewCognitoAuth(ctx context.Context, c *CognitoClient) CognitoAuth {
	logger := log.FromContext(ctx).WithName("authutil.NewCognitoAuth")
	logger.Info("New CognitoAuth created")
	return CognitoAuth{
		client: c,
	}
}

// CognitoClient with a cached token.
type CognitoClient struct {
	// A lock on this mutex is required to renew the token.
	mu     sync.Mutex
	client *resty.Client
	token  atomic.Pointer[Token]
	cfg    *CognitoConfig
}

type CognitoConfig struct {
	URL     *url.URL
	Timeout time.Duration
}

type Token struct {
	tokenString string
	// Expiration time as the number of seconds elapsed since January 1, 1970 UTC
	expiryTimeSeconds int64
}

type getTokenResponse struct {
	Access_token string
}

func NewCognitoClient(cfg *CognitoConfig) (*CognitoClient, error) {
	return &CognitoClient{
		client: resty.New().SetTimeout(cfg.Timeout),
		cfg:    cfg,
	}, nil
}

// AWS Cognito issued tokens are configured to expire in 1 hour
// Check if less than 5 min left in expiry, announce it to be refetched
func isTokenExpired(token Token) bool {
	return token.expiryTimeSeconds < time.Now().Add(time.Minute*5).Unix()
}

// Fetch the token. The token is valid for 1 hour.
func (cognito *CognitoClient) obtainToken(ctx context.Context) (*Token, error) {
	logger := log.FromContext(ctx)
	tokenResp := &getTokenResponse{}
	logger.V(9).Info("Fetching the token to access from AWS Cognito")

	// load client_id and client_secret
	client_id, err := os.ReadFile(clientIdFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read clientIdFile: %v", clientIdFile)
	}
	client_secret, err := os.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read clientSecretFile: %v", clientSecretFile)
	}

	url := cognito.cfg.URL.JoinPath("oauth2", "token")
	resp, err := cognito.client.R().
		SetBasicAuth(string(client_id), string(client_secret)).
		SetResult(tokenResp).
		SetContentLength(true).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type": "client_credentials",
			"client_id":  string(client_id),
		}).
		Post(url.String())
	if err != nil {
		logger.Error(err, "failed to send fetch token request to AWS Cognito")
		return nil, err
	}
	// Note: status in tokenResp is only used for logging perposes.
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch the token from AWS Cognito, received status %v - %v", resp.Status(), string(resp.Body()))
	}
	logger.V(9).Info("Obtained token from AWS Cognito")

	tokenParsed, err := jwt.ParseSigned(tokenResp.Access_token)
	if err != nil {
		return nil, err
	}

	var claims map[string]interface{} // generic map to store parsed token
	if err := tokenParsed.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, err
	}

	expClaim := claims["exp"]
	logger.V(9).Info("obtainToken: ", "token-expiry", time.Unix(int64(expClaim.(float64)), 0))

	return &Token{
		tokenString:       tokenResp.Access_token,
		expiryTimeSeconds: int64(expClaim.(float64)),
	}, nil
}

func (cognito *CognitoClient) GetGlobalAuthToken(ctx context.Context) (string, error) {
	if token := cognito.token.Load(); token != nil && !isTokenExpired(*token) {
		return token.tokenString, nil
	}

	cognito.mu.Lock()
	defer cognito.mu.Unlock()
	// Double check
	if token := cognito.token.Load(); token != nil && !isTokenExpired(*token) {
		return token.tokenString, nil
	}
	token, err := cognito.obtainToken(ctx)
	if err != nil {
		return "", err
	}
	cognito.token.Store(token)
	return token.tokenString, nil
}
