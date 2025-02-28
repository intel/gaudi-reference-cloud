// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package stoppable

import (
	"context"
)

// Stoppable accepts a blocking function and provides a non-blocking Start method and a blocking Stop method.
type Stoppable struct {
	runFunc func(context.Context) error
	cancel  context.CancelFunc
	errc    chan error
}

// Create a new Stoppable.
// The supplied function is expected to return when the context is cancelled.
func New(runFunc func(context.Context) error) *Stoppable {
	return &Stoppable{
		runFunc: runFunc,
		errc:    make(chan error, 1),
	}
}

// Run runFunc in a goroutine.
// Returns immediately.
func (s *Stoppable) Start(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)
	go func() {
		s.errc <- s.runFunc(ctx)
		close(s.errc)
	}()
}

// Cancel runFunc, wait for it to return, and return the error.
func (s *Stoppable) Stop(ctx context.Context) error {
	s.cancel()
	return <-s.errc
}
