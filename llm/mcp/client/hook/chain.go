package hook

import (
	"context"
	"sync"
)

// ClientHookChain manages the execution chain of client hooks, supports concurrent safety
type ClientHookChain struct {
	mu          sync.RWMutex
	beforeCalls []ClientToolBeforeCall
	afterCalls  []ClientToolAfterCall
	onErrors    []ClientToolOnError
}

// NewClientHookChain creates a new client hook chain
func NewClientHookChain() *ClientHookChain {
	return &ClientHookChain{
		beforeCalls: make([]ClientToolBeforeCall, 0),
		afterCalls:  make([]ClientToolAfterCall, 0),
		onErrors:    make([]ClientToolOnError, 0),
	}
}

// RegisterBeforeCall registers a before call hook
func (c *ClientHookChain) RegisterBeforeCall(h ClientToolBeforeCall) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.beforeCalls = append(c.beforeCalls, h)
}

// RegisterAfterCall registers an after call hook
func (c *ClientHookChain) RegisterAfterCall(h ClientToolAfterCall) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.afterCalls = append(c.afterCalls, h)
}

// RegisterOnError registers an error handler hook
func (c *ClientHookChain) RegisterOnError(h ClientToolOnError) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onErrors = append(c.onErrors, h)
}

// ExecuteBefore executes all before hooks, stops if any returns error
func (c *ClientHookChain) ExecuteBefore(ctx context.Context, hookCtx *ClientHookContext) error {
	c.mu.RLock()
	hooks := make([]ClientToolBeforeCall, len(c.beforeCalls))
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
func (c *ClientHookChain) ExecuteAfter(ctx context.Context, hookCtx *ClientHookContext) {
	c.mu.RLock()
	hooks := make([]ClientToolAfterCall, len(c.afterCalls))
	copy(hooks, c.afterCalls)
	c.mu.RUnlock()

	for _, h := range hooks {
		h(ctx, hookCtx)
	}
}

// ExecuteError executes all error handler hooks
func (c *ClientHookChain) ExecuteError(ctx context.Context, hookCtx *ClientHookContext) {
	c.mu.RLock()
	hooks := make([]ClientToolOnError, len(c.onErrors))
	copy(hooks, c.onErrors)
	c.mu.RUnlock()

	for _, h := range hooks {
		h(ctx, hookCtx)
	}
}

// RegisterHook registers a complete client hook interface
func (c *ClientHookChain) RegisterHook(h ClientToolHook) {
	c.RegisterBeforeCall(h.BeforeCall)
	c.RegisterAfterCall(h.AfterCall)
	c.RegisterOnError(h.OnError)
}
