// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package grpcutil

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"google.golang.org/grpc/metadata"
)

var (
	EmailClaim    = "email"
	ClientIdClaim = "client_id"
)

var (
	ErrNoMetadata    = errors.New("no metadata present in context")
	ErrNoJWTToken    = errors.New("JWT token not found in request")
	ErrTokenParse    = errors.New("error while parsing the token")
	ErrInvalidClaims = errors.New("unable to read claims from the JWT token")
	ErrClaimNotFound = errors.New("claim not found in the JWT token")
)

// Extracts a claim from the JWT token present in the context
func ExtractClaimFromCtx(ctx context.Context, jwtRequired bool, claim string) (string, error) {
	// Start logging and tracing
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("grpcutil.ExtractClaimFromCtx").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	// Extract metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Info("Metadata is not passed along with the request")
		if jwtRequired {
			return "", ErrNoMetadata
		}
		return "", nil
	}

	// Retrieve authorization header
	autHeader := md.Get("authorization")
	if len(autHeader) == 0 {
		log.Info("JWT token is not passed along with the request")
		if jwtRequired {
			return "", ErrNoJWTToken
		}
		return "", nil
	}

	// Parse and extract email from JWT token
	return ParseClaimFromJWT(ctx, autHeader[0], claim)
}

// Extracts an enterpriseId from the JWT token present in the context
func ExtractEnterpriseIDAndCountryCodefromCtx(ctx context.Context, jwtRequired bool) (string, string, error) {
	// Start logging and tracing
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("grpcutil.ExtractEnterpriseIDfromCtx").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	// Extract metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Info("Metadata is not passed along with the request")
		if jwtRequired {
			return "", "", ErrNoMetadata
		}
		return "", "", nil
	}

	// Retrieve authorization header
	autHeader := md.Get("authorization")
	if len(autHeader) == 0 {
		log.Info("JWT token is not passed along with the request")
		if jwtRequired {
			return "", "", ErrNoJWTToken
		}
		return "", "", nil
	}

	enterpriseId, err := ParseClaimFromJWT(ctx, autHeader[0], "enterpriseId")
	if err != nil {
		log.Error(err, "error while parsing the enterpriseId from token")
		return "", "", ErrTokenParse
	}

	countryCode, err := ParseClaimFromJWT(ctx, autHeader[0], "countryCode")
	if err != nil {
		log.Error(err, "error while parsing countryCode from the token")
		countryCode = ""
	}

	// Parse and extract enterpriseId from JWT token
	return enterpriseId, countryCode, nil
}

// Parses a specified claim from the JWT token in the context
func ParseClaimFromJWT(ctx context.Context, tokenStr string, claim string) (string, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("grpcutil.ParseClaimFromJWT").Start()
	defer span.End()

	jwtToken := strings.TrimPrefix(tokenStr, "Bearer ")
	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, jwt.MapClaims{})
	if err != nil {
		log.Error(err, "error while parsing the token")
		return "", ErrTokenParse
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidClaims
	}

	claimValue, ok := claims[claim].(string)
	if !ok || claimValue == "" {
		log.Info("TOKEN", claim, ErrClaimNotFound.Error())
		return "", ErrClaimNotFound
	}

	return claimValue, nil
}

// Extracts groups from JWT present in the context
func ExtractGroupsfromCtx(ctx context.Context, jwtRequired bool) (groups []string) {
	// Start logging and tracing
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("grpcutil.ExtractGroupsfromCtx").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	// Extract metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Info("Metadata is not passed along with the request")
	}

	// Retrieve authorization header
	autHeader := md.Get("authorization")
	if len(autHeader) != 0 {

		jwtToken := strings.TrimPrefix(autHeader[0], "Bearer ")
		token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, jwt.MapClaims{})
		if err != nil {
			log.Info("error while parsing the token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Info("failed to extract claims")
		}

		aInterface, ok := claims["groups"].([]interface{})
		if !ok {
			log.Info("groups not present in jtw")
		} else {
			for _, v := range aInterface {
				groups = append(groups, v.(string))
			}
		}

		// extract roles from jwt also
		aInterface, ok = claims["roles"].([]interface{})
		if !ok {
			log.Info("roles not present in jtw")
		} else {
			for _, v := range aInterface {
				groups = append(groups, v.(string))
			}
		}

	}
	return groups
}

// ExtractEmailEnterpriseIDAndGroups extracts email, enterprise ID, and groups from the JWT token present in the context
func ExtractEmailEnterpriseIDAndGroups(ctx context.Context) (email *string, enterpriseId *string, groups []string, err error) {
	// Start logging and tracing
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("grpcutil.ExtractEmailEnterpriseIDAndGroups").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	// Extract metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Info("Metadata is not passed along with the request")
		return nil, nil, nil, ErrNoMetadata
	}

	// Retrieve authorization header
	autHeader := md.Get("authorization")
	if len(autHeader) == 0 {
		return nil, nil, nil, ErrNoJWTToken
	}

	// Parse the JWT token
	jwtToken := strings.TrimPrefix(autHeader[0], "Bearer ")
	token, _, err := new(jwt.Parser).ParseUnverified(jwtToken, jwt.MapClaims{})
	if err != nil {
		log.Error(err, "error while parsing the token")
		return nil, nil, nil, ErrTokenParse
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, nil, ErrInvalidClaims
	}

	// Extract email
	emailClaim, ok := claims["email"].(string)
	if !ok || emailClaim == "" {
		errorMsg := "email is not part of the JWT token"
		return nil, nil, nil, fmt.Errorf(errorMsg)
	}
	email = &emailClaim

	// Extract enterprise ID
	enterpriseIdClaim, ok := claims["enterpriseId"].(string)
	if !ok || enterpriseIdClaim == "" {
		errorMsg := "enterpriseId is not part of the JWT token"
		return nil, nil, nil, fmt.Errorf(errorMsg)
	}
	enterpriseId = &enterpriseIdClaim

	// Extract groups
	if aInterface, ok := claims["groups"].([]interface{}); ok {
		for _, v := range aInterface {
			groups = append(groups, v.(string))
		}
	}

	// Extract roles and add to groups
	if aInterface, ok := claims["roles"].([]interface{}); ok {
		for _, v := range aInterface {
			groups = append(groups, v.(string))
		}
	}

	return email, enterpriseId, groups, nil
}
