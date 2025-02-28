// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"reflect"

	"google.golang.org/grpc"
)

//intercept otelgrpc to add cloudaccount ID to parent grpc span within a trace

func CustomUnaryServerInterceptor(param string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract the span from the context
		span := trace.SpanFromContext(ctx)

		val := reflect.ValueOf(req).Elem()
		field := val.FieldByName(param)
		if field.IsValid() {
			// fmt.Println("getting cloudaccount", field)
			span.SetAttributes(attribute.String(param, fmt.Sprintf("%v", field)))
		}
		return handler(ctx, req)
	}
}
