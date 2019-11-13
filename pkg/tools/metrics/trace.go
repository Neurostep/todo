package metrics

import (
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/go-kit/kit/log"
	"go.opencensus.io/trace"
)

type Config struct {
	TracingEnable bool
}

func SetupTracer(env string, cfg Config, logger log.Logger) (func(), error) {
	if !cfg.TracingEnable {
		logger.Log("event", "tracing not enabled")
		return func() {}, nil
	}

	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		OnError: func(err error) {
			logger.Log("event", "stackdriver_error", "error", err)
		},
		DefaultTraceAttributes: map[string]interface{}{
			"env": env,
		},
	})
	if err != nil {
		return func() {}, err
	}

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	return func() { trace.UnregisterExporter(exporter) }, nil
}
