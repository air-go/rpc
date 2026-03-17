package hook

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// ClientHookContext provides context information for client hook execution
type ClientHookContext struct {
	ServerName string                        // Server name
	ToolName   string                        // Tool name
	Request    *mcp.CallToolRequest          // Call request
	Response   *mcp.CallToolResult          // Call response
	Error      error                        // Error information
	Metadata   map[string]interface{}       // Extensible metadata
}

// ClientToolBeforeCall is called before tool is called
// Return error to interrupt the call flow
type ClientToolBeforeCall func(ctx context.Context, hookCtx *ClientHookContext) error

// ClientToolAfterCall is called after successful tool execution
type ClientToolAfterCall func(ctx context.Context, hookCtx *ClientHookContext)

// ClientToolOnError is called when tool execution fails
type ClientToolOnError func(ctx context.Context, hookCtx *ClientHookContext)

// ClientToolHook defines the hook interface for client-side tool calls
type ClientToolHook interface {
	// BeforeCall is executed before tool is called
	// Return error to interrupt the call flow
	BeforeCall(ctx context.Context, hookCtx *ClientHookContext) error

	// AfterCall is executed after successful tool call
	AfterCall(ctx context.Context, hookCtx *ClientHookContext)

	// OnError is executed when tool call fails
	OnError(ctx context.Context, hookCtx *ClientHookContext)
}
