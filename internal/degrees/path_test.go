package degrees_test

import (
	"github.com/clambin/tmdb/internal/degrees"
	"github.com/clambin/tmdb/pkg/tmdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPath_String(t *testing.T) {
	p := degrees.Path{
		degrees.Link{Person: tmdb.Person{Id: 1, Name: "actor1"}},
		degrees.Link{Person: tmdb.Person{Id: 2, Name: "actor2"}, Movie: degrees.Movie{ID: 1, Name: "movie1"}},
		degrees.Link{Person: tmdb.Person{Id: 3, Name: "actor3"}, Movie: degrees.Movie{ID: 2, Name: "movie2"}},
	}
	assert.Equal(t, "actor1 -> movie1 -> actor2 -> movie2 -> actor3 (3)", p.String())
}
