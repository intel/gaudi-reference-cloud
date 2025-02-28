// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package store_forward_logger

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// A logger that can store logged information and forward it later.
type StoreForwardLogger struct {
	logr.Logger
	buffer *bytes.Buffer
}

func New() *StoreForwardLogger {
	buffer := &bytes.Buffer{}
	var zapOpts = zap.Options{
		DestWriter:  buffer,
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	logger := zap.New(zap.UseFlagOptions(&zapOpts))
	return &StoreForwardLogger{
		buffer: buffer,
		Logger: logger,
	}
}

func (l StoreForwardLogger) WithName(name string) StoreForwardLogger {
	return StoreForwardLogger{
		Logger: l.Logger.WithName(name),
		buffer: l.buffer,
	}
}

func (l StoreForwardLogger) WithValues(keysAndValues ...any) StoreForwardLogger {
	return StoreForwardLogger{
		Logger: l.Logger.WithValues(keysAndValues...),
		buffer: l.buffer,
	}
}

func (l StoreForwardLogger) GetStoredString() string {
	return l.buffer.String()
}

// A StoreForwardErrGroup is an errgroup.Group with integrated logging.
// Logs for each goroutine are printed together without interleaving.
// Logs for successful goroutines are printed immediately upon completion of the goroutine.
// Logs for failed goroutines are printed at the end, making it convenient for users to find
// the cause of errors.
type StoreForwardErrGroup struct {
	errGroup           *errgroup.Group
	forwardToLogger    logr.Logger
	gCtx               context.Context
	mutex              sync.Mutex
	routinesWithErrors []storeForwardErrGroupRoutine
}

type storeForwardErrGroupRoutine struct {
	description string
	err         error
	logger      *StoreForwardLogger
}

func NewStoreForwardErrGroup(ctx context.Context) *StoreForwardErrGroup {
	forwardToLogger := log.FromContext(ctx)
	g, gCtx := errgroup.WithContext(ctx)
	return &StoreForwardErrGroup{
		errGroup:        g,
		forwardToLogger: forwardToLogger,
		gCtx:            gCtx,
	}
}

func (g *StoreForwardErrGroup) Go(description string, fn func(ctx context.Context) error) {
	// Create a new context that wraps the new logger.
	sflogger := New()
	sflogger.Info(fmt.Sprintf("BEGIN: Logs from %s", description))
	ctx := log.IntoContext(g.gCtx, sflogger.Logger)
	g.errGroup.Go(func() error {
		err := fn(ctx)

		routine := storeForwardErrGroupRoutine{
			description: description,
			err:         err,
			logger:      sflogger,
		}

		if routine.err != nil {
			isCanceled := g.gCtx.Err() == context.Canceled
			if isCanceled {
				sflogger.Info("Context canceled because another routine failed")
			} else {
				// An error occurred. Keep the logger so that it an be forwarded at the end of Wait().
				sflogger.Error(nil, fmt.Sprintf("END:   Logs from %s", description))
				sflogger.Error(err, fmt.Sprintf("Error from %s", description))
				// Lock the mutex to ensure only one Goroutine appends to routinesWithErrors.
				g.mutex.Lock()
				defer g.mutex.Unlock()
				g.routinesWithErrors = append(g.routinesWithErrors, routine)
				return fmt.Errorf("%s: %w", description, routine.err)
			}
		}

		// No error has occured or this was canceled due to an error in another routine. Forward logs now.
		sflogger.Info(fmt.Sprintf("END:   Logs from %s", description))
		// Log the entire stored log buffer with a single statement so that it does not get interleaved with another Goroutine.
		g.forwardToLogger.Info("Logs from " + description + ":\n" + strings.Trim(sflogger.GetStoredString(), "\n"))
		return routine.err
	})
}

func (g *StoreForwardErrGroup) Wait() error {
	err := g.errGroup.Wait()

	// A mutex lock should not be required because all goroutines have exited.
	// But this lock prevents Coverity from reporting a GUARDED_BY_VIOLATION violation.
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Forward any logs for routines that had errors.
	for _, routine := range g.routinesWithErrors {
		// Log the entire stored log buffer with a single statement so that it does not get interleaved with another Goroutine.
		g.forwardToLogger.Error(routine.err, "Logs from "+routine.description+":\n"+strings.Trim(routine.logger.GetStoredString(), "\n"))
	}

	return err
}
