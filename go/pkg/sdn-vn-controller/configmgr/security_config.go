// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package configmgr

type SecurityConfig struct {
	SqlSsl  SslConfig `koanf:"postgreSqlSSL"`
	OvnSsl  SslConfig `koanf:"ovnCentralSSL"`
	GrpcSsl SslConfig `koanf:"grpcServerSSL"`
}

type SslConfig struct {
	Enabled string `koanf:"enabled"`
	Verify  string `koanf:"verify"`
	Ca      string `koanf:"cA"`
	Cert    string `koanf:"certificate"`
	Key     string `koanf:"privateKey"`
}
