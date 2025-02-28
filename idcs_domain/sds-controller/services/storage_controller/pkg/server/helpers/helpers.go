// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package helpers

import (
	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/rs/zerolog/log"
)

func IntoAuthCreds(ctx *v1.AuthenticationContext) *backend.AuthCreds {
	if ctx == nil {
		return nil
	}

	switch t := ctx.GetScheme().(type) {
	case *v1.AuthenticationContext_Basic_:
		return &backend.AuthCreds{
			Scheme:      backend.Basic,
			Principal:   t.Basic.GetPrincipal(),
			Credentials: t.Basic.GetCredentials(),
		}
	case *v1.AuthenticationContext_Bearer_:
		return &backend.AuthCreds{
			Scheme:      backend.Bearer,
			Credentials: t.Bearer.GetToken(),
		}
	default:
		log.Warn().Msgf("Found unexpected auth scheme %T", t)
		return nil
	}
}

func ValueOrNil[T ~int | ~uint | ~uint64 | ~int64 | ~bool | ~string](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}

func IsAllKeyValueExists(m map[string]string, search map[string]string) bool {
	if len(m) < len(search) {
		return false
	}
	for k, v := range m {
		if val, ok := search[k]; ok && v != val {
			return false
		}
	}
	return true
}
