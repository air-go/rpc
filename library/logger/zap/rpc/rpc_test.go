package rpc

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/air-go/rpc/library/logger"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewRPCLogger(t *testing.T) {
	convey.Convey("TestNewRPCLogger", t, func() {
		convey.Convey("success", func() {
			l, err := NewRPCLogger(&RPCConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)
		})
	})
}

func TestRPCLoggerWrite(t *testing.T) {
	l, err := NewRPCLogger(&RPCConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, l)

	ctx := logger.InitFieldsContainer(context.Background())
	convey.Convey("TestRPCLoggerWrite", t, func() {
		convey.Convey("Info", func() {
			l.Info(ctx, "msg")
		})
		convey.Convey("Error", func() {
			l.Error(ctx, "msg")
		})
	})
}

func TestNewRPCLoggerWithWriter(t *testing.T) {
	ctx := logger.InitFieldsContainer(context.Background())

	convey.Convey("TestNewRPCLoggerWithWriter", t, func() {
		convey.Convey("injected writers receive logs", func() {
			dir := t.TempDir()
			infoFile := filepath.Join(dir, "info.log")
			errFile := filepath.Join(dir, "error.log")
			infoBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}

			l, err := NewRPCLogger(&RPCConfig{Level: "info", InfoFile: infoFile, ErrorFile: errFile},
				WithWriter(infoBuf, errBuf))
			assert.Nil(t, err)

			l.Info(ctx, "outbound")
			l.Error(ctx, "failed")

			assert.Contains(t, infoBuf.String(), "outbound")
			assert.Contains(t, errBuf.String(), "failed")
			assert.NoFileExists(t, infoFile)
			assert.NoFileExists(t, errFile)
		})
	})
}

func TestNewRPCLoggerRejectsSingleWriter(t *testing.T) {
	l, err := NewRPCLogger(&RPCConfig{}, WithWriter(nil, &bytes.Buffer{}))

	assert.NotNil(t, err)
	assert.Nil(t, l)
}
