package main

import (
	"errors"
	"flag"
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/clambin/tmdb/internal/proxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var (
	prometheusAddr = flag.String("metrics", ":9090", "Prometheus metric listener address")
	proxyAddr      = flag.String("addr", ":8888", "Proxy listener addr")
	cacheExpiry    = flag.Duration("cache", 24*time.Hour, "Time to cache tmdb data")
	debug          = flag.Bool("debug", false, "enable debug logging")
)

func main() {
	flag.Parse()

	p := proxy.New(*cacheExpiry, time.Hour)
	prometheus.MustRegister(p)

	var opts slog.HandlerOptions
	if *debug {
		opts.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &opts))

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(*prometheusAddr, nil); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start Prometheus metrics server", "err", err)
			os.Exit(1)
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
