package rpc

import (
	"context"
	"testing"

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

	convey.Convey("TestRPCLoggerWrite", t, func() {
		convey.Convey("Info", func() {
			l.Info(context.Background(), "msg")
		})
		convey.Convey("Error", func() {
			l.Error(context.Background(), "msg")
		})
	})
}
