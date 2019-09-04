package common

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(config *zap.Config) *zap.SugaredLogger {
	logger, _ := config.Build()
	return logger.Sugar()
}

func DefaultZapConfig(paths ...string) *zap.Config {
	return &zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		OutputPaths: paths,
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
}
