package rpc

import (
	"context"
	"io"

	"github.com/air-go/rpc/library/logger"
	zapLogger "github.com/air-go/rpc/library/logger/zap"
)

// RPCConfig is used to parse configuration file
// logger should be controlled with Options
type RPCConfig struct {
	InfoFile  string
	ErrorFile string
	Level     string
}

// RPCLogger is go-redis logger Hook
type RPCLogger struct {
	*zapLogger.ZapLogger
	config      *RPCConfig
	infoWriter  io.Writer
	errorWriter io.Writer
}

type RPCOption func(rl *RPCLogger)

// WithWriter sets both log destinations, overriding the rotating files created
// from RPCConfig.InfoFile and RPCConfig.ErrorFile; a single writer is rejected.
func WithWriter(infoWriter, errorWriter io.Writer) RPCOption {
	return func(rl *RPCLogger) {
		rl.infoWriter, rl.errorWriter = infoWriter, errorWriter
	}
}

// NewRPCLogger
func NewRPCLogger(config *RPCConfig, opts ...RPCOption) (rl *RPCLogger, err error) {
	rl = &RPCLogger{config: config}

	for _, o := range opts {
		o(rl)
	}

	infoWriter, errorWriter, err := logger.ResolveWriters(rl.infoWriter, rl.errorWriter, config.InfoFile, config.ErrorFile)
	if err != nil {
		return nil, err
	}
	rl.infoWriter, rl.errorWriter = infoWriter, errorWriter

	l, err := zapLogger.NewLogger(
		zapLogger.WithCallerSkip(4),
		zapLogger.WithModule(logger.ModuleRPC),
		zapLogger.WithInfoWriter(rl.infoWriter),
		zapLogger.WithErrorWriter(rl.errorWriter),
		zapLogger.WithLevel(config.Level),
	)
	if err != nil {
		return nil, err
	}
	rl.ZapLogger = l

	return
}

func (rl *RPCLogger) Info(ctx context.Context, msg string, fields ...logger.Field) {
	rl.logger().Info(ctx, msg, fields...)
}

func (rl *RPCLogger) Error(ctx context.Context, msg string, fields ...logger.Field) {
	rl.logger().Error(ctx, msg, fields...)
}

func (rl *RPCLogger) logger() *zapLogger.ZapLogger {
	return rl.ZapLogger
}
