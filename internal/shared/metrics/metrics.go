package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const Namespace = "novaflow"

const (
	OutcomeSuccess      = "success"
	OutcomeFailure      = "failure"
	OutcomeRetried      = "retried"
	OutcomeDeadLettered = "dead_lettered"
)

type Metrics struct {
	registry   *prometheus.Registry
	HTTP       *HTTP
	Worker     *Worker
	Scheduler  *Scheduler
	AI         *AI
	Publishing *Publishing
	Knowledge  *Knowledge
	Business   *Business
}

func New(service, version, environment string) *Metrics {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Name:      "build_info",
		Help:      "Build and runtime identity of the running service.",
	}, []string{"service", "version", "environment"})
	buildInfo.WithLabelValues(service, version, environment).Set(1)
	registry.MustRegister(buildInfo)

	m := &Metrics{
		registry:   registry,
		HTTP:       newHTTP(),
		Worker:     newWorker(),
		Scheduler:  newScheduler(),
		AI:         newAI(),
		Publishing: newPublishing(),
		Knowledge:  newKnowledge(),
		Business:   newBusiness(),
	}

	registry.MustRegister(m.HTTP.collectors()...)
	registry.MustRegister(m.Worker.collectors()...)
	registry.MustRegister(m.Scheduler.collectors()...)
	registry.MustRegister(m.AI.collectors()...)
	registry.MustRegister(m.Publishing.collectors()...)
	registry.MustRegister(m.Knowledge.collectors()...)
	registry.MustRegister(m.Business.collectors()...)

	return m
}

func (m *Metrics) Registry() *prometheus.Registry {
	if m == nil {
		return nil
	}
	return m.registry
}

func (m *Metrics) Register(c prometheus.Collector) error {
	if m == nil {
		return nil
	}
	return m.registry.Register(c)
}

func Outcome(err error) string {
	if err != nil {
		return OutcomeFailure
	}
	return OutcomeSuccess
}
