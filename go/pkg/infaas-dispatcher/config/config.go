// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import "time"

type DispatcherConfig struct {
	ListenPort                   uint16
	SupportedModels              []string
	BacklogSize                  uint32
	DefaultGenerateStreamTimeout *time.Duration
	MetricsPort                  uint16
}
