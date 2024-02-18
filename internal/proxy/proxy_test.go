package proxy_test

import (
	"bytes"
	"github.com/clambin/tmdb/internal/proxy"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTMDBProxy(t *testing.T) {
	var server server
	s := httptest.NewServer(&server)

	p := proxy.New(time.Hour, time.Hour)
	p.TargetHost = s.URL
	for range 3 {
		var w httptest.ResponseRecorder
		r, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		p.ServeHTTP(&w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	}
	assert.Equal(t, 1, server.count)

	s.Close()
	var w httptest.ResponseRecorder
	r, _ := http.NewRequest(http.MethodGet, "/bar", nil)
	p.ServeHTTP(&w, r)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	assert.NoError(t, testutil.CollectAndCompare(p, bytes.NewBufferString(`
# HELP tmdb_proxy_cache_hit Number of times the cache was used
# TYPE tmdb_proxy_cache_hit counter
tmdb_proxy_cache_hit 2
# HELP tmdb_proxy_cache_total Number of times the cache was tried
# TYPE tmdb_proxy_cache_total counter
tmdb_proxy_cache_total 4

# HELP tmdb_proxy_cache_count Total number of cache entries (excluding expired items)
# TYPE tmdb_proxy_cache_count gauge
tmdb_proxy_cache_count 1
# HELP tmdb_proxy_cache_size Total number of cache entries (including expired items)
# TYPE tmdb_proxy_cache_size gauge
tmdb_proxy_cache_size 1

`)))
}

var _ http.Handler = &server{}

type server struct {
	count int
}

func (s *server) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	s.count++
	w.WriteHeader(http.StatusOK)
}
