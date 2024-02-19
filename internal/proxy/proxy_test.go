package proxy_test

import (
	"bytes"
	"github.com/clambin/tmdb/internal/proxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
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

			p := proxy.New(time.Hour, time.Hour)
			p.TargetHost = s.URL

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

	p := proxy.New(time.Hour, time.Hour)
	p.TargetHost = s.URL

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
# HELP tmdb_proxy_cache_size Total number of cache entries (excluding expired items)
# TYPE tmdb_proxy_cache_size gauge
tmdb_proxy_cache_size 1
# HELP tmdb_proxy_cache_total Number of times the cache was tried
# TYPE tmdb_proxy_cache_total counter
tmdb_proxy_cache_total 4
# HELP tmdb_proxy_cache_total_size Total number of cache entries (including expired items)
# TYPE tmdb_proxy_cache_total_size gauge
tmdb_proxy_cache_total_size 1
`)))
}
