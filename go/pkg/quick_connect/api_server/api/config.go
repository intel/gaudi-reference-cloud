// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/**
 * Configurations used by quick_connect.
 *
 **/

package api

import (
	"time"
)

// Application configuration
type Config struct {
	ComputeApiServerAddress string `koanf:"computeApiServerAddr"`
	// ListenPort is the port the quick connect proxy is listening on
	ListenPort uint16 `koanf:"listenPort"`
	// TargetPort is the port the backend is listening on
	TargetPort        uint16            `koanf:"targetPort"`
	ClientCertificate ClientCertificate `koanf:"clientCertificate"`
}

type ClientCertificate struct {
	// CommonName is the CN of the client certificate provided when connecting to a target
	CommonName string `koanf:"commonName"`
	// TTL is the TTL of the client certificate, such as "2m"
	TTL time.Duration `koanf:"ttl"`
}
