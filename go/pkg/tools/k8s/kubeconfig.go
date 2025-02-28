// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func LoadKubeConfigFile(ctx context.Context, filePath string) (*restclient.Config, error) {
	configBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("LoadKubeConfigFile: ReadFile: unable to read %s: %w", filePath, err)
	}
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
	if err != nil {
		return nil, fmt.Errorf("LoadKubeConfigFile: ReadFile: unable to parse KubeConfig file %s: %w", filePath, err)
	}
	return restConfig, nil
}

// Return a restclient.Config for multiple contexts.
// This uses the default loading rules, which means it will read all files in the KUBECONFIG environment variable.
func LoadKubeConfigContexts(ctx context.Context, contexts []string) (map[string]*restclient.Config, error) {
	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return nil, err
	}
	restConfigs := make(map[string]*restclient.Config)
	for _, context := range contexts {
		config.CurrentContext = context
		restConfig, err := clientcmd.NewDefaultClientConfig(*config, nil).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", context, err)
		}
		restConfigs[context] = restConfig
	}
	return restConfigs, nil
}

// Load all KubeConfig files from a directory.
// Returns a map from the relative KubeConfig file name.
func LoadKubeConfigFiles(ctx context.Context, fsys fs.FS, filenamePattern string) (map[string]*restclient.Config, error) {
	log := log.FromContext(ctx).WithName("LoadKubeConfigFiles")
	log.Info("BEGIN")
	filenames, err := fs.Glob(fsys, filenamePattern)
	if err != nil {
		return nil, fmt.Errorf("LoadKubeConfigFiles: Glob: %w", err)
	}
	log.V(1).Info("Glob", "filenames", filenames)
	restConfigs := make(map[string]*restclient.Config)
	for _, filename := range filenames {
		fileInfo, err := fs.Stat(fsys, filename)
		if err != nil {
			return nil, err
		}
		if fileInfo.IsDir() {
			continue
		}
		log.V(0).Info("Read", "filename", filename)
		reader, err := fsys.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("LoadKubeConfigFiles: Open: %w", err)
		}
		defer reader.Close()
		configBuffer := new(bytes.Buffer)
		if _, err = configBuffer.ReadFrom(reader); err != nil {
			return nil, fmt.Errorf("LoadKubeConfigFiles.readFileMessage: ReadFrom: %w", err)
		}
		if err := reader.Close(); err != nil {
			return nil, fmt.Errorf("LoadKubeConfigFiles: Close: %w", err)
		}
		restConfig, err := clientcmd.RESTConfigFromKubeConfig(configBuffer.Bytes())
		if err != nil {
			return nil, fmt.Errorf("LoadKubeConfigFiles: RESTConfigFromKubeConfig: %w", err)
		}
		log.V(0).Info("Config", "filename", filename, "restConfig", restConfig)
		restConfigs[filename] = restConfig
	}
	log.Info("END")
	return restConfigs, nil
}

// Write a set of KubeConfig files to a directory.
// Map keys will be the filenames.
// Inverse of LoadKubeConfigFiles.
func WriteKubeConfigFiles(ctx context.Context, directory string, restConfigs map[string]*restclient.Config) error {
	for filename, restConfig := range restConfigs {
		kubeConfigBytes, err := KubeConfigFromRESTConfig(restConfig)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, filename), kubeConfigBytes, 0600); err != nil {
			return err
		}
	}
	return nil
}

// Convert a rest.Config to KubeConfig file contents.
// Inverse of https://pkg.go.dev/k8s.io/client-go/tools/clientcmd#RESTConfigFromKubeConfig
func KubeConfigFromRESTConfig(restConfig *restclient.Config) ([]byte, error) {
	context := "default-context"
	cluster := "default-cluster"
	user := "user"
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters[cluster] = &clientcmdapi.Cluster{
		Server:                   restConfig.Host,
		CertificateAuthorityData: restConfig.CAData,
	}
	contexts := make(map[string]*clientcmdapi.Context)
	contexts[context] = &clientcmdapi.Context{
		Cluster:  cluster,
		AuthInfo: user,
	}
	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos[user] = &clientcmdapi.AuthInfo{
		ClientCertificateData: restConfig.CertData,
		ClientKeyData:         restConfig.KeyData,
	}
	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: context,
		AuthInfos:      authinfos,
	}
	return clientcmd.Write(clientConfig)
}
