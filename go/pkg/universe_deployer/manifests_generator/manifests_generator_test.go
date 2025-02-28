// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package manifests_generator

import (
	"context"
	"os"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/reader"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test Manifests Generator for a specific environment.
// Caller can optionally specify a Universe Config file to override the components to include.
// Actual commit hashes are disregarded.
// Instead, this tests with the currently checked out commit (HEAD).
// This ensures that if commit hashes in the Universe Config are updated to the current commit,
// Manifests Generator will succeed.
func RunManifestsGenerator(idcEnv string, universeConfigFile string) error {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "universe_deployer_manifests_generator_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	deploymentArtifactsDir, err := os.Getwd()
	if err != nil {
		return err
	}

	envConfigReader := reader.Reader{
		DeploymentArtifactsDir: deploymentArtifactsDir,
	}
	multipleEnvConfig, err := envConfigReader.Read(ctx, idcEnv)
	if err != nil {
		return err
	}
	envConfig := multipleEnvConfig.Environments[idcEnv]

	var universeConfig *universe_config.UniverseConfig
	if universeConfigFile == "" {
		var err error
		universeConfig, err = universe_config.NewUniverseConfigFromEnvConfig(ctx, envConfig.EnvConfig)
		if err != nil {
			return err
		}
	} else {
		var err error
		universeConfig, err = universe_config.NewUniverseConfigFromFile(ctx, universeConfigFile)
		if err != nil {
			return err
		}
	}

	manifestsGenerator := &ManifestsGenerator{
		DeploymentArtifactsDir:       deploymentArtifactsDir,
		SecretsDir:                   filepath.Join(tempDir, "local/secrets"),
		UniverseConfig:               universeConfig,
		MultipleEnvConfig:            multipleEnvConfig,
		Commit:                       "5bd958c6ce65eea78b07da6d736f145b54f18e3a",
		Components:                   []string{},
		Snapshot:                     true,
		ManifestsTar:                 filepath.Join(tempDir, "manifests.tar"),
		OverrideDefaultChartRegistry: "",
	}
	if _, err := manifestsGenerator.GenerateManifests(ctx); err != nil {
		return err
	}
	return nil
}

func getIdcEnvs() []string {
	return []string{
		"dev3",
		"kind-singlecluster",
		"minimal",
		"qa1",
		"test-e2e-compute-bm",
		"test-e2e-compute-vm",
	}
}

var _ = Describe("Manifests Generator Tests", func() {
	It("Manifests Generator should succeed (staging)", func() {
		Expect(RunManifestsGenerator("staging", "universe_deployer/environments/staging-head.json")).Should(Succeed())
	})

	It("Manifests Generator should succeed (prod)", func() {
		Expect(RunManifestsGenerator("prod", "universe_deployer/environments/prod-head.json")).Should(Succeed())
	})

	idcEnvs := getIdcEnvs()
	for _, idcEnv := range idcEnvs {
		It("Manifests Generator should succeed ("+idcEnv+")", func() {
			Expect(RunManifestsGenerator(idcEnv, "")).Should(Succeed())
		})
	}
})
