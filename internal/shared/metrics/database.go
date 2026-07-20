package metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type poolCollector struct {
	pool *pgxpool.Pool

	acquired        *prometheus.Desc
	idle            *prometheus.Desc
	total           *prometheus.Desc
	max             *prometheus.Desc
	constructing    *prometheus.Desc
	acquireCount    *prometheus.Desc
	acquireDuration *prometheus.Desc
	canceledAcquire *prometheus.Desc
	emptyAcquire    *prometheus.Desc
}

func NewPoolCollector(pool *pgxpool.Pool) prometheus.Collector {
	const subsystem = "postgres"
	desc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(prometheus.BuildFQName(Namespace, subsystem, name), help, nil, nil)
	}
	return &poolCollector{
		pool:            pool,
		acquired:        desc("connections_acquired", "Connections currently checked out of the pool."),
		idle:            desc("connections_idle", "Idle connections in the pool."),
		total:           desc("connections_total", "Total connections currently held by the pool."),
		max:             desc("connections_max", "Maximum connections the pool may open."),
		constructing:    desc("connections_constructing", "Connections currently being established."),
		acquireCount:    desc("acquires_total", "Cumulative successful connection acquisitions."),
		acquireDuration: desc("acquire_duration_seconds_total", "Cumulative time spent waiting to acquire a connection."),
		canceledAcquire: desc("acquires_canceled_total", "Acquisitions canceled by context."),
		emptyAcquire:    desc("acquires_empty_total", "Acquisitions that had to wait for a new connection."),
	}
}

func (c *poolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.acquired
	ch <- c.idle
	ch <- c.total
	ch <- c.max
	ch <- c.constructing
	ch <- c.acquireCount
	ch <- c.acquireDuration
	ch <- c.canceledAcquire
	ch <- c.emptyAcquire
}

func (c *poolCollector) Collect(ch chan<- prometheus.Metric) {
	if c.pool == nil {
		return
	}
	stat := c.pool.Stat()

	gauge := func(d *prometheus.Desc, v float64) {
		ch <- prometheus.MustNewConstMetric(d, prometheus.GaugeValue, v)
	}
	counter := func(d *prometheus.Desc, v float64) {
		ch <- prometheus.MustNewConstMetric(d, prometheus.CounterValue, v)
	}

	gauge(c.acquired, float64(stat.AcquiredConns()))
	gauge(c.idle, float64(stat.IdleConns()))
	gauge(c.total, float64(stat.TotalConns()))
	gauge(c.max, float64(stat.MaxConns()))
	gauge(c.constructing, float64(stat.ConstructingConns()))
	counter(c.acquireCount, float64(stat.AcquireCount()))
	counter(c.acquireDuration, stat.AcquireDuration().Seconds())
	counter(c.canceledAcquire, float64(stat.CanceledAcquireCount()))
	counter(c.emptyAcquire, float64(stat.EmptyAcquireCount()))
}
