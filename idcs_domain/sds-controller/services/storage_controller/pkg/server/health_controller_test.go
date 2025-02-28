// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	_ "reflect"
	"testing"
	"time"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/stretchr/testify/assert"
)

type MockBackend struct {
	backend.HealthInterface
	health backend.HealthStatus
	sleep  time.Duration
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		health: backend.Healthy,
		sleep:  0 * time.Second,
	}
}

func NewUnhealthyMockBackend() *MockBackend {
	return &MockBackend{
		health: backend.Unhealthy,
		sleep:  0 * time.Second,
	}
}

func NewDegragedMockBackend() *MockBackend {
	return &MockBackend{
		health: backend.Degraded,
		sleep:  0 * time.Second,
	}
}

func NewSlowMockBackend() *MockBackend {
	return &MockBackend{
		health: backend.Healthy,
		sleep:  10 * time.Second,
	}
}

func (m *MockBackend) GetStatus(ctx context.Context) (*backend.ClusterStatus, error) {
	time.Sleep(m.sleep)
	return &backend.ClusterStatus{
		HealthStatus: m.health,
	}, nil
}

const TEST_UUID = "0000-0000-0000-0000"

func TestNewHealthController(t *testing.T) {
	mockBackend := NewMockBackend()
	tests := []struct {
		name      string
		backends  map[string]backend.HealthInterface
		interval  time.Duration
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "Creation of new HealthController",
			interval:  2 * time.Second,
			backends:  map[string]backend.HealthInterface{TEST_UUID: mockBackend},
			assertion: assert.NoError,
		},
		{
			name:      "Creation of new healthcontroller with 0 interval",
			interval:  0 * time.Second,
			backends:  map[string]backend.HealthInterface{TEST_UUID: mockBackend},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHealthController(tt.backends, tt.interval)
			tt.assertion(t, err)
		})
	}
}

func TestHealthController_CheckHealth(t *testing.T) {
	healthyBackend := NewMockBackend()
	unhealthyBackend := NewUnhealthyMockBackend()
	degradedBackend := NewDegragedMockBackend()
	slowBackend := NewSlowMockBackend()

	tests := []struct {
		name    string
		backend backend.HealthInterface
		uuid    string
		want    backend.HealthStatus
	}{
		{
			name:    "Healthy Backend",
			backend: healthyBackend,
			uuid:    TEST_UUID,
			want:    backend.Healthy,
		},
		{
			name:    "Unhealthy Backend",
			backend: unhealthyBackend,
			uuid:    TEST_UUID,
			want:    backend.Unhealthy,
		},
		{
			name:    "Degraded Backend",
			backend: degradedBackend,
			uuid:    TEST_UUID,
			want:    backend.Degraded,
		},
		{
			name:    "Slow Backend",
			backend: slowBackend,
			uuid:    TEST_UUID,
			want:    backend.Unhealthy,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backends := map[string]backend.HealthInterface{tt.uuid: tt.backend}
			results := make(chan HealthCheckResult)
			ctx := context.Background()
			hc := &HealthController{
				backends: backends,
				interval: 2 * time.Second,
				results:  results,
			}
			go hc.CheckHealth(ctx, tt.backend, tt.uuid)
			result := <-results
			if result.HealthStatus != tt.want {
				t.Errorf("CheckHealth() = %v, want %v", result, tt.want)
			}
		})
	}
}
