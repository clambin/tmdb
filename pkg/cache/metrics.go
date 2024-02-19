package cache

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	cacheAttempts  prometheus.Counter
	cacheHits      prometheus.Counter
	cacheSize      prometheus.Gauge
	cacheTotalSize prometheus.Gauge
}

func newMetrics() *metrics {
	return &metrics{
		cacheAttempts: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_total",
			Help:        "Number of times the cache was tried",
			ConstLabels: nil,
		}),
		cacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_hit",
			Help:        "Number of times the cache was used",
			ConstLabels: nil,
		}),
		cacheSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_size",
			Help:        "Total number of cache entries (excluding expired items)",
			ConstLabels: nil,
		}),
		cacheTotalSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_total_size",
			Help:        "Total number of cache entries (including expired items)",
			ConstLabels: nil,
		}),
	}
}
