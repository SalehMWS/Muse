package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTP struct {
	requests     *prometheus.CounterVec
	duration     *prometheus.HistogramVec
	responseSize *prometheus.HistogramVec
	inFlight     prometheus.Gauge
}

func newHTTP() *HTTP {
	return &HTTP{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total HTTP requests by method, route and status code.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency in seconds.",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.3, 0.5, 1, 2.5, 5, 10},
		}, []string{"method", "route"}),
		responseSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "http",
			Name:      "response_size_bytes",
			Help:      "HTTP response body size in bytes.",
			Buckets:   prometheus.ExponentialBuckets(64, 4, 8),
		}, []string{"method", "route"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "HTTP requests currently being served.",
		}),
	}
}

func (h *HTTP) collectors() []prometheus.Collector {
	return []prometheus.Collector{h.requests, h.duration, h.responseSize, h.inFlight}
}

func (h *HTTP) RequestStarted() {
	if h == nil {
		return
	}
	h.inFlight.Inc()
}

func (h *HTTP) RequestFinished(method, route string, status int, size int, elapsed time.Duration) {
	if h == nil {
		return
	}
	h.inFlight.Dec()
	h.requests.WithLabelValues(method, route, strconv.Itoa(status)).Inc()
	h.duration.WithLabelValues(method, route).Observe(elapsed.Seconds())
	h.responseSize.WithLabelValues(method, route).Observe(float64(size))
}
