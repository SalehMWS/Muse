package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Knowledge struct {
	ingests   *prometheus.CounterVec
	chunks    prometheus.Counter
	embedding *prometheus.HistogramVec
	search    *prometheus.HistogramVec
	queries   *prometheus.CounterVec
	topK      prometheus.Histogram
}

func newKnowledge() *Knowledge {
	return &Knowledge{
		ingests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "knowledge",
			Name:      "documents_ingested_total",
			Help:      "Documents submitted for indexing, by outcome.",
		}, []string{"outcome"}),
		chunks: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "knowledge",
			Name:      "chunks_indexed_total",
			Help:      "Chunks embedded and written to the vector store.",
		}),
		embedding: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "knowledge",
			Name:      "embedding_duration_seconds",
			Help:      "Time spent producing embeddings, by embedder.",
			Buckets:   []float64{0.005, 0.05, 0.1, 0.5, 1, 2.5, 5, 10, 30},
		}, []string{"embedder"}),
		search: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "knowledge",
			Name:      "search_duration_seconds",
			Help:      "Vector store search latency, by store.",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
		}, []string{"store"}),
		queries: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "knowledge",
			Name:      "queries_total",
			Help:      "Retrieval queries by outcome.",
		}, []string{"outcome"}),
		topK: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: "knowledge",
			Name:      "query_results",
			Help:      "Number of chunks returned per retrieval query.",
			Buckets:   []float64{0, 1, 2, 3, 5, 8, 10, 20},
		}),
	}
}

func (k *Knowledge) collectors() []prometheus.Collector {
	return []prometheus.Collector{k.ingests, k.chunks, k.embedding, k.search, k.queries, k.topK}
}

func (k *Knowledge) Ingested(outcome string, chunks int) {
	if k == nil {
		return
	}
	k.ingests.WithLabelValues(outcome).Inc()
	if chunks > 0 {
		k.chunks.Add(float64(chunks))
	}
}

func (k *Knowledge) Embedded(embedder string, elapsed time.Duration) {
	if k == nil {
		return
	}
	k.embedding.WithLabelValues(embedder).Observe(elapsed.Seconds())
}

func (k *Knowledge) Searched(store string, elapsed time.Duration) {
	if k == nil {
		return
	}
	k.search.WithLabelValues(store).Observe(elapsed.Seconds())
}

func (k *Knowledge) Queried(outcome string, results int) {
	if k == nil {
		return
	}
	k.queries.WithLabelValues(outcome).Inc()
	k.topK.Observe(float64(results))
}
