package telemetry

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (p *contextTelemetry) Log(ctx context.Context) *zap.Logger {
	return logfn(ctx, p.op.log)
}

func (p *ginContextTelemetry) Log(ctx *gin.Context) *zap.Logger {
	return logfn(ctx.Request.Context(), p.op.log)
}

func (w *withoutContextTelemetry) Log() *zap.Logger {
	return w.op.log
}

func NewLogger(config Config) (*zap.Logger, error) {
	zcore, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("error init zap logger: %w", err)
	}

	logger := zap.New(&zapWithSentry{zcore.Core()})
	if config.GetEnvironment() != "production" {
		return zap.NewDevelopment()
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:              config.GetSentryDsn(),
		Environment:      config.GetEnvironment(),
		TracesSampleRate: 0.2,
		Debug:            false,
		Release:          config.GetCommitTag(),
		ServerName:       config.GetServiceName(),
	})
	if err != nil {
		return nil, fmt.Errorf("error init sentry sdk: %w", err)
	}

	return logger.With(zap.String("service", config.GetServiceName()), zap.String("commitTag", config.GetCommitTag())), nil
}
