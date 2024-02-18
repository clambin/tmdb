package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/clambin/tmdb/internal/degrees"
	"github.com/clambin/tmdb/pkg/tmdb"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

var (
	authKey = flag.String("authkey", "", "TMDB API authentication key")
	proxy   = flag.String("proxy", "", "Use TMDB Proxy")
)

func main() {
	from, to, err := getArgs()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}
	tmdbClient := tmdb.New(*authKey, http.DefaultClient)
	if *proxy != "" {
		tmdbClient.BaseURL = *proxy
	}

	c := degrees.New(tmdbClient, slog.Default())

	for path := range c.Degrees(context.Background(), from, to, 4) {
		fmt.Println(path.String())
	}
}

func getArgs() (int, int, error) {
	flag.Parse()
	if *authKey == "" {
		*authKey = os.Getenv("TMDB_AUTHKEY")
		if *authKey == "" {
			return 0, 0, errors.New("no TMDB authentication key provided")
		}
	}
	var from, to int
	var err error
	for i, arg := range flag.Args() {
		if i == 0 {
			if from, err = strconv.Atoi(arg); err != nil {
				return 0, 0, fmt.Errorf("invalid from id: %w", err)
			}
		}
		if i == 1 {
			if to, err = strconv.Atoi(arg); err != nil {
				return 0, 0, fmt.Errorf("invalid to id: %w", err)
			}
		}
	}
	return from, to, err
}
