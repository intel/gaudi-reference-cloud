// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package metrics

import (
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PromMetrics struct {
	OutstandingRequests *prometheus.GaugeVec
	FailedRequests      *prometheus.GaugeVec
	RequestsDurations   *prometheus.HistogramVec
	constLabels         prometheus.Labels
	logger              logr.Logger
}

func NewPromMetrics(log logr.Logger, serviceName string, namespace string) *PromMetrics {

	constLabels := prometheus.Labels{
		"service": serviceName,
	}

	commonLabels := []string{"operation", "model"}

	outstandingRequests := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "requests_active_total", // gauge of current requests
		Help:        "Number of requests currently being processed",
		ConstLabels: constLabels,
		Namespace:   namespace,
	}, commonLabels)

	failedRequests := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "requests_failed_total", // counter of failures
		Help:        "Total number of failed requests",
		ConstLabels: constLabels,
		Namespace:   namespace,
	}, append(commonLabels, "error")) // "error" instead of "error_label" for consistency

	requestsDurations := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "request_duration_seconds", // uses base unit (seconds)
			Help:      "Time taken to handle requests in seconds",
			Buckets:   []float64{0.1, 1, 2, 3, 5, 10, 15, 25, 35, 60},
			Namespace: namespace,
		},
		append(commonLabels, "error"),
	)

	return &PromMetrics{
		OutstandingRequests: outstandingRequests,
		FailedRequests:      failedRequests,
		RequestsDurations:   requestsDurations,
		constLabels:         constLabels,
		logger:              log,
	}
}
