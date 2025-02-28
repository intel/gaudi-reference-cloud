// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package grpc_rest_gateway

// Configuration for a grpc_rest_gateway.
type Config struct {
	ListenPort uint16 `koanf:"listenPort"`
	// Format should be "localhost:30002"
	TargetAddr     string   `koanf:"targetAddr"`
	Deployment     string   `koanf:"deployment"`
	AllowedOrigins []string `koanf:"allowedOrigins"`
}
