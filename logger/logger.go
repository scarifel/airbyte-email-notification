package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	encoderCfg.MessageKey = "message"

	level := zap.InfoLevel
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		level = zap.DebugLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     encoderCfg,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	if level == zap.DebugLevel {
		config.Development = true
		config.DisableCaller = false
		config.DisableStacktrace = false
	}

	logger = zap.Must(config.Build())
}

func LogLevelString() string {
	return logger.Sugar().Level().String()
}

func Info(msg string, fields ...any) {
	logger.Sugar().Infow(msg, fields...)
}

func Error(msg string, fields ...any) {
	logger.Sugar().Errorw(msg, fields...)
}

func Warn(msg string, fields ...any) {
	logger.Sugar().Warnw(msg, fields...)
}

func Fatal(msg string, fields ...any) {
	logger.Sugar().Fatalw(msg, fields...)
}

func Debug(msg string, fields ...any) {
	logger.Sugar().Debugw(msg, fields...)
}