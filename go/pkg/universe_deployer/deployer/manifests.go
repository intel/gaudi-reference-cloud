// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"context"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

func (e *Deployer) GenerateManifests(ctx context.Context) error {
	if e.GenerateManifestsComplete {
		return nil
	}
	manifests, err := e.ManifestsGenerator.GenerateManifests(ctx)
	if err != nil {
		return err
	}
	e.Manifests = manifests
	// Copy to BuildArtifactsDir to make it available in Jenkins.
	if e.Options.BuildArtifactsDir != "" {
		dest := filepath.Join(e.Options.BuildArtifactsDir, "manifests.tar")
		if err := util.CopyFile(e.ManifestsTar, dest); err != nil {
			return err
		}
	}
	e.GenerateManifestsComplete = true
	return nil
}
