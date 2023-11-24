package context

import (
	"context"
)

type contextKey uint64

const (
	contextLogID contextKey = iota
	contextLogContainer
	contextTraceID
	contextResponseWriter
)

// WithLogID inject log id to context
func WithLogID(ctx context.Context, val interface{}) context.Context {
	return context.WithValue(ctx, contextLogID, val)
}

// ValueLogID extract log id from context
func ValueLogID(ctx context.Context) string {
	val := ctx.Value(contextLogID)
	logID, ok := val.(string)
	if !ok {
		return ""
	}
	return logID
}

// WithTraceID inject trace_id id to context
func WithTraceID(ctx context.Context, val interface{}) context.Context {
	return context.WithValue(ctx, contextTraceID, val)
}

// ValueTraceID extract trace id from context
func ValueTraceID(ctx context.Context) string {
	val := ctx.Value(contextTraceID)
	logID, ok := val.(string)
	if !ok {
		return ""
	}
	return logID
}

// WithLogContainer inject log container to context
func WithLogContainer(ctx context.Context, val interface{}) context.Context {
	return context.WithValue(ctx, contextLogContainer, val)
}

// ValueLogContainer extract log container from context
func ValueLogContainer(ctx context.Context) any {
	return ctx.Value(contextLogContainer)
}

// WithResponseWriter inject response write to context
func WithResponseWriter(ctx context.Context, val interface{}) context.Context {
	return context.WithValue(ctx, contextResponseWriter, val)
}

// ValueResponseWriter extract response write from context
func ValueResponseWriter(ctx context.Context) interface{} {
	return ctx.Value(contextResponseWriter)
}
