// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package configmgr

type OvnConfig struct {
	Ipaddr  string `koanf:"ovnAddress"`
	Ovnport string `koanf:"ovnPort"`
}
