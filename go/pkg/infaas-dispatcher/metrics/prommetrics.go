// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package metrics

import (
	"github.com/go-logr/logr"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "infaas_dispatcher"

type PromMetrics struct {
	OutstandingRequests          *prometheus.GaugeVec
	ConnectedAgents              *prometheus.GaugeVec
	FailedRequests               *prometheus.GaugeVec
	modelPendingRequestsCounters map[string]prometheus.GaugeFunc
	RequestsDurations            *prometheus.HistogramVec
	GrpcMetrics                  *grpc_prometheus.ServerMetrics
	constLabels                  prometheus.Labels
	namespace                    string
	logger                       logr.Logger
}

func NewPromMetrics(log logr.Logger, cfg config.DispatcherConfig, serviceName string) *PromMetrics {

	constLabels := prometheus.Labels{
		"service": serviceName,
	}
	commonLabels := []string{"operation", "model"}

	outstandingRequests := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "requests_active_total", // gauge of current requests
		Help:        "Number of requests currently being processed",
		Namespace:   namespace,
		ConstLabels: constLabels,
	}, commonLabels)

	failedRequests := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "requests_failed_total", // counter of failures
		Help:        "Total number of failed requests",
		Namespace:   namespace,
		ConstLabels: constLabels,
	}, append(commonLabels, "error")) // "error" instead of "error_label" for consistency

	connectedAgents := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "agents_connected", // current state metric
		Help:        "Number of currently connected agents",
		Namespace:   namespace,
		ConstLabels: constLabels,
	}, commonLabels)

	requestsDurations := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "request_duration_seconds", // uses base unit (seconds)
			Help:      "Time taken to handle requests in seconds",
			Namespace: namespace,
			Buckets:   []float64{0.1, 1, 2, 3, 5, 10, 15, 25, 35},
		},
		append(commonLabels, "error"),
	)

	rpcDurationHistogram := grpc_prometheus.WithServerHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets([]float64{1, 2, 3, 5, 10, 15, 25, 35}),
	)
	counterOptions := grpc_prometheus.WithServerCounterOptions(
		grpc_prometheus.WithConstLabels(prometheus.Labels{
			"service": serviceName,
		}),
	)
	grpcMetrics := grpc_prometheus.NewServerMetrics(rpcDurationHistogram, counterOptions)
	prometheus.MustRegister(grpcMetrics)

	return &PromMetrics{
		OutstandingRequests:          outstandingRequests,
		ConnectedAgents:              connectedAgents,
		FailedRequests:               failedRequests,
		modelPendingRequestsCounters: make(map[string]prometheus.GaugeFunc, len(cfg.SupportedModels)),
		RequestsDurations:            requestsDurations,
		GrpcMetrics:                  grpcMetrics,
		constLabels:                  constLabels,
		namespace:                    namespace,
		logger:                       log,
	}
}

func (p *PromMetrics) CreateGaugeForChannel(channelRequestsNumCallback func() float64, model string) {
	modelPendingRequestsCounter := promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:      "queue_length",
			Help:      "Current number of requests in the model's queue",
			Namespace: p.namespace,
			ConstLabels: prometheus.Labels{
				"service": p.constLabels["service"],
				"model":   model,
			},
		},
		channelRequestsNumCallback,
	)

	p.modelPendingRequestsCounters[model] = modelPendingRequestsCounter
}
