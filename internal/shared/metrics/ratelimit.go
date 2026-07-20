package metrics

import "github.com/prometheus/client_golang/prometheus"

type RateLimit struct {
	decisions *prometheus.CounterVec
	errors    *prometheus.CounterVec
}

func newRateLimit() *RateLimit {
	return &RateLimit{
		decisions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "rate_limit",
			Name:      "decisions_total",
			Help:      "Rate limit decisions by scope and outcome.",
		}, []string{"scope", "outcome"}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "rate_limit",
			Name:      "errors_total",
			Help:      "Rate limiter backend failures by scope.",
		}, []string{"scope"}),
	}
}

func (r *RateLimit) collectors() []prometheus.Collector {
	return []prometheus.Collector{r.decisions, r.errors}
}

func (r *RateLimit) Allowed(scope string) {
	if r == nil {
		return
	}
	r.decisions.WithLabelValues(scope, OutcomeAllowed).Inc()
}

func (r *RateLimit) Limited(scope string) {
	if r == nil {
		return
	}
	r.decisions.WithLabelValues(scope, OutcomeLimited).Inc()
}

func (r *RateLimit) Failed(scope string) {
	if r == nil {
		return
	}
	r.errors.WithLabelValues(scope).Inc()
}
