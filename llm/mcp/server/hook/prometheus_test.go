package hook

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

var (
	prometheusServerHookOnce sync.Once
	prometheusServerHook     *PrometheusServerHook
)

func getPrometheusServerHook() *PrometheusServerHook {
	prometheusServerHookOnce.Do(func() {
		prometheusServerHook = NewPrometheusServerHook()
	})
	return prometheusServerHook
}

func TestNewPrometheusServerHook(t *testing.T) {
	t.Skip("Skipping to avoid duplicate metrics registration")
	// _ = NewPrometheusServerHook()
}

func TestPrometheusServerHook_BeforeCall(t *testing.T) {
	hook := getPrometheusServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Metadata: make(map[string]interface{}),
	}

	err := hook.BeforeCall(context.Background(), hookCtx)
	if err != nil {
		t.Errorf("BeforeCall returned unexpected error: %v", err)
	}

	// 验证 start_time 已存储
	startTime, ok := hookCtx.Metadata["start_time"].(time.Time)
	if !ok || startTime.IsZero() {
		t.Error("Expected start_time to be stored in metadata")
	}
}

func TestPrometheusServerHook_AfterCall(t *testing.T) {
	hook := getPrometheusServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Metadata: make(map[string]interface{}),
	}

	// BeforeCall 记录开始时间
	_ = hook.BeforeCall(context.Background(), hookCtx)

	// 模拟一些处理时间
	time.Sleep(10 * time.Millisecond)

	// AfterCall 应该记录指标
	hook.AfterCall(context.Background(), hookCtx)

	// 验证指标被记录（这里我们只是验证不 panic）
	if hook.callsTotal == nil || hook.duration == nil {
		t.Error("Expected metrics to be initialized")
	}
}

func TestPrometheusServerHook_AfterCall_WithoutStartTime(t *testing.T) {
	hook := getPrometheusServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Metadata: make(map[string]interface{}),
	}

	// 没有调用 BeforeCall，直接调用 AfterCall 应该不 panic
	hook.AfterCall(context.Background(), hookCtx)
}

func TestPrometheusServerHook_OnError(t *testing.T) {
	hook := getPrometheusServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Error:    errors.New("test error"),
		Metadata: make(map[string]interface{}),
	}

	// BeforeCall 记录开始时间
	_ = hook.BeforeCall(context.Background(), hookCtx)

	// 模拟一些处理时间
	time.Sleep(10 * time.Millisecond)

	// OnError 应该记录错误指标
	hook.OnError(context.Background(), hookCtx)
}

func TestPrometheusServerHook_OnError_WithoutStartTime(t *testing.T) {
	hook := getPrometheusServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Error:    errors.New("test error"),
		Metadata: make(map[string]interface{}),
	}

	// 没有调用 BeforeCall，直接调用 OnError 应该不 panic
	hook.OnError(context.Background(), hookCtx)
}

func TestPrometheusServerHook_FullLifecycle(t *testing.T) {
	hook := getPrometheusServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Metadata: make(map[string]interface{}),
	}

	// 模拟成功生命周期
	_ = hook.BeforeCall(context.Background(), hookCtx)
	time.Sleep(5 * time.Millisecond)
	hook.AfterCall(context.Background(), hookCtx)
}

func TestPrometheusServerHook_ErrorLifecycle(t *testing.T) {
	hook := getPrometheusServerHook()

	hookCtx := &HookContext{
		ToolName: "test_tool",
		Metadata: make(map[string]interface{}),
	}

	// 模拟错误生命周期
	_ = hook.BeforeCall(context.Background(), hookCtx)
	time.Sleep(5 * time.Millisecond)
	hookCtx.Error = errors.New("execution error")
	hook.OnError(context.Background(), hookCtx)
}

func TestPrometheusServerHook_MetricsLabels(t *testing.T) {
	hook := getPrometheusServerHook()

	// 测试不同的工具名称
	toolNames := []string{"tool1", "tool2", "tool3"}

	for _, toolName := range toolNames {
		hookCtx := &HookContext{
			ToolName: toolName,
			Metadata: make(map[string]interface{}),
		}

		_ = hook.BeforeCall(context.Background(), hookCtx)
		time.Sleep(1 * time.Millisecond)
		hook.AfterCall(context.Background(), hookCtx)
	}

	// 测试错误指标
	hookCtx := &HookContext{
		ToolName: "tool1",
		Error:    errors.New("error"),
		Metadata: make(map[string]interface{}),
	}
	_ = hook.BeforeCall(context.Background(), hookCtx)
	hook.OnError(context.Background(), hookCtx)
}

func TestPrometheusServerHook_ServerToolHookInterface(t *testing.T) {
	// 确保 PrometheusServerHook 实现了 ServerToolHook 接口
	var _ ServerToolHook = (*PrometheusServerHook)(nil)
}
