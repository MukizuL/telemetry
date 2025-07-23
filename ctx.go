package telemetry

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	otelTrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type ContextTelemetry interface {
	GetSpan(ctx context.Context) otelTrace.Span
	NewSpan(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, otelTrace.Span)
	Traceparent(ctx context.Context) string
	WithTraceparent(ctx context.Context) context.Context

	Log(ctx context.Context) *zap.Logger

	Wrap(ctx context.Context, spanName, traceparent string) context.Context
	Clone(ctx context.Context) context.Context
}

type contextTelemetry struct {
	op *otelProvider
}

func newContextTelemetry(op *otelProvider) *contextTelemetry {
	return &contextTelemetry{op: op}
}

// WithTraceparent For dapr client
func (p *contextTelemetry) WithTraceparent(ctx context.Context) context.Context {
	md := metadata.Pairs("traceparent", p.Traceparent(ctx))
	return metadata.NewOutgoingContext(ctx, md)
}

func (p *contextTelemetry) GetSpan(ctx context.Context) otelTrace.Span {
	return otelTrace.SpanFromContext(ctx)
}

func (p *contextTelemetry) NewSpan(ctx context.Context, spanName string, attributes ...attribute.KeyValue) (context.Context, otelTrace.Span) {
	return newSpan(ctx, spanName, p.op.config.GetServiceName(), attributes...)
}

func (p *contextTelemetry) Traceparent(ctx context.Context) string {
	return traceparent(ctx)
}

func (p *contextTelemetry) Wrap(ctx context.Context, spanName, traceparent string) context.Context {
	propagator := otel.GetTextMapPropagator()
	if propagator == nil {
		propagator = propagation.TraceContext{}
	}

	c := propagator.Extract(ctx, propagation.MapCarrier{
		"traceparent": traceparent,
	})

	tracer := getTracer(p.op.config.GetServiceName())
	cont, _ := tracer.Start(c, spanName)

	return cont
}

func (p *contextTelemetry) Clone(ctx context.Context) context.Context {
	newCtx := context.Background()

	propagator := otel.GetTextMapPropagator()
	if propagator == nil {
		propagator = propagation.TraceContext{}
	}

	mc := propagation.MapCarrier{}
	propagator.Inject(ctx, &mc)

	newCtx = propagator.Extract(newCtx, &mc)

	return newCtx
}
