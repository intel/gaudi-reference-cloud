// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bufbuild/protovalidate-go"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	api "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	vast_api "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1/vast"
	weka_api "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1/weka"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	minio_storage "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/minio"
	vast_storage "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast"
	weka_storage "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka"
	conf "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/server/vast"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/server/weka"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	Config *conf.Config
}

var GrpcMetrics *grpcprom.ServerMetrics = grpcprom.NewServerMetrics(
	grpcprom.WithServerHandlingTimeHistogram(
		grpcprom.WithHistogramBuckets([]float64{0.1, 0.3, 0.6, 1, 2, 4, 5, 10}),
	),
)

var BackendInfoMetric *prometheus.GaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "backend_info",
	Help: "Details for relative backend from sds-controller configuration",
}, []string{"name", "uuid", "type", "location"})

func (s *Server) CreateGrpcServer() (*grpc.Server, error) {
	if s.Config == nil {
		return nil, errors.New("config cannot be empty")
	}
	opts := make([]grpc.ServerOption, 0)
	if s.Config.GrpcTLS != nil {
		log.Info().
			Str("Certificate location", s.Config.GrpcTLS.CertFile).
			Str("Key location", s.Config.GrpcTLS.KeyFile).
			Msg("Found TLS credentials in the params, creating TLS server connection")

		creds, err := credentials.NewServerTLSFromFile(s.Config.GrpcTLS.CertFile, s.Config.GrpcTLS.KeyFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	validator, err := protovalidate.New()

	if err != nil {
		log.Fatal().Msg("Could not initiate grpc validator")
	}

	opts = append(opts, grpc.ChainUnaryInterceptor(
		GrpcMetrics.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(interceptorLogger(log.Logger), loggingOpts...),
		protovalidate_middleware.UnaryServerInterceptor(validator),
	))

	grpcServer := grpc.NewServer(opts...)

	backends := s.registerBackends()

	healthBackends := make(map[string]backend.HealthInterface)
	for k, v := range backends {
		var i backend.HealthInterface = v
		healthBackends[k] = i
	}

	hc, err := NewHealthController(healthBackends, time.Second*time.Duration(s.Config.HealthInterval))

	if err != nil {
		log.Fatal().Msg("Could not initiate HealthController")
	} else {
		ctx := context.Background()
		go hc.Start(ctx)
	}

	api.RegisterClusterServiceServer(grpcServer, &ClusterHandler{
		Clusters: s.Config.Clusters,
		Backends: backends,
	})
	api.RegisterNamespaceServiceServer(grpcServer, &NamespaceHandler{
		Backends: backends,
	})
	api.RegisterUserServiceServer(grpcServer, &UserHandler{
		Backends: backends,
	})
	api.RegisterS3ServiceServer(grpcServer, &S3Handler{
		Backends: backends,
	})

	weka_api.RegisterStatefulClientServiceServer(grpcServer, &weka.StatefulClientHandler{
		Backends: backends,
	})

	weka_api.RegisterFilesystemServiceServer(grpcServer, &weka.FilesystemHandler{
		Backends: backends,
	})

	vast_api.RegisterFilesystemServiceServer(grpcServer, &vast.FilesystemHandler{
		Backends: backends,
	})

	return grpcServer, err
}

func (s *Server) registerBackends() map[string]backend.Interface {
	backends := make(map[string]backend.Interface)

	for _, value := range s.Config.Clusters {
		BackendInfoMetric.WithLabelValues(value.Name, value.UUID, string(value.Type), value.Location).Set(1)
		switch value.Type {
		case conf.Weka:
			backend, err := weka_storage.NewBackend(value)
			if err != nil {
				log.Fatal().Err(err).Msg("Could not initialize Weka backend")
			} else {
				log.Info().Str("uuid", value.UUID).Any("type", value.Type).Msg("Registered new cluster")
				backends[value.UUID] = backend
			}
		case conf.MinIO:
			backend, err := minio_storage.NewBackend(value)
			if err != nil {
				log.Fatal().Err(err).Msg("Could not initialize MinIO backend")
			} else {
				log.Info().Str("uuid", value.UUID).Any("type", value.Type).Msg("Registered new cluster")
				backends[value.UUID] = backend
			}
		case conf.Vast:
			backend, err := vast_storage.NewBackend(value)
			if err != nil {
				log.Fatal().Err(err).Msg("Could not initialize Vast backend")
			} else {
				log.Info().Str("uuid", value.UUID).Any("type", value.Type).Msg("Registered new cluster")
				backends[value.UUID] = backend
			}
		default:
			log.Error().Any("Unknown cluster type", value.Type).Send()
		}
	}

	return backends
}

func interceptorLogger(zl zerolog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l := zl.With().Fields(fields).Logger()

		switch lvl {
		case logging.LevelDebug:
			l.Debug().Ctx(ctx).Msg(msg)
		case logging.LevelInfo:
			l.Info().Ctx(ctx).Msg(msg)
		case logging.LevelWarn:
			l.Warn().Ctx(ctx).Msg(msg)
		case logging.LevelError:
			l.Error().Ctx(ctx).Msg(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
