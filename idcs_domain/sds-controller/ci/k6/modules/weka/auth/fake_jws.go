// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Based on example https://github.com/deepmap/oapi-codegen/blob/master/examples/authenticated-api/echo/server/fake_jws.go
package auth

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/deepmap/oapi-codegen/v2/pkg/ecdsafile"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
)

// Fake hardcoded is an ECDSA private key for tests
const PrivateKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBuNFnPOfdf24+Q+CgIJtdVGju5+DV/LeG67sRQGAdLJoAoGCCqGSM49
AwEHoUQDQgAE1dxxJn14YzZe6bav7odkeCxIxyrVKIBpAT6bV8ALHx4YnGLkBpC+
TtLwFIfVOiZi7xwGSN9QIJPjgjUiU5o+Yg==
-----END EC PRIVATE KEY-----`

const KeyID = `fake-key-id`
const FakeIssuer = "fake-issuer"
const FakeAudience = "example-users"
const PermissionsClaim = "perm"
const OrganizationClaim = "org"

type FakeAuthenticator struct {
	PrivateKey *ecdsa.PrivateKey
	KeySet     jwk.Set
}

var _ JWSValidator = (*FakeAuthenticator)(nil)

// NewFakeAuthenticator creates an authenticator example which uses a hard coded
// ECDSA key to validate JWT's that it has signed itself.
func NewFakeAuthenticator() (*FakeAuthenticator, error) {
	privKey, err := ecdsafile.LoadEcdsaPrivateKey([]byte(PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("loading PEM private key: %w", err)
	}

	set := jwk.NewSet()
	pubKey := jwk.NewECDSAPublicKey()

	err = pubKey.FromRaw(&privKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("parsing jwk key: %w", err)
	}

	err = pubKey.Set(jwk.AlgorithmKey, jwa.ES256)
	if err != nil {
		return nil, fmt.Errorf("setting key algorithm: %w", err)
	}

	err = pubKey.Set(jwk.KeyIDKey, KeyID)
	if err != nil {
		return nil, fmt.Errorf("setting key ID: %w", err)
	}

	set.Add(pubKey)

	return &FakeAuthenticator{PrivateKey: privKey, KeySet: set}, nil
}

// ValidateJWS ensures that the critical JWT claims needed to ensure that we
// trust the JWT are present and with the correct values.
func (f *FakeAuthenticator) ValidateJWS(jwsString string) (jwt.Token, error) {
	return jwt.Parse([]byte(jwsString), jwt.WithKeySet(f.KeySet),
		jwt.WithAudience(FakeAudience), jwt.WithIssuer(FakeIssuer))
}

// SignToken takes a JWT and signs it with our private key, returning a JWS.
func (f *FakeAuthenticator) SignToken(t jwt.Token) ([]byte, error) {
	hdr := jws.NewHeaders()
	if err := hdr.Set(jws.AlgorithmKey, jwa.ES256); err != nil {
		return nil, fmt.Errorf("setting algorithm: %w", err)
	}
	if err := hdr.Set(jws.TypeKey, "JWT"); err != nil {
		return nil, fmt.Errorf("setting type: %w", err)
	}
	if err := hdr.Set(jws.KeyIDKey, KeyID); err != nil {
		return nil, fmt.Errorf("setting Key ID: %w", err)
	}
	return jwt.Sign(t, jwa.ES256, f.PrivateKey, jwt.WithHeaders(hdr))
}

// CreateJWSWithClaims is a helper function to create JWT's with the specified
// claims.
func (f *FakeAuthenticator) CreateJWSWithClaims(sub string, org string, claims []string) ([]byte, error) {
	t := jwt.New()
	err := t.Set(jwt.IssuerKey, FakeIssuer)
	if err != nil {
		return nil, fmt.Errorf("setting issuer: %w", err)
	}
	err = t.Set(jwt.AudienceKey, FakeAudience)
	if err != nil {
		return nil, fmt.Errorf("setting audience: %w", err)
	}
	err = t.Set(jwt.SubjectKey, sub)
	if err != nil {
		return nil, fmt.Errorf("setting subject: %w", err)
	}
	err = t.Set(PermissionsClaim, claims)
	if err != nil {
		return nil, fmt.Errorf("setting permissions: %w", err)
	}
	err = t.Set(OrganizationClaim, org)
	if err != nil {
		return nil, fmt.Errorf("setting organization: %w", err)
	}
	return f.SignToken(t)
}
