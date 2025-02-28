// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"fmt"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	iks "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/kubernetes_provider/iks"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	IKSKubernetesProviderName = "iks"
)

type kubernetesProvider interface {
	// InitCluster ensures that all the prerequisites for the cluster are met before creating the controlplane nodes.
	// Every provider has different requirements for creating a cluster, we will handle all of that here.
	InitCluster(context.Context, *corev1.Secret, *privatecloudv1alpha1.Cluster, string, string, string, string, int, int, int) error
	// GetCluster gets the current status of a cluster.
	GetCluster(context.Context) (*privatecloudv1alpha1.ClusterStatus, error)
	// CleanUpCluster ensures a correct deletion of resources related to the cluster being deleted.
	CleanUpCluster(context.Context, string) error
	// GetNodes gets all the nodes and their state from a Kubernetes cluster.
	GetNodes(context.Context, string) ([]privatecloudv1alpha1.NodeStatus, error)
	// GetNode gets a node and its status from a Kubernetes cluster.
	GetNode(context.Context, string) (privatecloudv1alpha1.NodeStatus, error)
	// DeleteNode ensures a node is correctly deleted from a Kubernetes cluster.
	DeleteNode(context.Context, string) error
	// GetBootstrapScript returns the bootstrap script to use for node registration into a Kubernetes cluster
	// based on the nodegroup type.
	GetBootstrapScript(privatecloudv1alpha1.NodegroupType) (string, error)
	// CreateBootstrapTokenSecret creates a new bootstrap token used for worker node registration.
	CreateBootstrapTokenSecret(context.Context, *corev1.Secret) error
	// ApproveKubeletServingCertificateSigningRequests approves kubelet serving certificate signing requests of
	// nodes with a name that starts with the prefix provided.
	ApproveKubeletServingCertificateSigningRequests(context.Context, string) error
	// DrainNode cordons and drains a node, usually called before deletion.
	DrainNode(context.Context, string) error
	// Create a kubernetes secret.
	CreateSecret(context.Context, string, *corev1.Secret) error
	// Get a kubernetes secret.
	GetSecret(context.Context, string, string) (*corev1.Secret, error)
	// Create a kubernetes namespace.
	CreateNamespace(context.Context, *corev1.Namespace) error
	// Create a kubernetes storageclass.
	//CreateStorageClass(ctx context.Context, storageClass *storagev1.StorageClass) error
}

func newKubernetesProvider(provider string, config *Config, clusterDeleted bool, kubernetesClient *kubernetes.Clientset) (kubernetesProvider, error) {
	if provider == IKSKubernetesProviderName {
		iksProvider, err := iks.NewIKSProvider(
			config.KubernetesProviders.IKS.ControlplaneBootstrapScript,
			config.KubernetesProviders.IKS.WorkerBootstrapScript,
			clusterDeleted,
			kubernetesClient,
			config.CertExpirations.CaCertExpirationPeriod,
			config.CertExpirations.ControlPlaneCertExpirationPeriod,
		)
		if err != nil {
			return nil, err
		}
		return iksProvider, nil
	}

	return nil, fmt.Errorf("kubernetes provider not found")
}
