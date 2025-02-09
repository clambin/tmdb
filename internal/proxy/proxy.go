package proxy

import (
	"cmp"
	"errors"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/redis/go-redis/v9"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func TMDBProxyHandler(target string, redisClient RedisClient, ttl time.Duration, cacheMetrics roundtripper.CacheMetrics, logger *slog.Logger) http.Handler {
	cache := Cache{
		Namespace: "github.com/clambin/tmdb",
		Client:    redisClient,
	}
	tp := http.DefaultTransport.(*http.Transport).Clone()
	tp.MaxIdleConns = 100
	tp.MaxIdleConnsPerHost = 100
	tp.MaxConnsPerHost = 100

	client := tmdbClient{
		TargetHost: cmp.Or(target, "https://api.themoviedb.org", target),
		httpClient: &http.Client{
			Transport: tp,
			Timeout:   time.Second * 10,
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp *http.Response
		var err error
		var fetched bool
		if resp, err = cache.Get(r.Context(), r); err != nil {
			fetched = true
			if !errors.Is(err, redis.Nil) {
				logger.Warn("failed to get cached response", "error", err)
			}

			if resp, err = client.call(r); err == nil && resp.StatusCode == http.StatusOK {
				err = cache.Set(r.Context(), r, resp, ttl)
			}
		}

		if cacheMetrics != nil {
			cacheMetrics.Measure(r, !fetched)
		}

		if err != nil {
			http.Error(w, "failed to get request", http.StatusBadGateway)
			return
		}

		w.WriteHeader(resp.StatusCode)
		// TODO: set Content Encoding to "identity", i.e. no compression?
		copyHeader(w.Header(), resp.Header)
		_, _ = io.Copy(w, resp.Body)
		_ = resp.Body.Close()
	})
}

type tmdbClient struct {
	TargetHost string
	httpClient *http.Client
}

func (p tmdbClient) call(r *http.Request) (*http.Response, error) {
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

func HealthHandler(client RedisClient, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := client.Ping(r.Context()).Err(); err != nil {
			logger.Warn("failed to ping cache", "err", err)
			http.Error(w, "failed to ping cache", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
