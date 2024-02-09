package proxy

import "github.com/prometheus/client_golang/prometheus"

var _ prometheus.Collector = &metrics{}

type metrics struct {
	totalCounter prometheus.Counter
	hitCounter   prometheus.Counter
}

func newMetrics() *metrics {
	return &metrics{
		totalCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_total",
			Help:        "Number of times the cache was tried",
			ConstLabels: nil,
		}),
		hitCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "tmdb",
			Subsystem:   "proxy",
			Name:        "cache_hit",
			Help:        "Number of times the cache was used",
			ConstLabels: nil,
		}),
	}
}

func (m *metrics) add() {
	m.totalCounter.Add(1.0)
}

func (m *metrics) hit() {
	m.hitCounter.Add(1.0)
}

func (m *metrics) Describe(ch chan<- *prometheus.Desc) {
	m.totalCounter.Describe(ch)
	m.hitCounter.Describe(ch)
}

func (m *metrics) Collect(ch chan<- prometheus.Metric) {
	m.totalCounter.Collect(ch)
	m.hitCounter.Collect(ch)
}
