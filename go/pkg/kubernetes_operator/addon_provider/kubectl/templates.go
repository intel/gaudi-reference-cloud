// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package kubectl

import (
	"fmt"
	"text/template"
)

const (
	KubeProxyTemplateConfigName         = "kube-proxy"
	CorednsTemplateConfigName           = "coredns"
	CalicoTemplateConfigName            = "calico-config"
	KonnectivityAgentTemplateConfigName = "konnectivity-agent"
	WekaStorageclassTemplateConfigName  = "weka-storageclass"

	KubeProxyClusterCIDRKey = "clusterCIDR"
	KubeProxyClusterVIPKey  = "clusterVIP"

	CorednsClusterIPKey = "clusterIP"

	CalicoClusterCIDRKey = "clusterCIDR"

	KonnectivityProxyServerHostKey = "proxyServerHost"
	KonnectivityProxyServerPortKey = "proxyServerPort"

	WekaStorageclassReclaimPolicyKey           = "reclaimPolicy"
	WekaStorageclassFilesystemGroupNameKey     = "filesystemGroupName"
	WekaStorageclassInitialFilesystemSizeGBKey = "initialFilesystemSizeGB"
	WekaStorageclassSecretNameKey              = "secretName"
	WekaStorageclassSecretNamespaceKey         = "secretNamespace"
)

type kubeProxyTemplateConfig struct {
	ClusterCIDR string
	ClusterVIP  string
}

type corednsTemplateConfig struct {
	ClusterIP string
}

type calicoTemplateConfig struct {
	ClusterCIDR string
}

type konnectivityAgentTemplateConfig struct {
	ProxyServerHost string
	ProxyServerPort string
}

type wekaStorageclassTemplateConfig struct {
	ReclaimPolicy           string
	FilesystemGroupName     string
	InitialFilesystemSizeGB string
	SecretName              string
	SecretNamespace         string
}

type defaultTemplateConfig struct{}

func getTemplateConfig(clusterName string, addonName string, args map[string]string) (any, error) {
	if addonName == clusterName+"-"+WekaStorageclassTemplateConfigName {
		var tempConfig wekaStorageclassTemplateConfig

		reclaimPolicy, ok := args[WekaStorageclassReclaimPolicyKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", WekaStorageclassReclaimPolicyKey)
		}

		filesystemGroupName, ok := args[WekaStorageclassFilesystemGroupNameKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", WekaStorageclassFilesystemGroupNameKey)
		}

		initialFilesystemSizeGB, ok := args[WekaStorageclassInitialFilesystemSizeGBKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", WekaStorageclassInitialFilesystemSizeGBKey)
		}

		secretName, ok := args[WekaStorageclassSecretNameKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", WekaStorageclassSecretNameKey)
		}

		secretNamespace, ok := args[WekaStorageclassSecretNamespaceKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", WekaStorageclassSecretNamespaceKey)
		}

		tempConfig.ReclaimPolicy = reclaimPolicy
		tempConfig.FilesystemGroupName = filesystemGroupName
		tempConfig.InitialFilesystemSizeGB = initialFilesystemSizeGB
		tempConfig.SecretName = secretName
		tempConfig.SecretNamespace = secretNamespace

		return tempConfig, nil
	}

	if addonName == clusterName+"-"+KubeProxyTemplateConfigName {
		var tempConfig kubeProxyTemplateConfig

		clusterCIDR, ok := args[KubeProxyClusterCIDRKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", KubeProxyClusterCIDRKey)
		}
		tempConfig.ClusterCIDR = clusterCIDR

		clusterVIP, ok := args[KubeProxyClusterVIPKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", KubeProxyClusterVIPKey)
		}
		tempConfig.ClusterVIP = clusterVIP
		return tempConfig, nil
	}

	if addonName == clusterName+"-"+CorednsTemplateConfigName {
		var tempConfig corednsTemplateConfig

		clusterIP, ok := args[CorednsClusterIPKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", CorednsClusterIPKey)
		}
		tempConfig.ClusterIP = clusterIP
		return tempConfig, nil
	}

	if addonName == clusterName+"-"+CalicoTemplateConfigName {
		var tempConfig calicoTemplateConfig

		clusterCIDR, ok := args[CalicoClusterCIDRKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", CalicoClusterCIDRKey)
		}
		tempConfig.ClusterCIDR = clusterCIDR
		return tempConfig, nil
	}

	if addonName == clusterName+"-"+KonnectivityAgentTemplateConfigName {
		var tempConfig konnectivityAgentTemplateConfig

		proxyServerHost, ok := args[KonnectivityProxyServerHostKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", KonnectivityProxyServerHostKey)
		}
		tempConfig.ProxyServerHost = proxyServerHost

		proxyServerPort, ok := args[KonnectivityProxyServerPortKey]
		if !ok {
			return defaultTemplateConfig{}, fmt.Errorf("required arg not found: %s", KonnectivityProxyServerPortKey)
		}
		tempConfig.ProxyServerPort = proxyServerPort

		return tempConfig, nil
	}

	return defaultTemplateConfig{}, nil
}

func getTemplate(templateName, templateText string) (*template.Template, error) {
	parsedTemplate, err := template.New(templateName).Parse(templateText)
	if err != nil {
		return nil, err
	}

	return parsedTemplate, nil
}
