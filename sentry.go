package telemetry

import (
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"
)

type zapWithSentry struct {
	zapcore.Core
}

func (c *zapWithSentry) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

func (c *zapWithSentry) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if entry.Level == zapcore.PanicLevel || entry.Level == zapcore.ErrorLevel || entry.Level == zapcore.FatalLevel {
		sentry.CaptureEvent(parseEntry(entry, fields))
	}

	return c.Core.Write(entry, fields)
}

func parseEntry(entry zapcore.Entry, fields []zapcore.Field) *sentry.Event {
	msg := entry.Message
	sentryLevel := sentry.LevelError
	if entry.Level == zapcore.FatalLevel {
		sentryLevel = sentry.LevelFatal
	}

	var exceptions []sentry.Exception
	for _, field := range fields {
		if field.Key == "error" {
			exceptions = append(exceptions, sentry.Exception{
				Type:       string(sentryLevel),
				Value:      msg,
				Stacktrace: sentry.ExtractStacktrace(field.Interface.(error)),
			})
		}
	}

	return &sentry.Event{
		Message:   msg,
		Level:     sentryLevel,
		Exception: exceptions,
	}
}
