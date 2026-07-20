package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	EventUserRegistered     = "user_registered"
	EventUserLoggedIn       = "user_logged_in"
	EventInstagramConnected = "instagram_connected"
	EventContentCreated     = "content_created"
	EventCaptionGenerated   = "caption_generated"
	EventPostPublished      = "post_published"
	EventScheduleCreated    = "schedule_created"
	EventDocumentIngested   = "document_ingested"
)

type Business struct {
	events *prometheus.CounterVec
}

func newBusiness() *Business {
	return &Business{
		events: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: "business",
			Name:      "events_total",
			Help:      "Product-level events, by event name.",
		}, []string{"event"}),
	}
}

func (b *Business) collectors() []prometheus.Collector {
	return []prometheus.Collector{b.events}
}

func (b *Business) Record(event string) {
	if b == nil {
		return
	}
	b.events.WithLabelValues(event).Inc()
}
