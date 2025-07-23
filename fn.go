package telemetry

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelTrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const LogFieldsKey = "logFields"

func getTracer(name string) otelTrace.Tracer {
	return otel.Tracer(name)
}

func AddTelemetryField(ctx context.Context, name string, value interface{}) context.Context {
	var logFields []zap.Field
	fields := ctx.Value(LogFieldsKey)
	if fields != nil {
		logFields = fields.([]zap.Field)
	}

	logFields = append(logFields, zap.Any(name, value))

	ctx = context.WithValue(ctx, LogFieldsKey, logFields)

	return ctx
}

func logfn(ctx context.Context, log *zap.Logger) *zap.Logger {
	span := otelTrace.SpanFromContext(ctx)
	logFields := []zap.Field{
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()),
	}

	fields := ctx.Value(LogFieldsKey)
	if fields != nil {
		logFields = append(logFields, fields.([]zap.Field)...)
	}

	return log.With(logFields...)
}

func newSpan(ctx context.Context, spanName, serviceName string, attributes ...attribute.KeyValue) (context.Context, otelTrace.Span) {
	tracer := getTracer(serviceName)
	childCtx, childSpan := tracer.Start(ctx, spanName)

	for _, attr := range attributes {
		childSpan.SetAttributes(attr)
	}

	return childCtx, childSpan
}

func traceparent(ctx context.Context) string {
	span := otelTrace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if !sc.IsValid() {
		return ""
	}

	return fmt.Sprintf("00-%s-%s-%s",
		sc.TraceID().String(),
		sc.SpanID().String(),
		sc.TraceFlags().String(),
	)
}
