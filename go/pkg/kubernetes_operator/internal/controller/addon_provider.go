// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"fmt"
	"strings"

	helm "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/addon_provider/helm"
	kubectl "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/addon_provider/kubectl"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"k8s.io/client-go/rest"
)

const (
	kubectlAddonProvider = "kubectl"
	helmAddonProvider    = "helm"
	vastAddonProvider    = "vast"
)

type addonProvider interface {
	Put(context.Context, *privatecloudv1alpha1.Addon) error
	Get(context.Context, string, string) (*privatecloudv1alpha1.AddonStatus, error)
	Delete(context.Context, *privatecloudv1alpha1.Addon) error
}

func NewAddonProvider(provider string, restConfig *rest.Config, s3Info kubectl.S3AddonConfig) (addonProvider, error) {
	providerTypeSplit := strings.Split(provider, "-")
	if len(providerTypeSplit) == 2 {
		provider = strings.Split(provider, "-")[0]
	}

	if provider == kubectlAddonProvider {
		kubectlProvider, err := kubectl.NewAddonProvider(restConfig, s3Info)
		if err != nil {
			return nil, err
		}

		return kubectlProvider, nil
	}

	if provider == helmAddonProvider {
		helmProvider, err := helm.NewHelmAddonProvider(restConfig)
		if err != nil {
			return nil, err
		}

		return helmProvider, nil
	}

	if provider == vastAddonProvider {
		vastProvider, err := helm.NewVastAddonProvider(restConfig)
		if err != nil {
			return nil, err
		}

		return vastProvider, nil
	}

	return nil, fmt.Errorf("addon provider not found")
}
