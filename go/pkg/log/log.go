// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Package log is a structured logging library.
// Users can obtain a logr.Logger from the context.
// If the context does not have a logr.Logger, a global logger that uses Zap will be provided.
// The application will be accept command line options to set the Zap logger parameters such as `--zap-log-level`.
//
// See also https://github.com/go-logr/logr.
//
// Usage:
//
//		import "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
//		import "flag"
//
//		func main() {
//			log.BindFlags()
//			flag.Parse()
//			log.SetDefaultLogger()
//
//			ctx := context.Background()
//			log := log.FromContext(ctx)
//			log.Info("main initializing")
//			...
//		}
//
//		func OtherFunc(ctx context.Context, param1 string) {
//			log := log.FromContext(ctx).WithName("OtherFunc").WithValues("param1", param1)
//			log.Info("BEGIN", "key1", "value1", "key2", "value2")
//			log.V(9).Info("debug info", "key1", 37)
//		}
//

package log

import (
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type LoggingKeys string

const (
	CloudAcctIdLoggingKey = "CloudAcctId"
)

var (
	validLoggingKeys = map[string]bool{
		CloudAcctIdLoggingKey: true,
	}
)

var zapOpts = zap.Options{
	Development: true,
	TimeEncoder: zapcore.RFC3339TimeEncoder,
}

// This init function sets the default log level based on the environment variable ZAP_LOG_LEVEL.
// Flags will override this log level.
// Log level "-127" provides the maximum verbosity.
// Log level can also be a string such as DEBUG, INFO, WARN, etc.
func init() {
	logLevelStr := os.Getenv("ZAP_LOG_LEVEL")
	if logLevelStr != "" {
		level, err := logLevelFromString(logLevelStr)
		if err != nil {
			panic(fmt.Errorf("invalid value for environment variable ZAP_LOG_LEVEL: %w", err))
		}
		zapOpts.Level = level
	}
}

func logLevelFromString(logLevelStr string) (zapcore.Level, error) {
	levelInt, err := strconv.Atoi(logLevelStr)
	if err == nil {
		return zapcore.Level(levelInt), nil
	} else {
		var level zapcore.Level
		err := level.Set(logLevelStr)
		return level, err
	}
}

func BindFlags() {
	zapOpts.BindFlags(flag.CommandLine)
}

func SetDefaultLogger() {
	logger := zap.New(zap.UseFlagOptions(&zapOpts))
	ctrllog.SetLogger(logger)
}

func SetDefaultLoggerDebug() {
	zapOpts.Level = zapcore.DebugLevel
	SetDefaultLogger()
}

func GetContextWithLoggingKeys(ctx context.Context, key string, value string) context.Context {
	_, ok := validLoggingKeys[key]
	if ok {
		return context.WithValue(ctx, LoggingKeys("key"), value)
	}
	return ctx
}

func FromContextWithKeys(ctx context.Context, keysAndValues ...interface{}) logr.Logger {
	logger := ctrllog.FromContext(ctx, keysAndValues...)
	if ctx != nil {
		for key := range validLoggingKeys {
			if value, ok := ctx.Value(LoggingKeys(key)).(string); ok {
				logger = logger.WithValues(key, value)
			}
		}
	}
	return logger
}

func FromContext(ctx context.Context, keysAndValues ...interface{}) logr.Logger {
	return ctrllog.FromContext(ctx, keysAndValues...)
}

func IntoContext(ctx context.Context, log logr.Logger) context.Context {
	return ctrllog.IntoContext(ctx, log)
}

func IntoContextWithLogger(ctx context.Context, log logr.Logger) (context.Context, logr.Logger) {
	return ctrllog.IntoContext(ctx, log), log
}

// This method internally calls `logSanitizedResponse`
// `logSanitizedResponse`  deals with structure of pointers but doesn't return any explicit errors in case something goes wrong
// We don't want to abort the workflow in case of unexpected errors and therefore, we recover from that in this method and return
// On returning, resumes API workflow execution
func LogResponseOrError(log logr.Logger, req any, resp any, err error) {
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		// handle unexpected errors in logSanitizedResponse
		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Errorf("error occurred while sanitizing response object for logs"), logkeys.Error)
			}
		}()
		logSanitizedResponse(log, resp)
	}
}

// works only for protobuf objects.
func logSanitizedResponse(log logr.Logger, resp any) {
	// preprocess response
	respVal := reflect.ValueOf(resp)
	if respVal.IsValid() && !respVal.IsZero() {
		respCopy := proto.Clone(respVal.Interface().(proto.Message))
		switch reflect.ValueOf(respCopy).Type().Kind() {
		case reflect.Ptr:
			TrimObjectForLogs(reflect.ValueOf(respCopy).Elem())
		case reflect.Struct:
			TrimObjectForLogs(reflect.ValueOf(&respCopy).Elem())
		default:
			log.Info("Response", logkeys.Response, resp)
		}
		// log updated response copy
		log.Info("Response", logkeys.Response, respCopy)
	} else {
		// log original response
		log.Info("Response", logkeys.Response, resp)
	}
}
