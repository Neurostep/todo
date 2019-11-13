package logging

import (
	"context"

	"github.com/go-kit/kit/log"
	"go.opencensus.io/trace"
)

func FromContext(ctx context.Context, defaultLogger log.Logger) log.Logger {
	span := trace.FromContext(ctx)
	if span == nil {
		return defaultLogger
	}
	// https://cloud.google.com/logging/docs/agent/configuration#special-fields
	return log.With(defaultLogger, "logging.googleapis.com/spanId", span.SpanContext().TraceID)
}
