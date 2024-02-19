package proxy

import (
	"github.com/clambin/tmdb/pkg/cache"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"time"
)

var _ prometheus.Collector = &TMDBProxy{}

type TMDBProxy struct {
	TargetHost string
	client     *http.Client
	rt         *cache.RoundTripper
}

func New(expiry, cleanupInterval time.Duration) *TMDBProxy {
	rt := cache.NewRoundTripper(expiry, cleanupInterval, http.DefaultTransport)
	return &TMDBProxy{
		TargetHost: "https://api.themoviedb.org",
		client:     &http.Client{Transport: rt},
		rt:         rt,
	}
}

func (p *TMDBProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := p.TargetHost + r.URL.Path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	req, _ := http.NewRequestWithContext(r.Context(), r.Method, target, nil)
	copyHeader(req.Header, r.Header)
	// TODO: this prevents compression
	// for some reason http.ReadResponse returns the compressed body, and then the end client can't read the body.
	req.Header.Set("Accept-Encoding", "identity")

	resp, err := p.client.Do(req)
	if err != nil {
		http.Error(w, "cache problem", http.StatusBadGateway)
		return
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

	w.WriteHeader(resp.StatusCode)
	copyHeader(w.Header(), resp.Header)
	_, _ = io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (p *TMDBProxy) Describe(ch chan<- *prometheus.Desc) {
	p.rt.Describe(ch)
}

func (p *TMDBProxy) Collect(ch chan<- prometheus.Metric) {
	p.rt.Collect(ch)
}
