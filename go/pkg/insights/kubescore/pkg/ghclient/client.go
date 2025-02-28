// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ghclient

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GHClient :
type GHClient struct {
	ClientV3 *github.Client
	// ClientV4 *githubv4.Client
}

// Setup : setup github client for v3 and v4
func (cli *GHClient) Setup(ctx context.Context, authToken string) error {
	//setup v2 client for go-client apis
	if err := os.Setenv("NO_PROXY", ""); err != nil {
		fmt.Printf("error setting env NO_PROXY")
	}
	if err := os.Setenv("no_proxy", ""); err != nil {
		fmt.Printf("error setting env no_proxy")
	}
	if err := os.Setenv("https_proxy", ""); err != nil {
		fmt.Printf("error setting env https_proxy")
	}
	if err := os.Setenv("HTTPS_PROXY", ""); err != nil {
		fmt.Printf("error setting env HTTPS_PROXY")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	cli.ClientV3 = github.NewClient(tc)

	// cli.ClientV4 = githubv4.NewClient(tc)

	return nil
}
