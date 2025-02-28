// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func LinearBackoff(durationTotal time.Duration, durationBetweenAttempts time.Duration) wait.Backoff {
	return wait.Backoff{
		Duration: durationBetweenAttempts,
		Factor:   1.0,
		Jitter:   0,
		Steps:    int(durationTotal.Milliseconds() / durationBetweenAttempts.Milliseconds()),
	}
}
