// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package controller

type FirewallRuleProviderConfig struct {

	// BaseURL is the URL of the Firewall API
	BaseURL string `json:"baseURL"`

	// UsernameFile is a file containing the username used to connect the API.
	UsernameFile string `json:"userNameFile"`

	// UsernPasswordFileameFile is a file containing the password used to connect the API.
	PasswordFile string `json:"passwordFile"`

	// Environment informs if this is Prod or Stag
	Environment string `json:"environment"`

	// Region informs which IDC datacenter to target
	Region                  string `json:"region"`
	MaxConcurrentReconciles int    `json:"maxConcurrentReconciles"`
}
