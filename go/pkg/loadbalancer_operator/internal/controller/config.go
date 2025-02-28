// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

type LoadbalancerProviderConfig struct {
	// URL of the api endpoint for LB API
	BaseURL string `json:"baseURL"`

	// Type of LB provider.
	ProviderType string `json:"providerType"`
	Domain       string `json:"domain"`

	// UsernameFile is a file containing the username used to connect the API.
	UsernameFile string `json:"userNameFile"`

	// PasswordFile is a file containing the password used to connect the API.
	PasswordFile string `json:"passwordFile"`

	// PasswordFile is a file containing the password used to connect the API.
	AZClusterKubeconfigFile string `json:"AZClusterKubeconfigFile"`

	// LB environment to use for requests.
	Environment int `json:"environment"`

	// LB usergroup to use for requests.
	UserGroup int `json:"usergroup"`

	LoadbalancerMaxConcurrentReconciles int `json:"loadbalancerMaxConcurrentReconciles"`

	// The region in which the operator is deployed.
	RegionId string `json:"regionId"`

	// The availability zone in which the operator is deployed.
	AvailabilityZoneId string `json:"availabilityZoneId"`
}

type LoadbalancerPoolConfig struct {
}
