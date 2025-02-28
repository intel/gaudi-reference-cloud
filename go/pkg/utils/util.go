// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"fmt"
	"math"
	"sort"
)

func GenerateAcceleratorVNetName(instanceName string, instanceGroup string, clusterGroupID string) string {
	if instanceGroup == "" {
		return fmt.Sprintf("accelerator-net-%s-%s", instanceName, clusterGroupID)
	}
	return fmt.Sprintf("accelerator-net-%s-%s", instanceGroup, clusterGroupID)
}

func GenerateBGPClusterVNetName(superComputeGroupID, instanceGroup string) string {
	if superComputeGroupID != "" {
		return fmt.Sprintf("%s-accelerator-net-%s", superComputeGroupID, instanceGroup)
	}
	return fmt.Sprintf("accelerator-net-%s", instanceGroup)
}

func GenerateStorageVnetName(az string) string {
	return fmt.Sprintf("%s-storage", az)
}

func GetReservedHostCount() int {
	// gateway (.1), reserved (.2, .3)
	return 3
}

// Get the maximum CIDR prefix length that can accommodate a given number of usable IP addresses.
func GetMaximumPrefixLength(usableIPs int32) int32 {
	var maxVNetPrefixLength int32 = 32

	// For 1 or less usable IPs, return /32, which is 1 IP.
	if usableIPs <= 1 {
		return maxVNetPrefixLength
	}

	// Adjusting the count to include network and broadcast addresses (+2), and reserved addresses.
	totalIPs := usableIPs + 2 + int32(GetReservedHostCount())

	// Find the smallest power of 2 greater than or equal to totalIPs
	cidrPrefix := maxVNetPrefixLength - int32(math.Ceil(math.Log2(float64(totalIPs))))

	if cidrPrefix < 0 {
		cidrPrefix = 0
	} else if cidrPrefix > maxVNetPrefixLength {
		cidrPrefix = maxVNetPrefixLength
	}
	return cidrPrefix
}

// Return the list of keys from a map, sorted by name.
func ToList(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// containsTrue checks if the slice contains the string "true".
func ContainsTrue(values []string) bool {
	for _, value := range values {
		if value == "true" {
			return true
		}
	}
	return false
}
