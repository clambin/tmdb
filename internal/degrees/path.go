package degrees

import (
	"github.com/clambin/tmdb/pkg/tmdb"
	"strconv"
	"strings"
)

type Link struct {
	Entry
	tmdb.Person
}

type Entry struct {
	ID   int
	Name string
}

func (l Link) String() string {
	if l.Entry.Name == "" {
		return l.Person.Name
	}
	return l.Entry.Name + " -> " + l.Person.Name
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
	w.WriteString(" (" + strconv.Itoa(len(p)-1) + ")")
	return w.String()
}
