// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import "time"

type LoadBalancerProviderConfig struct {
	BaseURL                    string        `json:"baseURL"`
	ProviderType               string        `json:"providerType"`
	Domain                     string        `json:"domain"`
	UserName                   string        `json:"userName"`
	Secret                     string        `json:"secret"`
	IlbMaxConcurrentReconciles int           `json:"ilbMaxConcurrentReconciles"`
	HighwireTimeout            time.Duration `json:"highwireTimeout"`
}
