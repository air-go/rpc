package prometheus

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var defaultMiddleWareMetricsLables []string = []string{"method", "error_number", "http_status"}

type httpMetrics struct {
	once           sync.Once
	successCounter *prometheus.CounterVec
	successSummary *prometheus.SummaryVec
	errorCounter   *prometheus.CounterVec
	errorSummary   *prometheus.SummaryVec
	panicCounter   prometheus.Counter
	working        prometheus.Gauge
	filters        []Filter
}

var _ Metrics = (*httpMetrics)(nil)

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

func (m *httpMetrics) WithLabelValues(meta *MetricsMeta) {
	duration := float64(meta.End.Sub(meta.Start)) / float64(time.Millisecond)
	labelValues := []string{meta.Method, strconv.Itoa(meta.ErrorNumber), strconv.Itoa(meta.Status)}

	// handle filter
	for _, filter := range m.filters {
		if !filter(meta) {
			return
		}
	}

	// success
	if meta.Status == 200 && meta.ErrorNumber <= 0 {
		m.successCounter.WithLabelValues(labelValues...).Add(1)
		m.successSummary.WithLabelValues(labelValues...).Observe(duration)
		return
	}

	// error
	m.errorCounter.WithLabelValues(labelValues...).Add(1)
	m.errorSummary.WithLabelValues(labelValues...).Observe(duration)
}

func (m *httpMetrics) WithPanicValues() {
	m.panicCounter.Inc()
}

func NewHTTPMetrics(filters ...Filter) Metrics {
	return &httpMetrics{
		filters: filters,
		successCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "http_server",
			Name:      "success_count",
			Help:      "http_server success_count",
		}, defaultMiddleWareMetricsLables),
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
		}, defaultMiddleWareMetricsLables),
		errorCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "http_server",
			Name:      "err_count",
			Help:      "http_server err_count",
		}, defaultMiddleWareMetricsLables),
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
		}, defaultMiddleWareMetricsLables),
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

// HTTPMetricsMiddleware must register after PanicMiddleware and LoggerMiddleware
func HTTPMetricsMiddleware() gin.HandlerFunc {
	metrics := NewHTTPMetrics()
	metrics.Register(nil)
	return func(c *gin.Context) {
		meta := &MetricsMeta{
			Method: c.FullPath(),
			Start:  time.Now(),
		}

		defer func() {
			if err := recover(); err != nil {
				metrics.WithPanicValues()
				// keep panic, used by PanicMiddleware
				panic(err)
			}
		}()

		c.Next()

		meta.End = time.Now()
		meta.Status = c.Writer.Status()
		meta.ErrorNumber = len(c.Errors)

		metrics.WithLabelValues(meta)
	}
}
