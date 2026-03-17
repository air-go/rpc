package hook

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockHook 用于测试的 mock hook
type MockHook struct {
	beforeCalled bool
	afterCalled  bool
	errorCalled  bool
	beforeError  error
}

func (m *MockHook) BeforeCall(ctx context.Context, hookCtx *HookContext) error {
	m.beforeCalled = true
	return m.beforeError
}

func (m *MockHook) AfterCall(ctx context.Context, hookCtx *HookContext) {
	m.afterCalled = true
}

func (m *MockHook) OnError(ctx context.Context, hookCtx *HookContext) {
	m.errorCalled = true
}

func TestNewHookChain(t *testing.T) {
	chain := NewHookChain()
	if chain == nil {
		t.Fatal("NewHookChain returned nil")
	}
}

func TestHookChain_RegisterBeforeCall(t *testing.T) {
	chain := NewHookChain()
	mock := &MockHook{}

	chain.RegisterBeforeCall(mock.BeforeCall)

	// 验证 hook 已注册（通过执行来验证）
	hookCtx := &HookContext{
		ToolName: "test",
		Metadata: make(map[string]interface{}),
	}
	err := chain.ExecuteBefore(context.Background(), hookCtx)

	if err != nil {
		t.Errorf("ExecuteBefore returned unexpected error: %v", err)
	}
	if !mock.beforeCalled {
		t.Error("BeforeCall was not called")
	}
}

func TestHookChain_RegisterAfterCall(t *testing.T) {
	chain := NewHookChain()
	mock := &MockHook{}

	chain.RegisterAfterCall(mock.AfterCall)

	hookCtx := &HookContext{
		ToolName: "test",
		Metadata: make(map[string]interface{}),
	}
	chain.ExecuteAfter(context.Background(), hookCtx)

	if !mock.afterCalled {
		t.Error("AfterCall was not called")
	}
}

func TestHookChain_RegisterOnError(t *testing.T) {
	chain := NewHookChain()
	mock := &MockHook{}

	chain.RegisterOnError(mock.OnError)

	hookCtx := &HookContext{
		ToolName: "test",
		Metadata: make(map[string]interface{}),
	}
	chain.ExecuteError(context.Background(), hookCtx)

	if !mock.errorCalled {
		t.Error("OnError was not called")
	}
}

func TestHookChain_RegisterHook(t *testing.T) {
	chain := NewHookChain()
	mock := &MockHook{}

	chain.RegisterHook(mock)

	hookCtx := &HookContext{
		ToolName: "test",
		Metadata: make(map[string]interface{}),
	}

	// 测试 BeforeCall
	err := chain.ExecuteBefore(context.Background(), hookCtx)
	if err != nil {
		t.Errorf("ExecuteBefore returned unexpected error: %v", err)
	}
	if !mock.beforeCalled {
		t.Error("BeforeCall was not called via RegisterHook")
	}

	// 测试 AfterCall
	chain.ExecuteAfter(context.Background(), hookCtx)
	if !mock.afterCalled {
		t.Error("AfterCall was not called via RegisterHook")
	}

	// 测试 OnError
	chain.ExecuteError(context.Background(), hookCtx)
	if !mock.errorCalled {
		t.Error("OnError was not called via RegisterHook")
	}
}

func TestHookChain_ExecuteBefore(t *testing.T) {
	chain := NewHookChain()

	callCount := 0
	chain.RegisterBeforeCall(func(ctx context.Context, hookCtx *HookContext) error {
		callCount++
		return nil
	})

	chain.RegisterBeforeCall(func(ctx context.Context, hookCtx *HookContext) error {
		callCount++
		return nil
	})

	hookCtx := &HookContext{
		ToolName: "test",
		Metadata: make(map[string]interface{}),
	}
	err := chain.ExecuteBefore(context.Background(), hookCtx)

	if err != nil {
		t.Errorf("ExecuteBefore returned error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 hooks to be called, got %d", callCount)
	}
}

func TestHookChain_ExecuteBefore_Error(t *testing.T) {
	chain := NewHookChain()

	expectedErr := errors.New("test error")
	chain.RegisterBeforeCall(func(ctx context.Context, hookCtx *HookContext) error {
		return expectedErr
	})

	hookCtx := &HookContext{
		ToolName: "test",
		Metadata: make(map[string]interface{}),
	}
	err := chain.ExecuteBefore(context.Background(), hookCtx)

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestHookChain_ExecuteAfter(t *testing.T) {
	chain := NewHookChain()

	callCount := 0
	chain.RegisterAfterCall(func(ctx context.Context, hookCtx *HookContext) {
		callCount++
	})

	chain.RegisterAfterCall(func(ctx context.Context, hookCtx *HookContext) {
		callCount++
	})

	hookCtx := &HookContext{
		ToolName: "test",
		Metadata: make(map[string]interface{}),
	}
	chain.ExecuteAfter(context.Background(), hookCtx)

	if callCount != 2 {
		t.Errorf("Expected 2 hooks to be called, got %d", callCount)
	}
}

func TestHookChain_ExecuteError(t *testing.T) {
	chain := NewHookChain()

	callCount := 0
	chain.RegisterOnError(func(ctx context.Context, hookCtx *HookContext) {
		callCount++
	})

	chain.RegisterOnError(func(ctx context.Context, hookCtx *HookContext) {
		callCount++
	})

	hookCtx := &HookContext{
		ToolName: "test",
		Metadata: make(map[string]interface{}),
	}
	chain.ExecuteError(context.Background(), hookCtx)

	if callCount != 2 {
		t.Errorf("Expected 2 hooks to be called, got %d", callCount)
	}
}

func TestHookContext(t *testing.T) {
	ctx := &HookContext{
		ToolName: "test_tool",
		Metadata: make(map[string]interface{}),
	}

	if ctx.ToolName != "test_tool" {
		t.Errorf("Expected ToolName 'test_tool', got '%s'", ctx.ToolName)
	}

	if ctx.Metadata == nil {
		t.Error("Metadata map should not be nil")
	}

	// 测试存储和检索数据
	ctx.Metadata["key"] = "value"
	if val, ok := ctx.Metadata["key"].(string); !ok || val != "value" {
		t.Error("Failed to store and retrieve metadata value")
	}

	// 测试存储时间
	startTime := time.Now()
	ctx.Metadata["start_time"] = startTime

	retrievedTime, ok := ctx.Metadata["start_time"].(time.Time)
	if !ok || !retrievedTime.Equal(startTime) {
		t.Error("Failed to store and retrieve time value")
	}
}

func TestHookChain_Concurrent(t *testing.T) {
	chain := NewHookChain()

	// 注册多个 hooks
	for i := 0; i < 10; i++ {
		chain.RegisterBeforeCall(func(ctx context.Context, hookCtx *HookContext) error {
			return nil
		})
	}

	// 并发测试
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			hookCtx := &HookContext{
				ToolName: "test",
				Metadata: make(map[string]interface{}),
			}
			chain.ExecuteBefore(context.Background(), hookCtx)
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}
