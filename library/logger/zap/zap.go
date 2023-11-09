package zap

import (
	"context"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/air-go/rpc/library/app"
	"github.com/air-go/rpc/library/logger"
)

type ZapLogger struct {
	*zap.Logger
	opts *Options
}

type Options struct {
	level       logger.Level
	callSkip    int
	module      string
	serviceName string
	infoWriter  io.Writer
	errorWriter io.Writer
}

var _ logger.Logger = (*ZapLogger)(nil)

type Option func(l *Options)

func defaultOptions() *Options {
	return &Options{
		level:       logger.InfoLevel,
		callSkip:    1,
		module:      "default",
		serviceName: "default",
		infoWriter:  os.Stdout,
		errorWriter: os.Stdout,
	}
}

func WithCallerSkip(skip int) Option {
	return func(o *Options) { o.callSkip = skip }
}

func WithModule(module string) Option {
	return func(o *Options) { o.module = module }
}

func WithServiceName(serviceName string) Option {
	return func(o *Options) { o.serviceName = serviceName }
}

func WithInfoWriter(w io.Writer) Option {
	return func(o *Options) { o.infoWriter = w }
}

func WithErrorWriter(w io.Writer) Option {
	return func(o *Options) { o.errorWriter = w }
}

func WithLevel(l string) Option {
	return func(o *Options) { o.level = logger.StringToLevel(l) }
}

func NewLogger(options ...Option) (l *ZapLogger, err error) {
	opts := defaultOptions()
	for _, o := range options {
		o(opts)
	}

	l = &ZapLogger{
		opts: opts,
	}

	encoder := l.formatEncoder()

	infoEnabler := l.infoEnabler()
	errorEnabler := l.errorEnabler()

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(opts.infoWriter), infoEnabler),
		zapcore.NewCore(encoder, zapcore.AddSync(opts.errorWriter), errorEnabler),
	)

	fields := []zapcore.Field{
		zap.String(logger.AppName, app.Name()),
		zap.String(logger.Module, l.opts.module),
		zap.String(logger.ServiceName, l.opts.serviceName),
	}

	l.Logger = zap.New(core,
		zap.AddCaller(),
		// zap.AddStacktrace(errorEnabler),
		zap.AddCallerSkip(l.opts.callSkip),
		zap.Fields(fields...),
	)

	return
}

func (l *ZapLogger) infoEnabler() zap.LevelEnablerFunc {
	return zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		// current write level less than logger level, not write
		if lvl < zapLevel(l.opts.level) {
			return false
		}
		// current write level must less than or equal to info
		return lvl <= zapcore.InfoLevel
	})
}

func (l *ZapLogger) errorEnabler() zap.LevelEnablerFunc {
	return zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		// current write level less than logger level, not write
		if lvl < zapLevel(l.opts.level) {
			return false
		}
		// current write level must large than or equal to warn
		return lvl >= zapcore.WarnLevel
	})
}

func (l *ZapLogger) formatEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "level",
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		TimeKey:       "time",
		CallerKey:     "file",
		FunctionKey:   "func",
		StacktraceKey: "stack",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})
}

func (l *ZapLogger) GetLevel() logger.Level {
	return l.opts.level
}

func zapLevel(level logger.Level) zapcore.Level {
	switch level {
	case logger.DebugLevel:
		return zapcore.DebugLevel
	case logger.InfoLevel:
		return zapcore.InfoLevel
	case logger.WarnLevel:
		return zapcore.WarnLevel
	case logger.ErrorLevel:
		return zapcore.ErrorLevel
	case logger.FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func (l *ZapLogger) Debug(ctx context.Context, msg string, fields ...logger.Field) {
	l.Logger.Debug(msg, l.extractFields(ctx, fields...)...)
}

func (l *ZapLogger) Info(ctx context.Context, msg string, fields ...logger.Field) {
	l.Logger.Info(msg, l.extractFields(ctx, fields...)...)
}

func (l *ZapLogger) Warn(ctx context.Context, msg string, fields ...logger.Field) {
	l.Logger.Warn(msg, l.extractFields(ctx, fields...)...)
}

func (l *ZapLogger) Error(ctx context.Context, msg string, fields ...logger.Field) {
	l.Logger.Error(msg, l.extractFields(ctx, fields...)...)
}

func (l *ZapLogger) Fatal(ctx context.Context, msg string, fields ...logger.Field) {
	l.Logger.Fatal(msg, l.extractFields(ctx, fields...)...)
}

func (l *ZapLogger) extractFields(ctx context.Context, fields ...logger.Field) []zap.Field {
	ctx = logger.ForkContext(ctx)
	logger.AddField(ctx, fields...)

	fs := []zap.Field{}
	logger.RangeFields(ctx, func(f logger.Field) {
		fs = append(fs, zap.Reflect(f.Key(), f.Value()))
	})

	return fs
}

func (l *ZapLogger) Close() error {
	return l.Sync()
}
