package degrees

import (
	"context"
	"github.com/clambin/tmdb/internal/tmdb"
	"log/slog"
)

type Client struct {
	TMDBClient
	IncludeTV bool
	logger    *slog.Logger
}

type TMDBClient interface {
	GetPersonCredits(ctx context.Context, id int) (tmdb.PersonCredits, error)
	SearchPersonPage(ctx context.Context, query string, page int) ([]tmdb.Person, int, error)
	GetPerson(ctx context.Context, id int) (tmdb.Person, error)
	GetMovie(ctx context.Context, id int) (tmdb.Movie, error)
	GetMovieCredits(ctx context.Context, id int) (tmdb.MovieCredits, error)
}

var _ TMDBClient = tmdb.Client{}

func New(c TMDBClient, logger *slog.Logger) *Client {
	return &Client{
		TMDBClient: c,
		logger:     logger,
	}
}
