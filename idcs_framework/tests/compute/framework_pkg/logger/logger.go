package logger

import (
	"bufio"
	"fmt"
	"io"
	"log"
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

var (
	sharedLogger *CustomLogger
	parallelonce sync.Once
)

type CustomLogger struct {
	*log.Logger
	writer *bufio.Writer
}

func InitializeLogger(nodeID int) {
	parallelonce.Do(func() {
		logFile := fmt.Sprintf("suite1_node%d_log.txt", nodeID)
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}

		writer := bufio.NewWriter(file)

		multiWriter := io.MultiWriter(writer, os.Stdout) // Write to both file and plain stdout
		sharedLogger = &CustomLogger{
			Logger: log.New(multiWriter, fmt.Sprintf("[Node %d] ", nodeID), log.LstdFlags),
			writer: writer,
		}

		sharedLogger.Println("Logger initialized.")
		if err := sharedLogger.writer.Flush(); err != nil {
			log.Printf("failed to flush logger writer: %v", err)
		}
	})
}

func GetLogger() *CustomLogger {
	if sharedLogger == nil {
		panic("Logger not initialized. Call InitializeLogger first.")
	}
	return sharedLogger
}

func (c *CustomLogger) Println(v ...interface{}) {
	c.Logger.Println(v...)
	if err := c.writer.Flush(); err != nil {
		log.Printf("failed to flush logger writer: %v", err)
	}
}
