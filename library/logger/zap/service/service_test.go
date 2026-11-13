package service

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/air-go/rpc/library/logger"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

type spyWriter struct {
	buf    bytes.Buffer
	closed bool
}

func (w *spyWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }

func (w *spyWriter) Close() error {
	w.closed = true
	return nil
}

func TestNewServiceLogger(t *testing.T) {
	convey.Convey("TestNewServiceLogger", t, func() {
		convey.Convey("success", func() {
			l, err := NewServiceLogger("test", &Config{})
			assert.Nil(t, err)
			assert.NotNil(t, l)
		})
	})
}

func TestNewServiceLoggerWithWriter(t *testing.T) {
	ctx := logger.InitFieldsContainer(context.Background())

	convey.Convey("TestNewServiceLoggerWithWriter", t, func() {
		convey.Convey("injected writers receive logs and no file is created", func() {
			dir := t.TempDir()
			infoFile := filepath.Join(dir, "info.log")
			errFile := filepath.Join(dir, "error.log")
			infoBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}

			l, err := NewServiceLogger("test", &Config{InfoFile: infoFile, ErrorFile: errFile, Level: "info"},
				WithWriter(infoBuf, errBuf))
			assert.Nil(t, err)

			l.Info(ctx, "hello")
			l.Error(ctx, "boom")

			assert.Contains(t, infoBuf.String(), "hello")
			assert.Contains(t, errBuf.String(), "boom")
			assert.NoFileExists(t, infoFile)
			assert.NoFileExists(t, errFile)
		})

		convey.Convey("each level is written exactly once", func() {
			buf := &bytes.Buffer{}

			l, err := NewServiceLogger("test", &Config{Level: "info"},
				WithWriter(buf, buf))
			assert.Nil(t, err)

			l.Info(ctx, "info-once")
			l.Warn(ctx, "warn-once")
			l.Error(ctx, "error-once")

			assert.Equal(t, 1, bytes.Count(buf.Bytes(), []byte("info-once")))
			assert.Equal(t, 1, bytes.Count(buf.Bytes(), []byte("warn-once")))
			assert.Equal(t, 1, bytes.Count(buf.Bytes(), []byte("error-once")))
		})

		convey.Convey("level gating is unchanged", func() {
			buf := &bytes.Buffer{}

			l, err := NewServiceLogger("test", &Config{Level: "error"},
				WithWriter(buf, buf))
			assert.Nil(t, err)

			l.Info(ctx, "dropped-info")
			l.Warn(ctx, "dropped-warn")
			l.Error(ctx, "kept")

			assert.NotContains(t, buf.String(), "dropped")
			assert.Equal(t, 1, bytes.Count(buf.Bytes(), []byte("kept")))
		})

		convey.Convey("zero value writer falls back to files", func() {
			dir := t.TempDir()
			infoFile := filepath.Join(dir, "info.log")
			errFile := filepath.Join(dir, "error.log")

			l, err := NewServiceLogger("test", &Config{InfoFile: infoFile, ErrorFile: errFile, Level: "info"},
				WithWriter(nil, nil))
			assert.Nil(t, err)

			l.Info(ctx, "msg")

			assert.FileExists(t, infoFile)
		})

		convey.Convey("Close does not close injected writers", func() {
			spy := &spyWriter{}

			l, err := NewServiceLogger("test", &Config{Level: "info"},
				WithWriter(spy, spy))
			assert.Nil(t, err)

			l.Info(ctx, "msg")
			assert.Nil(t, l.Close())
			assert.False(t, spy.closed)
		})

		convey.Convey("a single writer is rejected", func() {
			l, err := NewServiceLogger("test", &Config{Level: "info"}, WithWriter(&bytes.Buffer{}, nil))

			assert.NotNil(t, err)
			assert.Nil(t, l)
		})
	})
}
