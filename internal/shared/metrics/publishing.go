package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Publishing struct {
	publications *prometheus.CounterVec
	duration     *prometheus.HistogramVec
	apiCalls     *prometheus.CounterVec
	apiDuration  *prometheus.HistogramVec
}

func newPublishing() *Publishing {
	return &Publishing{
		publications: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "publishing",
			Name:      "publications_total",
			Help:      "Publish attempts by media type and outcome.",
		}, []string{"media_type", "outcome"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "publishing",
			Name:      "publish_duration_seconds",
			Help:      "End-to-end duration of a publish, container creation included.",
			Buckets:   []float64{0.5, 1, 2.5, 5, 10, 30, 60, 120},
		}, []string{"media_type"}),
		apiCalls: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "instagram",
			Name:      "api_requests_total",
			Help:      "Requests to the Instagram Graph API by operation and outcome.",
		}, []string{"operation", "outcome"}),
		apiDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "instagram",
			Name:      "api_request_duration_seconds",
			Help:      "Instagram Graph API request latency in seconds.",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		}, []string{"operation"}),
	}
}

func (p *Publishing) collectors() []prometheus.Collector {
	return []prometheus.Collector{p.publications, p.duration, p.apiCalls, p.apiDuration}
}

func (p *Publishing) Published(mediaType, outcome string, elapsed time.Duration) {
	if p == nil {
		return
	}
	if mediaType == "" {
		mediaType = "unknown"
	}
	p.publications.WithLabelValues(mediaType, outcome).Inc()
	p.duration.WithLabelValues(mediaType).Observe(elapsed.Seconds())
}

func (p *Publishing) APICall(operation, outcome string, elapsed time.Duration) {
	if p == nil {
		return
	}
	p.apiCalls.WithLabelValues(operation, outcome).Inc()
	p.apiDuration.WithLabelValues(operation).Observe(elapsed.Seconds())
}
