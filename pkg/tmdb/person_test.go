package tmdb_test

import (
	"context"
	"github.com/clambin/tmdb/pkg/tmdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestCastCredit_GetTitle(t *testing.T) {
	type fields struct {
		mediaType string
		name      string
		title     string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "movie",
			fields: fields{
				mediaType: "movie",
				title:     "Movie",
			},
			want: "Movie",
		},
		{
			name: "tv show",
			fields: fields{
				mediaType: "tv",
				name:      "TV Show",
			},
			want: "TV Show",
		},
		{
			name: "unknown",
			fields: fields{
				mediaType: "invalid",
				name:      "name",
				title:     "title",
			},
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := tmdb.CastCredit{
				Title:     tt.fields.title,
				MediaType: tt.fields.mediaType,
				Name:      tt.fields.name,
			}
			assert.Equal(t, tt.want, c.GetTitle())
		})
	}
}

func TestClient_SearchPersonPage(t *testing.T) {
	s := makeTestServer("GET /3/search/person", func(r *http.Request) string {
		return "search-" + r.FormValue("query") + "-" + r.FormValue("page") + ".json"
	})
	c := tmdb.New("", nil)
	c.BaseURL = s.URL

	ctx := context.Background()
	persons, err := c.SearchPersonAllPages(ctx, "tom hanks")
	require.NoError(t, err)
	require.Len(t, persons, 1)
	assert.Equal(t, "Tom Hanks", persons[0].Name)

	s.Close()
	_, err = c.SearchPersonAllPages(ctx, "tom hanks")
	assert.Error(t, err)
}

func TestClient_GetPerson(t *testing.T) {
	s := makeTestServer("GET /3/person/{id}", func(r *http.Request) string {
		return "get-person-" + r.PathValue("id") + ".json"
	})
	c := tmdb.New("", nil)
	c.BaseURL = s.URL

	ctx := context.Background()
	person, err := c.GetPerson(ctx, 31)
	require.NoError(t, err)
	assert.Equal(t, "Tom Hanks", person.Name)

	s.Close()
	_, err = c.GetPerson(ctx, 31)
	assert.Error(t, err)
}

func TestClient_GetPersonCredits(t *testing.T) {
	s := makeTestServer("GET /3/person/{id}/combined_credits", func(r *http.Request) string {
		return "get-person-credits-" + r.PathValue("id") + ".json"
	})
	c := tmdb.New("", nil)
	c.BaseURL = s.URL

	ctx := context.Background()
	credits, err := c.GetPersonCredits(ctx, 31)
	require.NoError(t, err)
	assert.Equal(t, 31, credits.Id)
	assert.Len(t, credits.Cast, 236)

	s.Close()
	_, err = c.GetPersonCredits(ctx, 31)
	assert.Error(t, err)
}
