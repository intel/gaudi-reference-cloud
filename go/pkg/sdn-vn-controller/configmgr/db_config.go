// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package configmgr

type DbConfig struct {
	DbAvailable string `koanf:"dbAvailable"`
	Ipaddr      string `koanf:"dbAddress"`
	Dbport      string `koanf:"dbPort"`
	Dbname      string `koanf:"name"`
	Dbuser      string `koanf:"user"`
	Dbpasswd    string `koanf:"password"`
}
