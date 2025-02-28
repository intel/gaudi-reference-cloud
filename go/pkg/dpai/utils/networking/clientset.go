// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package networking

import (
	"log"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// regsiter cert-manager schema with controller runtime
func GetCertManagerClient(clientConfig *rest.Config) (client.Client, error) {
	log.Printf("Registering the Controller Runtime client with cert-manager client schema with client config %v", clientConfig)

	cManagerScheme := runtime.NewScheme()

	if err := certmanagerv1.AddToScheme(cManagerScheme); err != nil {
		log.Printf("Error registering cert-manager client schema with controller runtime. Error message: %+v", err)
		return nil, err
	}

	cManagerClient, err := client.New(clientConfig, client.Options{Scheme: cManagerScheme})
	if err != nil {
		log.Fatalf("Failed to create cert-manager client: %+v", err)
		return nil, err
	}

	return cManagerClient, nil
}

func GetIksAzClusterClient(azClientSet *k8s.K8sAzClient) (client.Client, error) {
	log.Println("Registering the Controller Runtime client for the IKS cluster Lb Resource private.cloud.intel.com/v1alpha1, kind: Loadbalancer")
	loadBalancerSchema := runtime.NewScheme()

	err := loadbalancerv1alpha1.AddToScheme(loadBalancerSchema)

	if err != nil {
		return nil, err
	}

	loadbalancerClient, err := client.New(azClientSet.ClientConfig, client.Options{Scheme: loadBalancerSchema})
	if err != nil {
		log.Fatalf("Failed to create load-balancer client in controller-runtime: %+v", err)
		return nil, err
	}

	return loadbalancerClient, nil
}
