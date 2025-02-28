// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import "github.com/google/uuid"

func IsValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

func IsValidResourceID(resourceId string) bool {
	return IsValidUUID(resourceId)
}
