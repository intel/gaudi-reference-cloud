// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_replicator/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_replicator/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type StorageReplicatorService struct {
}

func (svc *StorageReplicatorService) Init(ctx context.Context, cfg *config.Config) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("StorageReplicatorService.Init")
	log.Info("initializing IDC storage replicator service...")

	var err error
	// Load Kubernetes client configuration
	var defaultKubeRestConfig *rest.Config
	if defaultKubeRestConfig, err = conf.GetKubeRestConfig(); err != nil {
		defaultKubeRestConfig = &rest.Config{}
	}
	log.V(9).Info("DefaultKubeRestConfig", logkeys.Configuration, defaultKubeRestConfig)

	fileReplicator, err := controller.NewStorageReplicatorService(ctx, cfg, defaultKubeRestConfig)
	if err != nil {
		log.Error(err, "error starting file system replicator scheduler")
		return err
	}
	fileReplicator.StartStorageReplicationScheduler(ctx)
	return nil
}

func (svc *StorageReplicatorService) Name() string {
	return "idc-storage-replicator"
}

func initKubeRestClient(kubeConfig *rest.Config) (*rest.RESTClient, error) {
	err := v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("error adding to scheme: %v", err)
	}

	kubeConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupName, Version: "v1alpha1"}
	kubeConfig.APIPath = "/apis"
	kubeConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	kubeConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	restClient, err := rest.UnversionedRESTClientFor(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating kube client: %v", err)
	}
	return restClient, nil
}
