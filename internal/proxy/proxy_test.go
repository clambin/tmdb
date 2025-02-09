package proxy

import (
	"bytes"
	"context"
	"errors"
	"github.com/clambin/go-common/cache"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestProxyHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		wantPath       string
		wantQuery      string
		wantStatusCode int
	}{
		{
			name:           "simple path",
			path:           "/foo",
			wantPath:       "/foo",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "with query",
			path:           "/foo?bar=snafu",
			wantPath:       "/foo",
			wantQuery:      "bar=snafu",
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Helper()
				if !assert.Equal(t, tt.wantPath, r.URL.Path) {
					http.Error(w, "invalid path", http.StatusBadRequest)
					return
				}
				if !assert.Equal(t, tt.wantQuery, r.URL.RawQuery) {
					http.Error(w, "invalid query", http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(s.Close)

			redisClient := fakeRedisClient{cache: cache.New[string, string](time.Hour, time.Hour)}
			h := TMDBProxyHandler(s.URL, &redisClient, time.Minute, nil, discardLogger)

			var w httptest.ResponseRecorder
			r, _ := http.NewRequest(http.MethodGet, "http://localhost"+tt.path, nil)
			h.ServeHTTP(&w, r)
			assert.Equal(t, tt.wantStatusCode, w.Code)
		})
	}
}

func TestTMDBProxyHandler_Metrics(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", 1024)))
	}))
	t.Cleanup(s.Close)

	cacheMetrics := roundtripper.NewCacheMetrics(roundtripper.CacheMetricsOptions{GetPath: func(r *http.Request) string {
		return "/"
	}})

	h := TMDBProxyHandler(s.URL, &fakeRedisClient{cache: cache.New[string, string](time.Hour, time.Hour)}, time.Minute, cacheMetrics, discardLogger)
	r, err := http.NewRequest(http.MethodGet, "/foo", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, err)

	assert.NoError(t, testutil.CollectAndCompare(cacheMetrics, bytes.NewBufferString(`
# HELP http_cache_hit_total Number of times the cache was used
# TYPE http_cache_hit_total counter
http_cache_hit_total{method="GET",path="/"} 1

# HELP http_cache_total Number of times the cache was consulted
# TYPE http_cache_total counter
http_cache_total{method="GET",path="/"} 2
`)))
}

func TestHealthHandler(t *testing.T) {
	var redisClient fakeRedisClient
	h := HealthHandler(&redisClient, discardLogger)

	r, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	redisClient.pingErr = errors.New("ping error")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ RedisClient = &fakeRedisClient{}

type fakeRedisClient struct {
	cache   *cache.Cache[string, string]
	pingErr error
}

func (f fakeRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	cmd.SetErr(f.pingErr)
	return cmd
}

func (f fakeRedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	f.cache.AddWithExpiry(key, value.(string), ttl)
	return redis.NewStatusCmd(ctx)
}

func (f fakeRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	value, ok := f.cache.Get(key)
	cmd := redis.NewStringCmd(ctx)
	if !ok {
		cmd.SetErr(redis.Nil)
	} else {
		cmd.SetVal(value)
	}
	return cmd
}
