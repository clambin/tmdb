package proxy

import (
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

var _ roundtripper.CacheMetrics = metrics{}

type metrics struct {
	cacheAttempts prometheus.Counter
	cacheHits     prometheus.Counter
}

func newMetrics() *metrics {
	return &metrics{
		cacheAttempts: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_total_count",
			Help:        "Number of times the cache was tried",
			ConstLabels: nil,
		}),
		cacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_hit_count",
			Help:        "Number of times the cache was used",
			ConstLabels: nil,
		}),
	}
}

func (m metrics) Measure(_ *http.Request, found bool) {
	if found {
		m.cacheHits.Inc()
	}
	m.cacheAttempts.Inc()
}

func (m metrics) Describe(ch chan<- *prometheus.Desc) {
	m.cacheHits.Describe(ch)
	m.cacheAttempts.Describe(ch)
}

func (m metrics) Collect(ch chan<- prometheus.Metric) {
	m.cacheHits.Collect(ch)
	m.cacheAttempts.Collect(ch)
}
