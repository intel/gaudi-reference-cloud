// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once
var Log *zap.Logger
var Logf *zap.SugaredLogger

func Init() {

	once.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)
		file := getLoggerFile()
		level := zap.NewAtomicLevelAt(zap.InfoLevel)

		consoleEncoder := getDevelopmentConsoleEncoder()
		fileEncoder := getProductionFileEncoder()

		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, stdout, level),
			zapcore.NewCore(fileEncoder, file, level).
				With(
					[]zapcore.Field{
						zap.Int("pid", os.Getpid()),
					},
				),
		)

		Log = zap.New(core)
		Logf = Log.Sugar()
		defer Log.Sync()
	})
}

func getLoggerFile() zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   "../logs/test.log",
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     14, // days
	})
}

func getDevelopmentConsoleEncoder() zapcore.Encoder {
	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	return zapcore.NewConsoleEncoder(developmentCfg)
}

func getProductionFileEncoder() zapcore.Encoder {
	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	return zapcore.NewJSONEncoder(productionCfg)
}
