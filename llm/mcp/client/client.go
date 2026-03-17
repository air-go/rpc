package client

import (
	"context"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/air-go/rpc/llm/mcp/client/hook"
)

// Client is the MCP client wrapper
type Client struct {
	mcpClient  *mcpclient.Client
	hooks      *hook.ClientHookChain
	serverName string
}

// ClientOption is the client configuration option
type ClientOption func(*Client)

// WithHooks configures the ClientHookChain
func WithHooks(h *hook.ClientHookChain) ClientOption {
	return func(c *Client) {
		c.hooks = h
	}
}

// NewClient creates a new MCP client (used for stdio/http transport)
// serverName is a required parameter to identify the MCP server
func NewClient(serverName string, opts ...ClientOption) *Client {
	c := &Client{
		hooks:      hook.NewClientHookChain(),
		serverName: serverName,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// NewInProcessClient creates an in-process MCP client connected to the server
// This provides zero network overhead local communication
// serverName is a required parameter to identify the MCP server
func NewInProcessClient(server *mcpserver.MCPServer, serverName string, opts ...ClientOption) (*Client, error) {
	mcpClient, err := mcpclient.NewInProcessClient(server)
	if err != nil {
		return nil, err
	}

	c := &Client{
		mcpClient:  mcpClient,
		hooks:      hook.NewClientHookChain(),
		serverName: serverName,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Initialize initializes the client
func (c *Client) Initialize(ctx context.Context, req mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return c.mcpClient.Initialize(ctx, req)
}

// CallTool calls a tool
func (c *Client) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	hookCtx := &hook.ClientHookContext{
		ServerName: c.serverName,
		ToolName:   req.Params.Name,
		Request:    &req,
		Metadata:   make(map[string]interface{}),
	}

	// Execute BeforeCall hooks
	if err := c.hooks.ExecuteBefore(ctx, hookCtx); err != nil {
		return nil, err
	}

	// Execute actual call
	result, err := c.mcpClient.CallTool(ctx, req)
	hookCtx.Response = result
	hookCtx.Error = err

	if err != nil {
		c.hooks.ExecuteError(ctx, hookCtx)
		return nil, err
	}

	// Execute AfterCall hooks
	c.hooks.ExecuteAfter(ctx, hookCtx)

	return result, nil
}

// ListTools lists available tools
func (c *Client) ListTools(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return c.mcpClient.ListTools(ctx, req)
}

// Close closes the client
func (c *Client) Close() error {
	return c.mcpClient.Close()
}
