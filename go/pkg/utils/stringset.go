// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

// returns a list containing the common elements between list1 and list2.
func Intersect(list1, list2 []string) []string {
	poolSet := make(map[string]struct{})
	for _, pool := range list1 {
		poolSet[pool] = struct{}{}
	}

	// Set for intersection
	intersectionSet := make(map[string]struct{})
	for _, pool := range list2 {
		if _, exists := poolSet[pool]; exists {
			// Add to set, ensuring uniqueness
			intersectionSet[pool] = struct{}{}
		}
	}

	// Convert set back to slice
	var intersection []string
	for pool := range intersectionSet {
		intersection = append(intersection, pool)
	}

	return intersection
}
