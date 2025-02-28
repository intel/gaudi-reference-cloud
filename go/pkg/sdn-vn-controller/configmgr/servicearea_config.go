// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package configmgr

type ServiceAreaConfig struct {
	RegionId string   `koanf:"regionID"`
	AzId     []string `koanf:"availabilityZones"`
}
