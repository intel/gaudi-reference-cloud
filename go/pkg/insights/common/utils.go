// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package common

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// Helper function to split a version string into its components
func ParseVersion(version string) ([]int, error) {
	// Strip the 'v' prefix
	version = strings.TrimPrefix(version, "v")

	// Split the version into parts by "."
	parts := strings.Split(version, ".")
	parsedParts := make([]int, len(parts))

	for i, part := range parts {
		// Convert each part to an integer
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid version part: %s", part)
		}
		parsedParts[i] = num
	}
	return parsedParts, nil
}

// Custom sort function for sorting versions (Descending)
func SortByTagDescending(list []*v1.K8SReleaseMD) []*v1.K8SReleaseMD {
	sort.Slice(list, func(i, j int) bool {
		versionI, errI := ParseVersion(list[i].ReleaseId)
		versionJ, errJ := ParseVersion(list[j].ReleaseId)

		if errI != nil || errJ != nil {
			// If there's an error parsing the version, fall back to lexicographic comparison
			return list[i].ReleaseId > list[j].ReleaseId
		}

		// Compare version parts in descending order (reversed comparison)
		for k := 0; k < len(versionI) && k < len(versionJ); k++ {
			if versionI[k] != versionJ[k] {
				return versionI[k] > versionJ[k] // Reverse comparison
			}
		}

		// If the versions are identical up to the length of the shortest, the longer version is considered larger
		return len(versionI) > len(versionJ) // Reverse comparison
	})
	return list
}
