package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel string

var LogLevelEnum = struct {
	Debug   LogLevel
	Info    LogLevel
	Warning LogLevel
	Error   LogLevel
	Panic   LogLevel
	Fatal   LogLevel
}{
	Debug:   "DEBUG",
	Info:    "INFO",
	Warning: "WARNING",
	Error:   "ERROR",
	Panic:   "PANIC",
	Fatal:   "FATAL",
}

var LogLevelList = []LogLevel{"DEBUG", "INFO", "WARNING", "ERROR", "PANIC", "FATAL"}

func (p LogLevel) Valid() bool {
	for _, v := range LogLevelList {
		if p == v {
			return true
		}
	}
	return false
}

func ComparableLogLevel(level LogLevel) (zapcore.Level, error) {
	m := map[LogLevel]zapcore.Level{
		"DEBUG":   zapcore.DebugLevel,
		"INFO":    zapcore.InfoLevel,
		"WARNING": zapcore.WarnLevel,
		"ERROR":   zapcore.ErrorLevel,
		"PANIC":   zapcore.PanicLevel,
		"FATAL":   zapcore.FatalLevel,
	}
	r, ok := m[level]
	if !ok {
		return -1, fmt.Errorf("unsupported log level %s", level)
	}
	return r, nil
}

func LogLevel2Str(level zapcore.Level) (LogLevel, error) {
	m := map[zapcore.Level]LogLevel{
		zapcore.DebugLevel: "DEBUG",
		zapcore.InfoLevel:  "INFO",
		zapcore.WarnLevel:  "WARNING",
		zapcore.ErrorLevel: "ERROR",
		zapcore.PanicLevel: "PANIC",
		zapcore.FatalLevel: "FATAL",
	}
	r, ok := m[level]
	if !ok {
		return "", fmt.Errorf("unsupported log level %d", level)
	}
	return r, nil
}

type Logger struct {
	logger *zap.Logger
	level  *zap.AtomicLevel
}

func (p *Logger) SetLevel(level LogLevel) error {
	l, err := ComparableLogLevel(level)
	if err != nil {
		return err
	}
	p.level.SetLevel(l)
	return nil
}

func (p *Logger) GetLogger() *zap.Logger {
	return p.logger
}

func (p *Logger) GetLevel() LogLevel {
	r, _ := LogLevel2Str(p.level.Level())
	return r
}

func NewLogger(l LogLevel, paths ...string) (*Logger, error) {
	level, err := ComparableLogLevel(l)
	if err != nil {
		return nil, err
	}
	logLevel := zap.NewAtomicLevelAt(level)
	config := DefaultZapConfig(&logLevel, paths...)
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &Logger{
		logger: logger,
		level:  &logLevel,
	}, nil
}

func DefaultZapConfig(level *zap.AtomicLevel, paths ...string) *zap.Config {
	return &zap.Config{
		Encoding:      "json",
		Level:         *level,
		OutputPaths:   paths,
		EncoderConfig: zap.NewProductionEncoderConfig(),
	}
}
