package logger

import (
	"context"
)

type contextKey uint64

const (
	contextLogID contextKey = iota
	contextHTTPLogFields
	contextTraceID
	ContextRPC
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

// WithFields inject common http log fields to context
func WithFields(ctx context.Context, fields []Field) context.Context {
	return context.WithValue(ctx, contextHTTPLogFields, fields)
}

// ValueFields extrect common http log fields from context
func ValueFields(ctx context.Context) []Field {
	val := ctx.Value(contextHTTPLogFields)
	fields, ok := val.([]Field)
	if !ok {
		return []Field{}
	}
	return fields
}

func AddField(ctx context.Context, fields ...Field) context.Context {
	arr := ValueFields(ctx)

	oldIndex := map[string]int{}
	for i, f := range arr {
		oldIndex[f.Key()] = i
	}

	for _, f := range fields {
		if i, ok := oldIndex[f.Key()]; ok {
			arr[i] = f
			continue
		}
		arr = append(arr, f)
	}

	return WithFields(ctx, arr)
}
