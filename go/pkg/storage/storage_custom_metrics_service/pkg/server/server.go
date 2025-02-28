// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	reg *prometheus.Registry
	// Weka
	metricIKS     *prometheus.GaugeVec
	metricFS      *prometheus.GaugeVec
	metricTotalFS *prometheus.GaugeVec
	metricFSRatio *prometheus.GaugeVec
	metricFSCount *prometheus.GaugeVec
	metricNSUsed  *prometheus.GaugeVec
	metricNSUsage *prometheus.GaugeVec
	metricNSRatio *prometheus.GaugeVec
	// Vast
	metricVastNSUsed  *prometheus.GaugeVec
	metricVastNSUsage *prometheus.GaugeVec
	metricVastNSRatio *prometheus.GaugeVec
	metricVastFSCount *prometheus.GaugeVec
	metricVastFSRatio *prometheus.GaugeVec
	metricVastFS      *prometheus.GaugeVec
	metricVastFSUsage *prometheus.GaugeVec
	// Minio
	metricBkCount               *prometheus.GaugeVec
	metricTotalBK               *prometheus.GaugeVec
	metricBkUsage               *prometheus.GaugeVec
	metricBkSize                *prometheus.GaugeVec
	metricClusterUsage          *prometheus.GaugeVec
	metricClusterSpaceTotal     *prometheus.GaugeVec
	metricClusterSpaceAvailable *prometheus.GaugeVec
	metricMinioAllocated        *prometheus.GaugeVec
	metricClusterCount          *prometheus.GaugeVec
	metricFSUsage               *prometheus.GaugeVec

	metrics []*prometheus.GaugeVec
)

type Service struct {
}

