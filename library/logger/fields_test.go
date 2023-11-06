package logger

import (
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func newFields(key, value string) *field {
	return &field{key: key, value: value}
}

func Test_field_Key(t *testing.T) {
	convey.Convey("Test_field_Key", t, func() {
		convey.Convey("success", func() {
			field := newFields("key", "value")
			assert.Equal(t, field.Key(), "key")
		})
	})
}

func Test_field_Value(t *testing.T) {
	convey.Convey("Test_field_Key", t, func() {
		convey.Convey("success", func() {
			field := newFields("key", "value")
			assert.Equal(t, field.Value(), "value")
		})
	})
}

func TestReflect(t *testing.T) {
	convey.Convey("TestReflect", t, func() {
		convey.Convey("success", func() {
			field := Reflect("key", "value")
			assert.Equal(t, field.Key(), "key")
			assert.Equal(t, field.Value(), "value")
		})
	})
}

func TestError(t *testing.T) {
	convey.Convey("TestError", t, func() {
		convey.Convey("success", func() {
			err := errors.New("err")
			field := Error(err)
			assert.Equal(t, field.Key(), "error")
			assert.Equal(t, field.Value(), err)
		})
	})
}
