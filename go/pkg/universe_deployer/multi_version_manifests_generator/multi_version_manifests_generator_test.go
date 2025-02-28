// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package multi_version_manifests_generator

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/artifactory"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func RunMultiVersionManifestsGenerator(universeConfigFilename string) error {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("RunMultiVersionManifestsGenerator")
	tempDir, err := os.MkdirTemp("", "universe_deployer_multi_commit_manifests_generator_test_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	deploymentArtifactsTar := filepaths.DeploymentArtifactsTar

	useFakeArtifactory := false
	var artifactRepository string
	stopServer := func() error { return nil }

	if useFakeArtifactory {
		// To use this mode, run `hack/fake-artifactory.sh` before running this test.
		port := 46038
		artifactRepository = fmt.Sprintf("http://localhost:%d/artifactory/idc_evidence-igk-local/idc/releases", port)
	} else {
		// Run an HTTP server that serves the deployment artifacts tar built as a dependency of this test.
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			return err
		}
		port := listener.Addr().(*net.TCPAddr).Port
		log.Info("Starting HTTP server", "port", port)
		mux := http.NewServeMux()
		handler := func(w http.ResponseWriter, r *http.Request) {
			log.Info("HTTP handler serving file", "deploymentArtifactsTar", deploymentArtifactsTar)
			http.ServeFile(w, r, deploymentArtifactsTar)
		}
		mux.HandleFunc("/", handler)
		srv := &http.Server{
			Handler: mux,
		}
		go func() {
			if err := srv.Serve(listener); err != nil {
				log.Error(err, "serve")
			}
		}()
		artifactRepository = fmt.Sprintf("http://localhost:%d/artifactory/idc_evidence-igk-local/idc/releases", port)
		stopServer = func() error {
			return srv.Shutdown(ctx)
		}
	}

	artifactRepositoryUrl, err := url.Parse(artifactRepository)
	if err != nil {
		return err
	}

	universeConfig, err := universe_config.NewUniverseConfigFromFile(ctx, universeConfigFilename)
	if err != nil {
		return err
	}
	headCommit := "6302c77c84297d0c112e2f576831712ee0b041b1"
	if _, err = universeConfig.ReplaceCommits(ctx, map[string]string{util.HEAD: headCommit}); err != nil {
		return err
	}
	if _, err = universeConfig.ReplaceConfigCommits(ctx, map[string]string{util.HEAD: headCommit}); err != nil {
		return err
	}

	artifactoryObj, err := artifactory.New(ctx)
	if err != nil {
		return err
	}

	multiVersionManifestsGenerator := &MultiVersionManifestsGenerator{
		Artifactory:                artifactoryObj,
		ArtifactRepositoryUrl:      *artifactRepositoryUrl,
		CacheDir:                   filepath.Join(tempDir, "cache"),
		HeadCommit:                 headCommit,
		HeadDeploymentArtifactsTar: deploymentArtifactsTar,
		CombinedManifestsTar:       filepath.Join(tempDir, "manifests.tar"),
		UniverseConfig:             universeConfig,
	}
	manifests, err := multiVersionManifestsGenerator.GenerateManifests(ctx)
	Expect(err).Should(Succeed())
	log.Info("manifests", "manifests", manifests)

	Expect(stopServer()).Should(Succeed())

	// Generate manifests again to ensure that caching works.
	manifests2, err := multiVersionManifestsGenerator.GenerateManifests(ctx)
	Expect(err).Should(Succeed())
	Expect(manifests2).Should(Equal(manifests))

	return nil
}

var _ = Describe("Manifests Generator Tests", func() {
	It("Manifests Generator should succeed", func() {
		Expect(RunMultiVersionManifestsGenerator("go/pkg/universe_deployer/multi_version_manifests_generator/testdata/staging-test.json")).Should(Succeed())
	})
})
