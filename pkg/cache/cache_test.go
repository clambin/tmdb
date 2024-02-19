package cache_test

import (
	"bytes"
	"github.com/clambin/tmdb/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestResponseCache(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "simple",
			method: http.MethodGet,
			path:   "/foo",
			body:   "a simple GET",
		},
		{
			name:   "method",
			method: http.MethodPost,
			path:   "/foo",
			body:   "a simple POST",
		},
		{
			name:   "with query",
			method: http.MethodGet,
			path:   "/foo?bar=snafu",
			body:   "GET with query",
		},
	}

	c := cache.NewResponseCache(time.Hour, time.Hour)

	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, tt.path, nil)
		key, _, ok, err := c.Get(req)
		require.NoError(t, err)
		require.False(t, ok)

		assert.NoError(t, c.Put(key, nil, &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(tt.body)),
		}))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest(tt.method, tt.path, nil)
			_, resp, ok, err := c.Get(req)
			require.NoError(t, err)
			require.True(t, ok)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.body, string(body))
			require.NoError(t, resp.Body.Close())

		})
	}
}

func BenchmarkCachePut(b *testing.B) {
	c := cache.NewResponseCache(time.Minute, 5*time.Minute)
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBufferString("this is a request"))
	resp := http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("this is a response")),
		Request:    req,
	}
	key := c.GetKey(req)

	b.ResetTimer()
	for range b.N {
		if err := c.Put(key, req, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCacheGet(b *testing.B) {
	c := cache.NewResponseCache(time.Minute, 5*time.Minute)
	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBufferString("this is a request"))
	resp := http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("this is a response")),
		Request:    req,
	}
	_ = c.Put(c.GetKey(req), req, &resp)

	b.ResetTimer()
	for range b.N {
		_, _, ok, err := c.Get(req)
		if err != nil {
			b.Fatal(err)
		}
		if !ok {
			b.Fatal("response not found in cache???")
		}
		_ = req.Body.Close()
	}
}
