// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ParseFileSizeInGB(size string) int64 {
	splits := strings.Split(size, "GB")
	if len(splits) != 2 {
		return -1
	}

	sizeInt, err := strconv.ParseInt(splits[0], 10, 64)
	if err != nil {
		return -1
	}
	return sizeInt
}

func ParseFileSize(size string) int64 {
	var unit string
	if strings.HasSuffix(size, "GB") {
		unit = "GB"
	} else if strings.HasSuffix(size, "TB") {
		unit = "TB"
	} else {
		return -1
	}
	// split string to extract numeric value
	splits := strings.Split(size, unit)
	if len(splits) != 2 {
		return -1
	}
	// convert value to bytes
	sizeInt, err := strconv.ParseInt(splits[0], 10, 64)
	if err != nil {
		return -1
	}

	if unit == "TB" {
		sizeInt *= 1000
	}
	return sizeInt
}

func ParseBucketSize(size string) error {
	_, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		return err
	}
	return nil
}

// Return multiple for unit type in GB
func ParseUnit(size string) int64 {
	// This always returns 1000 for now assuming the min and max quota are specified in TB
	// TODO: in the future add intelligent logic here in case the above assumption is not true
	return 1000
}

// Convert req size for Priavte req from GB to TB if applicable
func ProcesSize(size string) string {
	res := size
	if strings.HasSuffix(size, "GB") {
		splits := strings.Split(size, "GB")
		if len(splits) != 2 {
			return size
		}
		sizeInt, err := strconv.ParseInt(splits[0], 10, 64)
		if err != nil {
			return size
		}
		if sizeInt >= 1000 {
			sizeInt /= 1000
			res = strconv.FormatInt(sizeInt, 10) + "TB"
		}
	}
	return res
}

// Validate resourceName.
// resourceName is valid when name is starting and ending with lowercase alphanumeric
// and contains lowercase alphanumeric, '-' characters and should have at most 63 characters
func ValidateResourceName(name string) error {
	re := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	matches := re.FindAllString(name, -1)
	if matches == nil {
		return status.Error(codes.InvalidArgument, "invalid resource name")
	}
	return nil
}

func ValidateInstanceName(name string) error {
	re := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	matches := re.FindAllString(name, -1)
	if matches == nil {
		return status.Error(codes.InvalidArgument, "invalid resource name")
	}
	return nil
}

// Validate prefix.
func ValidatePrefix(name string) error {
	re := regexp.MustCompile(`^([A-z0-9\\/\\.]*)$`)
	matches := re.FindAllString(name, -1)
	if len(name) > 1024 {
		return status.Error(codes.InvalidArgument, "maximum prefix length allowed is 1024")
	}
	if matches == nil {
		return status.Error(codes.InvalidArgument, "invalid prefix")
	}
	return nil
}

func AddSizes(s1, s2 string) (string, error) {
	size1 := ParseFileSize(s1)
	size2 := ParseFileSize(s2)
	if size1 < 0 || size2 < 0 {
		return "", fmt.Errorf("failed to parse file sizes when adding: %s or %s", s1, s2)
	}
	newSize := strconv.Itoa(int(size1)+int(size2)) + "GB"
	return ProcesSize(newSize), nil
}

func ConvertBytesToGBOrTB(bytes uint64) string {
	GB := 1_000_000_000     // 1 GB = 1,000,000,000 Bytes
	TB := 1_000_000_000_000 // 1 TB = 1,000,000,000,000 Bytes

	gb := int(bytes) / GB
	tb := int(bytes) / TB

	if gb < 1000 {
		return strconv.Itoa(gb) + "GB"
	}
	return strconv.Itoa(tb) + "TB"

}
