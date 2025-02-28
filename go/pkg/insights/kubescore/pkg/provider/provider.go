// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package provider

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/ghclient"
)

type KubeProvider interface {
	GetReleases(context.Context, *ghclient.GHClient, int) ([]common.ReleaseMD, error)
	GetReleaseMeta(context.Context, string, *ghclient.GHClient) (common.ReleaseMD, error)
	GetReleaseAssets(context.Context, string, string, *ghclient.GHClient) ([]common.ReleaseAsset, error)
	GetReleaseImages(context.Context, string, *ghclient.GHClient) ([]common.ImageReport, error)
	GetReleaseComponentMeta(context.Context, *ghclient.GHClient, string, int) ([]common.ReleaseMD, error)
	GetReleaseSBOM(context.Context, string, string) error
}
