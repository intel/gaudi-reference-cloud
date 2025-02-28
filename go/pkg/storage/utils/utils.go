// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"fmt"
	"os"
)

func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func GetRoleID(roleIdPath string) (string, error) {
	b, err := os.ReadFile(roleIdPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s", err)
	}
	return string(b), nil
}

func BytesToTerabytes(bytes int64) float64 {
	terabytes := float64(bytes) / (1024 * 1024 * 1024 * 1024)
	return float64(int(terabytes*1000)) / 1000 // Truncate to 3 decimal places
}
