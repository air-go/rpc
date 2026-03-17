package hook

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName   = "mcp/server"
	spanNameTool = "mcp.tool.call"
)

// OtelServerHook is the server-side OpenTelemetry tracing hook
type OtelServerHook struct {
	tracer trace.Tracer
}

// NewOtelServerHook creates a new Otel server hook
func NewOtelServerHook() *OtelServerHook {
	return &OtelServerHook{
		tracer: otel.Tracer(tracerName),
	}
}

// BeforeCall creates a span before tool execution
func (h *OtelServerHook) BeforeCall(ctx context.Context, hookCtx *HookContext) error {
	ctx, span := h.tracer.Start(ctx, spanNameTool,
		trace.WithAttributes(
			attribute.String("mcp.tool.name", hookCtx.ToolName),
		),
	)

	// Store span in metadata for later use
	hookCtx.Metadata["otel_span"] = span
	hookCtx.Metadata["otel_ctx"] = ctx

	// Record argument information
	if hookCtx.Request != nil {
		args := hookCtx.Request.GetArguments()
		if len(args) > 0 {
			span.SetAttributes(
				attribute.Int("mcp.tool.args.count", len(args)),
			)
		}
	}

	return nil
}

// AfterCall ends the span and records response information
func (h *OtelServerHook) AfterCall(ctx context.Context, hookCtx *HookContext) {
	span, ok := hookCtx.Metadata["otel_span"].(trace.Span)
	if !ok || span == nil {
		return
	}
	defer span.End()

	if hookCtx.Response != nil {
		span.SetAttributes(
			attribute.String("mcp.tool.result.type", "success"),
			attribute.Int("mcp.tool.result.content_length", len(hookCtx.Response.Content)),
		)
	}
}

// OnError records the error to the span
func (h *OtelServerHook) OnError(ctx context.Context, hookCtx *HookContext) {
	span, ok := hookCtx.Metadata["otel_span"].(trace.Span)
	if !ok || span == nil {
		return
	}

	span.SetStatus(codes.Error, "tool execution failed")
	if hookCtx.Error != nil {
		span.RecordError(hookCtx.Error)
		span.SetAttributes(
			attribute.String("mcp.tool.error", hookCtx.Error.Error()),
		)
	}
	span.End()
}
