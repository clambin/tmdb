package tmdb_test

import (
	"context"
	"github.com/clambin/tmdb/pkg/tmdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestClient_GetTVSeries(t *testing.T) {
	s := makeTestServer("GET /3/tv/{id}", func(r *http.Request) string {
		return "get-tv-" + r.PathValue("id") + ".json"
	})
	c := tmdb.New("", nil)
	c.BaseURL = s.URL

	ctx := context.Background()
	credits, err := c.GetTVSeries(ctx, 1668)
	require.NoError(t, err)
	assert.Equal(t, 1668, credits.Id)

	s.Close()
	_, err = c.GetTVSeries(ctx, 1668)
	assert.Error(t, err)

}

func TestClient_GetTVSeriesCredits(t *testing.T) {
	s := makeTestServer("GET /3/tv/{id}/aggregate_credits", func(r *http.Request) string {
		return "get-tv-credits-" + r.PathValue("id") + ".json"
	})
	c := tmdb.New("", nil)
	c.BaseURL = s.URL

	ctx := context.Background()
	credits, err := c.GetTVSeriesCredits(ctx, 1668)
	require.NoError(t, err)
	assert.Equal(t, 1668, credits.Id)
	assert.NotEmpty(t, credits.Cast)

	s.Close()
	_, err = c.GetTVSeriesCredits(ctx, 1668)
	assert.Error(t, err)
}
