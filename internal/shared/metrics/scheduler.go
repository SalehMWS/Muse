package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Scheduler struct {
	claimed   prometheus.Counter
	enqueued  *prometheus.CounterVec
	drift     prometheus.Histogram
	tick      prometheus.Histogram
	tickError prometheus.Counter
}

func newScheduler() *Scheduler {
	return &Scheduler{
		claimed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "scheduler",
			Name:      "schedules_claimed_total",
			Help:      "Due schedules claimed from the database.",
		}),
		enqueued: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "scheduler",
			Name:      "schedules_processed_total",
			Help:      "Claimed schedules by processing outcome.",
		}, []string{"outcome"}),
		drift: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "scheduler",
			Name:      "execution_drift_seconds",
			Help:      "Delay between a schedule's due time and the moment it was processed.",
			Buckets:   []float64{1, 5, 10, 30, 60, 300, 900, 3600},
		}),
		tick: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "scheduler",
			Name:      "tick_duration_seconds",
			Help:      "Duration of a single scheduler poll cycle.",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
		}),
		tickError: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "scheduler",
			Name:      "tick_errors_total",
			Help:      "Scheduler poll cycles that failed to claim due schedules.",
		}),
	}
}

func (s *Scheduler) collectors() []prometheus.Collector {
	return []prometheus.Collector{s.claimed, s.enqueued, s.drift, s.tick, s.tickError}
}

func (s *Scheduler) TickCompleted(claimed int, elapsed time.Duration) {
	if s == nil {
		return
	}
	s.tick.Observe(elapsed.Seconds())
	s.claimed.Add(float64(claimed))
}

func (s *Scheduler) TickFailed(elapsed time.Duration) {
	if s == nil {
		return
	}
	s.tick.Observe(elapsed.Seconds())
	s.tickError.Inc()
}

func (s *Scheduler) Processed(outcome string, drift time.Duration) {
	if s == nil {
		return
	}
	s.enqueued.WithLabelValues(outcome).Inc()
	if drift > 0 {
		s.drift.Observe(drift.Seconds())
	}
}
