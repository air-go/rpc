package hook

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	lp "github.com/air-go/rpc/library/prometheus"
)

const (
	clientMetricNamespace = "mcp_client"
	clientMetricSubsystem = "tool"
)

// PrometheusClientHook is the client-side Prometheus metrics collection hook
type PrometheusClientHook struct {
	once       sync.Once
	registerer prometheus.Registerer
	callsTotal *prometheus.CounterVec
	duration   *prometheus.HistogramVec
}

// NewPrometheusClientHook creates a new Prometheus client hook
func NewPrometheusClientHook() *PrometheusClientHook {
	h := &PrometheusClientHook{
		registerer: lp.GetRegistry(),
		callsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: clientMetricNamespace,
				Subsystem: clientMetricSubsystem,
				Name:      "calls_total",
				Help:      "Total number of client tool calls",
			},
			[]string{"server_name", "tool_name", "status"},
		),
		duration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: clientMetricNamespace,
				Subsystem: clientMetricSubsystem,
				Name:      "duration_seconds",
				Help:      "Client tool call duration in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"server_name", "tool_name"},
		),
	}

	h.once.Do(func() {
		h.registerer.MustRegister(
			h.callsTotal,
			h.duration,
		)
	})

	return h
}

// BeforeCall records the start time
func (h *PrometheusClientHook) BeforeCall(ctx context.Context, hookCtx *ClientHookContext) error {
	hookCtx.Metadata["start_time"] = time.Now()
	return nil
}

// AfterCall records success metrics
func (h *PrometheusClientHook) AfterCall(ctx context.Context, hookCtx *ClientHookContext) {
	startTime, ok := hookCtx.Metadata["start_time"].(time.Time)
	if ok {
		duration := time.Since(startTime).Seconds()
		h.duration.WithLabelValues(hookCtx.ServerName, hookCtx.ToolName).Observe(duration)
	}
	h.callsTotal.WithLabelValues(hookCtx.ServerName, hookCtx.ToolName, "success").Inc()
}

// OnError records failure metrics
func (h *PrometheusClientHook) OnError(ctx context.Context, hookCtx *ClientHookContext) {
	startTime, ok := hookCtx.Metadata["start_time"].(time.Time)
	if ok {
		duration := time.Since(startTime).Seconds()
		h.duration.WithLabelValues(hookCtx.ServerName, hookCtx.ToolName).Observe(duration)
	}
	h.callsTotal.WithLabelValues(hookCtx.ServerName, hookCtx.ToolName, "error").Inc()
}
