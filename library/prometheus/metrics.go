package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics interface {
	Register(prometheus.Registerer)
}
