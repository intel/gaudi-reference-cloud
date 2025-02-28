// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenIssuer(t *testing.T) {
	jws, err := NewFakeAuthenticator()

	assert.NoError(t, err)

	tokenBytes, err := jws.CreateJWSWithClaims("user1", "org1", []string{"claim1", "claim2"})

	assert.NoError(t, err)

	_, err = jws.ValidateJWS(string(tokenBytes))

	assert.NoError(t, err)
}
