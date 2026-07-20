package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Worker struct {
	jobs       *prometheus.CounterVec
	duration   *prometheus.HistogramVec
	inFlight   prometheus.Gauge
	panics     prometheus.Counter
	queueDepth *prometheus.GaugeVec
}

func newWorker() *Worker {
	return &Worker{
		jobs: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "worker",
			Name:      "jobs_total",
			Help:      "Total jobs handled by the worker pool, by job type and outcome.",
		}, []string{"type", "outcome"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "worker",
			Name:      "job_duration_seconds",
			Help:      "Job execution time in seconds.",
			Buckets:   []float64{0.05, 0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 120},
		}, []string{"type"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: "worker",
			Name:      "jobs_in_flight",
			Help:      "Jobs currently being executed.",
		}),
		panics: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "worker",
			Name:      "panics_recovered_total",
			Help:      "Panics recovered inside job handlers.",
		}),
		queueDepth: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: "worker",
			Name:      "queue_depth",
			Help:      "Number of entries in a job stream.",
		}, []string{"queue"}),
	}
}

func (w *Worker) collectors() []prometheus.Collector {
	return []prometheus.Collector{w.jobs, w.duration, w.inFlight, w.panics, w.queueDepth}
}

func (w *Worker) JobStarted() {
	if w == nil {
		return
	}
	w.inFlight.Inc()
}

func (w *Worker) JobFinished(jobType, outcome string, elapsed time.Duration) {
	if w == nil {
		return
	}
	w.inFlight.Dec()
	w.jobs.WithLabelValues(jobType, outcome).Inc()
	w.duration.WithLabelValues(jobType).Observe(elapsed.Seconds())
}

func (w *Worker) JobOutcome(jobType, outcome string) {
	if w == nil {
		return
	}
	w.jobs.WithLabelValues(jobType, outcome).Inc()
}

func (w *Worker) PanicRecovered() {
	if w == nil {
		return
	}
	w.panics.Inc()
}

func (w *Worker) SetQueueDepth(queue string, depth float64) {
	if w == nil {
		return
	}
	w.queueDepth.WithLabelValues(queue).Set(depth)
}
