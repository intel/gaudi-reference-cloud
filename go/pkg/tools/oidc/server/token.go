// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/zitadel/oidc/pkg/crypto"
	"github.com/zitadel/oidc/pkg/op"
)

var kArrayClaims = map[string]bool{"amr": true, "roles": true, "wids": true, "groups": true}

func getTokenParams(req *http.Request, claims map[string]any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	now := time.Now().UTC()
	// These claims may be overridden by query parameters
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	claims["exp"] = now.Add(1 * time.Hour).Unix()
	claims["enterpriseId"] = generateRandom8Digit()
	claims["countryCode"] = "US"

	// Send back the query arguments as claims
	for key, val := range req.Form {
		if kArrayClaims[key] {
			claims[key] = val
		} else if len(val) != 1 {
			return fmt.Errorf("found %d values for %s, only one may be specified", len(val), key)
		} else {
			claims[key] = val[0]
		}
	}
	return nil
}

func GetTokenHandler(provider op.OpenIDProvider) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		claims := map[string]any{}
		// Request can override this
		claims["iss"] = provider.Issuer()

		if err := getTokenParams(req, claims); err != nil {
			http.Error(resp, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}

		tok, err := crypto.Sign(claims, provider.Signer().Signer())
		if err != nil {
			http.Error(resp, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}

		io.WriteString(resp, tok)
	}
}

func generateRandom8Digit() string {
	rand.Seed(time.Now().UnixNano())
	min := 10000000
	max := 99999999
	randomID := min + rand.Intn(max-min)
	return strconv.Itoa(randomID)
}
