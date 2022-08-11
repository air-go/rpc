package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsMeta struct {
	Method      string
	Start       time.Time
	End         time.Time
	Status      int
	ErrorNumber int
}

// Filter if hit filter return false, don't incr metrics's statistics
type Filter func(*MetricsMeta) bool

type Metrics interface {
	Register(prometheus.Registerer)
	WithLabelValues(*MetricsMeta)
	WithPanicValues()
}
