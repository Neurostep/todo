package metrics

import (
	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/go-kit/kit/log"
	"go.opencensus.io/stats/view"
)

func SetupMonitoring(env string, cfg Config, logger log.Logger) (*prometheus.Exporter, error) {
	pe, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		logger.Log("event", "prometheus_exporter_failed", "error", err)
		return nil, err
	}

	view.RegisterExporter(pe)
	return pe, nil
}
