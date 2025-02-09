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
	tp := http.DefaultTransport.(*http.Transport).Clone()
	tp.MaxIdleConns = 100
	tp.MaxIdleConnsPerHost = 100
	tp.MaxConnsPerHost = 100

	client := tmdbClient{
		TargetHost: cmp.Or(target, "https://api.themoviedb.org"),
		httpClient: &http.Client{
			Transport: tp,
			Timeout:   time.Second * 10,
		},
	}

	responses := responseCache{
		Namespace: "github.com/clambin/tmdb",
		Client:    redisClient,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var fetched bool
		resp, err := responses.Get(r.Context(), r)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("failed to get cached response", "error", err)
			}

			if resp, err = client.call(r); err == nil && resp.StatusCode == http.StatusOK {
				fetched = true
				err = responses.Set(r.Context(), r, resp, ttl)
			}
		}

		if err != nil {
			http.Error(w, "failed to get request", http.StatusBadGateway)
			return
		}

		if cacheMetrics != nil {
			cacheMetrics.Measure(r, !fetched)
		}

		w.WriteHeader(resp.StatusCode)
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
	// ask for non-compressed responses so we have a clear text copy in our cache
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
