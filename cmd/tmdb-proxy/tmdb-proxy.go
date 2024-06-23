package main

import (
	"errors"
	"flag"
	"github.com/clambin/go-common/http/middleware"
	"github.com/clambin/tmdb/internal/proxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"os"
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

	o := redis.Options{
		Addr:     *redisAddr,
		DB:       *redisDB,
		Username: *redisUsername,
		Password: *redisPassword,
	}

	logger.Info("Starting proxy",
		"version", version,
		slog.Group("redis", "addr", *redisAddr, "db", *redisDB),
	)

	p := proxy.New(&o, *cacheExpiry, logger)
	prometheus.MustRegister(p)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(*prometheusAddr, nil); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start Prometheus metrics server", "err", err)
			os.Exit(1)
		}
	}()

	go func() {
		m := http.NewServeMux()
		m.Handle("/readyz", p.Health())
		healthServer := http.Server{Addr: *healthAddr, Handler: m}
		if err := healthServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start /readyz listener", "err", err)
		}
	}()

	s := http.Server{
		Addr:    *proxyAddr,
		Handler: middleware.RequestLogger(logger, slog.LevelDebug, middleware.DefaultRequestLogFormatter)(p),
	}
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start TMDB proxy server", "err", err)
		os.Exit(1)
	}
}
