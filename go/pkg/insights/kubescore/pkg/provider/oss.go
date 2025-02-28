// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/actions/imageutils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/ghclient"
	"github.com/spdx/tools-golang/json"
	"github.com/spdx/tools-golang/tagvalue"
)

const (
	// kubernetes
	kubeapiserver = "registry.k8s.io/kube-apiserver"
	kubectrl      = "registry.k8s.io/kube-controller-manager"
	kubeschd      = "registry.k8s.io/kube-scheduler"
	kubeproxy     = "registry.k8s.io/kube-proxy"
	pause         = "registry.k8s.io/pause"
	etcd          = "registry.k8s.io/etcd"
	coredns       = "registry.k8s.io/coredns/coredns"
	// calcico
	node  = "quay.io/calico/node"
	typha = "quay.io/calico/typha"
)

type OSSProvider struct {
	KubeRepoURL       string
	CalicoRegistryURL string
}

// Calico
func (oss *OSSProvider) GetCalicoReleases(ctx context.Context, ghclient *ghclient.GHClient, lastTime *time.Time, topK int) ([]common.ReleaseMD, error) {
	return ghclient.GetAllReleases(ctx, oss.CalicoRegistryURL, lastTime, topK)
}

func (oss *OSSProvider) GetCalicoReleaseMeta(ctx context.Context, version string, ghclient *ghclient.GHClient) (common.ReleaseMD, error) {
	return ghclient.GetRelease(ctx, oss.CalicoRegistryURL, version)
}

// Kubernetes
func (oss *OSSProvider) GetReleases(ctx context.Context, ghclient *ghclient.GHClient, lastTime *time.Time, topK int) ([]common.ReleaseMD, error) {
	return ghclient.GetAllReleases(ctx, oss.KubeRepoURL, lastTime, topK)
}

func (oss *OSSProvider) GetReleaseComponentMeta(ctx context.Context, ghclient *ghclient.GHClient, repoURL string, topK int) ([]common.ReleaseMD, error) {
	//TODO: perform compatibility check
	return ghclient.GetAllReleases(ctx, repoURL, nil, topK)
}

func (oss *OSSProvider) GetReleaseMeta(ctx context.Context, version string, ghclient *ghclient.GHClient) (common.ReleaseMD, error) {
	return ghclient.GetRelease(ctx, oss.KubeRepoURL, version)
}

func (oss *OSSProvider) GetReleaseAssets(ctx context.Context, version, name string, ghclient *ghclient.GHClient) ([]common.ReleaseAsset, error) {

	return nil, nil
}

func (oss *OSSProvider) GetReleaseImages(ctx context.Context, version string, ghclient *ghclient.GHClient) ([]common.ImageReport, error) {

	imgs := []common.ImageReport{
		{
			URL:       kubeapiserver + ":" + version,
			Digest:    imageutils.GetImageDigest(kubeapiserver + ":" + version),
			CreatedAt: imageutils.GetImageBuildTime(kubeapiserver + ":" + version),
		},
		{
			URL:       kubectrl + ":" + version,
			Digest:    imageutils.GetImageDigest(kubectrl + ":" + version),
			CreatedAt: imageutils.GetImageBuildTime(kubectrl + ":" + version),
		},
		{
			URL:       kubeschd + ":" + version,
			Digest:    imageutils.GetImageDigest(kubeschd + ":" + version),
			CreatedAt: imageutils.GetImageBuildTime(kubeschd + ":" + version),
		},
		{
			URL:       kubeproxy + ":" + version,
			Digest:    imageutils.GetImageDigest(kubeproxy + ":" + version),
			CreatedAt: imageutils.GetImageBuildTime(kubeproxy + ":" + version),
		},
	}

	return imgs, nil
}

func (oss *OSSProvider) GetCalicoReleaseImages(ctx context.Context, version string, ghclient *ghclient.GHClient) ([]common.ImageReport, error) {
	imgs := []common.ImageReport{
		{
			URL:       node + ":" + version,
			Digest:    imageutils.GetImageDigest(node + ":" + version),
			CreatedAt: imageutils.GetImageBuildTime(node + ":" + version),
		},
		{
			URL:       typha + ":" + version,
			Digest:    imageutils.GetImageDigest(typha + ":" + version),
			CreatedAt: imageutils.GetImageBuildTime(typha + ":" + version),
		},
	}

	return imgs, nil
}

func (oss OSSProvider) GetReleaseSBOM(ctx context.Context, version, filepath string) error {

	server := "https://sbom.k8s.io/"
	verionURI := fmt.Sprintf("%s/release", version)
	retCode, res, err := common.MakeGetAPICall(ctx, server, verionURI, nil)
	if err != nil {
		return err
	}

	if retCode != http.StatusOK {
		return fmt.Errorf("unexpected return code: %d", retCode)
	}
	if res == nil {
		return fmt.Errorf("empty response")
	}

	tmpspdxfp, err := os.CreateTemp(os.TempDir(), version)
	if err != nil {
		return fmt.Errorf("error saving tag-vaule spdx file")
	}
	defer tmpspdxfp.Close()
	if _, err := tmpspdxfp.Write(res); err != nil {
		return fmt.Errorf("error saving tag-vaule spdx file")
	}
	fmt.Printf("spdx tagvalue filepath: %s\n", tmpspdxfp.Name())
	if err := ConvertSPDXTVToJSON(tmpspdxfp.Name(), filepath); err != nil {
		return fmt.Errorf("error converting spdx to json ")
	}
	return nil
}

func ConvertSPDXTVToJSON(tvfp, jsonfp string) error {
	// open the SPDX file
	r, err := os.Open(tvfp)
	if err != nil {
		return fmt.Errorf("error while opening %v for reading: %v", tvfp, err)
	}
	defer r.Close()

	// try to load the SPDX file's contents as a tag-value file
	doc, err := tagvalue.Read(r)
	if err != nil {
		return fmt.Errorf("error while parsing %v: %v", tvfp, err)
	}

	// create a new file for writing
	w, err := os.Create(jsonfp)
	if err != nil {
		return fmt.Errorf("error while opening %v for writing: %v", jsonfp, err)
	}
	defer w.Close()

	var opt []json.WriteOption
	// you can use WriteOption to change JSON format
	// uncomment the following code to test it
	opt = append(opt, json.Indent(" "))      // to create multiline json
	opt = append(opt, json.EscapeHTML(true)) // to escape HTML characters

	// try to save the document to disk as JSON file
	err = json.Write(doc, w, opt...)
	if err != nil {
		return fmt.Errorf("error while saving %v: %v", jsonfp, err)
	}
	return nil
}
