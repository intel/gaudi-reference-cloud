// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package health

import (
	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

type GrpcHealthListener struct {
	healthServer *health.Server
}

func NewGrpcHealthListener(healthServer *health.Server) *GrpcHealthListener {
	return &GrpcHealthListener{
		healthServer: healthServer,
	}
}

func (g *GrpcHealthListener) OnResultsUpdated(results map[string]gosundheit.Result) {
	for _, v := range results {
		if !v.IsHealthy() {
			g.healthServer.SetServingStatus(pb.MaasGateway_ServiceDesc.ServiceName, healthgrpc.HealthCheckResponse_NOT_SERVING)
			return
		}
	}

	g.healthServer.SetServingStatus(pb.MaasGateway_ServiceDesc.ServiceName, healthgrpc.HealthCheckResponse_SERVING)
}
