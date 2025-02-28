// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"github.com/Masterminds/semver"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateGetRelease(in *v1.GetReleaseRequest) error {
	err := in.Validate()
	if validationErr, ok := err.(pb.GetReleaseRequestValidationError); ok {
		if validationErr.Field() == "ReleaseId" {
			return status.Errorf(codes.FailedPrecondition, "Invalid ReleaseId")
		} else if validationErr.Field() == "ComponentName" {
			return status.Errorf(codes.FailedPrecondition, "Invalid Component")
		}
	}
	return nil
}

// Custom validator function to check if NewReleaseId is ahead of CurrReleaseId
func validateComparisonReleases(in *v1.ReleaseComparisonFilter) error {
	err := in.Validate()
	if err != nil {
		if validationErr, ok := err.(pb.ReleaseComparisonFilterValidationError); ok {
			if validationErr.Field() == "CurrReleaseId" {
				return status.Errorf(codes.FailedPrecondition, "Invalid currReleaseId")
			} else if validationErr.Field() == "NewReleaseId" {
				return status.Errorf(codes.FailedPrecondition, "Invalid newReleaseId")
			} else if validationErr.Field() == "ComponentName" {
				return status.Errorf(codes.FailedPrecondition, "Invalid Component")
			} else {
				return status.Errorf(codes.FailedPrecondition, "Invalid Input")
			}
		}
	}
	// Trim the 'v' from the beginning of the version string
	newReleaseId := in.NewReleaseId[1:]
	currReleaseId := in.CurrReleaseId[1:]

	// Parse the release IDs as semantic versions
	newVersion, err := semver.NewVersion(newReleaseId)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "Failed to parse new releaseId as semantic")
	}

	currVersion, err := semver.NewVersion(currReleaseId)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "Failed to parse curr releaseId as semantic")
	}
	// New releaseId must be greater than current releaseId
	if !newVersion.GreaterThan(currVersion) {
		return status.Errorf(codes.FailedPrecondition, "new releaseId must be ahead of current releaseId")
	}

	return nil
}
