// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"context"
	"fmt"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetKubernetesClient returns a new client based on the kubeconfig received as a parameter.
func GetKubernetesClient(kubeconfig []byte) (*kubernetes.Clientset, error) {
	rc, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func GetKubernetesRestConfig(host, userAgent string, caCert, cert, key []byte) *rest.Config {
	return &rest.Config{
		Host:    host,
		Timeout: time.Second * 60,
		TLSClientConfig: rest.TLSClientConfig{
			CertData: cert,
			KeyData:  key,
			CAData:   caCert,
		},
		UserAgent: userAgent,
	}
}

func GetKubernetesClientFromConfig(restConfig *rest.Config) (*kubernetes.Clientset, error) {
	// make config non-cacheable
	restConfig.Proxy = http.ProxyFromEnvironment
	httpClient, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, err
	}
	utilnet.CloseIdleConnectionsFor(httpClient.Transport)

	return clientset, nil
}

// GetKubernetesAPIHealthEndpoint calls any of the kubernetes api health endpoints, readyz, livez or (deprecated) healthz.
func GetKubernetesAPIHealthEndpoint(ctx context.Context, endpoint string, restClient rest.Interface) error {
	result := restClient.Get().AbsPath(endpoint).Do(ctx)
	if result.Error() != nil {
		return result.Error()
	}

	var statusCode int
	result.StatusCode(&statusCode)
	if statusCode != http.StatusOK {
		return fmt.Errorf("endpoint didn't return 200 status code")
	}

	return nil
}

// GetSecret gets a kubernetes secret.
func GetSecret(ctx context.Context, client client.Client, secretName string, namespace string) (corev1.Secret, error) {
	var clusterSecret corev1.Secret
	if err := client.Get(ctx, k8stypes.NamespacedName{
		Namespace: namespace,
		Name:      secretName,
	}, &clusterSecret); err != nil {
		return clusterSecret, err
	}

	return clusterSecret, nil
}

// GetDataFromSecret given a key, returns its value from the secret data.
func GetDataFromSecret(clusterSecret corev1.Secret, keyData string) ([]byte, error) {
	valueData, ok := clusterSecret.Data[keyData]
	if !ok {
		return valueData, fmt.Errorf("can not get %s from secret %s", keyData, clusterSecret.Name)
	}

	return valueData, nil
}
