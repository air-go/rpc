package logger

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestWithLogID(t *testing.T) {
	convey.Convey("TestWithLogID", t, func() {
		convey.Convey("success", func() {
			ctx := context.TODO()
			val := "id"
			ctx = WithLogID(ctx, val)
			id, ok := ctx.Value(contextLogID).(string)
			assert.Equal(t, ok, true)
			assert.Equal(t, id, val)
		})
	})
}

func TestValueLogID(t *testing.T) {
	convey.Convey("TestValueLogID", t, func() {
		convey.Convey("success", func() {
			ctx := context.TODO()
			val := "id"
			ctx = context.WithValue(ctx, contextLogID, val)
			id := ValueLogID(ctx)
			assert.Equal(t, id, val)
		})
		convey.Convey("empty", func() {
			ctx := context.TODO()
			id := ValueLogID(ctx)
			assert.Equal(t, id, "")
		})
	})
}

func TestWithTraceID(t *testing.T) {
	convey.Convey("TestWithTraceID", t, func() {
		convey.Convey("success", func() {
			ctx := context.TODO()
			val := "id"
			ctx = WithTraceID(ctx, val)
			id, ok := ctx.Value(contextTraceID).(string)
			assert.Equal(t, ok, true)
			assert.Equal(t, id, val)
		})
	})
}

func TestValueTraceID(t *testing.T) {
	convey.Convey("TestValueTraceID", t, func() {
		convey.Convey("success", func() {
			ctx := context.TODO()
			val := "id"
			ctx = context.WithValue(ctx, contextTraceID, val)
			id := ValueTraceID(ctx)
			assert.Equal(t, id, val)
		})
		convey.Convey("empty", func() {
			ctx := context.TODO()
			id := ValueTraceID(ctx)
			assert.Equal(t, id, "")
		})
	})
}

func TestWithFields(t *testing.T) {
	convey.Convey("TestWithFields", t, func() {
		convey.Convey("success", func() {
			ctx := context.TODO()
			fields := []Field{Reflect("a", "a")}
			ctx = WithFields(ctx, fields)
			res := ValueFields(ctx)
			assert.Equal(t, res, fields)
		})
	})
}

func TestValueFields(t *testing.T) {
	convey.Convey("TestValueFields", t, func() {
		convey.Convey("success", func() {
			ctx := context.TODO()
			fields := []Field{Reflect("a", "a")}
			ctx = context.WithValue(ctx, contextHTTPLogFields, fields)
			res := ValueFields(ctx)
			assert.Equal(t, res, fields)
		})
		convey.Convey("empty", func() {
			ctx := context.TODO()
			res := ValueFields(ctx)
			assert.Equal(t, res, []Field{})
		})
	})
}

func TestAddField(t *testing.T) {
	convey.Convey("TestAddField", t, func() {
		convey.Convey("success", func() {
			ctx := context.TODO()
			ctx = context.WithValue(ctx, contextHTTPLogFields, []Field{Reflect("a", "a")})
			ctx = AddField(ctx, []Field{Reflect("b", "b")}...)
			want := []Field{
				Reflect("a", "a"),
				Reflect("b", "b"),
			}

			res := ValueFields(ctx)
			assert.Equal(t, res, want)
		})
	})
}
