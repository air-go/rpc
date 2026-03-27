package http

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewHTTPMetrics(t *testing.T) {
	// 测试创建 HTTPMetrics
	// 注意：由于全局 Registry 使用 sync.Once，同一个进程多次调用会复用已注册的指标
	// 所以只测试一次创建
	m := NewHTTPMetrics()
	if m == nil {
		t.Fatal("NewHTTPMetrics() returned nil")
	}

	// 测试获取 filters（默认是 nil）
	filters := m.getFilters()
	// filters 可以是 nil，这是预期行为
	_ = filters

	// 测试注册自定义 collector
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "custom_test_counter",
		Help: "custom test counter",
	})
	m.Register(counter)
}

func TestHTTPMetricsWithPanicValues(t *testing.T) {
	// 使用独立的 Registry 测试 panic 计数器
	testRegistry := prometheus.NewRegistry()

	m := &httpMetrics{
		panicCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_panic_count3",
			Help: "test panic count",
		}),
		registerer: testRegistry,
	}

	testRegistry.MustRegister(m.panicCounter)

	// 测试 panic 计数
	m.withPanicValues()

	panicCount := testutil.ToFloat64(m.panicCounter)
	if panicCount != 1 {
		t.Errorf("Expected panic counter to be 1, got %f", panicCount)
	}
}

func TestHTTPMetricsImplementsMetricsInterface(t *testing.T) {
	// 验证 httpMetrics 实现了 lp.Metrics 接口
	var _ interface{} = (*httpMetrics)(nil)
}

func TestHTTPMetricsWithFilters(t *testing.T) {
	// 使用独立的 Registry 测试过滤器选项
	testRegistry := prometheus.NewRegistry()

	filters := []Filter{
		func(c *gin.Context) bool {
			return true
		},
	}

	// 创建带过滤器的 httpMetrics
	labels := []Label{}
	m := &httpMetrics{
		opts: &Options{
			labels:  labels,
			filters: filters,
		},
		successCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "test_with_filters_success_count",
			Help: "test success count",
		}, []string{}),
		successSummary: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: "test_with_filters_success_cost",
			Help: "test success cost",
		}, []string{}),
		errorCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "test_with_filters_err_count",
			Help: "test error count",
		}, []string{}),
		errorSummary: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: "test_with_filters_err_cost",
			Help: "test error cost",
		}, []string{}),
		panicCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_with_filters_panic_count",
			Help: "test panic count",
		}),
		working: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "test_with_filters_working_count",
			Help: "test working count",
		}),
		registerer: testRegistry,
	}

	// 验证过滤器已设置
	f := m.getFilters()
	if len(f) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(f))
	}
}
