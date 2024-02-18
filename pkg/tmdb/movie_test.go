package tmdb_test

import (
	"context"
	"github.com/clambin/tmdb/pkg/tmdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestClient_GetMovie(t *testing.T) {
	s := makeTestServer("GET /3/movie/{id}", func(r *http.Request) string {
		return "get-movie-" + r.PathValue("id") + ".json"
	})
	c := tmdb.New("", nil)
	c.BaseURL = s.URL

	ctx := context.Background()
	movie, err := c.GetMovie(ctx, 680)
	require.NoError(t, err)
	assert.Equal(t, 680, movie.Id)
	assert.Equal(t, "Pulp Fiction", movie.Title)

	s.Close()
	_, err = c.GetPersonCredits(ctx, 31)
	assert.Error(t, err)
}

func TestClient_GetMovieCredits(t *testing.T) {
	s := makeTestServer("GET /3/movie/{id}/credits", func(r *http.Request) string {
		return "get-movie-credits-" + r.PathValue("id") + ".json"
	})
	c := tmdb.New("", nil)
	c.BaseURL = s.URL

	ctx := context.Background()
	credits, err := c.GetMovieCredits(ctx, 680)
	require.NoError(t, err)
	assert.Equal(t, 680, credits.Id)
	assert.NotEmpty(t, credits.Cast)

	s.Close()
	_, err = c.GetPersonCredits(ctx, 31)
	assert.Error(t, err)
}
