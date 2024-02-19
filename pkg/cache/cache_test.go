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
	c := cache.NewResponseCache(time.Hour, time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/", bytes.NewBufferString("this is a request"))
	key, resp, ok, err := c.Get(req)
	require.NoError(t, err)
	assert.False(t, ok)

	resp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("this is a response")),
		Request:    req,
	}
	err = c.Put(key, req, resp)
	assert.NoError(t, err)

	key, resp, ok, err = c.Get(req)
	require.NoError(t, err)
	assert.True(t, ok)

	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "this is a response", string(body))
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
