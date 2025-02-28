// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ghclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/pkg/errors"

	"github.com/google/go-github/github"
)

var releaseFilers = []string{
	"rc",
	"alpha",
	"beta",
}

// GetAllReleases :
func (ghcli *GHClient) GetAllReleases(ctx context.Context, repoURL string, lastTime *time.Time, threshold int) ([]common.ReleaseMD, error) {
	logger := log.FromContext(ctx).WithName("GHClient.GetAllReleases")
	releases := []common.ReleaseMD{}
	repoOwner := ""
	if repoURL != "" {
		repoOwner = parseRepositoryOwner(repoURL)
	}
	license, err := ghcli.GetRepositoryLicense(ctx, repoURL)
	if err != nil {
		logger.Error(err, "error fetching repo license")
	}
	repo := parseRepositoryName(repoURL)
	pageCount := 0
	for {
		pageCount++
		rlist, ghres, err := ghcli.ClientV3.Repositories.ListReleases(ctx, repoOwner, repo, &github.ListOptions{Page: pageCount, PerPage: 100})
		if err != nil || ghres.StatusCode != 200 || len(rlist) == 0 {
			// logger.Error(err, "error making github api calls")
			if ghres != nil && ghres.StatusCode == http.StatusForbidden {
				logger.Info("error", "un-expected return code", ghres.StatusCode)
				return releases, err
			}
			break
		}
		for _, r := range rlist {
			ignoreRelease := false
			for _, f := range releaseFilers {
				if strings.Contains(r.GetTagName(), f) {
					ignoreRelease = true
				}
			}
			if !ignoreRelease {
				releases = append(releases, common.ReleaseMD{
					Tag:       r.GetTagName(),
					CreatedAt: r.GetCreatedAt().Time,
					CommitID:  r.GetTargetCommitish(),
					URL:       r.GetHTMLURL(),
					License:   license,
				})
			}
		}
	}

	sort.Slice(releases, func(i, j int) bool {
		return releases[j].CreatedAt.Before(releases[i].CreatedAt)
	})

	logger.Info("getreleases", "number of releases discovered", len(releases))

	cutOffIdx := len(releases) - 1
	if threshold != -1 && len(releases) > threshold {
		cutOffIdx = threshold
	}

	if lastTime != nil {
		cutOffIdx = 0
		for _, r := range releases {
			logger.Info("", "last time ", *lastTime, "currTime", r.CreatedAt)
			if r.CreatedAt.After(*lastTime) {
				cutOffIdx++
			} else {
				break
			}
		}
	}
	//sort release tags
	sortedReleases := sortByTagDescending(releases)
	return sortedReleases[:cutOffIdx], nil
}

// GetAllReleases :
func (ghcli *GHClient) GetAllReleasesGreaterThan(ctx context.Context, repoURL, currentVersion string) ([]common.ReleaseMD, error) {
	releases := []common.ReleaseMD{}

	rlist, err := ghcli.GetAllReleases(ctx, repoURL, nil, -1)
	if err != nil {
		return nil, fmt.Errorf("error getting release versions")
	}
	for _, r := range rlist {
		if common.IsGreater(r.Tag, currentVersion) {
			releases = append(releases, r)
		}
	}

	sort.Slice(releases, func(i, j int) bool {
		return common.IsGreater(releases[i].Tag, releases[j].Tag)
	})

	return releases, nil
}

// GetLatestRelease :
func (ghcli *GHClient) GetLatestRelease(ctx context.Context, repoURL string) (common.ReleaseMD, error) {
	release := common.ReleaseMD{}
	repoOwner := ""
	if repoURL != "" {
		repoOwner = parseRepositoryOwner(repoURL)
	}
	repo := parseRepositoryName(repoURL)
	result, ghresp, err := ghcli.ClientV3.Repositories.GetLatestRelease(ctx, repoOwner, repo)
	if err != nil {
		if ghresp.StatusCode == http.StatusForbidden {
			return release, err
		}
		return release, errors.Wrapf(err, "error quering releases")
	}
	if ghresp.StatusCode != 200 {
		return release, errors.Wrapf(err, "un-expected response code %d\n", ghresp.StatusCode)
	}
	release.Tag = result.GetTagName()
	release.CreatedAt = result.GetCreatedAt().Time
	release.CommitID = result.GetTargetCommitish()
	return release, nil
}

