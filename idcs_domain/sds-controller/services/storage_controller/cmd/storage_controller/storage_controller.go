// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	configLocation, exists := os.LookupEnv("STORAGE_CONTROLLER_CONFIG_FILE")
	if !exists {
		log.Fatal().Msg("STORAGE_CONTROLLER_CONFIG_FILE env variable is not set")
	}

	config, err := conf.LoadStorageConfig(configLocation)

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot parse config")
	}

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		server.HealthMetric,
		server.GrpcMetrics,
		server.BackendInfoMetric,
	)

	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	go http.ListenAndServe(":9099", nil)

	server := server.Server{
		Config: config,
	}

	grpcServer, err := server.CreateGrpcServer()

	if err != nil {
		log.Fatal().Err(err).Msg("Could not create server")
	}

	log.Info().Int("port", config.ListenPort).Msg("Starting server")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(config.ListenPort)))
	if err != nil {
		log.Fatal().Err(err).Msg("Could not listen on port")
	}

	err = grpcServer.Serve(lis)

	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
