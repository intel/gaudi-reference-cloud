// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"time"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

var HealthMetric *prometheus.GaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "backend_health_status",
	Help: "Health status of the relative backend",
}, []string{"uuid"})

type HealthCheckResult struct {
	UUID         string
	HealthStatus backend.HealthStatus
}

type HealthController struct {
	backends map[string]backend.HealthInterface
	interval time.Duration
	results  chan HealthCheckResult
}

func NewHealthController(backends map[string]backend.HealthInterface, interval time.Duration) (*HealthController, error) {

	return &HealthController{
		backends: backends,
		interval: interval,
		results:  make(chan HealthCheckResult),
	}, nil
}

func (hc *HealthController) CheckHealth(ctx context.Context, backend backend.HealthInterface, uuid string) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	go func() {
		select {
		case <-ctx.Done():
			switch ctx.Err(){
			case context.DeadlineExceeded:
				hc.results <- HealthCheckResult{UUID: uuid, HealthStatus: 3}
			}
		}
	}()

	resp, err := backend.GetStatus(ctx)
	if err != nil {
		hc.results <- HealthCheckResult{UUID: uuid, HealthStatus: 3}
		return
	}
	log.Info().Ctx(ctx).Str("uuid", uuid).Stringer("status", resp.HealthStatus).Msg("Got status from the backend")
	hc.results <- HealthCheckResult{UUID: uuid, HealthStatus: resp.HealthStatus}
}

func (hc *HealthController) Start(ctx context.Context) {

	log.Info().Msg("Starting HealthController")
	if hc.interval <= 0 {
		log.Info().Msg("Assuming default healthcheck interval")
		hc.interval = 10 * time.Second
	}

	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for uuid, backend := range hc.backends {
				go hc.CheckHealth(ctx, backend, uuid)
			}
		case result := <-hc.results:
			HealthMetric.WithLabelValues(result.UUID).Set(float64(result.HealthStatus))
		case <-ctx.Done():
			return
		}
	}
}
