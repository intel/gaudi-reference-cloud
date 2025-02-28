// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ghclient

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/common"
)

func parseRepositoryOwner(giturl string) string {
	var owner string
	gParts := strings.Split(giturl, "/")
	if len(gParts) >= 4 {
		owner = gParts[3]
	}
	return owner
}

func parseRepositoryName(giturl string) string {
	var repo string
	repoURL := strings.TrimRight(giturl, "/")
	urlParts := strings.Split(repoURL, "/")
	repo = urlParts[len(urlParts)-1]
	return repo
}

func contains(cache map[int]struct{}, key int) bool {
	if _, ok := cache[key]; ok {
		return true
	}
	return false
}

func downloadFromURL(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading from url")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error reading from url")
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading from url")
	}
	return bytes, nil
}

// Helper function to split a version string into its components
func parseVersion(version string) ([]int, error) {
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
func sortByTagDescending(list []common.ReleaseMD) []common.ReleaseMD {
	sort.Slice(list, func(i, j int) bool {
		versionI, errI := parseVersion(list[i].Tag)
		versionJ, errJ := parseVersion(list[j].Tag)

		if errI != nil || errJ != nil {
			// If there's an error parsing the version, fall back to lexicographic comparison
			return list[i].Tag > list[j].Tag
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
