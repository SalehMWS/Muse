package metrics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (m *Metrics) RegisterRoutes(app *fiber.App, path string) {
	if m == nil {
		return
	}
	if path == "" {
		path = "/metrics"
	}
	handler := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
	app.Get(path, adaptor.HTTPHandler(handler))
}
