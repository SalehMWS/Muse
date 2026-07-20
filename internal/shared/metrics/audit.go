package metrics

import "github.com/prometheus/client_golang/prometheus"

type Audit struct {
	events        *prometheus.CounterVec
	writeFailures *prometheus.CounterVec
}

func newAudit() *Audit {
	return &Audit{
		events: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "audit",
			Name:      "events_total",
			Help:      "Audit events recorded by action and result.",
		}, []string{"action", "result"}),
		writeFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "audit",
			Name:      "write_failures_total",
			Help:      "Audit events that could not be persisted, by action.",
		}, []string{"action"}),
	}
}

func (a *Audit) collectors() []prometheus.Collector {
	return []prometheus.Collector{a.events, a.writeFailures}
}

func (a *Audit) Recorded(action, result string) {
	if a == nil {
		return
	}
	a.events.WithLabelValues(action, result).Inc()
}

func (a *Audit) WriteFailed(action string) {
	if a == nil {
		return
	}
	a.writeFailures.WithLabelValues(action).Inc()
}
