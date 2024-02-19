package cache_test

import (
	"bytes"
	"github.com/clambin/tmdb/pkg/cache"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRoundTripper(t *testing.T) {
	var count int
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		count++
	}))

	rt := cache.NewRoundTripper(time.Hour, time.Hour, http.DefaultTransport)
	client := http.Client{Transport: rt}
	for range 3 {
		resp, err := client.Get(s.URL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
	assert.Equal(t, 1, count)

	s.Close()
	_, err := client.Get(s.URL)
	assert.NoError(t, err)

	assert.NoError(t, testutil.CollectAndCompare(rt, bytes.NewBufferString(`
# HELP tmdb_proxy_cache_hit Number of times the cache was used
# TYPE tmdb_proxy_cache_hit counter
tmdb_proxy_cache_hit 3
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
