package main

import (
	"context"
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
	id      = flag.Bool("id", false, "Don't look up actor names, use ID directly")
	depth   = flag.Int("depth", 1, "Maximum number of movies between both actors (1 finds common movies")
)

func main() {
	flag.Parse()
	if *authKey == "" {
		*authKey = os.Getenv("TMDB_AUTHKEY")
		if *authKey == "" {
			panic("no TMDB authentication key provided")
		}
	}

	tmdbClient := tmdb.New(*authKey, http.DefaultClient)
	if *proxy != "" {
		tmdbClient.BaseURL = *proxy
	}
	c := degrees.New(tmdbClient, slog.Default())

	from, to, err := getArgs(c)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}

	for path := range c.Degrees(context.Background(), from, to, *depth) {
		fmt.Println(path.String())
	}
}

func getArgs(c *degrees.Client) (int, int, error) {
	var from, to int
	var err error
	for i, arg := range flag.Args() {
		var actorID int
		if actorID, err = getActor(c, arg); err != nil {
			return 0, 0, fmt.Errorf("invalid actor %s: %w", arg, err)
		}
		if i == 0 {
			from = actorID
		}
		if i == 1 {
			to = actorID
		}

	}
	return from, to, err
}

func getActor(c *degrees.Client, query string) (int, error) {
	if *id {
		return strconv.Atoi(query)
	}
	actor, err := c.FindActor(context.Background(), query)
	return actor.Id, err
}
