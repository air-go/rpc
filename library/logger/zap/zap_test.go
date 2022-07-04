package zap

import (
	"context"
	"os"
	"testing"

	"github.com/air-go/rpc/library/logger"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func Test_defaultOptions(t *testing.T) {
	convey.Convey("Test_defaultOptions", t, func() {
		convey.Convey("success", func() {
			opts := defaultOptions()
			assert.Equal(t, opts.level, logger.InfoLevel)
			assert.Equal(t, opts.callSkip, 1)
			assert.Equal(t, opts.module, "default")
			assert.Equal(t, opts.serviceName, "default")
			assert.Equal(t, opts.infoWriter, os.Stdout)
			assert.Equal(t, opts.errorWriter, os.Stdout)
		})
	})
}

func TestWithOption(t *testing.T) {
	convey.Convey("TestWithOption", t, func() {
		convey.Convey("success", func() {
			option := &Options{}
			opts := []Option{
				WithCallerSkip(2),
				WithModule("error"),
				WithServiceName("test"),
				WithInfoWriter(os.Stderr),
				WithErrorWriter(os.Stderr),
				WithLevel("info"),
			}
			for _, o := range opts {
				o(option)
			}
			assert.Equal(t, option.callSkip, 2)
			assert.Equal(t, option.module, "error")
			assert.Equal(t, option.serviceName, "test")
			assert.Equal(t, option.infoWriter, os.Stderr)
			assert.Equal(t, option.errorWriter, os.Stderr)
			assert.Equal(t, option.level, logger.InfoLevel)
		})
	})
}

func TestNewLogger(t *testing.T) {
	convey.Convey("TestNewLogger", t, func() {
		convey.Convey("success", func() {
			l, err := NewLogger()
			assert.Nil(t, err)
			assert.NotEmpty(t, l)
		})
	})
}

func TestZapLogger_infoEnabler(t *testing.T) {
	l := &ZapLogger{
		opts: &Options{level: logger.InfoLevel},
	}
	convey.Convey("TestZapLogger_infoEnabler", t, func() {
		convey.Convey("current write level less than logger level", func() {
			f := l.infoEnabler()
			assert.Equal(t, false, f.Enabled(zapcore.DebugLevel))
		})
		convey.Convey("current write level large than logger level, lvl <= zapcore.InfoLevel", func() {
			f := l.infoEnabler()
			assert.Equal(t, true, f.Enabled(zapcore.InfoLevel))
		})
		convey.Convey("current write level large than logger level, lvl > zapcore.InfoLevel", func() {
			f := l.infoEnabler()
			assert.Equal(t, false, f.Enabled(zapcore.WarnLevel))
		})
	})
}

func TestZapLogger_errorEnabler(t *testing.T) {
	l := &ZapLogger{
		opts: &Options{level: logger.WarnLevel},
	}
	convey.Convey("TestZapLogger_errorEnabler", t, func() {
		convey.Convey("current write level less than logger level", func() {
			f := l.errorEnabler()
			assert.Equal(t, false, f.Enabled(zapcore.InfoLevel))
		})
		convey.Convey("current write level large than logger level, lvl >= zapcore.WarnLevel", func() {
			f := l.errorEnabler()
			assert.Equal(t, true, f.Enabled(zapcore.WarnLevel))
		})
		convey.Convey("current write level large than logger level, lvl < zapcore.WarnLevel", func() {
			f := l.errorEnabler()
			assert.Equal(t, false, f.Enabled(zapcore.InfoLevel))
		})
	})
}

func TestZapLogger_formatEncoder(t *testing.T) {
	convey.Convey("TestZapLogger_formatEncoder", t, func() {
		convey.Convey("success", func() {
			assert.NotEmpty(t, true, (&ZapLogger{}).formatEncoder())
		})
	})
}

func TestZapLogger_GetLevel(t *testing.T) {
	l := &ZapLogger{
		opts: &Options{level: logger.WarnLevel},
	}
	convey.Convey("TestZapLogger_GetLevel", t, func() {
		convey.Convey("success", func() {
			assert.NotEmpty(t, logger.WarnLevel, l.GetLevel())
		})
	})
}

func Test_zapLevel(t *testing.T) {
	convey.Convey("Test_zapLevel", t, func() {
		convey.Convey("DebugLevel", func() {
			assert.Equal(t, zapcore.DebugLevel, zapLevel(logger.DebugLevel))
		})
		convey.Convey("InfoLevel", func() {
			assert.Equal(t, zapcore.InfoLevel, zapLevel(logger.InfoLevel))
		})
		convey.Convey("WarnLevel", func() {
			assert.Equal(t, zapcore.WarnLevel, zapLevel(logger.WarnLevel))
		})
		convey.Convey("ErrorLevel", func() {
			assert.Equal(t, zapcore.ErrorLevel, zapLevel(logger.ErrorLevel))
		})
		convey.Convey("FatalLevel", func() {
			assert.Equal(t, zapcore.FatalLevel, zapLevel(logger.FatalLevel))
		})
		convey.Convey("default", func() {
			assert.Equal(t, zapcore.InfoLevel, zapLevel(logger.UnknownLevel))
		})
	})
}

func TestZapLoggerWrite(t *testing.T) {
	convey.Convey("TestZapLoggerWrite", t, func() {
		convey.Convey("Debug", func() {
			StdLogger.Debug(context.Background(), "msg")
		})
		convey.Convey("Info", func() {
			StdLogger.Info(context.Background(), "msg")
		})
		convey.Convey("Warn", func() {
			StdLogger.Warn(context.Background(), "msg")
		})
		convey.Convey("Error", func() {
			StdLogger.Error(context.Background(), "msg")
		})
	})
}

func TestZapLogger_extractFields(t *testing.T) {
	l := &ZapLogger{}
	convey.Convey("TestZapLogger_extractFields", t, func() {
		convey.Convey("sum fields", func() {
			ctx := context.TODO()
			ctx = logger.WithFields(ctx, []logger.Field{logger.Reflect("key", "value")})
			fields := l.extractFields(ctx, []logger.Field{logger.Reflect("key1", "value1")}...)

			count := 2
			assert.Equal(t, count, len(fields))

			actual := 0
			for _, f := range fields {
				if f.Key == "key" && f.Interface == "value" {
					actual += 1
				}
				if f.Key == "key1" && f.Interface == "value1" {
					actual += 1
				}
			}
			assert.Equal(t, count, actual)
		})
		convey.Convey("cover fields", func() {
			ctx := context.TODO()
			ctx = logger.WithFields(ctx, []logger.Field{logger.Reflect("key", "value")})
			fields := l.extractFields(ctx, []logger.Field{logger.Reflect("key", "value1")}...)

			count := 1
			assert.Equal(t, count, len(fields))

			actual := 0
			for _, f := range fields {
				if f.Key == "key" && f.Interface == "value1" {
					actual += 1
				}
			}
			assert.Equal(t, count, actual)
		})
	})
}
