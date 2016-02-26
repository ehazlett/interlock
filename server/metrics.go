package server

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	eventsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "interlock",
			Subsystem: "totals",
			Name:      "events_processed",
			Help:      "Total number of events processed",
		},
	)

	lastReloadDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "interlock",
		Subsystem: "system",
		Name:      "last_reload_duration",
		Help:      "Duration of last reload in nanoseconds",
	})

	uptime = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "interlock",
			Subsystem: "totals",
			Name:      "uptime",
			Help:      "Uptime in seconds",
		},
	)
)

type Metrics struct {
	EventsProcessed    prometheus.Counter
	LastReloadDuration prometheus.Gauge
	Uptime             prometheus.Counter
}

func NewMetrics() *Metrics {
	prometheus.MustRegister(eventsProcessed)
	prometheus.MustRegister(lastReloadDuration)
	prometheus.MustRegister(uptime)

	return &Metrics{
		EventsProcessed:    eventsProcessed,
		LastReloadDuration: lastReloadDuration,
		Uptime:             uptime,
	}
}
