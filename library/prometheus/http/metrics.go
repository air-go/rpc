package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	lp "github.com/air-go/rpc/library/prometheus"
)

type httpMetrics struct {
	opts           *Options
	once           sync.Once
	successCounter *prometheus.CounterVec
	successSummary *prometheus.SummaryVec
	errorCounter   *prometheus.CounterVec
	errorSummary   *prometheus.SummaryVec
	panicCounter   prometheus.Counter
	working        prometheus.Gauge
}

var _ lp.Metrics = (*httpMetrics)(nil)

// opts ...OptionFunc)
func NewHTTPMetrics(opts ...OptionFunc) *httpMetrics {
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	labels := []string{}
	for _, l := range opt.labels {
		labels = append(labels, l.Name())
	}

	return &httpMetrics{
		opts: opt,
		successCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "http_server",
			Name:      "success_count",
			Help:      "http_server success_count",
		}, labels),
		successSummary: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: "http_server",
			Name:      "success_cost",
			Objectives: map[float64]float64{
				0.8:  0.05,
				0.9:  0.02,
				0.95: 0.01,
				0.99: 0.001,
			},
			Help: "http_server success_cost",
		}, labels),
		errorCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "http_server",
			Name:      "err_count",
			Help:      "http_server err_count",
		}, labels),
		errorSummary: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: "http_server",
			Name:      "err_cost",
			Help:      "http_server err_cost",
			Objectives: map[float64]float64{
				0.8:  0.05,
				0.9:  0.02,
				0.95: 0.01,
				0.99: 0.001,
			},
		}, labels),
		panicCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "http_server",
			Name:      "panic_count",
			Help:      "http_server panic_count",
		}),
		working: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "http_server",
			Subsystem: "working",
			Name:      "working_count",
			Help:      "http_server working_count",
		}),
	}
}

func (m *httpMetrics) Register(r prometheus.Registerer) {
	m.once.Do(func() {
		if r == nil {
			r = prometheus.DefaultRegisterer
		}
		r.MustRegister(
			m.successCounter,
			m.successSummary,
			m.errorCounter,
			m.errorSummary,
			m.panicCounter,
			m.working,
		)
	})
}

func (m *httpMetrics) withLabelValues(c *gin.Context, cost time.Duration) {
	values := []string{}
	for _, v := range m.opts.labels {
		values = append(values, v.GetValue(c))
	}

	// success
	if c.Writer.Status() == http.StatusOK && len(c.Errors) <= 0 {
		m.successCounter.WithLabelValues(values...).Add(1)
		m.successSummary.WithLabelValues(values...).Observe(cost.Seconds())
		return
	}

	// error
	m.errorCounter.WithLabelValues(values...).Add(1)
	m.errorSummary.WithLabelValues(values...).Observe(cost.Seconds())
}

func (m *httpMetrics) withPanicValues() {
	m.panicCounter.Inc()
}

func (m *httpMetrics) getFilters() []Filter {
	return m.opts.filters
}
