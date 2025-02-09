package proxy

import (
	"errors"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"io"
	"log/slog"
	"net/http"
	"time"
)

var _ prometheus.Collector = &TMDBProxy{}

type TMDBProxy struct {
	TargetHost string
	cache      Cache
	ttl        time.Duration
	httpClient *http.Client
	roundtripper.CacheMetrics
	logger *slog.Logger
}

func New(cfg *redis.Options, ttl time.Duration, logger *slog.Logger) *TMDBProxy {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxIdleConnsPerHost = 100
	t.MaxConnsPerHost = 100

	return &TMDBProxy{
		TargetHost: "https://api.themoviedb.org",
		cache: Cache{
			Namespace: "github.com/clambin/tmdb",
			Client:    redis.NewClient(cfg),
		},
		ttl: ttl,
		httpClient: &http.Client{
			Timeout:   time.Second * 10,
			Transport: t,
		},
		CacheMetrics: newMetrics(),
		logger:       logger,
	}
}

func (p *TMDBProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	cached := true
	if resp, err = p.cache.Get(r.Context(), r); err != nil {
		cached = false
		if !errors.Is(err, redis.Nil) {
			p.logger.Warn("failed to get cached response", "error", err)
		}

		if resp, err = p.call(r); err == nil && resp.StatusCode == http.StatusOK {
			err = p.cache.Set(r.Context(), r, resp, p.ttl)
		}
	}
	p.CacheMetrics.Measure(r, cached)

	if err != nil {
		http.Error(w, "failed to get request", http.StatusBadGateway)
		return
	}

	w.WriteHeader(resp.StatusCode)
	copyHeader(w.Header(), resp.Header)
	_, _ = io.Copy(w, resp.Body)
	_ = resp.Body.Close()
}

func (p *TMDBProxy) call(r *http.Request) (*http.Response, error) {
	target := p.TargetHost + r.URL.Path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	req, _ := http.NewRequestWithContext(r.Context(), r.Method, target, nil)
	copyHeader(req.Header, r.Header)
	// TODO: this prevents compression
	// for some reason http.ReadResponse returns the compressed body, and then the end client can't read the body.
	req.Header.Set("Accept-Encoding", "identity")

	return p.httpClient.Do(req)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (p *TMDBProxy) Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := p.cache.Client.Ping(r.Context()).Err(); err != nil {
			p.logger.Warn("failed to ping cache", "err", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
