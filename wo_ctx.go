package telemetry

import "go.uber.org/zap"

type WithoutContextTelemetry interface {
	Log() *zap.Logger
}

type withoutContextTelemetry struct {
	op *otelProvider
}

func newWithoutContext(op *otelProvider) *withoutContextTelemetry {
	return &withoutContextTelemetry{op: op}
}
