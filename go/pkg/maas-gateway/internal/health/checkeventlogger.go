// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package health

import (
	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/go-logr/logr"
)

type CheckEventsLogger struct {
	logger logr.Logger
}

func NewCheckEventsLogger(logger logr.Logger) *CheckEventsLogger {
	return &CheckEventsLogger{
		logger: logger.WithName("CheckEventsLogger"),
	}
}

func (l *CheckEventsLogger) printErrorResult(msg string, name string, res gosundheit.Result) {
	l.logger.Error(res.Error, msg, "result", "error",
		"name", name,
		"details", res.Details,
		"timestamp", res.Timestamp,
		"duration", res.Duration,
		"contiguousFailures", res.ContiguousFailures,
		"timeOfFirstFailure", res.TimeOfFirstFailure,
	)
}

func (l *CheckEventsLogger) printSuccessResult(msg string, name string, res gosundheit.Result) {
	l.logger.Info(msg, "result", "success",
		"name", name,
		"timestamp", res.Timestamp,
		"duration", res.Duration,
	)
}

func (l *CheckEventsLogger) printResult(msg string, name string, res gosundheit.Result) {
	if res.Error != nil {
		l.printErrorResult(msg, name, res)
	} else {
		l.printSuccessResult(msg, name, res)
	}
}

func (l *CheckEventsLogger) OnCheckRegistered(name string, res gosundheit.Result) {
	l.printResult("Check registered with initial result:", name, res)
}

func (l *CheckEventsLogger) OnCheckStarted(name string) {
	l.logger.Info("Check started...", "name", name)
}

func (l *CheckEventsLogger) OnCheckCompleted(name string, res gosundheit.Result) {
	l.printResult("Check completed with result:", name, res)
}
