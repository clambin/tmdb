package degrees_test

import (
	"context"
	"errors"
	"github.com/clambin/tmdb/internal/degrees"
	"github.com/clambin/tmdb/internal/degrees/mocks"
	"github.com/clambin/tmdb/pkg/tmdb"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestClient_FindActor(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		persons []tmdb.Person
		err     error
		wantID  int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:  "actor",
			query: "foo",
			persons: []tmdb.Person{
				{Id: 1, Name: "foo", KnownForDepartment: "Acting", Popularity: 1},
				{Id: 2, Name: "foo", KnownForDepartment: "Directing", Popularity: 1},
				{Id: 3, Name: "foo", KnownForDepartment: "Acting", Popularity: 2},
				{Id: 4, Name: "foo", KnownForDepartment: "Directing", Popularity: 2},
			},
			wantID:  3,
			wantErr: assert.NoError,
		},
		{
			name:  "fallback to non-actor",
			query: "foo",
			persons: []tmdb.Person{
				{Id: 2, Name: "foo", KnownForDepartment: "Directing", Popularity: 1},
				{Id: 4, Name: "foo", KnownForDepartment: "Directing", Popularity: 2},
			},
			wantID:  4,
			wantErr: assert.NoError,
		},
		{
			name:    "nothing found",
			query:   "foo",
			wantErr: assert.Error,
		},
		{
			name:    "error",
			query:   "foo",
			err:     errors.New("failed"),
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			api := mocks.NewTMDBClient(t)
			api.EXPECT().SearchPersonAllPages(ctx, tt.query).Return(tt.persons, tt.err)
			c := degrees.New(api, slog.Default())

			person, err := c.FindActor(ctx, tt.query)
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantID, person.Id)
		})
	}
}
