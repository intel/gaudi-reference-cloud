// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mineralriver

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Option func(*TelemetryOptions)

func WithService(service string) Option {
	return func(s *TelemetryOptions) {
		s.service = service
	}
}

func WithLogLevel(logLevel string) Option {
	return func(s *TelemetryOptions) {
		s.logLevel = logLevel
	}
}

func WithCustomResources(customResources map[string]string) Option {
	return func(s *TelemetryOptions) {
		s.customResources = customResources
	}
}

type TelemetryOptions struct {
	service         string
	logLevel        string
	customResources map[string]string
	oltpEndpoint    string
}

func New(option ...Option) *TelemetryOptions {
	s := &TelemetryOptions{
		service:      os.Getenv("OTEL_SERVICE_NAME"),
		oltpEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}
	for _, o := range option {
		o(s)
	}
	return s
}

// init tracer to connect to otel collector
func (s *TelemetryOptions) InitTracer(ctx context.Context) *sdktrace.TracerProvider {
	if s.oltpEndpoint == "" {
		return s.defaultTracer(ctx)
	}

	client := otlptracehttp.NewClient(otlptracehttp.WithEndpoint(s.oltpEndpoint))
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Print(err)
		return s.defaultTracer(ctx)
	}

	resources, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("module", s.service),
			attribute.String("loglevel", s.logLevel),
		),
	)
	if err != nil {
		log.Print(err)
		return s.defaultTracer(ctx)
	}

	for key, value := range s.customResources {
		resources_new, err := resource.New(
			ctx,
			resource.WithAttributes(
				attribute.String(key, value),
			),
		)
		if err != nil {
			log.Print(err)
			return s.defaultTracer(ctx)
		}
		resources, err = resource.Merge(resources_new, resources)
		if err != nil {
			log.Print(err)
			return s.defaultTracer(ctx)
		}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	log.Printf("mineral-river: Tracing to %s as service %q", s.oltpEndpoint, s.service)
	return tp
}

func (s *TelemetryOptions) defaultTracer(ctx context.Context) *sdktrace.TracerProvider {
	log.Print("mineral-river: Using default (noop) tracer")
	return sdktrace.NewTracerProvider()
}
