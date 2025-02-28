// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

type Config struct {
	ListenPort uint16 `koanf:"listenPort"`
}

func (cc *Config) GetListenPort() uint16 {
	return cc.ListenPort
}

func (cc *Config) SetListenPort(port uint16) {
	cc.ListenPort = port
}
