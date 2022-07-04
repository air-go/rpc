package logger

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestLevel_String(t *testing.T) {
	convey.Convey("TestLevel_String", t, func() {
		convey.Convey("debug", func() {
			assert.Equal(t, DebugLevel.String(), "debug")
		})
		convey.Convey("info", func() {
			assert.Equal(t, InfoLevel.String(), "info")
		})
		convey.Convey("warn", func() {
			assert.Equal(t, WarnLevel.String(), "warn")
		})
		convey.Convey("error", func() {
			assert.Equal(t, ErrorLevel.String(), "error")
		})
		convey.Convey("fatal", func() {
			assert.Equal(t, FatalLevel.String(), "fatal")
		})
		convey.Convey("unknown", func() {
			assert.Equal(t, UnknownLevel.String(), "unknown")
		})
	})
}

func TestStringToLevel(t *testing.T) {
	convey.Convey("TestStringToLevel", t, func() {
		convey.Convey("debug", func() {
			assert.Equal(t, DebugLevel, StringToLevel("debug"))
		})
		convey.Convey("info", func() {
			assert.Equal(t, InfoLevel, StringToLevel("info"))
		})
		convey.Convey("warn", func() {
			assert.Equal(t, WarnLevel, StringToLevel("warn"))
		})
		convey.Convey("error", func() {
			assert.Equal(t, ErrorLevel, StringToLevel("error"))
		})
		convey.Convey("fatal", func() {
			assert.Equal(t, FatalLevel, StringToLevel("fatal"))
		})
		convey.Convey("unknown", func() {
			assert.Equal(t, UnknownLevel, StringToLevel("unknown"))
		})
	})
}
