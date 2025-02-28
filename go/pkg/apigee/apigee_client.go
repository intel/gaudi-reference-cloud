// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package apigee

import (
	"context"
	"fmt"
	"net/http"

	"sync"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"gopkg.in/square/go-jose.v2/jwt"
)

// ICPClient with a cached token.
type ApigeeClient struct {
	mu       sync.RWMutex
	Client   *resty.Client
	token    atomic.Pointer[Token]
	url      string
	username string
	password string
}

type Token struct {
	tokenString string
	exp         int64
}

type getTokenResponse struct {
	Access_token string
}

func NewClient(url string, username string, password string) (*ApigeeClient, error) {
	cli := resty.New().SetTimeout(1 * time.Minute)
	return &ApigeeClient{Client: cli, url: url, username: username, password: password}, nil
}

func (apg *ApigeeClient) GetCurrentToken(ctx context.Context) (string, error) {
	if token := apg.token.Load(); token != nil && !isTokenExpired(*token) {
		return token.tokenString, nil
	}

	apg.mu.Lock()
	defer apg.mu.Unlock()
	// Double check
	if token := apg.token.Load(); token != nil && !isTokenExpired(*token) {
		return token.tokenString, nil
	}
	token, err := apg.obtainToken(ctx)
	if err != nil {
		return "", err
	}
	apg.token.Store(token)
	return token.tokenString, nil
}

// Fetch the token. On the dev setup the token is valid for 24 hours.
func (apg *ApigeeClient) obtainToken(ctx context.Context) (*Token, error) {
	logger := log.FromContext(ctx)
	tokenResp := &getTokenResponse{}
	logger.V(9).Info("Fetching Apigee access token")
	resp, err := apg.Client.R().
		SetResult(tokenResp).
		SetContentLength(true).
		SetFormData(map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     apg.username,
			"client_secret": apg.password,
		}).
		Post(apg.url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Apigee token: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		// TODO : Unmarshal the error to errorResponse
		return nil, fmt.Errorf("Error in fetching Apigee Token, received response status %v", resp.Status())
	}
	logger.V(9).Info("Obtained Apigee access token")
	expClaim, err := extractExpClaim(ctx, tokenResp.Access_token)
	if err != nil {
		return nil, err
	}

	return &Token{
		tokenString: tokenResp.Access_token,
		exp:         expClaim,
	}, nil
}

// Helper method
func isTokenExpired(token Token) bool {
	// 5 min buffer for expiry
	return token.exp < time.Now().Add(time.Minute*5).Unix()
}

func extractExpClaim(ctx context.Context, token string) (int64, error) {
	log := log.FromContext(ctx)
	if token == "" {
		return 0, fmt.Errorf("Failed to fetch Apigee token")
	}
	var claims map[string]interface{} // generic map to store parsed token
	tokenParsed, _ := jwt.ParseSigned(token)
	if err := tokenParsed.UnsafeClaimsWithoutVerification(&claims); err != nil {
		log.Error(err, "Failed to read the Apigee token")
		return 0, err
	}
	expClaim := claims["exp"]
	if expClaim == nil {
		return 0, fmt.Errorf("exp claim missing in the Apigee token")
	}

	return int64(expClaim.(float64)), nil
}
