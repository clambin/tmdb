package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

var _ http.RoundTripper = &RoundTripper{}
var _ prometheus.Collector = &RoundTripper{}

type RoundTripper struct {
	cache   *ResponseCache
	metrics *metrics
	next    http.RoundTripper
}

func NewRoundTripper(expiry, cleanup time.Duration, rt http.RoundTripper) *RoundTripper {
	return &RoundTripper{
		cache:   NewResponseCache(expiry, cleanup),
		metrics: newMetrics(),
		next:    rt,
	}
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.metrics.cacheAttempts.Inc()
	cacheKey, resp, found, err := r.cache.Get(req)
	if err != nil || found {
		if found {
			r.metrics.cacheHits.Inc()
		}
		return resp, err
	}

	if resp, err = r.next.RoundTrip(req); err == nil {
		err = r.cache.Put(cacheKey, req, resp)
	}

	return resp, err
}

func (r *RoundTripper) Describe(ch chan<- *prometheus.Desc) {
	r.metrics.cacheHits.Describe(ch)
	r.metrics.cacheAttempts.Describe(ch)
	r.metrics.cacheSize.Describe(ch)
	r.metrics.cacheTotalSize.Describe(ch)
}

func (r *RoundTripper) Collect(ch chan<- prometheus.Metric) {
	r.metrics.cacheSize.Set(float64(r.cache.cache.Len()))
	r.metrics.cacheTotalSize.Set(float64(r.cache.cache.Size()))

	r.metrics.cacheAttempts.Collect(ch)
	r.metrics.cacheHits.Collect(ch)
	r.metrics.cacheSize.Collect(ch)
	r.metrics.cacheTotalSize.Collect(ch)
}
