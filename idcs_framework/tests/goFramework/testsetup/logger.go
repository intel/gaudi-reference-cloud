package testsetup

import (
	"fmt"
	"os"
	//"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log variable is a globally accessible variable which will be initialized when the InitializeZapCustomLogger function is executed successfully.
	Log  *zap.Logger
	Logf *zap.SugaredLogger
)

/*
InitializeZapCustomLogger Funtion initializes a logger using uber-go/zap package in the application.
*/
func InitializeZapCustomLogger() {
	wd, _ := os.Getwd()
	//wd = filepath.Clean(filepath.Join(wd, "../../../results"))
	logFileName := wd + "/results.log"
	//fmt.Println("Log file name with path is :" + logFileName)
	e := os.Remove(logFileName)
	if e != nil {
		fmt.Println("Log File doesnt exists", e)
	}
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	logFile, _ := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)
	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Logf = Log.Sugar()
}
