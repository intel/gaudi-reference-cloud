// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package address_translation

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type AddressTranslationType string

const (
	Internet  AddressTranslationType = "internet"
	Storage   AddressTranslationType = "storage"
	Transient AddressTranslationType = "transient"
)

var mapAddressTranslationType = map[string]AddressTranslationType{
	"internet":  Internet,
	"storage":   Storage,
	"transient": Transient,
}

func IsValidTranslationType(translationType string) (AddressTranslationType, error) {
	normalizedTranslationType, exists := mapAddressTranslationType[strings.ToLower(translationType)]

	if !exists {
		// Generate message and return error
		validTypes := make([]string, 0, len(mapAddressTranslationType))
		for key := range mapAddressTranslationType {
			validTypes = append(validTypes, key)
		}
		errorMsg := fmt.Sprintf("Invalid translation type. Valid types are: %v", validTypes)
		return "", status.Error(codes.InvalidArgument, errorMsg)
	}
	return normalizedTranslationType, nil
}
