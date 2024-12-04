package prometheus

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	lc "github.com/air-go/rpc/library/context"
)

type Metrics interface {
	Register(...prometheus.Collector)
}

var CustomCollector = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "custom_error",
		Name:      "error_count",
		Help:      "error",
	},
	[]string{"message", "logid", "traceid", "error"},
)

type options struct{}

type CustomCollectorOption func(o *options)

// ReportCustomError
// Registration is required before use.
// metrics.Register(CustomCollector)
func ReportCustomError(ctx context.Context, msg string, err error, opts ...CustomCollectorOption) {
	if msg == "" {
		return
	}

	option := &options{}
	for _, o := range opts {
		o(option)
	}

	e := ""
	if err != nil {
		e = err.Error()
	}

	values := []string{msg, lc.ValueLogID(ctx), lc.ValueTraceID(ctx), e}
	CustomCollector.WithLabelValues(values...).Inc()
}
