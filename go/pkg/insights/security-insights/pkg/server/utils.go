// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// Function to compare two reports for a component and return three lists of vulnerabilities
func compareVulnerabilities(currReport, newReport []*pb.Vulnerability) (common, uniqueToCurr, uniqueToNew []*pb.Vulnerability) {
	// Create maps to hold the vulnerabilities for quick lookup by CVE ID
	currVulnsMap := make(map[string]*pb.Vulnerability)
	newVulnsMap := make(map[string]*pb.Vulnerability)

	// Maps to track vulnerabilities added to each list (prevent duplicates)
	commonSet := make(map[string]struct{})
	uniqueToCurrSet := make(map[string]struct{})
	uniqueToNewSet := make(map[string]struct{})

	// Populate the maps with vulnerabilities from both reports
	for _, v := range currReport {
		currVulnsMap[v.Id] = v
	}
	for _, v := range newReport {
		newVulnsMap[v.Id] = v
	}

	// Find common vulnerabilities
	for _, v := range currReport {
		if _, found := newVulnsMap[v.Id]; found {
			if _, exists := commonSet[v.Id]; !exists {
				common = append(common, v)
				commonSet[v.Id] = struct{}{}
			}
		} else {
			if _, exists := uniqueToCurrSet[v.Id]; !exists {
				uniqueToCurr = append(uniqueToCurr, v)
				uniqueToCurrSet[v.Id] = struct{}{}
			}
		}
	}
	// Find new vulnerabilities
	for _, v := range newReport {
		if _, found := currVulnsMap[v.Id]; !found {
			if _, exists := uniqueToNewSet[v.Id]; !exists {
				uniqueToNew = append(uniqueToNew, v)
				uniqueToNewSet[v.Id] = struct{}{}
			}
		}
	}

	return
}
