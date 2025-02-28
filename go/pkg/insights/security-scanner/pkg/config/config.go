// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

type Config struct {
	SchedulerInterval uint16 `koanf:"schedulerInterval"`
	GithubKey         string `koanf:"githubKey"`
	SecurityInsights  struct {
		URL string `koanf:"url"`
	} `koanf:"securityInsightsService"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		return &Config{SchedulerInterval: 15}
	}
	return Cfg
}
