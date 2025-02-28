// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubernetes 1.24 kube-scheduler (https://github.com/kubernetes/kubernetes/tree/73da4d3652771d6c6dfe904fe8fae594a1a72e2b/pkg/scheduler).
// To see changes made, run diff-kube-scheduler.sh.

/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"context"
	"errors"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"

	toolsk8s "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s"
	"k8s.io/client-go/rest"
)

const (
	DurationToExpireAssumedPod = 3 * time.Minute
	// Harvester kubeconfig file name Pattern
	HarvesterConfPattern = "*.hconf"
	// KubeVirt kubeconfig file name Pattern
	KubeVirtConfPattern = "*.kconf"
)

// ErrConfigNotFound is returned when no Kubernetes configuration is found.
var ErrConfigNotFound = errors.New("kubernetes configuration not found")

var GpuModelToResourceName = map[string]string{
	"HL-225":       "habana.com/GAUDI2_AI_TRAINING_ACCELERATOR",
	"gpu-max-1100": "intel.com/PONTE_VECCHIO_XT_1_TILE_DATA_CENTER_GPU_MAX_1100",
}

// IsScalarResourceName validates the resource for Extended, Hugepages, Native and AttachableVolume resources
func IsScalarResourceName(name corev1.ResourceName) bool {
	for _, resourceName := range GpuModelToResourceName {
		if corev1.ResourceName(resourceName) == name {
			return true
		}
	}
	return false
}

// If bmKubeConfigDir is empty, then use the in-cluster config
func GetKubeRestConfig(bmKubeConfigDir string) (*rest.Config, error) {
	if bmKubeConfigDir != "" {
		fsys := os.DirFS(bmKubeConfigDir)
		configs, err := toolsk8s.LoadKubeConfigFiles(context.Background(), fsys, HarvesterConfPattern)
		if err != nil || len(configs) == 0 {
			return nil, ErrConfigNotFound
		}
		// Return the first kubeConfig found
		for _, value := range configs {
			return value, nil
		}
	}
	return loadInClusterConfig()
}

func loadInClusterConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, ErrConfigNotFound
	}
	return config, nil
}
