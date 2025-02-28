// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ghclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// GetRepositoryMD :
func (ghcli *GHClient) GetRepositoryLicense(ctx context.Context, repo string) (string, error) {
	license := "unknown"
	owner := ""
	if repo != "" {
		owner = parseRepositoryOwner(repo)
	}
	repoName := parseRepositoryName(repo)
	r, resp, err := ghcli.ClientV3.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		fmt.Println(err)
		if resp != nil && resp.StatusCode == http.StatusForbidden {
			return license, err
		}
		return license, errors.Wrapf(err, "error fetching repository metadata")
	}
	// fmt.Printf("ghclient status code: %d\n", res.StatusCode)
	if r != nil {
		if r.GetLicense() != nil {
			license = *r.GetLicense().Name
		}
	}
	return license, nil
}
