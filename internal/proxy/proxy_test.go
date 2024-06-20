package proxy

import (
	"bytes"
	"context"
	"github.com/clambin/go-common/cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestTMDBProxy_ServeHTTP(t *testing.T) {
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
			defer s.Close()

			p := New(&redis.Options{}, time.Hour, slog.Default())
			p.TargetHost = s.URL
			p.cache.Client = &fakeRedisClient{cache: cache.New[string, string](time.Hour, time.Hour)}

			var w httptest.ResponseRecorder
			r, _ := http.NewRequest(http.MethodGet, "http://localhost"+tt.path, nil)
			p.ServeHTTP(&w, r)
			assert.Equal(t, tt.wantStatusCode, w.Code)
		})
	}
}

func TestTMDBProxy_Collect(t *testing.T) {
	var count atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		count.Add(1)
	}))

	p := New(&redis.Options{}, time.Hour, slog.Default())
	p.TargetHost = s.URL
	p.cache.Client = &fakeRedisClient{cache: cache.New[string, string](time.Hour, time.Hour)}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(p)

	for range 3 {
		var w httptest.ResponseRecorder
		r, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		p.ServeHTTP(&w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	}
	assert.Equal(t, int32(1), count.Load())

	s.Close()
	var w httptest.ResponseRecorder
	r, _ := http.NewRequest(http.MethodGet, "/bar", nil)
	p.ServeHTTP(&w, r)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	assert.NoError(t, testutil.GatherAndCompare(reg, bytes.NewBufferString(`
# HELP tmdb_proxy_cache_hit Number of times the cache was used
# TYPE tmdb_proxy_cache_hit counter
tmdb_proxy_cache_hit 2

# HELP tmdb_proxy_cache_total Number of times the cache was tried
# TYPE tmdb_proxy_cache_total counter
tmdb_proxy_cache_total 4
`)))
}

var _ RedisClient = &fakeRedisClient{}

type fakeRedisClient struct {
	cache *cache.Cache[string, string]
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
