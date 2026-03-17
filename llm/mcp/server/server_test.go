package server

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/air-go/rpc/llm/mcp/client"
	"github.com/air-go/rpc/llm/mcp/server/hook"
)

func TestNewServer(t *testing.T) {
	s := NewServer("test-server", "1.0.0")
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.name != "test-server" {
		t.Errorf("Expected name 'test-server', got '%s'", s.name)
	}
	if s.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", s.version)
	}
	if s.mcpServer == nil {
		t.Error("Expected mcpServer to be initialized")
	}
	if s.hooks == nil {
		t.Error("Expected hooks to be initialized")
	}
}

func TestNewServer_WithOptions(t *testing.T) {
	hookChain := hook.NewHookChain()
	s := NewServer("test-server", "1.0.0", WithHooks(hookChain))

	if s.hooks != hookChain {
		t.Error("Expected custom hookChain to be set")
	}
}

func TestMCPServer_AddTool(t *testing.T) {
	s := NewServer("test-server", "1.0.0")

	called := false
	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		called = true
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "result"},
			},
		}, nil
	}

	tool := mcp.NewTool("test_tool",
		mcp.WithDescription("Test tool"),
	)
	s.AddTool(tool, handler)

	// 创建客户端并调用工具来测试
	c, err := client.NewInProcessClient(s.GetMCPServer(), "test-server")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()
	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "test_tool",
		},
	}

	result, err := c.CallTool(ctx, req)
	if err != nil {
		t.Errorf("CallTool returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if !called {
		t.Error("Handler was not called")
	}
}

func TestMCPServer_AddTool_WithHooks(t *testing.T) {
	beforeCalled := false
	afterCalled := false

	hookChain := hook.NewHookChain()
	hookChain.RegisterBeforeCall(func(ctx context.Context, hookCtx *hook.HookContext) error {
		beforeCalled = true
		return nil
	})
	hookChain.RegisterAfterCall(func(ctx context.Context, hookCtx *hook.HookContext) {
		afterCalled = true
	})

	s := NewServer("test-server", "1.0.0", WithHooks(hookChain))

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "result"},
			},
		}, nil
	}

	tool := mcp.NewTool("test_tool",
		mcp.WithDescription("Test tool"),
	)
	s.AddTool(tool, handler)

	// 创建客户端
	c, err := client.NewInProcessClient(s.GetMCPServer(), "test-server")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()
	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	// 调用工具
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "test_tool",
		},
	}

	_, err = c.CallTool(ctx, req)
	if err != nil {
		t.Errorf("CallTool returned error: %v", err)
	}

	if !beforeCalled {
		t.Error("BeforeCall hook was not called")
	}
	if !afterCalled {
		t.Error("AfterCall hook was not called")
	}
}

func TestMCPServer_AddTool_WithError(t *testing.T) {
	errorCalled := false

	hookChain := hook.NewHookChain()
	hookChain.RegisterOnError(func(ctx context.Context, hookCtx *hook.HookContext) {
		errorCalled = true
	})

	s := NewServer("test-server", "1.0.0", WithHooks(hookChain))

	expectedErr := errors.New("test error")
	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, expectedErr
	}

	tool := mcp.NewTool("test_tool",
		mcp.WithDescription("Test tool"),
	)
	s.AddTool(tool, handler)

	// 创建客户端
	c, err := client.NewInProcessClient(s.GetMCPServer(), "test-server")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()
	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	// 调用工具
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "test_tool",
		},
	}

	_, err = c.CallTool(ctx, req)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errorCalled {
		t.Error("OnError hook was not called")
	}
}

func TestMCPServer_AddTool_HookInterrupt(t *testing.T) {
	handlerCalled := false

	hookChain := hook.NewHookChain()
	hookChain.RegisterBeforeCall(func(ctx context.Context, hookCtx *hook.HookContext) error {
		return errors.New("interrupted by hook")
	})

	s := NewServer("test-server", "1.0.0", WithHooks(hookChain))

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		handlerCalled = true
		return &mcp.CallToolResult{}, nil
	}

	tool := mcp.NewTool("test_tool",
		mcp.WithDescription("Test tool"),
	)
	s.AddTool(tool, handler)

	// 创建客户端
	c, err := client.NewInProcessClient(s.GetMCPServer(), "test-server")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()
	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	// 调用工具
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "test_tool",
		},
	}

	_, err = c.CallTool(ctx, req)
	if err == nil {
		t.Error("Expected error from hook, got nil")
	}

	if handlerCalled {
		t.Error("Handler should not have been called when hook returned error")
	}
}

func TestMCPServer_GetMCPServer(t *testing.T) {
	s := NewServer("test-server", "1.0.0")

	underlyingServer := s.GetMCPServer()
	if underlyingServer == nil {
		t.Error("GetMCPServer returned nil")
	}
	if underlyingServer != s.mcpServer {
		t.Error("GetMCPServer should return the same mcpServer")
	}
}

func TestMCPServer_WrapHandler(t *testing.T) {
	s := NewServer("test-server", "1.0.0")

	called := false
	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		called = true
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "result"},
			},
		}, nil
	}

	wrapped := s.wrapHandler(handler)
	if wrapped == nil {
		t.Fatal("wrapHandler returned nil")
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "test_tool",
		},
	}

	result, err := wrapped(context.Background(), req)
	if err != nil {
		t.Errorf("Wrapped handler returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if !called {
		t.Error("Handler was not called")
	}
}

func TestMCPServer_WrapHandler_WithHookContext(t *testing.T) {
	s := NewServer("test-server", "1.0.0")

	// 测试 hook context 的设置
	var capturedHookCtx *hook.HookContext

	hookChain := hook.NewHookChain()
	hookChain.RegisterBeforeCall(func(ctx context.Context, hookCtx *hook.HookContext) error {
		capturedHookCtx = hookCtx
		return nil
	})

	s.hooks = hookChain

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{}, nil
	}

	wrapped := s.wrapHandler(handler)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "test_tool",
		},
	}

	_, _ = wrapped(context.Background(), req)

	if capturedHookCtx == nil {
		t.Fatal("Expected hookCtx to be set")
	}
	if capturedHookCtx.ToolName != "test_tool" {
		t.Errorf("Expected ToolName 'test_tool', got '%s'", capturedHookCtx.ToolName)
	}
	if capturedHookCtx.Request == nil {
		t.Error("Expected Request to be set")
	}
	if capturedHookCtx.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}
}

func TestMCPServer_MultipleTools(t *testing.T) {
	s := NewServer("test-server", "1.0.0")

	// 添加多个工具
	tools := []struct {
		name string
		fn   func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{
			name: "tool1",
			fn: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						mcp.TextContent{Type: "text", Text: "tool1 result"},
					},
				}, nil
			},
		},
		{
			name: "tool2",
			fn: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						mcp.TextContent{Type: "text", Text: "tool2 result"},
					},
				}, nil
			},
		},
	}

	for _, tool := range tools {
		toolDef := mcp.NewTool(tool.name, mcp.WithDescription(tool.name))
		s.AddTool(toolDef, tool.fn)
	}

	// 创建客户端来测试
	c, err := client.NewInProcessClient(s.GetMCPServer(), "test-server")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()
	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	// 测试每个工具都能正常工作
	for _, tool := range tools {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: tool.name,
			},
		}

		result, err := c.CallTool(ctx, req)
		if err != nil {
			t.Errorf("Tool %s returned error: %v", tool.name, err)
		}
		if result == nil {
			t.Fatalf("Tool %s returned nil result", tool.name)
		}
	}
}
