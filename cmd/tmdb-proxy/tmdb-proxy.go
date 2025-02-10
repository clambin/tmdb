package main

import (
	"context"
	"flag"
	"github.com/clambin/go-common/httputils"
	"github.com/clambin/go-common/httputils/middleware"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/clambin/tmdb/internal/proxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	version        = "change-me"
	debug          = flag.Bool("debug", false, "enable debug logging")
	prometheusAddr = flag.String("metrics.addr", ":9090", "Prometheus metric listener address")
	proxyAddr      = flag.String("proxy.addr", ":8888", "Proxy addr")
	healthAddr     = flag.String("health.addr", ":8080", "Health check addr")
	cacheExpiry    = flag.Duration("cache.ttl", 24*time.Hour, "Time to cache tmdb data")
	redisAddr      = flag.String("cache.redis.addr", "localhost:6379", "Redis address")
	redisDB        = flag.Int("cache.redis.db", 0, "Redis database number")
	redisUsername  = flag.String("cache.redis.username", "", "Redis username")
	redisPassword  = flag.String("cache.redis.password", "", "Redis password")
)

func main() {
	flag.Parse()

	var opts slog.HandlerOptions
	if *debug {
		opts.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &opts))

	logger.Info("Starting proxy",
		"version", version,
		slog.Group("redis", "addr", *redisAddr, "db", *redisDB),
	)

	cacheMetrics := roundtripper.NewCacheMetrics(roundtripper.CacheMetricsOptions{
		Namespace:   "tmdb",
		Subsystem:   "proxy",
		ConstLabels: nil,
		GetPath:     func(r *http.Request) string { return "/" },
	})
	prometheus.MustRegister(cacheMetrics)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		DB:       *redisDB,
		Username: *redisUsername,
		Password: *redisPassword,
	})

	requestLogger := middleware.RequestLogger(logger, slog.LevelDebug, middleware.DefaultRequestLogFormatter)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var g errgroup.Group
	g.Go(func() error {
		return httputils.RunServer(ctx, &http.Server{
			Addr:    *prometheusAddr,
			Handler: promhttp.Handler(),
		})
	})
	g.Go(func() error {
		return httputils.RunServer(ctx, &http.Server{
			Addr:    *healthAddr,
			Handler: proxy.HealthHandler(redisClient, logger.With("handler", "health")),
		})
	})
	g.Go(func() error {
		return httputils.RunServer(ctx, &http.Server{
			Addr: *proxyAddr,
			Handler: requestLogger(
				proxy.TMDBProxyHandler("", redisClient, *cacheExpiry, cacheMetrics, logger.With("handler", "proxy")),
			),
		})
	})

	if err := g.Wait(); err != nil {
		slog.Error("failed to start TMDB proxy server", "err", err)
		os.Exit(1)
	}
}
