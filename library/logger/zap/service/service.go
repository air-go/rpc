package service

import (
	"io"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/zap"
)

type Config struct {
	InfoFile  string
	ErrorFile string
	Level     string
}

type ServiceLogger struct {
	*zap.ZapLogger
	config      *Config
	infoWriter  io.Writer
	errorWriter io.Writer
}

type Option func(l *ServiceLogger)

// WithWriter sets both log destinations, overriding the rotating files created
// from Config.InfoFile and Config.ErrorFile; a single writer is rejected.
func WithWriter(infoWriter, errorWriter io.Writer) Option {
	return func(l *ServiceLogger) {
		l.infoWriter, l.errorWriter = infoWriter, errorWriter
	}
}

func NewServiceLogger(serviceName string, config *Config, opts ...Option) (*ServiceLogger, error) {
	l := &ServiceLogger{config: config}
	for _, o := range opts {
		o(l)
	}

	infoWriter, errorWriter, err := logger.ResolveWriters(l.infoWriter, l.errorWriter, config.InfoFile, config.ErrorFile)
	if err != nil {
		return nil, err
	}
	l.infoWriter, l.errorWriter = infoWriter, errorWriter

	zl, err := zap.NewLogger(
		zap.WithModule(logger.ModuleHTTP),
		zap.WithServiceName(serviceName),
		zap.WithInfoWriter(l.infoWriter),
		zap.WithErrorWriter(l.errorWriter),
		zap.WithLevel(config.Level),
	)
	if err != nil {
		return nil, err
	}
	l.ZapLogger = zl

	return l, nil
}
