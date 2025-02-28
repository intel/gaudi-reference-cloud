// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"fmt"
	"net/url"
)

func DeploymentArtifactsRetentionDays() int {
	return 90
}

// Return the URL that the deployment artifacts tar for a commit can be downloaded from.
// artifactRepositoryUrl should be like "https://internal-placeholder.com/idc_evidence-igk-local/idc/releases".
func DeploymentArtifactsTarUrl(artifactRepositoryUrl url.URL, component string, commit string) (url.URL, error) {
	baseFileName := fmt.Sprintf("universe_deployer_deployment_artifacts_%s_%s.tar", component, commit)
	artifactUrl := artifactRepositoryUrl
	artifactUrl.Path = fmt.Sprintf("%s/%s/%s/%s", artifactUrl.Path, component, commit, baseFileName)
	return artifactUrl, nil
}
