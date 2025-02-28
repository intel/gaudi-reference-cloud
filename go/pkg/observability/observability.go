// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package observability

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	mineralriver "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/mineral-river"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Observability struct {
	mrTelemetryOptions *mineralriver.TelemetryOptions
}

type TracerProvider struct {
	sdktrace.TracerProvider
}

func New(ctx context.Context) *Observability {
	mrTelemetryOptions := mineralriver.New()
	return &Observability{
		mrTelemetryOptions: mrTelemetryOptions,
	}
}

func (o *Observability) InitTracer(ctx context.Context) *TracerProvider {
	otel.SetLogger(log.FromContext(ctx))
	return &TracerProvider{
		*o.mrTelemetryOptions.InitTracer(ctx),
	}
}
