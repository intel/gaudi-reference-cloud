// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"

	middleware "github.com/oapi-codegen/echo-middleware"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticator(t *testing.T) {
	e := echo.New()

	jws, err := NewFakeAuthenticator()

	assert.NoError(t, err)

	tokenBytes, err := jws.CreateJWSWithClaims("user1", "org1", []string{"one:r", "two:w"})

	assert.NoError(t, err)

	authenticator := NewAuthenticator(jws)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/cluster", nil)
	req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", string(tokenBytes)))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	//nolint:staticcheck
	ctx := context.WithValue(c.Request().Context(), middleware.EchoContextKey, c)

	input := openapi3filter.AuthenticationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: c.Request(),
		},
		Scopes: []string{"one:r", "two:w"},
	}

	err = authenticator(ctx, &input)

	assert.NoError(t, err)

	input = openapi3filter.AuthenticationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: c.Request(),
		},
		Scopes: []string{"one:r", "three:w"},
	}

	err = authenticator(ctx, &input)

	assert.Error(t, err)
}
