package degrees

import (
	"github.com/clambin/tmdb/pkg/tmdb"
	"strconv"
	"strings"
)

type Link struct {
	Movie
	tmdb.Person
}

type Movie struct {
	ID   int
	Name string
}

func (l Link) String() string {
	if l.Movie.Name == "" {
		return l.Person.Name
	}
	return l.Movie.Name + " -> " + l.Person.Name
}

type Path []Link

func (p Path) String() string {
	var w strings.Builder
	for i := range p {
		if i > 0 {
			w.WriteString(" -> ")
		}
		w.WriteString(p[i].String())
	}
	w.WriteString(" (" + strconv.Itoa(len(p)) + ")")
	return w.String()
}
