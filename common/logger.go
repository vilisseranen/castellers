package common

import (
	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger

func InitializeLogger() error {
	cfg := zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	var err error
	logger, err = cfg.Build()
	sugar = logger.Sugar()
	return err
}

func Log(level, message string, args ...interface{}) {
	switch level {
	case "DEBUG":
		sugar.Debugf(message, args...)
	case "INFO":
		sugar.Infof(message, args...)
	case "WARN":
		sugar.Warnf(message, args...)
	case "ERROR":
		sugar.Errorf(message, args...)
	case "FATAL":
		sugar.Fatalf(message, args...)
	}
}

func Debug(message string, args ...interface{}) {
	Log("DEBUG", message, args)
}

func Info(message string, args ...interface{}) {
	Log("INFO", message, args)
}

func Warn(message string, args ...interface{}) {
	Log("WARN", message, args)
}

func Error(message string, args ...interface{}) {
	Log("ERROR", message, args)
}

func Fatal(message string, args ...interface{}) {
	Log("FATAL", message, args)
}
