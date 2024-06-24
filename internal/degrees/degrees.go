package degrees

import (
	"context"
	"github.com/clambin/go-common/set"
	"github.com/clambin/tmdb/pkg/tmdb"
	"log/slog"
	"sync"
	"sync/atomic"
)

type TMDBClient interface {
	GetPersonCredits(ctx context.Context, id int) (tmdb.PersonCredits, error)
	SearchPersonPage(ctx context.Context, query string, page int) ([]tmdb.Person, int, error)
	GetPerson(ctx context.Context, id int) (tmdb.Person, error)
	GetMovie(ctx context.Context, id int) (tmdb.Movie, error)
	GetMovieCredits(ctx context.Context, id int) (tmdb.MovieCredits, error)
	SearchPersonAllPages(ctx context.Context, query string) ([]tmdb.Person, error)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

type PathFinder struct {
	TMDBClient TMDBClient
	From       tmdb.Person
	To         tmdb.Person
	Logger     *slog.Logger

	maxPathLength       atomic.Int32
	toActorMovieCredits actorMovieCredits
	lock                sync.RWMutex
	visitedActors       set.Set[int]
}

type actorMovieCredits struct {
	movieIDs   set.Set[int]
	movieNames map[int]string
}

func (f *PathFinder) Find(ctx context.Context, ch chan Path, maxPathLength int) {
	defer close(ch)
	if err := f.init(ctx, maxPathLength); err != nil {
		f.Logger.Error("failed to initialize path finder", "error", err)
		return
	}
	f.findActor(ctx, ch, f.From, Path{Link{Person: f.From}})
}

func (f *PathFinder) init(ctx context.Context, maxPathLength int) (err error) {
	f.maxPathLength.Store(int32(maxPathLength))
	f.visitedActors = set.New[int]()
	f.toActorMovieCredits, err = f.getActorMovieCredits(ctx, f.To.Id)
	return err
}

func (f *PathFinder) findActor(ctx context.Context, ch chan Path, from tmdb.Person, path Path) {
	logger := f.Logger.With("path", path)
	logger.Debug("checking actor", "actor", from.Name)

	fromActorMovieCredits, err := f.getActorMovieCredits(ctx, from.Id)
	if err != nil {
		logger.Error("failed to get actor credits", "err", err)
		return
	}

	f.visit(from.Id)

	commonMovieIDs := set.Intersection(fromActorMovieCredits.movieIDs, f.toActorMovieCredits.movieIDs)
	if len(commonMovieIDs) > 0 {
		for commonMovieID := range commonMovieIDs {
			f.report(ch, append(path, Link{
				Person: f.To,
				Movie:  Movie{ID: commonMovieID, Name: f.toActorMovieCredits.movieNames[commonMovieID]},
			}))
		}
		return
	}

	if f.maxPathReached(path) {
		return
	}

	var wg sync.WaitGroup

	for movieID := range fromActorMovieCredits.movieIDs {
		credits, err := f.TMDBClient.GetMovieCredits(ctx, movieID)
		if err != nil {
			logger.Error("failed to get movie credits", "id", movieID, "err", err)
			return
		}
		for _, cast := range credits.Cast {
			// other goroutines may find a shorter path concurrently
			if f.maxPathReached(path) {
				break
			}
			// don't traverse actors we've already visited
			if f.visited(cast.Id) {
				continue
			}
			p := tmdb.Person{Id: cast.Id, Name: cast.Name}
			newPath := make(Path, len(path)+1)
			copy(newPath, append(path, Link{
				Person: p,
				Movie:  Movie{ID: credits.Id, Name: fromActorMovieCredits.movieNames[credits.Id]},
			}))
			wg.Add(1)
			go func() {
				f.findActor(ctx, ch, p, newPath)
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func (f *PathFinder) report(ch chan Path, path Path) {
	if pathLen := len(path); pathLen < int(f.maxPathLength.Load()) {
		f.Logger.Debug("shorter path found!", "maxPathLength", pathLen)
		f.maxPathLength.Store(int32(pathLen))
	}
	ch <- path
}

func (f *PathFinder) maxPathReached(path Path) bool {
	return len(path)+1 >= int(f.maxPathLength.Load())
}

func (f *PathFinder) getActorMovieCredits(ctx context.Context, id int) (actorMovieCredits, error) {
	result := actorMovieCredits{
		movieIDs:   set.New[int](),
		movieNames: make(map[int]string),
	}
	cr, err := f.TMDBClient.GetPersonCredits(ctx, id)
	if err == nil {
		for _, credit := range cr.Cast {
			if credit.MediaType == "movie" {
				result.movieIDs.Add(credit.Id)
				result.movieNames[credit.Id] = credit.GetTitle()
			}
		}
	}
	return result, err
}

func (f *PathFinder) visit(id int) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.visitedActors.Add(id)
}

func (f *PathFinder) visited(id int) bool {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.visitedActors.Contains(id)
}
