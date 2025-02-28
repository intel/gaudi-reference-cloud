// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package observability

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// LogAndSpan is a convenience type for using a Logger and Span together.
// Example usage:
//
//		ctx, log, span := observability.LogAndSpanFromContext(ctx).WithName("Instance.Create").WithValues("ResourceId", ResourceId).Start()
//		defer span.End()
//	    log.Info("Creating instance", "request", request)
type LogAndSpan struct {
	ctx           context.Context
	logger        logr.Logger
	tracer        trace.Tracer
	name          string
	keysAndValues []interface{}
}

const traceIdLogKey string = "traceId"
const tracerName string = "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"

// Extract a Logger and Span from the context.
// If context does not have a Span, a no-op tracer will be used.
// It also adds the trace.id to the Logger.
func LogAndSpanFromContext(ctx context.Context) LogAndSpan {
	logger := log.FromContext(ctx)
	span := trace.SpanFromContext(ctx)
	return LogAndSpan{
		ctx:    ctx,
		logger: logger,
		tracer: span.TracerProvider().Tracer(tracerName),
	}
}

// Extract a Logger and Span from the context.
// If context does not have a Span, use the global tracer provider.
// It also adds the trace.id to the Logger.
func LogAndSpanFromContextOrGlobal(ctx context.Context) LogAndSpan {
	logger := log.FromContext(ctx)
	span := trace.SpanFromContext(ctx)
	var tracer trace.Tracer
	if isNoopSpan(span) {
		tracerProvider := otel.GetTracerProvider()
		tracer = tracerProvider.Tracer(tracerName)
	} else {
		tracer = span.TracerProvider().Tracer(tracerName)
	}
	return LogAndSpan{
		ctx:    ctx,
		logger: logger,
		tracer: tracer,
	}
}

// Add a name to the Logger and Span.
func (ls LogAndSpan) WithName(name string) LogAndSpan {
	ls.name = name
	return ls
}

// Add key/value pairs to the Logger and Span.
func (ls LogAndSpan) WithValues(keysAndValues ...interface{}) LogAndSpan {
	if len(keysAndValues)%2 != 0 {
		panic("the length of keysAndValues must be even")
	}
	ls.keysAndValues = append(ls.keysAndValues, keysAndValues...)
	return ls
}

// Start a new Span and return the Context, Logger, and Span.
func (ls LogAndSpan) Start() (context.Context, logr.Logger, trace.Span) {
	ctx := ls.ctx
	var kv []attribute.KeyValue
	for i := 0; i < len(ls.keysAndValues)-1; i += 2 {
		key := fmt.Sprintf("%v", ls.keysAndValues[i])
		value := fmt.Sprintf("%v", ls.keysAndValues[i+1])
		kv = append(kv, attribute.String(key, value))
	}
	ctx, span := ls.tracer.Start(ctx, ls.name, trace.WithAttributes(kv...))
	ctx = trace.ContextWithSpan(ctx, span)

	traceID := span.SpanContext().TraceID().String()
	logger := ls.logger.WithValues(traceIdLogKey, traceID).WithValues(ls.keysAndValues...)
	ctx = log.IntoContext(ctx, logger)
	logger = ls.logger.WithName(ls.name).WithValues(traceIdLogKey, traceID).WithValues(ls.keysAndValues...)

	return ctx, logger, span
}

func isNoopSpan(span trace.Span) bool {
	return span == trace.SpanFromContext(context.Background())
}
