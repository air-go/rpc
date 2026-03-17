package hook

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// HookContext provides context information for hook execution
type HookContext struct {
	ToolName string                        // Tool name
	Request  *mcp.CallToolRequest          // Call request
	Response *mcp.CallToolResult          // Call response
	Error    error                        // Error information
	Metadata map[string]interface{}       // Extensible metadata
}

// ToolBeforeCall is called before tool execution
// Return error to interrupt the call flow
type ToolBeforeCall func(ctx context.Context, hookCtx *HookContext) error

// ToolAfterCall is called after successful tool execution
type ToolAfterCall func(ctx context.Context, hookCtx *HookContext)

// ToolOnError is called when tool execution fails
type ToolOnError func(ctx context.Context, hookCtx *HookContext)

// ServerToolHook defines the hook interface for server-side tool calls
type ServerToolHook interface {
	// BeforeCall is executed before tool is called
	// Return error to interrupt the call flow
	BeforeCall(ctx context.Context, hookCtx *HookContext) error

	// AfterCall is executed after successful tool call
	AfterCall(ctx context.Context, hookCtx *HookContext)

	// OnError is executed when tool call fails
	OnError(ctx context.Context, hookCtx *HookContext)
}
