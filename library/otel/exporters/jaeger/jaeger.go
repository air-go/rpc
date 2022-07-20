package jaeger

import (
	"time"

	propagators "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/contrib/samplers/jaegerremote"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/air-go/rpc/library/app"
)

type JaegerConfig struct {
	Host                    string  `json:"host"`
	Port                    string  `json:"port"`
	SamplingServerURL       string  `json:"samplingServerURL"`
	SamplingRefreshInterval int     `json:"samplingRefreshInterval"`
	SamplerProportion       float64 `json:"samplerProportion"`
}

type Jaeger struct {
	Exporter    *jaeger.Exporter
	Sampler     *jaegerremote.Sampler
	Propagation propagators.Jaeger
}

func NewJaeger( config *JaegerConfig) (*Jaeger, error) {
	exporter, err := jaeger.New(
		// jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("")),
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(config.Host),
			jaeger.WithAgentPort(config.Port),
		),
	)
	if err != nil {
		return nil, err
	}

	return &Jaeger{
		Exporter: exporter,
		Sampler: jaegerremote.New(
			app.Name(),
			jaegerremote.WithSamplingServerURL(config.SamplingServerURL),
			jaegerremote.WithSamplingRefreshInterval(time.Duration(config.SamplingRefreshInterval)*time.Second),
			jaegerremote.WithInitialSampler(trace.TraceIDRatioBased(config.SamplerProportion)),
		),
		Propagation: propagators.Jaeger{},
	}, nil
}
