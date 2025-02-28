// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

// func TestCompatibilityNetworkProxy(t *testing.T) {
// 	k8sReleases := []common.ReleaseMD{
// 		{Tag: "v1.25.0"},
// 		{Tag: "v1.26.0"},
// 		{Tag: "v1.26.2"},
// 		{Tag: "v1.27.1"},
// 		{Tag: "v1.27.8"},
// 		{Tag: "v0.5.0"},
// 	}

// 	releaseTime := time.Now()
// 	compReleases := []common.ReleaseMD{
// 		{
// 			Tag:       "v0.0.2",
// 			CreatedAt: releaseTime,
// 		},
// 		{
// 			Tag:       "v0.0.5",
// 			CreatedAt: releaseTime,
// 		},
// 	}

// 	// policy := config.ThirdPartyComponentPolicy{
// 	// 	MinimumVersion: "v0.0.0",
// 	// 	K8sVersions:    "> v1.27.0",
// 	// }

// 	expected := []common.ReleaseComponentMD{
// 		{
// 			ReleaseId:        "v1.27.1",
// 			ComponentVersion: "v0.0.2",
// 			License:          "Apache 2.0",
// 			ReleaseTime:      releaseTime,
// 		},
// 		{
// 			ReleaseId:        "v1.27.1",
// 			ComponentVersion: "v0.0.5",
// 			License:          "Apache 2.0",
// 			ReleaseTime:      releaseTime,
// 		},
// 		{
// 			ReleaseId:        "v1.27.8",
// 			ComponentVersion: "v0.0.2",
// 			License:          "Apache 2.0",
// 			ReleaseTime:      releaseTime,
// 		},
// 		{
// 			ReleaseId:        "v1.27.8",
// 			ComponentVersion: "v0.0.5",
// 			License:          "Apache 2.0",
// 			ReleaseTime:      releaseTime,
// 		},
// 	}
// 	// results := findCompabilitySupport(k8sReleases, compReleases, policy)
// 	// if !cmp.Equal(results, expected) {
// 	// 	t.Fail()
// 	// }
// 	// fmt.Println(results)
// }
