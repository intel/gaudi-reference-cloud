// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package common

import (
	"github.com/Masterminds/semver"
)

// IsEqual :
func IsEqual(base, target string) bool {
	baseV, err := semver.NewVersion(base)
	if err != nil {
		// fmt.Printf("error parsing version `%v`: %v\n", base, err)
		return false
	}
	targetV, err := semver.NewVersion(target)
	if err != nil {
		// fmt.Printf("error parsing version `%v`: %v\n", target, err)
		return false
	}
	return baseV.Equal(targetV)
}

func IsGreater(base, target string) bool {
	baseV, err := semver.NewVersion(base)
	if err != nil {
		// fmt.Printf("error parsing version `%v`: %v\n", base, err)
		return false
	}
	targetV, err := semver.NewVersion(target)
	if err != nil {
		// fmt.Printf("error parsing version `%v`: %v\n", target, err)
		return false
	}
	return baseV.GreaterThan(targetV)
}

func IsGreaterMajor(base, target string) bool {
	baseV, err := semver.NewVersion(base)
	if err != nil {
		// fmt.Printf("error parsing version `%v`: %v\n", base, err)
		return false
	}
	targetV, err := semver.NewVersion(target)
	if err != nil {
		// fmt.Printf("error parsing version `%v`: %v\n", target, err)
		return false
	}
	return baseV.Major() > targetV.Major()
}

func IsValid(version string) bool {
	_, err := semver.NewVersion(version)
	return err == nil
}

func CompareWithConstraints(version, constraint string) bool {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return false
	}
	// Check if the version meets the constraints. The a variable will be true.
	return c.Check(v)
}
