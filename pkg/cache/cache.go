package cache

import (
	"bufio"
	"bytes"
	"github.com/clambin/go-common/cache"
	"net/http"
	"net/http/httputil"
	"time"
)

type ResponseCache struct {
	cache  *cache.Cache[string, []byte]
	GetKey func(r *http.Request) string
}

func NewResponseCache(expiration, cleanup time.Duration) *ResponseCache {
	return &ResponseCache{
		cache:  cache.New[string, []byte](expiration, cleanup),
		GetKey: func(r *http.Request) string { return r.Method + "|" + r.URL.Path },
	}
}

// Get attempts to retrieve a http.Response from the cache for the request r.  On return, key will hold the key used to store the response
// (to be passed to Put), resp will contain the cached response (if found) and ok indicates if the response was found in the cache.
//
// Clients must call resp.Body.Close when finished reading resp.Body. After that call, clients can inspect resp.Trailer to find
// key/value pairs included in the response trailer.
func (c *ResponseCache) Get(r *http.Request) (key string, resp *http.Response, ok bool, err error) {
	key = c.GetKey(r)
	body, found := c.cache.Get(key)
	if !found {
		return key, nil, false, nil
	}

	response := bufio.NewReader(bytes.NewReader(body))
	resp, err = http.ReadResponse(response, r)
	return key, resp, found, err
}

// Put stores a http.Response in the case, using the provided key.  The request is currently not used.
func (c *ResponseCache) Put(key string, _ *http.Request, resp *http.Response) error {
	buf, err := httputil.DumpResponse(resp, true)
	if err == nil {
		c.cache.Add(key, buf)
	}
	return err
}
