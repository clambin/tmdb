package proxy

import (
	"bufio"
	"bytes"
	"github.com/clambin/go-common/cache"
	"net/http"
	"net/http/httputil"
	"time"
)

type Cache struct {
	cache *cache.Cache[string, []byte]
}

func newCache(expiry, cleanupInterval time.Duration) *Cache {
	return &Cache{
		cache: cache.New[string, []byte](expiry, cleanupInterval),
	}
}

func (c *Cache) Get(req *http.Request) (string, *http.Response, bool, error) {
	key := getCacheKey(req)
	body, found := c.cache.Get(key)
	if !found {
		return key, nil, false, nil
	}

	r := bufio.NewReader(bytes.NewReader(body))
	resp, err := http.ReadResponse(r, req)
	return key, resp, found, err
}

func (c *Cache) Put(key string, _ *http.Request, resp *http.Response) error {
	buf, err := httputil.DumpResponse(resp, true)
	if err == nil {
		c.cache.Add(key, buf)
	}
	return err
}

func getCacheKey(r *http.Request) string {
	return r.Method + " | " + r.URL.String()
}
