// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iks

import (
	"text/template"
)

const (
	KubeconfigTemplateText = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {{.Ca}}
    server: https://{{.Server}}:{{.Port}}
  name: {{.ClusterName}}
contexts:
- context:
    cluster: {{.ClusterName}}
    user: {{.User}}
  name: default
current-context: "default"
kind: Config
preferences: {}
users:
- name: {{.User}}
  user:
    client-certificate-data: {{.Cert}}
    client-key-data: {{.Key}}
`
)

var (
	KubeconfigTemplate = template.Must(template.New("kubeconfig").Parse(KubeconfigTemplateText))
)

type ClusterTemplateStruct struct {
	APIVersion string `yaml:"apiVersion"`
	Clusters   []struct {
		Cluster struct {
			CertificateAuthorityData string `yaml:"certificate-authority-data"`
			Server                   string `yaml:"server"`
		} `yaml:"cluster"`
		Name string `yaml:"name"`
	} `yaml:"clusters"`
	Contexts []struct {
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
	CurrentContext string `yaml:"current-context"`
	Kind           string `yaml:"kind"`
	Preferences    struct {
	} `yaml:"preferences"`
	Users []struct {
		Name string `yaml:"name"`
		User struct {
			ClientCertificateData string `yaml:"client-certificate-data"`
			ClientKeyData         string `yaml:"client-key-data"`
		} `yaml:"user"`
	} `yaml:"users"`
}

type KubeconfigTemplateConfig struct {
	ClusterName string
	User        string
	Server      string
	Port        string
	Ca          string
	Cert        string
	Key         string
}
