package logger

import (

	"go.uber.org/zap"
)

// глобальный логгер
var Log *zap.Logger = zap.NewNop()



func NewServiceLogger(level string) (*zap.Logger, error) {
	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = logLevel
	zlogger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zlogger, nil
}
