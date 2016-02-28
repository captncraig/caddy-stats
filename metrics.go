package metrics

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "caddy"

func define(subsystem string) {
	if subsystem == "" {
		subsystem = "http"
	}
	requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: Subsystem,
		Name:      "request_count_total",
		Help:      "Counter of HTTP(S) requests made.",
	}, []string{"host"})

	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: Subsystem,
		Name:      "request_duration_seconds",
		Help:      "Histogram of the time (in seconds) each request took.",
	}, []string{"host"})

	responseSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: Subsystem,
		Name:      "response_size_bytes",
		Help:      "Size of the returns response in bytes.",
		Buckets:   []float64{0, 500, 1000, 2000, 3000, 4000, 5000, 10000, 20000, 30000, 50000, 100000},
	}, []string{"host"})

	responseStatus = prometheus.NewCounterVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: Subsystem,
		Name:      "status_code_count_total",
		Help:      "Counter of response status codes.",
	}, []string{"host", "status"})
}
