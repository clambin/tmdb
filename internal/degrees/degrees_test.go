package degrees_test

import (
	"context"
	"errors"
	"github.com/clambin/tmdb/internal/degrees"
	"github.com/clambin/tmdb/internal/degrees/mocks"
	"github.com/clambin/tmdb/pkg/tmdb"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"strconv"
	"testing"
)

func TestClient_Degrees(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(context.Context, *mocks.TMDBClient)
		fromID   int
		toID     int
		maxDepth int
		want     []string
	}{
		{
			name: "common movie",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPersonCredits(ctx, 1).Return(makePersonCredits(1, 1), nil) //.Once()
				getter.EXPECT().GetPersonCredits(ctx, 2).Return(makePersonCredits(2, 1), nil) //.Once()
			},
			fromID:   1,
			toID:     2,
			maxDepth: 2,
			want: []string{
				"actor1 -> movie1 -> actor2 (2)",
			},
		},
		{
			name: "match",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPersonCredits(ctx, 1).Return(makePersonCredits(1, 1), nil)       //.Once()
				getter.EXPECT().GetPersonCredits(ctx, 2).Return(makePersonCredits(2, 1, 2), nil)    //.Once()
				getter.EXPECT().GetPersonCredits(ctx, 3).Return(makePersonCredits(3, 2, 3, 4), nil) //.Once()
				getter.EXPECT().GetPersonCredits(ctx, 4).Return(makePersonCredits(4, 3, 4), nil)    //.Once()
				getter.EXPECT().GetMovieCredits(ctx, 1).Return(makeMovieCredits(1, 1, 2), nil)      //.Once()
				getter.EXPECT().GetMovieCredits(ctx, 2).Return(makeMovieCredits(2, 2, 3), nil)      //.Once()
			},
			fromID:   1,
			toID:     4,
			maxDepth: 4,
			want: []string{
				"actor1 -> movie1 -> actor2 -> movie2 -> actor3 -> movie3 -> actor4 (4)",
				"actor1 -> movie1 -> actor2 -> movie2 -> actor3 -> movie4 -> actor4 (4)",
			},
		},
		{
			name: "too short",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPersonCredits(ctx, 1).Return(makePersonCredits(1, 1), nil) //.Once()
				getter.EXPECT().GetPersonCredits(ctx, 4).Return(makePersonCredits(4, 4), nil)
			},
			fromID:   1,
			toID:     4,
			maxDepth: 1,
			want:     []string{},
		},
		{
			name: "GetPersonCredits fails",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPersonCredits(ctx, 4).Return(tmdb.PersonCredits{}, errors.New("failed")) //.Once()
			},
			fromID:   1,
			toID:     4,
			maxDepth: 1,
			want:     []string{},
		},
		{
			name: "GetMovieCredits fails",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPersonCredits(ctx, 4).Return(makePersonCredits(4, 4), nil)
				getter.EXPECT().GetPersonCredits(ctx, 1).Return(makePersonCredits(1, 1), nil)             //.Once()
				getter.EXPECT().GetMovieCredits(ctx, 1).Return(tmdb.MovieCredits{}, errors.New("failed")) //.Once()
			},
			fromID:   1,
			toID:     4,
			maxDepth: 3,
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			getter := mocks.NewTMDBClient(t)
			if tt.setup != nil {
				tt.setup(ctx, getter)
			}

			l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

			ch := make(chan degrees.Path)
			f := degrees.PathFinder{
				TMDBClient: getter,
				From:       makePerson(tt.fromID),
				To:         makePerson(tt.toID),
				Logger:     l,
			}
			go f.Find(ctx, ch, tt.maxDepth)
			var count int
			for path := range ch {
				t.Log(path.String())
				assert.Contains(t, tt.want, path.String())
				count++
			}
			assert.Equal(t, len(tt.want), count)
		})
	}
}

func makePerson(id int) tmdb.Person {
	return tmdb.Person{Id: id, Name: "actor" + strconv.Itoa(id)}
}

func makePersonCredits(id int, movieIDs ...int) tmdb.PersonCredits {
	c := tmdb.PersonCredits{Id: id}
	for _, movieID := range movieIDs {
		c.Cast = append(c.Cast, tmdb.CastCredit{
			Id:        movieID,
			MediaType: "movie",
			Title:     "movie" + strconv.Itoa(movieID),
		})
	}
	return c
}

func makeMovieCredits(id int, actorIDs ...int) tmdb.MovieCredits {
	c := tmdb.MovieCredits{Id: id}
	for _, personID := range actorIDs {
		c.Cast = append(c.Cast, tmdb.MovieCastCredits{
			Id:   personID,
			Name: "actor" + strconv.Itoa(personID),
		})
	}
	return c
}
