package server

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/air-go/rpc/llm/mcp/server/hook"
)

// MCPServer is the MCP server wrapper
type MCPServer struct {
	mcpServer *server.MCPServer
	hooks     *hook.HookChain
	name      string
	version   string
}

// Option is the server configuration option
type Option func(*MCPServer)

// WithHooks configures the HookChain
func WithHooks(h *hook.HookChain) Option {
	return func(s *MCPServer) {
		s.hooks = h
	}
}

// NewServer creates a new MCP server
func NewServer(name, version string, opts ...Option) *MCPServer {
	mcpServer := server.NewMCPServer(name, version)
	hookChain := hook.NewHookChain()

	s := &MCPServer{
		mcpServer: mcpServer,
		hooks:     hookChain,
		name:      name,
		version:   version,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// AddTool registers a tool to the server
func (s *MCPServer) AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	// Wrap handler to integrate hooks
	wrappedHandler := s.wrapHandler(handler)
	s.mcpServer.AddTool(tool, wrappedHandler)
}

// wrapHandler wraps the tool handler to integrate HookChain
func (s *MCPServer) wrapHandler(handler server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		hookCtx := &hook.HookContext{
			ToolName: req.Params.Name,
			Request:  &req,
			Metadata: make(map[string]interface{}),
		}

		// Execute BeforeCall hooks
		if err := s.hooks.ExecuteBefore(ctx, hookCtx); err != nil {
			return nil, err
		}

		// Execute actual handler
		result, err := handler(ctx, req)
		hookCtx.Response = result
		hookCtx.Error = err

		if err != nil {
			s.hooks.ExecuteError(ctx, hookCtx)
			return nil, err
		}

		// Execute AfterCall hooks
		s.hooks.ExecuteAfter(ctx, hookCtx)

		return result, nil
	}
}

// GetMCPServer gets the underlying mcp-go server
// This is used to create an in-process client connection
func (s *MCPServer) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}
