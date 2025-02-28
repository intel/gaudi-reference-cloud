// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package imageutils

import (
	"encoding/json"

	"github.com/google/go-containerregistry/pkg/crane"
)

func GetImageBuildTime(imageURL string) string {
	type imageConfig struct {
		Created string `json:"created"`
	}

	imgCfg := imageConfig{}
	c, err := crane.Config(imageURL)
	if err != nil {
		return ""
	}
	if err := json.Unmarshal(c, &imgCfg); err != nil {
		return ""
	}
	return imgCfg.Created
}

func GetImageDigest(imageURL string) string {
	d, err := crane.Digest(imageURL)
	if err != nil {
		return ""
	}
	return d
}
