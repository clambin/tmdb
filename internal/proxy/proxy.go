package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"time"
)

var _ prometheus.Collector = &TMDBProxy{}

type TMDBProxy struct {
	HTTPClient *http.Client
	TargetHost string
	cache      *Cache
	metrics    *metrics
}

func New(expiry, cleanupInterval time.Duration) *TMDBProxy {
	c := newCache(expiry, cleanupInterval)
	return &TMDBProxy{
		HTTPClient: http.DefaultClient,
		TargetHost: "https://api.themoviedb.org",
		cache:      c,
		metrics:    newMetrics(c),
	}
}

func (p *TMDBProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key, resp, ok, err := p.cache.Get(r)
	if err != nil {
		http.Error(w, "cache problem", http.StatusInternalServerError)
		return
	}

	p.metrics.addRequest()
	if ok {
		p.metrics.addHit()
	} else {
		if resp, err = p.do(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if resp.StatusCode == http.StatusOK {
			_ = p.cache.Put(key, r, resp)
		}
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

	w.WriteHeader(resp.StatusCode)
	copyHeader(w.Header(), resp.Header)
	_, _ = io.Copy(w, resp.Body)
}

func (p *TMDBProxy) do(r *http.Request) (*http.Response, error) {
	req, _ := http.NewRequestWithContext(r.Context(), r.Method, p.TargetHost+r.URL.String(), nil)
	copyHeader(req.Header, r.Header)
	// TODO: this prevents compression
	// for some reason http.ReadResponse returns the compressed body, and then the end client can't read the body.
	req.Header.Set("Accept-Encoding", "identity")
	return p.HTTPClient.Do(req)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (p *TMDBProxy) Describe(ch chan<- *prometheus.Desc) {
	p.metrics.Describe(ch)
}

func (p *TMDBProxy) Collect(ch chan<- prometheus.Metric) {
	p.metrics.Collect(ch)
}
