// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Import this package to enable gRPC logging.
// Set environment variables GRPC_GO_LOG_SEVERITY_LEVEL and GRPC_GO_LOG_VERBOSITY_LEVEL.
// All logs are written to stderr.
package grpclog

import (
	_ "google.golang.org/grpc/grpclog"
)
