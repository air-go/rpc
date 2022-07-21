package otel

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
)

func ExtractHTTPBaggage(ctx context.Context, header http.Header) context.Context {
	b, err := baggage.Parse(header.Get(Baggage))
	if err != nil {
		return ctx
	}
	ctx = baggage.ContextWithBaggage(ctx, b)
	if header == nil {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}

func InjectHTTPBaggage(ctx context.Context, header http.Header) {
	if header == nil {
		return
	}
	header.Set(Baggage, baggage.FromContext(ctx).String())
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
}

func ExtractGRPCBaggage(ctx context.Context) context.Context {
	// TODO
	return ctx
}

func InjectGRPCBaggage() {
	// TODO
}
