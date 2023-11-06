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
