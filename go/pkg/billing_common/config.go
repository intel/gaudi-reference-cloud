// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

type CommonConfig struct {
	MaxDefaultHistory    int `koanf:"maxDefaultHistory"`
	InstanceSearchWindow int `koanf:"instanceSearchWindow"`
}
