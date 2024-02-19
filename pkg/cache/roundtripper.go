package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

var _ http.RoundTripper = &RoundTripper{}
var _ prometheus.Collector = &RoundTripper{}

var (
	cacheAttempts = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "tmdb",
		Subsystem:   "proxy",
		Name:        "cache_total",
		Help:        "Number of times the cache was tried",
		ConstLabels: nil,
	})
	cacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "tmdb",
		Subsystem:   "proxy",
		Name:        "cache_hit",
		Help:        "Number of times the cache was used",
		ConstLabels: nil,
	})
	cacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "tmdb",
		Subsystem:   "proxy",
		Name:        "cache_size",
		Help:        "Total number of cache entries (excluding expired items)",
		ConstLabels: nil,
	})
	cacheTotalSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "tmdb",
		Subsystem:   "proxy",
		Name:        "cache_total_size",
		Help:        "Total number of cache entries (including expired items)",
		ConstLabels: nil,
	})
)

type RoundTripper struct {
	cache *ResponseCache
	next  http.RoundTripper
}

func NewRoundTripper(expiry, cleanup time.Duration, rt http.RoundTripper) *RoundTripper {
	return &RoundTripper{
		cache: NewResponseCache(expiry, cleanup),
		next:  rt,
	}
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	cacheAttempts.Inc()
	cacheKey, resp, found, err := r.cache.Get(req)
	if err != nil {
		return nil, err
	}
	if found {
		cacheHits.Inc()
		return resp, nil
	}

	if resp, err = r.next.RoundTrip(req); err == nil {
		err = r.cache.Put(cacheKey, req, resp)
	}

	return resp, err
}

func (r *RoundTripper) Describe(ch chan<- *prometheus.Desc) {
	cacheHits.Describe(ch)
	cacheAttempts.Describe(ch)
	cacheSize.Describe(ch)
	cacheTotalSize.Describe(ch)
}

func (r *RoundTripper) Collect(ch chan<- prometheus.Metric) {
	cacheSize.Set(float64(r.cache.cache.Len()))
	cacheTotalSize.Set(float64(r.cache.cache.Size()))

	cacheAttempts.Collect(ch)
	cacheHits.Collect(ch)
	cacheSize.Collect(ch)
	cacheTotalSize.Collect(ch)
}
