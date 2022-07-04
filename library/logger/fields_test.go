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

func TestFind(t *testing.T) {
	convey.Convey("TestFind", t, func() {
		convey.Convey("success", func() {
			fields := []Field{newFields("key", "value")}
			f := Find("key", fields)
			assert.Equal(t, f.Key(), "key")
			assert.Equal(t, f.Value(), "value")
		})
		convey.Convey("not find", func() {
			fields := []Field{}
			f := Find("key", fields)
			assert.Equal(t, f.Key(), "nil")
			assert.Equal(t, f.Value(), "nil")
		})
	})
}
