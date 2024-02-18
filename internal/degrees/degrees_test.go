package degrees_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-common/set"
	"github.com/clambin/tmdb/internal/degrees"
	"github.com/clambin/tmdb/internal/degrees/mocks"
	"github.com/clambin/tmdb/internal/tmdb"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"strconv"
	"strings"
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
			name: "match",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPerson(ctx, 1).Return(makePerson(1), nil).Once()
				getter.EXPECT().GetPerson(ctx, 2).Return(makePerson(2), nil).Once()
				getter.EXPECT().GetPerson(ctx, 3).Return(makePerson(3), nil).Once()

				getter.EXPECT().GetPersonCredits(ctx, 1).Return(makePersonCredits(1, 1), nil).Once()
				getter.EXPECT().GetPersonCredits(ctx, 2).Return(makePersonCredits(2, 1, 2), nil).Once()
				getter.EXPECT().GetPersonCredits(ctx, 3).Return(makePersonCredits(3, 2, 3, 4), nil).Once()

				getter.EXPECT().GetMovieCredits(ctx, 1).Return(makeMovieCredits(1, 1, 2), nil).Once()
				getter.EXPECT().GetMovieCredits(ctx, 2).Return(makeMovieCredits(2, 2, 3), nil).Once()
				getter.EXPECT().GetMovieCredits(ctx, 3).Return(makeMovieCredits(3, 3, 4), nil).Once()
				getter.EXPECT().GetMovieCredits(ctx, 4).Return(makeMovieCredits(4, 3, 4), nil).Once()
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
			name: "short",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPerson(ctx, 1).Return(makePerson(1), nil).Once()
				getter.EXPECT().GetPerson(ctx, 2).Return(makePerson(2), nil).Once()
				getter.EXPECT().GetPerson(ctx, 3).Return(makePerson(3), nil).Once()

				getter.EXPECT().GetPersonCredits(ctx, 1).Return(makePersonCredits(1, 1), nil).Once()
				getter.EXPECT().GetPersonCredits(ctx, 2).Return(makePersonCredits(2, 1, 2), nil).Once()
				getter.EXPECT().GetPersonCredits(ctx, 3).Return(makePersonCredits(3, 2, 3, 4), nil).Once()

				getter.EXPECT().GetMovieCredits(ctx, 1).Return(makeMovieCredits(1, 1, 2), nil).Once()
				getter.EXPECT().GetMovieCredits(ctx, 2).Return(makeMovieCredits(2, 2, 3, 4), nil).Once()
				getter.EXPECT().GetMovieCredits(ctx, 3).Return(makeMovieCredits(3, 3, 4), nil).Once()
				getter.EXPECT().GetMovieCredits(ctx, 4).Return(makeMovieCredits(4, 3, 4), nil).Once()
			},
			fromID:   1,
			toID:     4,
			maxDepth: 4,
			want: []string{
				"actor1 -> movie1 -> actor2 -> movie2 -> actor3 -> movie3 -> actor4 (4)",
				"actor1 -> movie1 -> actor2 -> movie2 -> actor3 -> movie4 -> actor4 (4)",
				"actor1 -> movie1 -> actor2 -> movie2 -> actor4 (3)",
			},
		},
		{
			name: "too short",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPerson(ctx, 1).Return(makePerson(1), nil).Once()
				getter.EXPECT().GetPersonCredits(ctx, 1).Return(makePersonCredits(1, 1), nil).Once()
				getter.EXPECT().GetMovieCredits(ctx, 1).Return(makeMovieCredits(1, 1, 2), nil).Once()
			},
			fromID:   1,
			toID:     4,
			maxDepth: 2,
			want:     []string{},
		},
		{
			name: "GetPerson fails",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPerson(ctx, 1).Return(tmdb.Person{}, errors.New("failed")).Once()
			},
			fromID:   1,
			toID:     4,
			maxDepth: 2,
			want:     []string{},
		},
		{
			name: "GetPersonCredits fails",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPerson(ctx, 1).Return(makePerson(1), nil).Once()
				getter.EXPECT().GetPersonCredits(ctx, 1).Return(tmdb.PersonCredits{}, errors.New("failed")).Once()
			},
			fromID:   1,
			toID:     4,
			maxDepth: 2,
			want:     []string{},
		},
		{
			name: "GetMovieCredits fails",
			setup: func(ctx context.Context, getter *mocks.TMDBClient) {
				getter.EXPECT().GetPerson(ctx, 1).Return(makePerson(1), nil).Once()
				getter.EXPECT().GetPersonCredits(ctx, 1).Return(makePersonCredits(1, 1), nil).Once()
				getter.EXPECT().GetMovieCredits(ctx, 1).Return(tmdb.MovieCredits{}, errors.New("failed")).Once()
			},
			fromID:   1,
			toID:     4,
			maxDepth: 4,
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
			c := degrees.New(getter, l)

			var count int
			expected := set.New[string](tt.want...)
			for path := range c.Degrees(ctx, tt.fromID, tt.toID, tt.maxDepth) {
				assert.True(t, expected.Contains(path.String()), strings.Join(expected.ListOrdered(), "|"))
				expected.Remove(path.String())
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

func BenchmarkClient_Degrees(b *testing.B) {
	ctx := context.Background()

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	bc := newBenchClient()
	for range b.N {
		c := degrees.New(bc, l)
		for range c.Degrees(ctx, 1, 4, 4) {
		}
	}
}

func newBenchClient() *benchClient {
	c := new(benchClient)
	c.person = map[int]tmdb.Person{
		1: makePerson(1),
		2: makePerson(2),
		3: makePerson(3),
		4: makePerson(4),
	}
	c.personCredits = map[int]tmdb.PersonCredits{
		1: makePersonCredits(1, 1),
		2: makePersonCredits(2, 1, 2),
		3: makePersonCredits(3, 2, 3, 4),
		4: makePersonCredits(4, 3, 4),
	}
	c.movie = map[int]tmdb.Movie{
		3: {Id: 3, Title: "movie3"},
		4: {Id: 4, Title: "movie4"},
	}
	c.movieCredits = map[int]tmdb.MovieCredits{
		1: makeMovieCredits(1, 1, 2),
		2: makeMovieCredits(2, 2, 3),
		3: makeMovieCredits(3, 3, 4),
		4: makeMovieCredits(3, 3, 4),
	}
	return c
}

var _ degrees.TMDBClient = &benchClient{}

type benchClient struct {
	person        map[int]tmdb.Person
	personCredits map[int]tmdb.PersonCredits
	movie         map[int]tmdb.Movie
	movieCredits  map[int]tmdb.MovieCredits
}

func (b benchClient) GetPersonCredits(_ context.Context, id int) (tmdb.PersonCredits, error) {
	if credits, ok := b.personCredits[id]; ok {
		return credits, nil
	}
	return tmdb.PersonCredits{}, fmt.Errorf("personCredits not found: %d", id)
}

func (b benchClient) SearchPersonPage(_ context.Context, _ string, _ int) ([]tmdb.Person, int, error) {
	panic("implement me")
}

func (b benchClient) GetPerson(_ context.Context, id int) (tmdb.Person, error) {
	if person, ok := b.person[id]; ok {
		return person, nil
	}
	return tmdb.Person{}, fmt.Errorf("person not found: %d", id)
}

func (b benchClient) GetMovie(_ context.Context, id int) (tmdb.Movie, error) {
	if movie, ok := b.movie[id]; ok {
		return movie, nil
	}
	return tmdb.Movie{}, fmt.Errorf("movie not found: %d", id)
}

func (b benchClient) GetMovieCredits(_ context.Context, id int) (tmdb.MovieCredits, error) {
	if credits, ok := b.movieCredits[id]; ok {
		return credits, nil
	}
	return tmdb.MovieCredits{}, fmt.Errorf("movieCredits not found: %d", id)
}
