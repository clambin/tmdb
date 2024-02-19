package proxy_test

import (
	"bytes"
	"github.com/clambin/tmdb/internal/proxy"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestTMDBProxy(t *testing.T) {
	var count atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		count.Add(1)
	}))

	p := proxy.New(time.Hour, time.Hour)
	p.TargetHost = s.URL
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

	assert.NoError(t, testutil.CollectAndCompare(p, bytes.NewBufferString(`
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
