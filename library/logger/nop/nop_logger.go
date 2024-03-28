package nop

import (
	"context"

	"github.com/air-go/rpc/library/logger"
)

var Logger = &nopLogger{}

type nopLogger struct{}

var _ logger.Logger = (*nopLogger)(nil)

func (*nopLogger) Debug(ctx context.Context, msg string, fields ...logger.Field) {}

func (*nopLogger) Info(ctx context.Context, msg string, fields ...logger.Field) {}

func (*nopLogger) Warn(ctx context.Context, msg string, fields ...logger.Field) {}

func (*nopLogger) Error(ctx context.Context, msg string, fields ...logger.Field) {}

func (*nopLogger) DPanic(ctx context.Context, msg string, fields ...logger.Field) {}

func (*nopLogger) Panic(ctx context.Context, msg string, fields ...logger.Field) {}

func (*nopLogger) Fatal(ctx context.Context, msg string, fields ...logger.Field) {}

func (*nopLogger) GetLevel() logger.Level { return logger.UnknownLevel }

func (*nopLogger) Close() error { return nil }