type Config struct {
	ListenPort               uint16 `koanf:"listenPort"`
	StorageControllerAddr    string `koanf:"storageControllerAddr"`
	StorageKmsAddr           string `koanf:"storageKmsServerAddr"`
	StorageControllerUseMtls bool   `koanf:"storageControllerServerUseMtls"`
	Interval                 uint   `koanf:"metricUpdateIntervalMinutes"`
	ObjectMetricEnabled      bool   `koanf:"bucketMetricEnabled"`
	StorageAPIEndpoint       string `koanf:"storageAPIServerAddr"`

	TestMode bool
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (svc *Service) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("Service.Init")

	log.Info("initializing IDC Storage Custom Metrics Service...")

	strCntClient := storagecontroller.StorageControllerClient{}
	if err := strCntClient.Init(ctx, cfg.StorageControllerAddr, cfg.StorageControllerUseMtls); err != nil {
		log.Error(err, "error initializing storage controller client, exiting...")
		return fmt.Errorf("pre-conditioned failed")
	}
	log.Info("successfully initialized storage controller client ")

	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}

	log.Info("debug info: storageKms Address", logkeys.ServerAddr, cfg.StorageKmsAddr)
	// Connect to storage kms service.
	kmsclientConn, err := grpcutil.NewClient(ctx, cfg.StorageKmsAddr, dialOptions...)
	if err != nil {
		log.Error(err, "unable to obtain connection for storage kms", logkeys.ServerAddr, cfg.StorageKmsAddr)
		return fmt.Errorf("storage controller server grpc dial failed")
	}
	log.Info("api-server address", logkeys.ServerAddr, cfg.StorageAPIEndpoint)
	storageClientConn, err := grpcutil.NewClient(ctx, cfg.StorageAPIEndpoint)
	if err != nil {
		log.Error(err, "error creating storage service client")
		return err
	}
	storageKmsClient := pb.NewStorageKMSPrivateServiceClient(kmsclientConn)
	storageClient := pb.NewFilesystemPrivateServiceClient(storageClientConn)
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := storageClient.PingPrivate(pingCtx, &emptypb.Empty{}); err != nil {
		log.Error(err, "unable to ping storage service")
		return err
	}
	storageMonitor, err := NewStorageCustomMetricService(&strCntClient, storageKmsClient, storageClient, cfg.ObjectMetricEnabled)
	if err != nil {
		log.Error(err, "error starting storage custom metrics service")
		return err
	}
	log.Info("storage custom metrics service initialized")

	// Create a metrics registry.
	reg = prometheus.NewRegistry()

	// Weka
	metricNSUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "weka_cluster_org_used",
		Help: "cluster namespace utilization",
	}, []string{"cluster"})

	metricNSRatio = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "weka_cluster_org_used_ratio",
		Help: "cluster namespace utilization ratio",
	}, []string{"cluster"})

	metricNSUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "weka_org_space_usage",
		Help: "cluster namespace space utilization",
	}, []string{"cluster", "namespace"})

	metricIKS = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "weka_cluster_iks_filesystems_per_org",
		Help: "namespace filesystem count",
	}, []string{"cluster", "namespace"})

	metricFS = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "weka_cluster_filesystems_per_org",
		Help: "namespace filesystem count",
	}, []string{"cluster", "namespace"})

	metricTotalFS = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "weka_cluster_total_filesystems_per_org",
		Help: "namespace total filesystem count",
	}, []string{"cluster", "namespace"})

	metricFSRatio = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "weka_cluster_filesystems_ratio",
		Help: "cluster filesystem usage",
	}, []string{"cluster"})

	metricFSCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "weka_cluster_filesystem_count",
		Help: "cluster filesystem count",
	}, []string{"cluster"})

	metricFSUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "weka_filesystem_size",
		Help: "space used in volume",
	}, []string{"cluster", "cloudaccount", "name"})

	// Vast
	metricVastNSUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vast_cluster_org_used",
		Help: "cluster namespace utilization",
	}, []string{"cluster"})

	metricVastNSRatio = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vast_cluster_org_used_ratio",
		Help: "cluster namespace utilization ratio",
	}, []string{"cluster"})
	metricVastNSUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vast_org_space_usage",
		Help: "cluster namespace space utilization",
	}, []string{"cluster", "namespace"})

	metricVastFS = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vast_cluster_filesystems_per_org",
		Help: "namespace filesystem count",
	}, []string{"cluster", "namespace"})

	metricVastFSRatio = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vast_cluster_filesystems_ratio",
		Help: "cluster filesystem usage",
	}, []string{"cluster"})

	metricVastFSCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vast_cluster_filesystem_count",
		Help: "cluster filesystem count",
	}, []string{"cluster"})

	metricVastFSUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vast_filesystem_size",
		Help: "space used in volume",
	}, []string{"cluster", "cloudaccount", "name"})

	// Minio
	metricBkCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "minio_cluster_buckets_per_org",
		Help: "namespace bucket count",
	}, []string{"cluster", "cloudaccount"})

	metricTotalBK = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "minio_cluster_bucket_total_count",
		Help: "cluster bucket count",
	}, []string{"cluster"})

	metricBkUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "minio_cluster_bucket_usage",
		Help: "namespace bucket usage",
	}, []string{"cluster", "cloudaccount", "bucketname"})

	metricBkSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "minio_cluster_bucket_size",
		Help: "namespace bucket size",
	}, []string{"cluster", "cloudaccount", "bucketname"})

	// Common
	metricClusterUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_space_usage",
		Help: "cluster capacity",
	}, []string{"cluster", "type"})

	metricClusterSpaceTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_space_total",
		Help: "cluster capacity in GB",
	}, []string{"cluster", "type"})

	metricClusterSpaceAvailable = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_space_available",
		Help: "cluster capacity in GB",
	}, []string{"cluster", "type"})

	metricMinioAllocated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "minio_cluster_space_allocated",
		Help: "cluster capacity",
	}, []string{"cluster"})

	metricClusterCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_count",
		Help: "number of clusters",
	}, []string{"type"})

	// define metrics list
	metrics = []*prometheus.GaugeVec{
		// Weka
		metricIKS,
		metricFS,
		metricTotalFS,
		metricFSRatio,
		metricFSCount,
		metricFSUsage,
		metricNSUsed,
		metricNSUsage,
		metricNSRatio,
		// Vast
		metricVastNSUsed,
		metricVastNSUsage,
		metricVastNSRatio,
		metricVastFSCount,
		metricVastFSRatio,
		metricVastFS,
		metricVastFSUsage,
		// Minio
		metricBkCount,
		metricTotalBK,
		metricBkSize,
		metricBkUsage,
		// Common
		metricClusterUsage,
		metricClusterSpaceTotal,
		metricClusterSpaceAvailable,
		metricClusterCount,
	}

	// register metric with prometheus
	for _, m := range metrics {
		reg.MustRegister(m)
	}

	storageMonitor.scanClusters(ctx)

	// Expose metrics on endpoint
	if !cfg.TestMode {
		// Create a HTTP server for prometheus.
		httpServer := &http.Server{
			Handler:           promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
			Addr:              fmt.Sprintf("0.0.0.0:%d", 9092),
			ReadHeaderTimeout: 10 * time.Second,
		}
		// Start http server for prometheus.
		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				log.Error(err, "Unable to start a http server.")
				os.Exit(1)
			}
		}()
		log.Info("prometheus service started", logkeys.ListenPort, 9092)
	}
	duration := time.Duration(cfg.Interval) * time.Minute
	// Start a goroutine to update metric every set interval(minutes)
	go storageMonitor.StartMetricUpdater(ctx, duration, reg)

	return nil
}

func (*Service) Name() string {
	return "idc-storage-custom-metrics-service"
}

func StartStorageCustomMetricService() {
	ctx := context.Background()

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*Config](ctx, &Service{}, &Config{})
	if err != nil {
		logger := log.FromContext(ctx).WithName("StartStorageCustomMetricService")
		logger.Error(err, "init err")
	}
}
