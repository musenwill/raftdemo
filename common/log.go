package common

import (
	"go.uber.org/zap"
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
	}
}
