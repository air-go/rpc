package hook

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	clientTracerName = "mcp/client"
	clientSpanName   = "mcp.client.call_tool"
)

// OtelClientHook is the client-side OpenTelemetry tracing hook
type OtelClientHook struct {
	tracer trace.Tracer
}

// NewOtelClientHook creates a new Otel client hook
func NewOtelClientHook() *OtelClientHook {
	return &OtelClientHook{
		tracer: otel.Tracer(clientTracerName),
	}
}

// BeforeCall creates a span before sending the request
func (h *OtelClientHook) BeforeCall(ctx context.Context, hookCtx *ClientHookContext) error {
	ctx, span := h.tracer.Start(ctx, clientSpanName,
		trace.WithAttributes(
			attribute.String("mcp.server.name", hookCtx.ServerName),
			attribute.String("mcp.tool.name", hookCtx.ToolName),
		),
	)

	hookCtx.Metadata["otel_span"] = span
	hookCtx.Metadata["otel_ctx"] = ctx

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

// AfterCall ends the span and records response
func (h *OtelClientHook) AfterCall(ctx context.Context, hookCtx *ClientHookContext) {
	span, ok := hookCtx.Metadata["otel_span"].(trace.Span)
	if !ok || span == nil {
		return
	}
	defer span.End()

	if hookCtx.Response != nil {
		span.SetAttributes(
			attribute.String("mcp.client.result.type", "success"),
			attribute.Int("mcp.client.result.content_length", len(hookCtx.Response.Content)),
		)
	}
}

// OnError records the error to the span
func (h *OtelClientHook) OnError(ctx context.Context, hookCtx *ClientHookContext) {
	span, ok := hookCtx.Metadata["otel_span"].(trace.Span)
	if !ok || span == nil {
		return
	}

	span.SetStatus(codes.Error, "client tool call failed")
	if hookCtx.Error != nil {
		span.RecordError(hookCtx.Error)
		span.SetAttributes(
			attribute.String("mcp.client.error", hookCtx.Error.Error()),
		)
	}
	span.End()
}
