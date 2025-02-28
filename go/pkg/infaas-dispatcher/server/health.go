// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/friendsofgo/errors"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"time"
)

func (d *Dispatcher) setupHealthChecks() error {
	d.log.Info("setting up health checks...")
	d.health = gosundheit.New(gosundheit.WithHealthListeners(healthUpdater{d.healthService}))
	err := d.health.RegisterCheck(
		&checks.CustomCheck{
			CheckName: "queues-length",
			CheckFunc: func(ctx context.Context) (details interface{}, err error) {
				totalLen := 0
				totalCapacity := 0
				for _, ch := range d.model2PendingRequests {
					totalLen += len(ch)
					totalCapacity += cap(ch)
				}

				if totalLen > totalLen/2 {
					return "max request queue length exceeded", fmt.Errorf("total request queue length (=%d) exceedes max allowed length (=%d)", totalLen, totalCapacity/2)
				}

				return "total request queue len is ok", nil
			},
		},
		gosundheit.ExecutionPeriod(1*time.Second),
		gosundheit.ExecutionTimeout(1*time.Second),
	)

	return errors.Wrap(err, "failed to register health checks")
}

type healthUpdater struct {
	healthService *health.Server
}

func (l healthUpdater) OnResultsUpdated(results map[string]gosundheit.Result) {
	for _, v := range results {
		if !v.IsHealthy() {
			l.healthService.SetServingStatus("ready", healthgrpc.HealthCheckResponse_NOT_SERVING)
			return
		}
	}

	l.healthService.SetServingStatus("ready", healthgrpc.HealthCheckResponse_SERVING)
	l.healthService.SetServingStatus("alive", healthgrpc.HealthCheckResponse_SERVING)
	l.healthService.SetServingStatus("startup", healthgrpc.HealthCheckResponse_SERVING)
}
