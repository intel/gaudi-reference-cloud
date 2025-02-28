// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package k8s

import (
	"context"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sAzClient struct {
	ClientConfig *rest.Config
	ClientSet    *kubernetes.Clientset
	ctx          context.Context
	cancelFunc   context.CancelFunc
}

const (
	AZ_KUBE_PATH = "/vault/secrets/az"
)

func (k *K8sAzClient) ConfigureK8sClient() (*K8sAzClient, error) {
	log.Println("Initializing the K8s Client for the IKS AZ Cluster")
	// read the az cluster mounted by vault sidecar agent injector in the pod fs.
	config, err := clientcmd.BuildConfigFromFlags("", AZ_KUBE_PATH)
	if err != nil {
		log.Fatalf("Cannot find the kubeconfig File for AZ")
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)

	if err != nil {
		log.Fatalf("error generating the clientSet for the AZ Cluster KubeConfig")
		return nil, err
	}
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)

	return &K8sAzClient{
		ClientConfig: config,
		ClientSet:    clientSet,
		ctx:          ctx,
		cancelFunc:   cancelFunc,
	}, nil
}
