package stdout

import (
	"os"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Stdout struct {
	Exporter *stdouttrace.Exporter
	Sampler  trace.Sampler
}

func NewStdout() (*Stdout, error) {
	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(os.Stdout),
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, err
	}

	return &Stdout{
		Exporter: exporter,
		Sampler:  trace.AlwaysSample(),
	}, nil
}
