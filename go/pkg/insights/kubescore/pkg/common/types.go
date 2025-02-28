// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package common

import "time"

type ReleaseCmdOpts struct {
	Version        string
	OutputFormat   string
	ConfigFilepath string
	Config         RunConfig
	ListVersions   bool
	Distribution   string
}

type RecommendCmdOpts struct {
	CurrentVersion string
	OutputFormat   string
	ConfigFilepath string
	Config         RunConfig
}

const (
	GITHUB_API_KEY = "GITHUB_API_KEY"
	K8sRepoUrl     = "https://github.com/kubernetes/kubernetes"
	RKEReporURL    = "https://github.com/rancher/rke2"
)

type RunConfig struct {
	GitHub struct {
		APIKey string `yaml:"apiKey"`
	} `yaml:"githubConfig"`

	VulnerabilityScannner ScannerConfig `yaml:"vulnerabilityScannerConfig"`

	Database RunConfigDB `yaml:"db"`
}

type ScannerConfig struct {
	Snyk struct {
		Endpoint  string `yaml:"endpoint"`
		AuthToken string `yaml:"authToken"`
	} `yaml:"snyk"`
}

type RunConfigDB struct {
	Redis struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		DB       int    `yaml:"dbIdx"`
	} `yaml:"redis"`
}

// ReleaseMD :
type ReleaseMD struct {
	Tag       string    `json:"release_tag"`
	CreatedAt time.Time `json:"created_at"`
	CommitID  string    `json:"commit_id"`
	Name      string    `json:"name"`
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	License   string    `json:"license"`
}

type ReleaseReport struct {
	ReleaseTag  string           `json:"releaseTag"`
	ReleaseTime time.Time        `json:"releaseTime"`
	ReleaseMD   ReleaseMD        `json:"releaseMD"`
	SupportMD   ReleaseSupportMD `json:"supportMD"`
	Images      []ImageReport    `json:"images"`
}

type ImageReport struct {
	URL             string            `json:"url"`
	Digest          string            `json:"digest"`
	CreatedAt       string            `json:"createdAt"`
	Vulnerabilities VulnerabilityData `json:"vulnerabilities"`
}

type VulnerabilityData struct {
	Summary VulnerabilitySummary `json:"vulnerabilities"`
}

type VulnerabilitySummary struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

type ReleaseAsset struct {
	ReleaseID   string
	Name        string
	Type        string
	DownloadURL string
}

type RecommendationReport struct {
	CurrentRelease     string `json:"currentRelease"`
	RecommendedRelease string `json:"recommendedRelease"`
	ReleaseLagTime     string `json:"releaseLagTime"`
	ReleaseLagSpace    int    `json:"releaseLagSpace"`
}

type ReleaseSupportMD struct {
	Eol string `json:"eol"`
	Eos string `json:"support"`
	Lts bool   `json:"lts"`

	EOLTime time.Time
	EOSTime time.Time
}

type ReleaseComponentMD struct {
	ReleaseId        string
	ComponentName    string
	ComponentVersion string
	ReleaseTime      time.Time
	License          string
	Purl             string
	Type             string
}
