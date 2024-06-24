package proxy

import (
	"bufio"
	"bytes"
	"context"
	"github.com/redis/go-redis/v9"
	"net/http"
	"net/http/httputil"
	"time"
)

type Cache struct {
	Namespace string
	Client    RedisClient
}

type RedisClient interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Ping(ctx context.Context) *redis.StatusCmd
}

func (c *Cache) Set(ctx context.Context, req *http.Request, resp *http.Response, expiration time.Duration) error {
	buf, err := httputil.DumpResponse(resp, true)
	if err == nil {
		err = c.Client.Set(ctx, c.getKey(req), string(buf), expiration).Err()
	}
	return err
}

func (c *Cache) Get(ctx context.Context, req *http.Request) (*http.Response, error) {
	body, err := c.Client.Get(ctx, c.getKey(req)).Result()
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(bytes.NewReader([]byte(body)))
	return http.ReadResponse(r, req)
}

func (c *Cache) getKey(r *http.Request) string {
	return c.Namespace + "|" + r.Method + "|" + r.URL.String()
}
