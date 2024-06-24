package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/clambin/tmdb/internal/degrees"
	"github.com/clambin/tmdb/pkg/tmdb"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

var (
	debug   = flag.Bool("debug", false, "debug mode")
	authKey = flag.String("authkey", "", "TMDB API authentication key")
	proxy   = flag.String("proxy", "", "Use TMDB Proxy")
	id      = flag.Bool("id", false, "Don't look up actor names, use ID directly")
	depth   = flag.Int("depth", 1, "Maximum number of movies between both actors (1 finds common movies")
)

const maxConcurrentRequests = 15

func main() {
	flag.Parse()
	if *authKey == "" {
		*authKey = os.Getenv("TMDB_AUTHKEY")
		if *authKey == "" {
			panic("no TMDB authentication key provided")
		}
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxIdleConnsPerHost = 100
	t.MaxConnsPerHost = 100

	rt := roundtripper.New(
		roundtripper.WithLimiter(maxConcurrentRequests),
		roundtripper.WithRoundTripper(t),
	)

	tmdbClient := tmdb.New(*authKey, &http.Client{Transport: rt})
	if *proxy != "" {
		tmdbClient.BaseURL = *proxy
	}

	var opts slog.HandlerOptions
	if *debug {
		opts.Level = slog.LevelDebug
	}
	l := slog.New(slog.NewTextHandler(os.Stderr, &opts))

	from, to, err := getActors(tmdbClient)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	pf := degrees.PathFinder{
		TMDBClient: tmdbClient,
		From:       from,
		To:         to,
		Logger:     l,
		Mode:       degrees.ModeMovies | degrees.ModeTVSeries,
	}

	var found bool
	for d := 1; !found && d <= *depth; d++ {
		l.Debug("trying depth " + strconv.Itoa(d))
		ch := make(chan degrees.Path)
		go pf.Find(context.Background(), ch, d)
		for path := range ch {
			fmt.Println(path.String())
			found = true
		}
	}
}

func getActors(c degrees.TMDBClient) (from tmdb.Person, to tmdb.Person, err error) {
	ctx := context.Background()
	for i, arg := range flag.Args() {
		var p tmdb.Person
		if !*id {
			if p, err = degrees.FindActor(ctx, c, arg); err != nil {
				return tmdb.Person{}, tmdb.Person{}, fmt.Errorf("invalid actor %s: %w", arg, err)
			}
		} else {
			var personId int
			if personId, err = strconv.Atoi(arg); err != nil {
				return tmdb.Person{}, tmdb.Person{}, fmt.Errorf("invalid actor id %s: %w", arg, err)
			}
			if p, err = c.GetPerson(context.Background(), personId); err != nil {
				return tmdb.Person{}, tmdb.Person{}, fmt.Errorf("invalid actor %s: %w", arg, err)
			}
		}
		switch i {
		case 0:
			from = p
		case 1:
			to = p
		default:
		}
	}
	return from, to, nil
}
