package degrees

import (
	"context"
	"github.com/clambin/go-common/set"
	"github.com/clambin/tmdb/pkg/tmdb"
	"slices"
)

func (c *Client) Degrees(ctx context.Context, fromActorID, toActorID int, maxDepth int) chan Path {
	ch := make(chan Path)
	go func() {
		c.findActor(ctx, ch, fromActorID, toActorID, maxDepth-1, nil, nil, nil)
		close(ch)
	}()
	return ch
}

func (c *Client) findActor(ctx context.Context, ch chan Path, fromActorID int, toActorID int, maxDepth int, path Path, examinedActorIDs set.Set[int], examinedMovieIDs set.Set[int]) {
	c.logger.Debug("findActor", "from", fromActorID, "to", toActorID, "maxDepth", maxDepth, "path", path)

	actor, err := c.GetPerson(ctx, fromActorID)
	if err != nil {
		c.logger.Warn("tmdb call failed", "err", err, "call", "GetPerson")
		return
	}

	fromActorMovies, err := c.getActorMovieCredits(ctx, fromActorID)
	if err != nil {
		c.logger.Warn("tmdb call failed", "err", err, "call", "getActorMovieCredits")
		return
	}

	for movieID := range fromActorMovies.movieIDs {
		if examinedMovieIDs.Contains(movieID) {
			continue
		}

		newExaminedMovieIDs := examinedMovieIDs.Copy()
		newExaminedMovieIDs.Add(movieID)

		newPath := append(path, Link{
			Person: tmdb.Person{Id: fromActorID, Name: actor.Name},
			Movie:  Movie{ID: movieID, Name: fromActorMovies.movieNames[movieID]},
		})

		credits, err := c.GetMovieCredits(ctx, movieID)
		if err != nil {
			c.logger.Warn("tmdb call failed", "err", err, "call", "GetMovieCredits")
			return
		}

		cast := getCastIDs(credits.Cast)
		if name, ok := cast[toActorID]; ok {
			c.reportPath(ctx, ch, newPath, tmdb.Person{Id: toActorID, Name: name})
			continue
		}

		if maxDepth > 0 {
			for _, a := range credits.Cast {
				if a.Id != fromActorID && !examinedActorIDs.Contains(a.Id) {
					newExaminedActorIDs := examinedActorIDs.Copy()
					newExaminedActorIDs.Add(a.Id)
					c.findActor(ctx, ch, a.Id, toActorID, maxDepth-1, newPath, newExaminedActorIDs, newExaminedMovieIDs)
				}
			}
		}
	}
}

func getCastIDs(cast []tmdb.MovieCastCredits) map[int]string {
	names := make(map[int]string)
	for _, c := range cast {
		names[c.Id] = c.Name
	}
	return names
}

func (c *Client) reportPath(_ context.Context, ch chan Path, path Path, person tmdb.Person) {
	p := slices.Concat(path, Path{Link{Person: person}})
	c.logger.Debug("path found", "path", p)
	ch <- p
}

type actorMovieCredits struct {
	id         int
	movieIDs   set.Set[int]
	movieNames map[int]string
}

func (c *Client) getActorMovieCredits(ctx context.Context, id int) (actorMovieCredits, error) {
	result := actorMovieCredits{
		id:         id,
		movieIDs:   set.New[int](),
		movieNames: make(map[int]string),
	}
	cr, err := c.TMDBClient.GetPersonCredits(ctx, id)
	if err == nil {
		for _, credit := range cr.Cast {
			if credit.MediaType == "movie" || c.IncludeTV {
				result.movieIDs.Add(credit.Id)
				result.movieNames[credit.Id] = credit.GetTitle()
			}
		}
	}
	return result, err
}
