package hook

import (
	"context"
	"sync"
)

// HookChain manages the execution chain of hooks, supports concurrent safety
type HookChain struct {
	mu          sync.RWMutex
	beforeCalls []ToolBeforeCall
	afterCalls  []ToolAfterCall
	onErrors    []ToolOnError
}

// NewHookChain creates a new hook chain
func NewHookChain() *HookChain {
	return &HookChain{
		beforeCalls: make([]ToolBeforeCall, 0),
		afterCalls:  make([]ToolAfterCall, 0),
		onErrors:    make([]ToolOnError, 0),
	}
}

// RegisterBeforeCall registers a before call hook
func (c *HookChain) RegisterBeforeCall(h ToolBeforeCall) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.beforeCalls = append(c.beforeCalls, h)
}

// RegisterAfterCall registers an after call hook
func (c *HookChain) RegisterAfterCall(h ToolAfterCall) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.afterCalls = append(c.afterCalls, h)
}

// RegisterOnError registers an error handler hook
func (c *HookChain) RegisterOnError(h ToolOnError) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onErrors = append(c.onErrors, h)
}

// ExecuteBefore executes all before hooks, stops if any returns error
func (c *HookChain) ExecuteBefore(ctx context.Context, hookCtx *HookContext) error {
	c.mu.RLock()
	hooks := make([]ToolBeforeCall, len(c.beforeCalls))
	copy(hooks, c.beforeCalls)
	c.mu.RUnlock()

	for _, h := range hooks {
		if err := h(ctx, hookCtx); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteAfter executes all after hooks
func (c *HookChain) ExecuteAfter(ctx context.Context, hookCtx *HookContext) {
	c.mu.RLock()
	hooks := make([]ToolAfterCall, len(c.afterCalls))
	copy(hooks, c.afterCalls)
	c.mu.RUnlock()

	for _, h := range hooks {
		h(ctx, hookCtx)
	}
}

// ExecuteError executes all error handler hooks
func (c *HookChain) ExecuteError(ctx context.Context, hookCtx *HookContext) {
	c.mu.RLock()
	hooks := make([]ToolOnError, len(c.onErrors))
	copy(hooks, c.onErrors)
	c.mu.RUnlock()

	for _, h := range hooks {
		h(ctx, hookCtx)
	}
}

// RegisterHook registers a complete hook interface
func (c *HookChain) RegisterHook(h ServerToolHook) {
	c.RegisterBeforeCall(h.BeforeCall)
	c.RegisterAfterCall(h.AfterCall)
	c.RegisterOnError(h.OnError)
}
