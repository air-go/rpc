package prometheus

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	registry     *prometheus.Registry
	registryOnce sync.Once
)

// GetRegistry 获取全局 Registry，确保只初始化一次
func GetRegistry() *prometheus.Registry {
	registryOnce.Do(func() {
		registry = prometheus.NewRegistry()

		// 注册进程级通用指标
		registry.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
		)
	})

	return registry
}
