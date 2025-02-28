// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

type Config struct {
	SchedulerInterval uint16 `koanf:"schedulerInterval"`
	GitHubAPI         struct {
		Key string `koanf:"apiKey"`
	} `koanf:"githubConfig"`
	SecurityInsights struct {
		URL string `koanf:"url"`
	} `koanf:"securityInsightsService"`
	ThirdPartyComponents      []string                    `koanf:"thirdPartyComponents"`
	ThirdPartyComponentPolicy []ThirdPartyComponentPolicy `koanf:"thirdPartyComponentPolicy"`
}

type ThirdPartyComponentPolicy struct {
	ComponentName string `koanf:"componentName"`
	TopK          int    `koanf:"topK"`
	GitHubSource  string `koanf:"githubSource"`
	Policies      []struct {
		K8sVersions    string `koanf:"k8sVersion"`
		MinimumVersion string `koanf:"minVersion"`
	} `koanf:"policies"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		return &Config{SchedulerInterval: 15}
	}
	return Cfg
}
