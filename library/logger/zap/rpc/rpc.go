package rpc

import (
	"context"

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
	config *RPCConfig
}

type RPCOption func(rl *RPCLogger)

// NewRPCLogger
func NewRPCLogger(config *RPCConfig, opts ...RPCOption) (rl *RPCLogger, err error) {
	rl = &RPCLogger{config: config}

	for _, o := range opts {
		o(rl)
	}

	infoWriter, errWriter, err := logger.RotateWriter(config.InfoFile, config.ErrorFile)
	if err != nil {
		return
	}

	l, err := zapLogger.NewLogger(
		zapLogger.WithCallerSkip(4),
		zapLogger.WithModule(logger.ModuleRPC),
		zapLogger.WithInfoWriter(infoWriter),
		zapLogger.WithErrorWriter(errWriter),
		zapLogger.WithLevel(config.Level),
	)
	if err != nil {
		return
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
