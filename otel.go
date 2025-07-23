package telemetry

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Provider interface {
	Gin() GinContextTelemetry
	Context() ContextTelemetry
	WithoutContext() WithoutContextTelemetry
}

type otelProvider struct {
	tp *trace.TracerProvider

	gt     *ginContextTelemetry
	ct     *contextTelemetry
	log    *zap.Logger
	wc     *withoutContextTelemetry
	config Config
}

func (p *otelProvider) Gin() GinContextTelemetry {
	return p.gt
}

func (p *otelProvider) Context() ContextTelemetry {
	return p.ct
}

func (p *otelProvider) WithoutContext() WithoutContextTelemetry {
	return p.wc
}

type OtelProviderOut struct {
	fx.Out

	TP  Provider
	Log *zap.Logger
	Opt fx.Option
}

func NewOtelProvider(cfg Config) (OtelProviderOut, error) {
	log, err := NewLogger(cfg)
	if err != nil {
		return OtelProviderOut{}, fmt.Errorf(
			"failed to create logger: %w",
			err,
		)
	}

	provider := &otelProvider{
		log:    log,
		config: cfg,
	}

	ctx := context.Background()
	var (
		exporter trace.SpanExporter
	)

	if cfg.GetOtelServerURL() != "" {
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(cfg.GetOtelServerURL()),
			otlptracehttp.WithInsecure(),
		)
	} else {
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	}

	if err != nil {
		return OtelProviderOut{}, fmt.Errorf("failed to create exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.GetServiceName()),
		),
	)
	if err != nil {
		return OtelProviderOut{}, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	provider.tp = tp
	provider.gt = newGinContextTelemetry(provider)
	provider.ct = newContextTelemetry(provider)
	provider.wc = newWithoutContext(provider)

	return OtelProviderOut{
		TP:  provider,
		Log: log,
	}, nil
}

type OtelContextKey string

const TraceSpanKey OtelContextKey = "traceSpan"
