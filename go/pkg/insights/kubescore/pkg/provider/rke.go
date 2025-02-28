// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package provider

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/actions/imageutils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/ghclient"
)

type RkeProvider struct {
	RepoURL string
}

const (
	releaseAssetImageList = "rke2-images-all.linux-amd64.txt"
)

func (rke RkeProvider) GetReleases(ctx context.Context, ghclient *ghclient.GHClient, topK int) ([]common.ReleaseMD, error) {
	return ghclient.GetAllReleases(ctx, rke.RepoURL, nil, topK)
}

func (rke RkeProvider) GetReleaseMeta(ctx context.Context, version string, ghclient *ghclient.GHClient) (common.ReleaseMD, error) {
	return ghclient.GetRelease(ctx, rke.RepoURL, version)
}

func (rke RkeProvider) GetReleaseAssets(ctx context.Context, version, name string, ghclient *ghclient.GHClient) ([]common.ReleaseAsset, error) {

	return nil, nil
}

func (rke RkeProvider) GetReleaseImages(ctx context.Context, version string, ghclient *ghclient.GHClient) ([]common.ImageReport, error) {
	imagesBuf, err := ghclient.GetReleaseAsset(ctx, rke.RepoURL, version, releaseAssetImageList)
	if err != nil {
		fmt.Printf("error reading release asset for repo [%s], release [%s]\n", rke.RepoURL, releaseAssetImageList)
		return nil, errors.New("failed to read release images")
	}
	reader := bytes.NewReader(imagesBuf)
	fileScanner := bufio.NewScanner(reader)

	fileScanner.Split(bufio.ScanLines)
	imgs := []common.ImageReport{}
	for fileScanner.Scan() {
		imgs = append(imgs, common.ImageReport{
			URL:       fileScanner.Text(),
			CreatedAt: imageutils.GetImageBuildTime(fileScanner.Text()),
			Digest:    imageutils.GetImageDigest(fileScanner.Text()),
		})
	}

	return imgs, nil
}

func (rke RkeProvider) GetReleaseComponentMeta(ctx context.Context, ghclient *ghclient.GHClient, repoURL string, topK int) ([]common.ReleaseMD, error) {
	//TODO: perform compatibility check
	return nil, nil
}

func (rke RkeProvider) GetReleaseSBOM(ctx context.Context, version, filepath string) error {

	return nil
}
