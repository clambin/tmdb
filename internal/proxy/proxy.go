package proxy

import (
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"time"
)

var _ prometheus.Collector = &TMDBProxy{}

type TMDBProxy struct {
	TargetHost string
	client     *http.Client
	roundtripper.CacheMetrics
}

func New(expiry, cleanupInterval time.Duration) *TMDBProxy {
	cacheMetrics := newMetrics()
	rt := roundtripper.New(roundtripper.WithCache(roundtripper.CacheOptions{
		CacheTable:        nil,
		DefaultExpiration: expiry,
		CleanupInterval:   cleanupInterval,
		GetKey:            func(r *http.Request) string { return r.Method + "|" + r.URL.String() },
		CacheMetrics:      cacheMetrics,
	}))
	return &TMDBProxy{
		TargetHost:   "https://api.themoviedb.org",
		client:       &http.Client{Transport: rt},
		CacheMetrics: cacheMetrics,
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

	w.WriteHeader(resp.StatusCode)
	copyHeader(w.Header(), resp.Header)
	_, _ = io.Copy(w, resp.Body)
	_ = resp.Body.Close()
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
