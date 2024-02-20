package degrees

import (
	"github.com/clambin/tmdb/pkg/tmdb"
	"strconv"
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
	val := l.Person.Name
	if l.Movie.ID != 0 {
		val += " -> " + l.Movie.Name
	}
	return val
}

type Path []Link

func (p Path) String() string {
	var output string
	for i := range p {
		if i > 0 {
			output += " -> "
		}
		output += p[i].String()
	}
	output += " (" + strconv.Itoa(len(p)-1) + ")"
	return output
}
