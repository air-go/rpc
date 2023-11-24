package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitFieldsContainer(t *testing.T) {
	ctx := InitFieldsContainer(context.Background())

	AddField(ctx, Reflect("key", "val"))
	assert.Equal(t, "val", FindField(ctx, "key").Value())

	AddField(ctx, Reflect("key", "val1"))
	assert.Equal(t, "val1", FindField(ctx, "key").Value())

	AddField(ctx, Reflect("key1", "val"))
	assert.Equal(t, 2, len(ExtractFields(ctx)))

	DeleteField(ctx, "key1")

	assert.Equal(t, "key", ExtractFields(ctx)[0].Key())

	newCtx := ForkContext(ctx)
	DeleteField(newCtx, "key")

	assert.Equal(t, "key", ExtractFields(ctx)[0].Key())
	assert.Equal(t, 0, len(ExtractFields(newCtx)))

	AddField(ctx, Reflect(AppName, "app_name"))
	AddField(ctx, Reflect(LogID, "log_id"))
	AddField(ctx, Reflect(TraceID, "trace_id"))
	newCtx = ForkContextOnlyMeta(ctx)
	assert.Equal(t, 3, len(ExtractFields(newCtx)))
}
