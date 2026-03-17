package client

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/air-go/rpc/llm/mcp/client/hook"
)

func setupTestServer(t *testing.T) *server.MCPServer {
	s := server.NewMCPServer("test-server", "1.0.0")

	tool := mcp.NewTool("echo",
		mcp.WithDescription("Echo tool"),
		mcp.WithString("message",
			mcp.Description("Message to echo"),
			mcp.Required(),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		message := req.GetString("message", "")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "Echo: " + message,
				},
			},
		}, nil
	}

	s.AddTool(tool, handler)
	return s
}

func TestNewClient(t *testing.T) {
	c := NewClient("test-server")
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.serverName != "test-server" {
		t.Errorf("Expected serverName 'test-server', got '%s'", c.serverName)
	}
	if c.hooks == nil {
		t.Error("Expected hooks to be initialized")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	hookChain := hook.NewClientHookChain()
	c := NewClient(
		"test-server",
		WithHooks(hookChain),
	)

	if c.serverName != "test-server" {
		t.Errorf("Expected serverName 'test-server', got '%s'", c.serverName)
	}
	if c.hooks != hookChain {
		t.Error("Expected custom hookChain to be set")
	}
}

func TestNewInProcessClient(t *testing.T) {
	mcpServer := setupTestServer(t)

	c, err := NewInProcessClient(mcpServer, "test-server")
	if err != nil {
		t.Fatalf("NewInProcessClient returned error: %v", err)
	}
	if c == nil {
		t.Fatal("NewInProcessClient returned nil client")
	}
	if c.mcpClient == nil {
		t.Error("Expected mcpClient to be initialized")
	}
	defer c.Close()
}

func TestNewInProcessClient_WithOptions(t *testing.T) {
	mcpServer := setupTestServer(t)
	hookChain := hook.NewClientHookChain()

	c, err := NewInProcessClient(mcpServer, "test-server", WithHooks(hookChain))
	if err != nil {
		t.Fatalf("NewInProcessClient returned error: %v", err)
	}
	if c.serverName != "test-server" {
		t.Errorf("Expected serverName 'test-server', got '%s'", c.serverName)
	}
	if c.hooks != hookChain {
		t.Error("Expected custom hookChain to be set")
	}
	defer c.Close()
}

func TestClient_Initialize(t *testing.T) {
	mcpServer := setupTestServer(t)

	c, err := NewInProcessClient(mcpServer, "test-server")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()
	initReq := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "test-client",
				Version: "1.0.0",
			},
		},
	}

	result, err := c.Initialize(ctx, initReq)
	if err != nil {
		t.Errorf("Initialize returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.ServerInfo.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", result.ServerInfo.Name)
	}
}

func TestClient_CallTool(t *testing.T) {
	mcpServer := setupTestServer(t)

	c, err := NewInProcessClient(mcpServer, "test-server")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()

	// 初始化
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
		t.Fatalf("Failed to initialize: %v", err)
	}

	// 调用工具
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "echo",
			Arguments: map[string]interface{}{
				"message": "test message",
			},
		},
	}

	result, err := c.CallTool(ctx, req)
	if err != nil {
		t.Errorf("CallTool returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if len(result.Content) == 0 {
		t.Fatal("Expected content in result")
	}
}

func TestClient_CallTool_WithHooks(t *testing.T) {
	mcpServer := setupTestServer(t)

	beforeCalled := false
	afterCalled := false

	hookChain := hook.NewClientHookChain()
	hookChain.RegisterBeforeCall(func(ctx context.Context, hookCtx *hook.ClientHookContext) error {
		beforeCalled = true
		return nil
	})
	hookChain.RegisterAfterCall(func(ctx context.Context, hookCtx *hook.ClientHookContext) {
		afterCalled = true
	})

	c, err := NewInProcessClient(mcpServer, "test-server", WithHooks(hookChain))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()

	// 初始化
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
		t.Fatalf("Failed to initialize: %v", err)
	}

	// 调用工具
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "echo",
			Arguments: map[string]interface{}{
				"message": "test message",
			},
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

func TestClient_CallTool_WithError(t *testing.T) {
	errorCalled := false

	hookChain := hook.NewClientHookChain()
	hookChain.RegisterOnError(func(ctx context.Context, hookCtx *hook.ClientHookContext) {
		errorCalled = true
	})

	mcpServer := server.NewMCPServer("test-server", "1.0.0")

	errorTool := mcp.NewTool("error_tool",
		mcp.WithDescription("Tool that returns error"),
	)
	errorHandler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, errors.New("tool error")
	}
	mcpServer.AddTool(errorTool, errorHandler)

	c, err := NewInProcessClient(mcpServer, "test-server", WithHooks(hookChain))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()

	// 初始化
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
		t.Fatalf("Failed to initialize: %v", err)
	}

	// 调用会出错的工具
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "error_tool",
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

func TestClient_ListTools(t *testing.T) {
	mcpServer := setupTestServer(t)

	c, err := NewInProcessClient(mcpServer, "test-server")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()

	// 初始化
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
		t.Fatalf("Failed to initialize: %v", err)
	}

	// 列出工具
	result, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Errorf("ListTools returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if len(result.Tools) == 0 {
		t.Error("Expected at least one tool")
	}

	foundEcho := false
	for _, tool := range result.Tools {
		if tool.Name == "echo" {
			foundEcho = true
			break
		}
	}
	if !foundEcho {
		t.Error("Expected to find 'echo' tool")
	}
}

func TestClient_Close(t *testing.T) {
	mcpServer := setupTestServer(t)

	c, err := NewInProcessClient(mcpServer, "test-server")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = c.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}