func (ghcli *GHClient) GetReleaseTimestamp(ctx context.Context, repoURL, releaseID string) (time.Time, error) {
	ts := time.Time{}
	owner := ""
	if repoURL != "" {
		owner = parseRepositoryOwner(repoURL)
	}
	repo := parseRepositoryName(repoURL)

	release, ghresp, err := ghcli.ClientV3.Repositories.GetReleaseByTag(ctx, owner, repo, releaseID)
	if err != nil {
		fmt.Printf("error reading release time: %v", err)
		return ts, errors.Wrapf(err, "error quering releases")
	}
	if ghresp.StatusCode != 200 {
		fmt.Println(ghresp.StatusCode)
		return ts, errors.Wrapf(err, "un-expected response code %d\n", ghresp.StatusCode)
	}

	return release.GetPublishedAt().Time, nil
}

func (ghcli *GHClient) GetReleaseAsset(ctx context.Context, repoURL, releaseTag, assetName string) ([]byte, error) {
	owner := ""
	if repoURL != "" {
		owner = parseRepositoryOwner(repoURL)
	}
	repo := parseRepositoryName(repoURL)

	rlmd, err := ghcli.GetRelease(ctx, repoURL, releaseTag)
	if err != nil {
		fmt.Printf("error fetching release")
	}
	result, ghresp, err := ghcli.ClientV3.Repositories.ListReleaseAssets(ctx, owner, repo, rlmd.ID, &github.ListOptions{})
	if err != nil {
		if ghresp.StatusCode == http.StatusForbidden {
			return nil, err
		}
		return nil, errors.Wrapf(err, "error quering releases")
	}
	if ghresp.StatusCode != 200 {
		return nil, errors.Wrapf(err, "un-expected response code %d\n", ghresp.StatusCode)
	}

	var assetID int64
	for _, ra := range result {
		if strings.Compare(ra.GetName(), assetName) == 0 {
			assetID = ra.GetID()
			break
		}
	}

	if assetID == 0 {
		fmt.Printf("release asset not found for release [%s], assetname [%s]\n", releaseTag, assetName)
		return nil, nil
	}

	reader, redirectURL, err := ghcli.ClientV3.Repositories.DownloadReleaseAsset(ctx, owner, repo, assetID)
	if err != nil {
		if ghresp.StatusCode == http.StatusForbidden {
			return nil, err
		}
		return nil, errors.Wrapf(err, "error quering releases")
	}
	if reader != nil {
		return io.ReadAll(reader)
	} else if redirectURL != "" {
		return downloadFromURL(redirectURL)
	}

	return nil, errors.New("empty response")
}

func (ghcli *GHClient) GetRelease(ctx context.Context, repoURL, tag string) (common.ReleaseMD, error) {
	rmd := common.ReleaseMD{}
	repoOwner := ""
	if repoURL != "" {
		repoOwner = parseRepositoryOwner(repoURL)
	}
	repo := parseRepositoryName(repoURL)
	result, ghresp, err := ghcli.ClientV3.Repositories.GetReleaseByTag(ctx, repoOwner, repo, tag)
	if err != nil {
		if ghresp.StatusCode == http.StatusForbidden {
			return rmd, err
		}
		return rmd, errors.Wrapf(err, "error quering releases")
	}
	if ghresp.StatusCode != 200 {
		return rmd, errors.Wrapf(err, "un-expected response code %d\n", ghresp.StatusCode)
	}

	rmd.Tag = result.GetTagName()
	rmd.Name = result.GetName()
	rmd.ID = result.GetID()
	rmd.CreatedAt = result.GetCreatedAt().Time
	rmd.CommitID = result.GetTargetCommitish()

	return rmd, nil
}
