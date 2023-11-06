package rpc

import (
	"context"
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
