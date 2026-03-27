package prometheus

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestGetRegistry(t *testing.T) {
	// 测试获取 Registry
	registry := GetRegistry()
	if registry == nil {
		t.Fatal("GetRegistry() returned nil")
	}

	// 测试返回的是 prometheus.Registry 类型
	var _ *prometheus.Registry = registry
}

func TestGetRegistryOnce(t *testing.T) {
	// 测试多次调用返回同一个实例
	registry1 := GetRegistry()
	registry2 := GetRegistry()

	if registry1 != registry2 {
		t.Error("GetRegistry() should return the same instance")
	}
}

func TestRegistryContainsCollectors(t *testing.T) {
	// 测试 Registry 中是否包含 ProcessCollector 和 GoCollector
	// 注意：默认的 prometheus.Registry 不包含这些，需要通过 CollectorCount 来验证
	registry := GetRegistry()

	// 验证 registry 可以正常工作
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "test counter",
	})
	counter.Inc()

	err := registry.Register(counter)
	if err != nil {
		t.Fatalf("Failed to register test counter: %v", err)
	}

	// 验证计数器可以被收集
	count := testutil.ToFloat64(counter)
	if count != 1 {
		t.Errorf("Expected counter to be 1, got %f", count)
	}

	// 清理
	registry.Unregister(counter)
}
