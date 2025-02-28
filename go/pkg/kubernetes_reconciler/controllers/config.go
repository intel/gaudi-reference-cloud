// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controllers

import (
	"time"
)

type Config struct {
	PollPeriodicity                      time.Duration `json:"pollPeriodicity"`
	IKSAPIAddress                        string        `json:"iksAPIAddress"`
	ClusterMaxConcurrentReconciles       int           `json:"clusterMaxConcurrentReconciles"`
	ClusterSecretMaxConcurrentReconciles int           `json:"clusterSecretMaxConcurrentReconciles"`
}
