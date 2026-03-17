package hook

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"go.opentelemetry.io/otel/trace"
)

func TestNewOtelServerHook(t *testing.T) {
	hook := NewOtelServerHook()
	if hook == nil {
		t.Fatal("NewOtelServerHook returned nil")
	}
	if hook.tracer == nil {
		t.Error("Expected tracer to be initialized")
	}
}

func TestOtelServerHook_BeforeCall(t *testing.T) {
	hook := NewOtelServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Request: &mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "test_tool",
			},
		},
		Metadata: make(map[string]interface{}),
	}

	err := hook.BeforeCall(context.Background(), hookCtx)
	if err != nil {
		t.Errorf("BeforeCall returned unexpected error: %v", err)
	}

	// 验证 span 已存储在 metadata 中
	span, ok := hookCtx.Metadata["otel_span"].(trace.Span)
	if !ok || span == nil {
		t.Error("Expected otel_span to be stored in metadata")
	}

	// 验证 context 已存储
	ctx, ok := hookCtx.Metadata["otel_ctx"].(context.Context)
	if !ok || ctx == nil {
		t.Error("Expected otel_ctx to be stored in metadata")
	}
}

func TestOtelServerHook_BeforeCall_WithArguments(t *testing.T) {
	hook := NewOtelServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Request: &mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "test_tool",
				Arguments: map[string]interface{}{
					"arg1": "value1",
					"arg2": "value2",
				},
			},
		},
		Metadata: make(map[string]interface{}),
	}

	err := hook.BeforeCall(context.Background(), hookCtx)
	if err != nil {
		t.Errorf("BeforeCall returned unexpected error: %v", err)
	}

	// 验证 span 存在
	span, ok := hookCtx.Metadata["otel_span"].(trace.Span)
	if !ok || span == nil {
		t.Error("Expected otel_span to be stored in metadata")
	}
}

func TestOtelServerHook_AfterCall(t *testing.T) {
	hook := NewOtelServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Response: &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "result"},
			},
		},
		Metadata: make(map[string]interface{}),
	}

	// BeforeCall 设置 span
	_ = hook.BeforeCall(context.Background(), hookCtx)

	// AfterCall 应该结束 span
	hook.AfterCall(context.Background(), hookCtx)

	// span 应该被结束，这里我们只是验证不 panic
	span, ok := hookCtx.Metadata["otel_span"].(trace.Span)
	if ok && span != nil {
		// span 被结束了，这是正常的
	}
}

func TestOtelServerHook_AfterCall_WithoutSpan(t *testing.T) {
	hook := NewOtelServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Metadata: make(map[string]interface{}),
	}

	// 没有调用 BeforeCall，直接调用 AfterCall 应该不 panic
	hook.AfterCall(context.Background(), hookCtx)
}

func TestOtelServerHook_OnError(t *testing.T) {
	hook := NewOtelServerHook()

	testError := errors.New("test error")

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Error:    testError,
		Metadata: make(map[string]interface{}),
	}

	// BeforeCall 设置 span
	_ = hook.BeforeCall(context.Background(), hookCtx)

	// OnError 应该记录错误并结束 span
	hook.OnError(context.Background(), hookCtx)

	// 验证不 panic
	span, ok := hookCtx.Metadata["otel_span"].(trace.Span)
	if ok && span != nil {
		// span 被结束了，这是正常的
	}
}

func TestOtelServerHook_OnError_WithoutSpan(t *testing.T) {
	hook := NewOtelServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Error:    errors.New("test error"),
		Metadata: make(map[string]interface{}),
	}

	// 没有调用 BeforeCall，直接调用 OnError 应该不 panic
	hook.OnError(context.Background(), hookCtx)
}

func TestOtelServerHook_FullLifecycle(t *testing.T) {
	hook := NewOtelServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Request: &mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "test_tool",
			},
		},
		Metadata: make(map[string]interface{}),
	}

	// 模拟完整生命周期
	err := hook.BeforeCall(context.Background(), hookCtx)
	if err != nil {
		t.Errorf("BeforeCall returned error: %v", err)
	}

	// 执行成功
	hookCtx.Response = &mcp.CallToolResult{}
	hook.AfterCall(context.Background(), hookCtx)
}

func TestOtelServerHook_ErrorLifecycle(t *testing.T) {
	hook := NewOtelServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Request: &mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "test_tool",
			},
		},
		Metadata: make(map[string]interface{}),
	}

	// 模拟错误生命周期
	err := hook.BeforeCall(context.Background(), hookCtx)
	if err != nil {
		t.Errorf("BeforeCall returned error: %v", err)
	}

	// 执行失败
	hookCtx.Error = errors.New("execution error")
	hook.OnError(context.Background(), hookCtx)
}

func TestOtelServerHook_ServerToolHookInterface(t *testing.T) {
	// 确保 OtelServerHook 实现了 ServerToolHook 接口
	var _ ServerToolHook = (*OtelServerHook)(nil)
}
