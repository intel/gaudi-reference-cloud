// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mineralriver

import (
	"context"
	"log"
	"testing"
)

func TestTelemetry(t *testing.T) {
	customResourceOptions := make(map[string]string)
	customResourceOptions["custom1"] = "custom1"
	customResourceOptions["custom2"] = "custom2"
	mr := New(
		WithCustomResources(customResourceOptions),
	)
	tp := mr.InitTracer(context.Background())
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
}
