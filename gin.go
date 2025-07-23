package telemetry

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	otelTrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type GinContextTelemetry interface {
	GetSpan(ctx *gin.Context) otelTrace.Span
	NewSpan(ctx *gin.Context, name string, attributes ...attribute.KeyValue) otelTrace.Span
	Traceparent(ctx *gin.Context) string
	WithTraceparent(ctx *gin.Context) *gin.Context

	Log(ctx *gin.Context) *zap.Logger

	Wrap(ctx *gin.Context)

	Middleware() gin.HandlerFunc
	Clone(ctx *gin.Context) context.Context
}

type ginContextTelemetry struct {
	op *otelProvider
}

func newGinContextTelemetry(p *otelProvider) *ginContextTelemetry {
	return &ginContextTelemetry{
		op: p,
	}
}

func (p *ginContextTelemetry) Wrap(ctx *gin.Context) {
	propagator := otel.GetTextMapPropagator()
	if propagator == nil {
		propagator = propagation.TraceContext{}
	}
	tracer := getTracer(p.op.config.GetServiceName())
	c, span := tracer.Start(
		propagator.Extract(ctx.Request.Context(),
			propagation.HeaderCarrier(ctx.Request.Header)),
		ctx.FullPath(),
	)
	ctx.Request = ctx.Request.WithContext(c)
	ctx.Set(string(TraceSpanKey), span)
}

func (p *ginContextTelemetry) GetSpan(ctx *gin.Context) otelTrace.Span {
	span, exists := ctx.Get(string(TraceSpanKey))
	if !exists {
		return otelTrace.SpanFromContext(ctx.Request.Context())
	}

	if s, ok := span.(otelTrace.Span); ok {
		return s
	}

	return otelTrace.SpanFromContext(ctx.Request.Context())
}

func (p *ginContextTelemetry) NewSpan(ctx *gin.Context, name string, attributes ...attribute.KeyValue) otelTrace.Span {
	childCtx, span := newSpan(ctx.Request.Context(), name, p.op.config.GetServiceName(), attributes...)

	ctx.Request.WithContext(childCtx)

	return span
}

func (p *ginContextTelemetry) Traceparent(ctx *gin.Context) string {
	return traceparent(ctx.Request.Context())
}

func (p *ginContextTelemetry) WithTraceparent(ctx *gin.Context) *gin.Context {
	md := metadata.Pairs("traceparent", p.Traceparent(ctx))
	ctx.Request.WithContext(metadata.NewOutgoingContext(ctx.Request.Context(), md))

	return ctx
}

func (p *ginContextTelemetry) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p.Wrap(ctx)
		ctx.Next()
		p.GetSpan(ctx).End()
	}
}

func (p *ginContextTelemetry) Clone(ctx *gin.Context) context.Context {
	newCtx := context.Background()

	propagator := otel.GetTextMapPropagator()
	if propagator == nil {
		propagator = propagation.TraceContext{}
	}

	mc := propagation.MapCarrier{}
	propagator.Inject(ctx.Request.Context(), &mc)

	newCtx = propagator.Extract(newCtx, &mc)

	return newCtx
}
