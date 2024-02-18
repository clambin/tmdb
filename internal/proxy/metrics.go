package proxy

import "github.com/prometheus/client_golang/prometheus"

var _ prometheus.Collector = &metrics{}

type metrics struct {
	cache  *Cache
	total  prometheus.Counter
	hits   prometheus.Counter
	count  prometheus.Gauge
	length prometheus.Gauge
}

func newMetrics(c *Cache) *metrics {
	return &metrics{
		cache: c,
		total: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_total",
			Help:        "Number of times the cache was tried",
			ConstLabels: nil,
		}),
		hits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_hit",
			Help:        "Number of times the cache was used",
			ConstLabels: nil,
		}),
		length: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_count",
			Help:        "Total number of cache entries (excluding expired items)",
			ConstLabels: nil,
		}),
		count: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_size",
			Help:        "Total number of cache entries (including expired items)",
			ConstLabels: nil,
		}),
	}
}

func (m *metrics) addRequest() {
	m.total.Add(1.0)
}

func (m *metrics) addHit() {
	m.hits.Add(1.0)
}

func (m *metrics) Describe(ch chan<- *prometheus.Desc) {
	m.total.Describe(ch)
	m.hits.Describe(ch)
	m.count.Describe(ch)
	m.length.Describe(ch)
}

func (m *metrics) Collect(ch chan<- prometheus.Metric) {
	m.total.Collect(ch)
	m.hits.Collect(ch)
	m.count.Set(float64(m.cache.cache.Size()))
	m.count.Collect(ch)
	m.length.Set(float64(m.cache.cache.Len()))
	m.length.Collect(ch)
}
