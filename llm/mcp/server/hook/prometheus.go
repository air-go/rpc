package hook

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricNamespace = "mcp_server"
	metricSubsystem = "tool"
)

// PrometheusServerHook is the server-side Prometheus metrics collection hook
type PrometheusServerHook struct {
	callsTotal *prometheus.CounterVec
	duration   *prometheus.HistogramVec
}

// NewPrometheusServerHook creates a new Prometheus server hook
func NewPrometheusServerHook() *PrometheusServerHook {
	return &PrometheusServerHook{
		callsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubsystem,
				Name:      "calls_total",
				Help:      "Total number of tool calls",
			},
			[]string{"tool_name", "status"},
		),
		duration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubsystem,
				Name:      "duration_seconds",
				Help:      "Tool execution duration in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"tool_name"},
		),
	}
}

// BeforeCall records the start time
func (h *PrometheusServerHook) BeforeCall(ctx context.Context, hookCtx *HookContext) error {
	hookCtx.Metadata["start_time"] = time.Now()
	return nil
}

// AfterCall records success metrics
func (h *PrometheusServerHook) AfterCall(ctx context.Context, hookCtx *HookContext) {
	startTime, ok := hookCtx.Metadata["start_time"].(time.Time)
	if ok {
		duration := time.Since(startTime).Seconds()
		h.duration.WithLabelValues(hookCtx.ToolName).Observe(duration)
	}
	h.callsTotal.WithLabelValues(hookCtx.ToolName, "success").Inc()
}

// OnError records failure metrics
func (h *PrometheusServerHook) OnError(ctx context.Context, hookCtx *HookContext) {
	startTime, ok := hookCtx.Metadata["start_time"].(time.Time)
	if ok {
		duration := time.Since(startTime).Seconds()
		h.duration.WithLabelValues(hookCtx.ToolName).Observe(duration)
	}
	h.callsTotal.WithLabelValues(hookCtx.ToolName, "error").Inc()
}
