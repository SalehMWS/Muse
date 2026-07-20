package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type AI struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	tokens   *prometheus.CounterVec
}

func newAI() *AI {
	return &AI{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "ai",
			Name:      "requests_total",
			Help:      "Completion requests sent to the configured LLM provider.",
		}, []string{"provider", "model", "outcome"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "ai",
			Name:      "request_duration_seconds",
			Help:      "LLM provider request latency in seconds.",
			Buckets:   []float64{0.25, 0.5, 1, 2.5, 5, 10, 20, 30, 60},
		}, []string{"provider", "model"}),
		tokens: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "ai",
			Name:      "tokens_total",
			Help:      "Tokens reported by the LLM provider, by direction.",
		}, []string{"provider", "model", "kind"}),
	}
}

func (a *AI) collectors() []prometheus.Collector {
	return []prometheus.Collector{a.requests, a.duration, a.tokens}
}

func (a *AI) Request(provider, model, outcome string, elapsed time.Duration) {
	if a == nil {
		return
	}
	a.requests.WithLabelValues(provider, model, outcome).Inc()
	a.duration.WithLabelValues(provider, model).Observe(elapsed.Seconds())
}

func (a *AI) Tokens(provider, model string, prompt, completion int) {
	if a == nil {
		return
	}
	if prompt > 0 {
		a.tokens.WithLabelValues(provider, model, "prompt").Add(float64(prompt))
	}
	if completion > 0 {
		a.tokens.WithLabelValues(provider, model, "completion").Add(float64(completion))
	}
}
