package degrees

import (
	"cmp"
	"context"
	"errors"
	"github.com/clambin/tmdb/pkg/tmdb"
	"slices"
)

var ErrPersonNotFound = errors.New("person not found")

func FindActor(ctx context.Context, c TMDBClient, query string) (tmdb.Person, error) {
	persons, err := c.SearchPersonAllPages(ctx, query)
	if err != nil {
		return tmdb.Person{}, err
	}
	slices.SortFunc(persons, func(a, b tmdb.Person) int {
		return -cmp.Compare(a.Popularity, b.Popularity)
	})
	for _, p := range persons {
		if p.KnownForDepartment == "Acting" {
			return p, nil
		}
	}
	if len(persons) == 0 {
		return tmdb.Person{}, ErrPersonNotFound
	}
	return persons[0], nil
}
